# danh_gia_perplexity_sampling.py
"""Bài 04 - Đánh giá bằng perplexity & sampling: load checkpoint đã train ở
bài 03, tính perplexity trên val set, rồi generate text bằng temperature +
top-k sampling, verify greedy (temperature=0) deterministic."""
import json
import math
from pathlib import Path

import torch
import torch.nn as nn
import torch.nn.functional as F

torch.manual_seed(0)

HERE = Path(__file__).parent
DATA_DIR = HERE / ".." / "01-chuan-bi-du-lieu-tokenizer"
CKPT_DIR = HERE / ".." / "03-training-loop-checkpoint-amp"
DEVICE = "mps" if torch.backends.mps.is_available() else "cpu"


# --- kiến trúc GPT y hệt bài 02/03 (mỗi bài script độc lập, copy lại định nghĩa) ---
class GPTConfig:
    def __init__(self, vocab_size, block_size=64, n_layer=4, n_head=4, n_embd=128, dropout=0.1):
        self.vocab_size, self.block_size = vocab_size, block_size
        self.n_layer, self.n_head, self.n_embd, self.dropout = n_layer, n_head, n_embd, dropout


class CausalSelfAttention(nn.Module):
    def __init__(self, cfg):
        super().__init__()
        self.n_head, self.head_dim = cfg.n_head, cfg.n_embd // cfg.n_head
        self.qkv_proj = nn.Linear(cfg.n_embd, 3 * cfg.n_embd)
        self.out_proj = nn.Linear(cfg.n_embd, cfg.n_embd)
        self.dropout = nn.Dropout(cfg.dropout)
        mask = torch.triu(torch.ones(cfg.block_size, cfg.block_size), diagonal=1).bool()
        self.register_buffer("causal_mask", mask)

    def forward(self, x):
        b, t, d = x.shape
        q, k, v = self.qkv_proj(x).chunk(3, dim=-1)
        split = lambda z: z.view(b, t, self.n_head, self.head_dim).transpose(1, 2)
        q, k, v = split(q), split(k), split(v)
        scores = (q @ k.transpose(-2, -1)) / math.sqrt(self.head_dim)
        scores = scores.masked_fill(self.causal_mask[:t, :t], float("-inf"))
        out = (self.dropout(torch.softmax(scores, dim=-1)) @ v).transpose(1, 2).contiguous().view(b, t, d)
        return self.out_proj(out)


class MLP(nn.Module):
    def __init__(self, cfg):
        super().__init__()
        self.fc1 = nn.Linear(cfg.n_embd, 4 * cfg.n_embd)
        self.fc2 = nn.Linear(4 * cfg.n_embd, cfg.n_embd)
        self.dropout = nn.Dropout(cfg.dropout)

    def forward(self, x):
        return self.dropout(self.fc2(F.gelu(self.fc1(x))))


class Block(nn.Module):
    def __init__(self, cfg):
        super().__init__()
        self.ln1, self.attn = nn.LayerNorm(cfg.n_embd), CausalSelfAttention(cfg)
        self.ln2, self.mlp = nn.LayerNorm(cfg.n_embd), MLP(cfg)

    def forward(self, x):
        x = x + self.attn(self.ln1(x))
        return x + self.mlp(self.ln2(x))


class GPT(nn.Module):
    def __init__(self, cfg):
        super().__init__()
        self.cfg = cfg
        self.token_emb = nn.Embedding(cfg.vocab_size, cfg.n_embd)
        self.pos_emb = nn.Embedding(cfg.block_size, cfg.n_embd)
        self.dropout = nn.Dropout(cfg.dropout)
        self.blocks = nn.ModuleList([Block(cfg) for _ in range(cfg.n_layer)])
        self.ln_f = nn.LayerNorm(cfg.n_embd)
        self.lm_head = nn.Linear(cfg.n_embd, cfg.vocab_size, bias=False)
        self.lm_head.weight = self.token_emb.weight

    def forward(self, idx, targets=None):
        b, t = idx.shape
        pos = torch.arange(t, device=idx.device)
        x = self.dropout(self.token_emb(idx) + self.pos_emb(pos))
        for block in self.blocks:
            x = block(x)
        logits = self.lm_head(self.ln_f(x))
        loss = None
        if targets is not None:
            loss = F.cross_entropy(logits.view(-1, logits.size(-1)), targets.view(-1))
        return logits, loss

    @torch.no_grad()
    def generate(self, idx, max_new_tokens, temperature=1.0, top_k=None):
        self.eval()
        for _ in range(max_new_tokens):
            idx_cond = idx[:, -self.cfg.block_size :]
            logits, _ = self(idx_cond)
            logits = logits[:, -1, :] / max(temperature, 1e-8)
            if top_k is not None:
                v, _ = torch.topk(logits, min(top_k, logits.size(-1)))
                logits[logits < v[:, [-1]]] = float("-inf")
            probs = F.softmax(logits, dim=-1)
            if temperature == 0:
                next_id = probs.argmax(dim=-1, keepdim=True)
            else:
                next_id = torch.multinomial(probs, num_samples=1)
            idx = torch.cat([idx, next_id], dim=1)
        return idx


def apply_bpe(chunk, merges):
    symbols = list(chunk)
    for pair in merges:
        i, new_symbols = 0, []
        while i < len(symbols):
            if i < len(symbols) - 1 and (symbols[i], symbols[i + 1]) == pair:
                new_symbols.append(symbols[i] + symbols[i + 1])
                i += 2
            else:
                new_symbols.append(symbols[i])
                i += 1
        symbols = new_symbols
    return symbols


if __name__ == "__main__":
    import re

    tok = json.loads((DATA_DIR / "tokenizer.json").read_text(encoding="utf-8"))
    merges = [tuple(p) for p in tok["merges"]]
    vocab = tok["vocab"]
    token_to_id = {t: i for i, t in enumerate(vocab)}

    def encode(text):
        return [token_to_id[t] for c in re.findall(r"\S+|\s+", text) for t in apply_bpe(c, merges)]

    def decode(ids):
        return "".join(vocab[i] for i in ids)

    ckpt = torch.load(CKPT_DIR / "checkpoint.pt", map_location=DEVICE, weights_only=False)
    cfg = GPTConfig(**ckpt["cfg"])
    model = GPT(cfg).to(DEVICE)
    model.load_state_dict(ckpt["model"])
    model.eval()
    print(f"Đã load checkpoint từ bài 03, train {ckpt['step']} step, vocab_size={cfg.vocab_size}")

    # --- perplexity trên val set ---
    tokenized = torch.load(DATA_DIR / "tokenized_corpus.pt")
    val_ids = tokenized["val_ids"]
    losses = []
    with torch.no_grad():
        for i in range(0, len(val_ids) - cfg.block_size - 1, cfg.block_size):
            x = val_ids[i : i + cfg.block_size].unsqueeze(0).to(DEVICE)
            y = val_ids[i + 1 : i + cfg.block_size + 1].unsqueeze(0).to(DEVICE)
            _, loss = model(x, y)
            losses.append(loss.item())
    avg_loss = sum(losses) / len(losses)
    perplexity = math.exp(avg_loss)
    print(f"\nVal loss trung bình: {avg_loss:.4f}")
    print(f"Perplexity: {perplexity:.1f}  (so với random guessing = vocab_size = {cfg.vocab_size})")
    assert perplexity == perplexity, "perplexity là NaN"
    assert perplexity < cfg.vocab_size, "model phải tốt hơn random guessing"
    print("OK: perplexity hữu hạn và tốt hơn random guessing")

    # --- sampling: temperature + top-k ---
    prompt = "Attention là"
    prompt_ids = torch.tensor([encode(prompt)], device=DEVICE)

    print(f"\nPrompt: {prompt!r}")
    for temp, top_k in [(0.0, None), (0.8, 20), (1.2, None)]:
        out = model.generate(prompt_ids.clone(), max_new_tokens=40, temperature=temp, top_k=top_k)
        text = decode(out[0].tolist())
        label = "greedy (temp=0)" if temp == 0 else f"temp={temp}, top_k={top_k}"
        print(f"  [{label}]\n    {text!r}")

    # --- verify: greedy sampling (temperature=0) phải deterministic ---
    out1 = model.generate(prompt_ids.clone(), max_new_tokens=20, temperature=0.0)
    out2 = model.generate(prompt_ids.clone(), max_new_tokens=20, temperature=0.0)
    assert torch.equal(out1, out2), "greedy sampling PHẢI deterministic, ra kết quả khác nhau"
    print("\nOK: greedy sampling (temperature=0) deterministic - 2 lần chạy ra đúng 1 kết quả")

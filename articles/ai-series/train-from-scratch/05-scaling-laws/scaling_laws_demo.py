# scaling_laws_demo.py
"""Bài 05 - Scaling laws: thí nghiệm nhỏ đo val loss cuối cùng khi tăng (a)
kích thước model (n_embd) và (b) lượng data train, cùng 1 số step cố định -
minh hoạ trực quan xu hướng "nhiều tham số/nhiều data hơn -> loss thấp hơn"
đứng sau scaling laws thật (Kaplan 2020, Chinchilla 2022), KHÔNG phải fit
công thức thật (cần quy mô compute lớn hơn nhiều)."""
import json
import math
import time
from pathlib import Path

import torch
import torch.nn as nn
import torch.nn.functional as F

torch.manual_seed(0)

HERE = Path(__file__).parent
DATA_DIR = HERE / ".." / "01-chuan-bi-du-lieu-tokenizer"
DEVICE = "mps" if torch.backends.mps.is_available() else "cpu"


# --- kiến trúc GPT y hệt bài 02/03/04 (mỗi bài script độc lập) ---
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


def get_batch(data, block_size, batch_size, device):
    ix = torch.randint(0, len(data) - block_size - 1, (batch_size,))
    x = torch.stack([data[i : i + block_size] for i in ix])
    y = torch.stack([data[i + 1 : i + block_size + 1] for i in ix])
    return x.to(device), y.to(device)


def train_and_eval(vocab_size, n_embd, train_data, val_ids, n_steps=300, block_size=64, batch_size=32):
    torch.manual_seed(0)
    cfg = GPTConfig(vocab_size=vocab_size, block_size=block_size, n_layer=4, n_head=4, n_embd=n_embd)
    model = GPT(cfg).to(DEVICE)
    opt = torch.optim.AdamW(model.parameters(), lr=3e-4, weight_decay=0.01)
    for _ in range(n_steps):
        x, y = get_batch(train_data, block_size, batch_size, DEVICE)
        _, loss = model(x, y)
        opt.zero_grad()
        loss.backward()
        opt.step()
    model.eval()
    val_losses = []
    with torch.no_grad():
        for i in range(0, len(val_ids) - block_size - 1, block_size):
            x = val_ids[i : i + block_size].unsqueeze(0).to(DEVICE)
            y = val_ids[i + 1 : i + block_size + 1].unsqueeze(0).to(DEVICE)
            _, l = model(x, y)
            val_losses.append(l.item())
    n_params = sum(p.numel() for p in model.parameters())
    return sum(val_losses) / len(val_losses), n_params


if __name__ == "__main__":
    print(f"Device: {DEVICE}")
    tokenized = torch.load(DATA_DIR / "tokenized_corpus.pt")
    train_ids, val_ids = tokenized["train_ids"], tokenized["val_ids"]
    vocab_size = len(json.loads((DATA_DIR / "tokenizer.json").read_text(encoding="utf-8"))["vocab"])

    # --- (a) tăng kích thước model, giữ nguyên toàn bộ data train ---
    print("\n=== (a) Tăng n_embd (kích thước model), cùng 300 step, cùng toàn bộ train data ===")
    print(f"{'n_embd':>8} | {'params':>10} | {'val_loss':>9}")
    model_results = []
    t0 = time.time()
    for n_embd in (32, 64, 128, 256):
        val_loss, n_params = train_and_eval(vocab_size, n_embd, train_ids, val_ids)
        model_results.append((n_embd, n_params, val_loss))
        print(f"{n_embd:>8} | {n_params:>10,} | {val_loss:>9.4f}")
    print(f"(chạy trong {time.time() - t0:.1f}s)")

    assert model_results[-1][2] < model_results[0][2], "model lớn nhất phải có val_loss thấp hơn model nhỏ nhất"

    # --- (b) tăng lượng data train, giữ nguyên kích thước model ---
    print("\n=== (b) Tăng % data train (n_embd=128 cố định, 300 step) ===")
    print(f"{'% data':>8} | {'n_token':>9} | {'val_loss':>9}")
    data_results = []
    for frac in (0.25, 0.5, 1.0):
        n_tok = int(len(train_ids) * frac)
        subset = train_ids[:n_tok]
        val_loss, _ = train_and_eval(vocab_size, 128, subset, val_ids)
        data_results.append((frac, n_tok, val_loss))
        print(f"{frac * 100:>7.0f}% | {n_tok:>9} | {val_loss:>9.4f}")

    assert data_results[-1][2] < data_results[0][2], "100% data phải có val_loss thấp hơn 25% data"
    print("\nOK: val_loss giảm khi tăng model size VÀ khi tăng lượng data train (đúng xu hướng scaling law)")

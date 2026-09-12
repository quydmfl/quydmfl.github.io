# gpt_from_scratch.py
"""Bài 02 - Cài đặt GPT từ đầu bằng PyTorch: kiến trúc decoder-only style GPT-2
(token+positional embedding, N block causal self-attention + MLP, pre-norm,
weight tying), verify causal mask thật (sửa token tương lai không ảnh hưởng
logit ở vị trí quá khứ) và đếm tham số."""
import math

import torch
import torch.nn as nn
import torch.nn.functional as F

torch.manual_seed(0)


class GPTConfig:
    def __init__(self, vocab_size, block_size=64, n_layer=4, n_head=4, n_embd=128, dropout=0.1):
        self.vocab_size = vocab_size
        self.block_size = block_size
        self.n_layer = n_layer
        self.n_head = n_head
        self.n_embd = n_embd
        self.dropout = dropout


class CausalSelfAttention(nn.Module):
    """Multi-head self-attention với causal mask - mỗi vị trí chỉ được nhìn
    chính nó và các vị trí TRƯỚC nó (khác bài 04 sub-series Nền tảng, không có mask)."""

    def __init__(self, cfg):
        super().__init__()
        assert cfg.n_embd % cfg.n_head == 0
        self.n_head = cfg.n_head
        self.head_dim = cfg.n_embd // cfg.n_head
        self.qkv_proj = nn.Linear(cfg.n_embd, 3 * cfg.n_embd)
        self.out_proj = nn.Linear(cfg.n_embd, cfg.n_embd)
        self.dropout = nn.Dropout(cfg.dropout)
        # causal mask: mask[i, j] = True nếu vị trí j > i (tương lai so với i)
        mask = torch.triu(torch.ones(cfg.block_size, cfg.block_size), diagonal=1).bool()
        self.register_buffer("causal_mask", mask)

    def forward(self, x):
        b, t, d = x.shape
        qkv = self.qkv_proj(x)
        q, k, v = qkv.chunk(3, dim=-1)

        def split_heads(z):
            return z.view(b, t, self.n_head, self.head_dim).transpose(1, 2)

        q, k, v = split_heads(q), split_heads(k), split_heads(v)
        scores = (q @ k.transpose(-2, -1)) / math.sqrt(self.head_dim)
        scores = scores.masked_fill(self.causal_mask[:t, :t], float("-inf"))
        weights = self.dropout(torch.softmax(scores, dim=-1))
        out = (weights @ v).transpose(1, 2).contiguous().view(b, t, d)
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
    """Pre-norm: LayerNorm trước attention/MLP, residual sau - giống GPT-2,
    ổn định hơn post-norm (LayerNorm sau) khi stack nhiều layer."""

    def __init__(self, cfg):
        super().__init__()
        self.ln1 = nn.LayerNorm(cfg.n_embd)
        self.attn = CausalSelfAttention(cfg)
        self.ln2 = nn.LayerNorm(cfg.n_embd)
        self.mlp = MLP(cfg)

    def forward(self, x):
        x = x + self.attn(self.ln1(x))
        x = x + self.mlp(self.ln2(x))
        return x


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
        self.lm_head.weight = self.token_emb.weight  # weight tying (GPT-2 style)

    def forward(self, idx, targets=None):
        b, t = idx.shape
        assert t <= self.cfg.block_size, f"sequence dài {t} > block_size {self.cfg.block_size}"
        pos = torch.arange(t, device=idx.device)
        x = self.dropout(self.token_emb(idx) + self.pos_emb(pos))
        for block in self.blocks:
            x = block(x)
        x = self.ln_f(x)
        logits = self.lm_head(x)
        loss = None
        if targets is not None:
            loss = F.cross_entropy(logits.view(-1, logits.size(-1)), targets.view(-1))
        return logits, loss


if __name__ == "__main__":
    cfg = GPTConfig(vocab_size=613, block_size=64, n_layer=4, n_head=4, n_embd=128, dropout=0.1)
    model = GPT(cfg)
    model.eval()  # tắt dropout để test causal mask deterministic

    n_params = sum(p.numel() for p in model.parameters())
    n_params_notied = n_params + cfg.vocab_size * cfg.n_embd  # nếu KHÔNG tie weight
    print(f"Số tham số: {n_params:,} (nếu không tie token_emb/lm_head: {n_params_notied:,})")

    batch, seq_len = 2, 10
    idx = torch.randint(0, cfg.vocab_size, (batch, seq_len))
    logits, _ = model(idx)
    print(f"Input shape : {tuple(idx.shape)}")
    print(f"Output shape: {tuple(logits.shape)}")
    assert logits.shape == (batch, seq_len, cfg.vocab_size)

    # --- verify causal mask thật: sửa token ở vị trí cuối (tương lai so với
    # mọi vị trí trước) không được làm thay đổi logit ở các vị trí trước ---
    idx2 = idx.clone()
    idx2[:, -1] = (idx2[:, -1] + 1) % cfg.vocab_size  # đổi token cuối cùng
    logits2, _ = model(idx2)

    diff_past = (logits[:, :-1] - logits2[:, :-1]).abs().max().item()
    diff_last = (logits[:, -1] - logits2[:, -1]).abs().max().item()
    print(f"\nMax diff logit ở các vị trí TRƯỚC vị trí bị sửa: {diff_past:.3e}")
    print(f"Max diff logit ở CHÍNH vị trí bị sửa           : {diff_last:.3e}")

    assert diff_past < 1e-6, "causal mask SAI: sửa token tương lai làm đổi logit quá khứ"
    assert diff_last > 1e-4, "sửa token tương lai phải làm đổi chính logit của vị trí đó"
    print("\nOK: causal mask đúng - sửa token cuối không ảnh hưởng logit các vị trí trước nó")

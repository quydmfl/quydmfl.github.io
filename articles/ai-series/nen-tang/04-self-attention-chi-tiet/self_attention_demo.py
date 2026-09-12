# self_attention_demo.py
"""Bài 04 - Self-Attention chi tiết: cài multi-head attention từ đầu (Q/K/V,
scaled dot-product, multi-head, positional encoding), rồi đối chiếu với
torch.nn.MultiheadAttention (dùng lại đúng weight) để verify cài đúng."""
import math
import torch
import torch.nn as nn

torch.manual_seed(0)

embed_dim, num_heads, seq_len, batch = 16, 4, 5, 2
head_dim = embed_dim // num_heads

# --- Positional encoding (sinusoidal, cố định) ---
def sinusoidal_positional_encoding(seq_len, dim):
    pos = torch.arange(seq_len).unsqueeze(1).float()
    i = torch.arange(dim).unsqueeze(0).float()
    angle_rates = 1 / (10000 ** (2 * (i // 2) / dim))
    angles = pos * angle_rates
    pe = torch.zeros(seq_len, dim)
    pe[:, 0::2] = torch.sin(angles[:, 0::2])
    pe[:, 1::2] = torch.cos(angles[:, 1::2])
    return pe

x = torch.randn(batch, seq_len, embed_dim)
x = x + sinusoidal_positional_encoding(seq_len, embed_dim)

# --- Layer tham chiếu: torch.nn.MultiheadAttention ---
ref_attn = nn.MultiheadAttention(embed_dim, num_heads, bias=False, batch_first=True)

def scaled_dot_product_attention(q, k, v, mask=None):
    """q,k,v: (batch, heads, seq, head_dim)"""
    scores = q @ k.transpose(-2, -1) / math.sqrt(q.shape[-1])
    if mask is not None:
        scores = scores.masked_fill(mask == 0, float("-inf"))
    weights = torch.softmax(scores, dim=-1)
    return weights @ v, weights

def manual_multihead_attention(x, in_proj_weight, out_proj_weight, num_heads):
    b, s, d = x.shape
    head_dim = d // num_heads
    qkv = x @ in_proj_weight.T          # (b, s, 3*d)
    q, k, v = qkv.chunk(3, dim=-1)

    def split_heads(t):
        return t.view(b, s, num_heads, head_dim).transpose(1, 2)  # (b, heads, s, head_dim)

    q, k, v = split_heads(q), split_heads(k), split_heads(v)
    out, weights = scaled_dot_product_attention(q, k, v)
    out = out.transpose(1, 2).contiguous().view(b, s, d)          # gộp lại các head
    return out @ out_proj_weight.T, weights

manual_out, manual_weights = manual_multihead_attention(
    x, ref_attn.in_proj_weight, ref_attn.out_proj.weight, num_heads
)
ref_out, ref_weights = ref_attn(x, x, x, need_weights=True, average_attn_weights=True)

max_diff_out = (manual_out - ref_out).abs().max().item()
max_diff_w = (manual_weights.mean(dim=1) - ref_weights).abs().max().item()

print(f"Output shape: {manual_out.shape}")
print(f"Max diff output (manual vs torch.nn.MultiheadAttention): {max_diff_out:.3e}")
print(f"Max diff attention weights (avg qua head): {max_diff_w:.3e}")

assert max_diff_out < 1e-5, "manual attention KHÔNG khớp torch.nn.MultiheadAttention"
assert max_diff_w < 1e-5, "attention weights KHÔNG khớp"
print("\nOK: manual multi-head attention khớp torch.nn.MultiheadAttention (sai số < 1e-5)")

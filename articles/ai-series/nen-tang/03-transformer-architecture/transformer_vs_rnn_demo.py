# transformer_vs_rnn_demo.py
"""Bài 03 - Transformer Architecture: minh hoạ vì sao attention song song hoá được
còn RNN thì không - so sánh thời gian xử lý 1 sequence bằng RNN loop tuần tự
(mỗi bước phải chờ bước trước) vs self-attention (1 phép matmul duy nhất)."""
import time
import torch

torch.manual_seed(0)

d = 64  # chiều hidden/feature

def rnn_forward(x, Wh, Wx):
    """x: (seq_len, d). Mỗi bước PHẢI chờ hidden của bước trước -> vòng lặp Python
    tuần tự, không thể chạy song song giữa các timestep."""
    seq_len = x.shape[0]
    hidden = torch.zeros(d)
    for t in range(seq_len):
        hidden = torch.tanh(hidden @ Wh + x[t] @ Wx)
    return hidden

def attention_forward(x, Wq, Wk, Wv):
    """x: (seq_len, d). Toàn bộ sequence xử lý trong 1 phép matmul lớn -
    không có vòng lặp tuần tự theo timestep."""
    q, k, v = x @ Wq, x @ Wk, x @ Wv
    scores = (q @ k.T) / (d ** 0.5)
    weights = torch.softmax(scores, dim=-1)
    return weights @ v

Wh = torch.randn(d, d) * 0.1
Wx = torch.randn(d, d) * 0.1
Wq = torch.randn(d, d) * 0.1
Wk = torch.randn(d, d) * 0.1
Wv = torch.randn(d, d) * 0.1

print(f"{'seq_len':>8} | {'RNN loop (ms)':>14} | {'Attention matmul (ms)':>22}")
for seq_len in (32, 128, 512, 2048):
    x = torch.randn(seq_len, d)

    t0 = time.perf_counter()
    rnn_forward(x, Wh, Wx)
    t_rnn = (time.perf_counter() - t0) * 1000

    t0 = time.perf_counter()
    attention_forward(x, Wq, Wk, Wv)
    t_attn = (time.perf_counter() - t0) * 1000

    print(f"{seq_len:>8} | {t_rnn:>14.2f} | {t_attn:>22.2f}")

seq_len = 2048
x = torch.randn(seq_len, d)
t0 = time.perf_counter(); rnn_forward(x, Wh, Wx); t_rnn_big = (time.perf_counter() - t0) * 1000
t0 = time.perf_counter(); attention_forward(x, Wq, Wk, Wv); t_attn_big = (time.perf_counter() - t0) * 1000
assert t_rnn_big > t_attn_big, "kỳ vọng RNN loop tuần tự chậm hơn attention song song ở seq dài"
print(f"\nOK: ở seq_len={seq_len}, RNN loop ({t_rnn_big:.1f}ms) chậm hơn attention matmul ({t_attn_big:.1f}ms)")

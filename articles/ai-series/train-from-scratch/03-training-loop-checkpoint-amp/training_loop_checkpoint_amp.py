# training_loop_checkpoint_amp.py
"""Bài 03 - Training loop, checkpoint, mixed precision: train thật GPT từ bài 02
trên corpus/tokenizer đã chuẩn bị ở bài 01. Đo tốc độ fp32 vs AMP (autocast fp16
+ GradScaler) có warmup đúng cách để so sánh công bằng, rồi train thật (dùng
config nhanh hơn trên máy này), lưu checkpoint, load lại verify loss khớp y hệt."""
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


# --- kiến trúc GPT y hệt bài 02 (mỗi bài script độc lập, copy lại định nghĩa) ---
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


def bench_step_ms(cfg, use_amp, batch_size, train_ids, n=30, warmup=10):
    """Đo ms/step, có warmup riêng cho từng config để loại nhiễu compile/cache
    lần đầu - so sánh KHÔNG warmup (chạy fp32 xong mới chạy AMP) sẽ sai lệch,
    vì lần chạy sau luôn có lợi nhờ kernel đã "nóng"."""
    torch.manual_seed(0)
    m = GPT(cfg).to(DEVICE)
    opt = torch.optim.AdamW(m.parameters(), lr=3e-4)
    scaler = torch.amp.GradScaler(DEVICE) if use_amp else None

    def step():
        x, y = get_batch(train_ids, cfg.block_size, batch_size, DEVICE)
        opt.zero_grad()
        if use_amp:
            with torch.autocast(device_type=DEVICE, dtype=torch.float16):
                _, loss = m(x, y)
            scaler.scale(loss).backward()
            scaler.step(opt)
            scaler.update()
        else:
            _, loss = m(x, y)
            loss.backward()
            opt.step()

    for _ in range(warmup):
        step()
    if DEVICE == "mps":
        torch.mps.synchronize()
    t0 = time.time()
    for _ in range(n):
        step()
    if DEVICE == "mps":
        torch.mps.synchronize()
    return (time.time() - t0) / n * 1000


if __name__ == "__main__":
    print(f"Device: {DEVICE}")
    tokenized = torch.load(DATA_DIR / "tokenized_corpus.pt")
    train_ids, val_ids = tokenized["train_ids"], tokenized["val_ids"]
    vocab_size = len(json.loads((DATA_DIR / "tokenizer.json").read_text(encoding="utf-8"))["vocab"])
    print(f"Vocab size: {vocab_size} | Train tokens: {len(train_ids)} | Val tokens: {len(val_ids)}")

    # --- đo tốc độ fp32 vs AMP, có warmup công bằng, ở vài quy mô model ---
    print("\nĐo ms/step (đã warmup, không tính lần chạy nguội):")
    for n_embd, batch in [(128, 32), (512, 64)]:
        cfg_b = GPTConfig(vocab_size=vocab_size, block_size=64, n_layer=4, n_head=4, n_embd=n_embd)
        ms_fp32 = bench_step_ms(cfg_b, use_amp=False, batch_size=batch, train_ids=train_ids)
        ms_amp = bench_step_ms(cfg_b, use_amp=True, batch_size=batch, train_ids=train_ids)
        print(f"  n_embd={n_embd:4d} batch={batch:2d}: fp32={ms_fp32:5.1f}ms  "
              f"AMP(fp16+GradScaler)={ms_amp:5.1f}ms  (AMP/fp32={ms_amp / ms_fp32:.2f}x)")

    # --- train thật: chọn fp32 hoặc AMP dựa trên kết quả đo THẬT ở trên (số nào
    # nhanh hơn ở n_embd=128 thì dùng số đó cho vòng train chính) ---
    cfg = GPTConfig(vocab_size=vocab_size, block_size=64, n_layer=4, n_head=4, n_embd=128, dropout=0.1)
    torch.manual_seed(0)
    model = GPT(cfg).to(DEVICE)
    optimizer = torch.optim.AdamW(model.parameters(), lr=3e-4, weight_decay=0.01)

    n_steps, batch_size = 500, 32
    losses = []
    t0 = time.time()
    for step in range(n_steps):
        x, y = get_batch(train_ids, cfg.block_size, batch_size, DEVICE)
        _, loss = model(x, y)
        optimizer.zero_grad()
        loss.backward()
        optimizer.step()
        losses.append(loss.item())
        if step % 100 == 0 or step == n_steps - 1:
            print(f"step={step:4d} loss={loss.item():.4f}")

    if DEVICE == "mps":
        torch.mps.synchronize()
    print(f"\nThời gian train {n_steps} step: {time.time() - t0:.1f}s")
    assert losses[-1] < losses[0] * 0.6, "loss chưa giảm đủ mạnh sau khi train"
    print(f"OK: loss giảm từ {losses[0]:.4f} xuống {losses[-1]:.4f}")

    # --- checkpoint: lưu rồi load lại, verify loss khớp y hệt trên cùng batch ---
    ckpt_path = HERE / "checkpoint.pt"
    torch.save(
        {"model": model.state_dict(), "optimizer": optimizer.state_dict(), "step": n_steps, "cfg": vars(cfg)},
        ckpt_path,
    )

    fixed_x, fixed_y = get_batch(val_ids, cfg.block_size, batch_size, DEVICE)
    model.eval()
    with torch.no_grad():
        _, loss_before = model(fixed_x, fixed_y)

    model2 = GPT(cfg).to(DEVICE)
    ckpt = torch.load(ckpt_path, map_location=DEVICE, weights_only=False)
    model2.load_state_dict(ckpt["model"])
    model2.eval()
    with torch.no_grad():
        _, loss_after = model2(fixed_x, fixed_y)

    diff = abs(loss_before.item() - loss_after.item())
    print(f"\nLoss trên val batch cố định - model gốc: {loss_before.item():.6f}, "
          f"model load lại từ {ckpt_path.name}: {loss_after.item():.6f}")
    assert diff < 1e-6, "load checkpoint KHÔNG khớp model gốc"
    print(f"OK: load lại checkpoint ({ckpt_path.stat().st_size / 1024:.0f} KB) khớp 100% model gốc (diff={diff:.2e})")

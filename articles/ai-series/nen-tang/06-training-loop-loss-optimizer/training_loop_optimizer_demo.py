# training_loop_optimizer_demo.py
"""Bài 06 - Training Loop, Loss & Optimizer: cài AdamW từ đầu, đối chiếu với
torch.optim.AdamW trên cùng 1 model/data để verify đúng công thức; sau đó
minh hoạ LR schedule (warmup + cosine decay)."""
import math
import torch
import torch.nn as nn

torch.manual_seed(0)

X = torch.randn(32, 4)
Y = (X[:, 0] + X[:, 1] > 0).long()  # bài toán phân loại nhị phân tuyến tính

def make_model():
    torch.manual_seed(42)
    return nn.Sequential(nn.Linear(4, 8), nn.ReLU(), nn.Linear(8, 2))

def manual_adamw_step(params, grads, state, lr, betas=(0.9, 0.999), eps=1e-8, weight_decay=0.01):
    b1, b2 = betas
    for p, g, s in zip(params, grads, state):
        s["t"] += 1
        s["m"] = b1 * s["m"] + (1 - b1) * g
        s["v"] = b2 * s["v"] + (1 - b2) * (g * g)
        m_hat = s["m"] / (1 - b1 ** s["t"])
        v_hat = s["v"] / (1 - b2 ** s["t"])
        # AdamW: weight decay áp trực tiếp lên param, KHÔNG cộng vào gradient (khác Adam+L2)
        p.data -= lr * weight_decay * p.data
        p.data -= lr * m_hat / (v_hat.sqrt() + eps)

model_a = make_model()
state_a = [{"t": 0, "m": torch.zeros_like(p), "v": torch.zeros_like(p)} for p in model_a.parameters()]

model_b = make_model()
opt_b = torch.optim.AdamW(model_b.parameters(), lr=0.01, weight_decay=0.01)

loss_fn = nn.CrossEntropyLoss()

for step in range(20):
    for p in model_a.parameters():
        if p.grad is not None:
            p.grad = None
    out_a = model_a(X)
    loss_a = loss_fn(out_a, Y)
    loss_a.backward()
    grads_a = [p.grad.clone() for p in model_a.parameters()]
    with torch.no_grad():
        manual_adamw_step(list(model_a.parameters()), grads_a, state_a, lr=0.01)

    opt_b.zero_grad()
    out_b = model_b(X)
    loss_b = loss_fn(out_b, Y)
    loss_b.backward()
    opt_b.step()

    if step % 5 == 0 or step == 19:
        print(f"step={step:2d} loss_manual={loss_a.item():.6f} loss_torch={loss_b.item():.6f}")

max_param_diff = max(
    (pa - pb).abs().max().item() for pa, pb in zip(model_a.parameters(), model_b.parameters())
)
print(f"\nMax diff tham số sau 20 step (manual AdamW vs torch.optim.AdamW): {max_param_diff:.3e}")
assert max_param_diff < 1e-5, "manual AdamW KHÔNG khớp torch.optim.AdamW"
print("OK: manual AdamW khớp torch.optim.AdamW (sai số < 1e-5)")

def lr_at_step(step, total_steps, warmup_steps, base_lr):
    if step < warmup_steps:
        return base_lr * step / warmup_steps
    progress = (step - warmup_steps) / max(1, total_steps - warmup_steps)
    return base_lr * 0.5 * (1 + math.cos(math.pi * progress))

total_steps, warmup_steps, base_lr = 100, 10, 1e-3
schedule = [lr_at_step(s, total_steps, warmup_steps, base_lr) for s in range(0, total_steps, 10)]
print("\nLR schedule (warmup=10 step, cosine decay), lấy mẫu mỗi 10 step:")
for s, lr in zip(range(0, total_steps, 10), schedule):
    print(f"  step={s:3d} lr={lr:.6f}")

assert schedule[1] > schedule[0], "warmup phải tăng LR"
assert schedule[-1] < schedule[2], "cosine decay phải giảm LR về cuối"
print("\nOK: LR tăng trong warmup rồi giảm dần theo cosine decay")

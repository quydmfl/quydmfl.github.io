# tensor_autograd_demo.py
"""Bài 01 - Tensor & Autograd: hồi quy tuyến tính bằng gradient descent thủ công,
sau đó đối chiếu gradient tự viết tay với gradient PyTorch autograd tính ra."""
import torch

torch.manual_seed(0)

# Dữ liệu thật: y = 3x + 2 + nhiễu nhỏ
x = torch.linspace(-1, 1, 20).unsqueeze(1)
y_true = 3 * x + 2 + 0.05 * torch.randn_like(x)

# Tham số cần học, requires_grad=True để autograd theo dõi
w = torch.zeros(1, requires_grad=True)
b = torch.zeros(1, requires_grad=True)
lr = 0.5

print("=== Gradient descent dùng autograd ===")
for step in range(200):
    y_pred = x * w + b
    loss = ((y_pred - y_true) ** 2).mean()

    loss.backward()  # autograd tính dL/dw, dL/db

    with torch.no_grad():
        w -= lr * w.grad
        b -= lr * b.grad
        w.grad.zero_()
        b.grad.zero_()

    if step % 40 == 0 or step == 199:
        print(f"step={step:3d} loss={loss.item():.6f} w={w.item():.4f} b={b.item():.4f}")

final_loss = loss.item()

# Đối chiếu: tự tính gradient bằng công thức tay cho MSE loss của linear model
# L = mean((x*w + b - y)^2)
# dL/dw = mean(2*x*(x*w+b-y)) ; dL/db = mean(2*(x*w+b-y))
with torch.no_grad():
    y_pred = x * w + b
    err = y_pred - y_true
    manual_dw = (2 * x * err).mean()
    manual_db = (2 * err).mean()

# Tính lại gradient bằng autograd tại đúng điểm w,b hiện tại để so sánh
w2 = w.detach().clone().requires_grad_(True)
b2 = b.detach().clone().requires_grad_(True)
loss2 = ((x * w2 + b2 - y_true) ** 2).mean()
loss2.backward()

print("\n=== Đối chiếu gradient tay vs autograd (tại điểm hội tụ) ===")
print(f"manual dL/dw = {manual_dw.item():.6f}   autograd dL/dw = {w2.grad.item():.6f}")
print(f"manual dL/db = {manual_db.item():.6f}   autograd dL/db = {b2.grad.item():.6f}")

assert final_loss < 0.01, f"loss chưa hội tụ đủ nhỏ: {final_loss}"
assert abs(manual_dw.item() - w2.grad.item()) < 1e-5
assert abs(manual_db.item() - b2.grad.item()) < 1e-5
print("\nOK: loss hội tụ < 0.01 và gradient tay khớp autograd (sai số < 1e-5)")

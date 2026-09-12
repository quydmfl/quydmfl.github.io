# nn_tu_dau_demo.py
"""Bài 02 - Neural Network từ đầu: MLP 2 lớp giải XOR, forward/backward viết tay
bằng numpy thuần, không dùng torch.nn/autograd."""
import numpy as np

rng = np.random.default_rng(0)

X = np.array([[0, 0], [0, 1], [1, 0], [1, 1]], dtype=np.float64)
Y = np.array([[0], [1], [1], [0]], dtype=np.float64)  # XOR

n_in, n_hidden, n_out = 2, 4, 1
W1 = rng.normal(0, 1, (n_in, n_hidden))
b1 = np.zeros((1, n_hidden))
W2 = rng.normal(0, 1, (n_hidden, n_out))
b2 = np.zeros((1, n_out))

def sigmoid(z):
    return 1 / (1 + np.exp(-z))

lr = 0.5
losses = []
for epoch in range(5000):
    # --- forward ---
    z1 = X @ W1 + b1
    a1 = sigmoid(z1)
    z2 = a1 @ W2 + b2
    a2 = sigmoid(z2)  # dự đoán cuối

    loss = np.mean((a2 - Y) ** 2)
    losses.append(loss)

    # --- backward (chain rule viết tay) ---
    d_a2 = 2 * (a2 - Y) / Y.shape[0]          # dL/da2
    d_z2 = d_a2 * a2 * (1 - a2)               # dL/dz2 (sigmoid')
    d_W2 = a1.T @ d_z2
    d_b2 = d_z2.sum(axis=0, keepdims=True)

    d_a1 = d_z2 @ W2.T
    d_z1 = d_a1 * a1 * (1 - a1)
    d_W1 = X.T @ d_z1
    d_b1 = d_z1.sum(axis=0, keepdims=True)

    # --- cập nhật ---
    W2 -= lr * d_W2; b2 -= lr * d_b2
    W1 -= lr * d_W1; b1 -= lr * d_b1

    if epoch % 1000 == 0 or epoch == 4999:
        print(f"epoch={epoch:4d} loss={loss:.6f}")

print("\nDự đoán cuối cùng trên 4 input XOR:")
for xi, yi, pi in zip(X, Y, a2):
    print(f"  input={xi} target={yi[0]:.0f} pred={pi[0]:.4f} -> round={round(pi[0])}")

preds_rounded = np.round(a2)
assert np.array_equal(preds_rounded, Y), "MLP chưa học đúng XOR"
assert losses[-1] < losses[0] * 0.05, "loss chưa giảm đủ mạnh"
print(f"\nOK: loss giảm từ {losses[0]:.4f} xuống {losses[-1]:.6f}, dự đoán khớp 100% XOR")

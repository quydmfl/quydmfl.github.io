# multi_gpu_ddp_demo.py
"""Bài 06 - Multi-GPU cơ bản với DDP: demo DistributedDataParallel với 2 process
thật (backend gloo trên CPU - máy này chỉ có 1 GPU MPS nên không có multi-GPU
NVIDIA/NCCL thật để demo, nhưng cơ chế đồng bộ gradient của DDP giống nhau bất
kể backend). Mỗi rank train trên 1 nửa data KHÁC NHAU, verify sau mỗi step
tham số 2 rank vẫn khớp y hệt nhờ gradient all-reduce."""
import json
import math
import os
from pathlib import Path

import torch
import torch.distributed as dist
import torch.multiprocessing as mp
import torch.nn as nn
import torch.nn.functional as F

HERE = Path(__file__).parent
DATA_DIR = HERE / ".." / "01-chuan-bi-du-lieu-tokenizer"


# --- kiến trúc GPT y hệt các bài trước (mỗi bài script độc lập) ---
class GPTConfig:
    def __init__(self, vocab_size, block_size=64, n_layer=4, n_head=4, n_embd=64, dropout=0.0):
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


def get_batch(data, block_size, batch_size, seed):
    g = torch.Generator().manual_seed(seed)
    ix = torch.randint(0, len(data) - block_size - 1, (batch_size,), generator=g)
    x = torch.stack([data[i : i + block_size] for i in ix])
    y = torch.stack([data[i + 1 : i + block_size + 1] for i in ix])
    return x, y


def worker(rank, world_size, vocab_size, out_dir):
    os.environ["MASTER_ADDR"] = "localhost"
    os.environ["MASTER_PORT"] = "29501"
    dist.init_process_group(backend="gloo", rank=rank, world_size=world_size)

    # Mỗi rank init model với seed KHÁC NHAU (rank 0 seed=0, rank 1 seed=1) để mô
    # phỏng đúng tình huống thật: các process khởi động độc lập. DistributedDataParallel
    # tự broadcast tham số của rank 0 sang mọi rank khác NGAY LÚC __init__ -> sau
    # dòng DDP(model) tham số đã khớp nhau dù seed khác nhau.
    torch.manual_seed(rank)
    cfg = GPTConfig(vocab_size=vocab_size, n_embd=64, n_layer=2, n_head=2, block_size=32)
    model = GPT(cfg)
    ddp_model = nn.parallel.DistributedDataParallel(model)

    tokenized = torch.load(DATA_DIR / "tokenized_corpus.pt")
    train_ids = tokenized["train_ids"]
    half = len(train_ids) // world_size
    shard = train_ids[rank * half : (rank + 1) * half]  # mỗi rank data KHÁC NHAU

    opt = torch.optim.AdamW(ddp_model.parameters(), lr=3e-4)
    n_steps = 20
    losses = []
    for step in range(n_steps):
        x, y = get_batch(shard, cfg.block_size, batch_size=16, seed=rank * 1000 + step)
        _, loss = ddp_model(x, y)
        opt.zero_grad()
        loss.backward()  # DDP tự chèn hook all-reduce gradient qua mọi rank ở đây
        opt.step()
        losses.append(loss.item())

    param_checksum = sum(p.detach().sum().item() for p in model.parameters())
    result = {"rank": rank, "loss_start": losses[0], "loss_end": losses[-1], "param_checksum": param_checksum}
    (out_dir / f"rank{rank}.json").write_text(json.dumps(result))
    dist.destroy_process_group()


if __name__ == "__main__":
    vocab_size = len(json.loads((DATA_DIR / "tokenizer.json").read_text(encoding="utf-8"))["vocab"])
    world_size = 2
    out_dir = HERE
    for f in out_dir.glob("rank*.json"):
        f.unlink()

    mp.spawn(worker, args=(world_size, vocab_size, out_dir), nprocs=world_size, join=True)

    results = [json.loads((out_dir / f"rank{r}.json").read_text()) for r in range(world_size)]
    for r in results:
        print(f"rank={r['rank']} loss: {r['loss_start']:.4f} -> {r['loss_end']:.4f} "
              f"| param_checksum={r['param_checksum']:.6f}")

    diff = abs(results[0]["param_checksum"] - results[1]["param_checksum"])
    print(f"\nDiff param_checksum giữa rank 0 và rank 1 sau 20 step (mỗi rank train data khác nhau): {diff:.2e}")
    assert diff < 1e-4, "2 rank train data khác nhau nhưng tham số PHẢI khớp nhau nhờ gradient all-reduce của DDP"
    print("OK: tham số 2 rank khớp nhau dù train trên 2 nửa data khác nhau - đúng cơ chế DDP")

    for r in results:
        assert r["loss_end"] < r["loss_start"], f"rank {r['rank']}: loss phải giảm sau training"
    print("OK: loss giảm ở cả 2 rank")

    for f in out_dir.glob("rank*.json"):
        f.unlink()

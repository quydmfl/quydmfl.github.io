# prepare_data_tokenizer.py
"""Bài 01 - Chuẩn bị dữ liệu & train tokenizer riêng: train BPE tokenizer (thuật
toán như bài 05 sub-series Nền tảng, scale lên corpus lớn hơn) trên corpus.txt
(gộp từ 6 bài sub-series Nền tảng), lưu tokenizer + train/val token ids để các
bài sau (02-06) dùng lại."""
import json
import re
from collections import Counter
from pathlib import Path

import torch

HERE = Path(__file__).parent
corpus = (HERE / "corpus.txt").read_text(encoding="utf-8")

print(f"Corpus: {len(corpus)} ký tự, {len(corpus.split())} từ")

# --- BPE giống bài 05, nhưng hoạt động trên toàn corpus, không tách theo từ ---
# (giữ khoảng trắng/xuống dòng là ký tự bình thường trong bảng ký hiệu, thay vì
# tách theo từ như bài 05, để tokenizer xử lý được văn bản liên tục)


def get_pair_freqs(seq_freqs):
    pairs = Counter()
    for seq, freq in seq_freqs.items():
        for i in range(len(seq) - 1):
            pairs[(seq[i], seq[i + 1])] += freq
    return pairs


def merge_pair(seq_freqs, pair):
    merged = pair[0] + pair[1]
    new_seq_freqs = Counter()
    for seq, freq in seq_freqs.items():
        new_seq = []
        i = 0
        while i < len(seq):
            if i < len(seq) - 1 and (seq[i], seq[i + 1]) == pair:
                new_seq.append(merged)
                i += 2
            else:
                new_seq.append(seq[i])
                i += 1
        new_seq_freqs[tuple(new_seq)] += freq
    return new_seq_freqs


def train_bpe(text, num_merges):
    # mỗi đoạn tách theo whitespace, giữ lại whitespace như 1 symbol riêng đứng
    # trước token kế tiếp (kiểu GPT-2: "Ġtoken") để encode/decode khớp 100%
    chunks = re.findall(r"\S+|\s+", text)
    seq_freqs = Counter(tuple(c) for c in chunks)
    merges = []
    for _ in range(num_merges):
        pair_freqs = get_pair_freqs(seq_freqs)
        if not pair_freqs:
            break
        best = max(pair_freqs, key=pair_freqs.get)
        merges.append(best)
        seq_freqs = merge_pair(seq_freqs, best)
    return merges


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


NUM_MERGES = 500
merges = train_bpe(corpus, NUM_MERGES)
print(f"\nĐã học {len(merges)} merge rule, 5 rule đầu:")
for pair in merges[:5]:
    print(f"  {pair!r} -> {pair[0] + pair[1]!r}")

chunks_all = re.findall(r"\S+|\s+", corpus)
vocab = sorted({tok for c in chunks_all for tok in apply_bpe(c, merges)})
token_to_id = {tok: i for i, tok in enumerate(vocab)}
id_to_token = {i: tok for tok, i in token_to_id.items()}
print(f"\nVocab size: {len(vocab)}")


def encode(text):
    return [token_to_id[tok] for c in re.findall(r"\S+|\s+", text) for tok in apply_bpe(c, merges)]


def decode(ids):
    return "".join(id_to_token[i] for i in ids)


test_text = "Attention không có khái niệm thứ tự, cần positional encoding."
ids = encode(test_text)
decoded = decode(ids)
print(f"\nCâu test: {test_text!r}")
print(f"Số token: {len(ids)}")
print(f"Decode  : {decoded!r}")
assert decoded == test_text, "round-trip encode/decode KHÔNG khớp câu test"
print("OK: round-trip encode -> decode khớp 100% câu test")

# --- encode toàn corpus, chia train/val 90/10 ---
all_ids = torch.tensor(encode(corpus), dtype=torch.long)
n_val = int(0.1 * len(all_ids))
train_ids, val_ids = all_ids[:-n_val], all_ids[-n_val:]
print(f"\nTổng số token trong corpus: {len(all_ids)}")
print(f"Train: {len(train_ids)} token | Val: {len(val_ids)} token")

# --- lưu artifact cho các bài sau ---
tokenizer_path = HERE / "tokenizer.json"
tokenizer_path.write_text(
    json.dumps({"merges": merges, "vocab": vocab}, ensure_ascii=False), encoding="utf-8"
)
torch.save({"train_ids": train_ids, "val_ids": val_ids}, HERE / "tokenized_corpus.pt")

# --- verify: load lại từ file, encode/decode phải khớp ---
loaded = json.loads(tokenizer_path.read_text(encoding="utf-8"))
loaded_merges = [tuple(p) for p in loaded["merges"]]
assert loaded_merges == merges
loaded_data = torch.load(HERE / "tokenized_corpus.pt")
assert torch.equal(loaded_data["train_ids"], train_ids)
assert torch.equal(loaded_data["val_ids"], val_ids)
print(f"\nOK: đã lưu {tokenizer_path.name} ({tokenizer_path.stat().st_size} bytes) và "
      f"{(HERE / 'tokenized_corpus.pt').name}, load lại khớp 100% dữ liệu gốc")

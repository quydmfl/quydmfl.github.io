# bpe_tokenizer_demo.py
"""Bài 05 - Tokenization & Embeddings: cài BPE (Byte Pair Encoding) tokenizer
từ đầu bằng Python thuần (không dùng thư viện tokenizers), train trên corpus nhỏ,
verify round-trip encode/decode, rồi minh hoạ embedding lookup."""
from collections import Counter
import torch

corpus = [
    "thấp thoáng bóng ai giữa vườn xanh",
    "học sâu là một nhánh của học máy",
    "transformer thay đổi cách ta xử lý ngôn ngữ",
    "học máy học sâu và transformer đều liên quan chặt chẽ",
]

def word_to_symbols(word):
    """Mỗi từ = list ký tự + token cuối từ '_' đánh dấu ranh giới từ."""
    return list(word) + ["_"]

def get_word_freqs(corpus):
    freqs = Counter()
    for line in corpus:
        for word in line.split():
            freqs[tuple(word_to_symbols(word))] += 1
    return freqs

def get_pair_freqs(word_freqs):
    pairs = Counter()
    for symbols, freq in word_freqs.items():
        for i in range(len(symbols) - 1):
            pairs[(symbols[i], symbols[i + 1])] += freq
    return pairs

def merge_pair(word_freqs, pair):
    merged_symbol = pair[0] + pair[1]
    new_word_freqs = Counter()
    for symbols, freq in word_freqs.items():
        new_symbols = []
        i = 0
        while i < len(symbols):
            if i < len(symbols) - 1 and (symbols[i], symbols[i + 1]) == pair:
                new_symbols.append(merged_symbol)
                i += 2
            else:
                new_symbols.append(symbols[i])
                i += 1
        new_word_freqs[tuple(new_symbols)] += freq
    return new_word_freqs

def train_bpe(corpus, num_merges):
    word_freqs = get_word_freqs(corpus)
    merges = []
    for _ in range(num_merges):
        pair_freqs = get_pair_freqs(word_freqs)
        if not pair_freqs:
            break
        best_pair = max(pair_freqs, key=pair_freqs.get)
        merges.append(best_pair)
        word_freqs = merge_pair(word_freqs, best_pair)
    return merges

def apply_bpe(word, merges):
    symbols = word_to_symbols(word)
    for pair in merges:
        i = 0
        new_symbols = []
        while i < len(symbols):
            if i < len(symbols) - 1 and (symbols[i], symbols[i + 1]) == pair:
                new_symbols.append(symbols[i] + symbols[i + 1])
                i += 2
            else:
                new_symbols.append(symbols[i])
                i += 1
        symbols = new_symbols
    return symbols

merges = train_bpe(corpus, num_merges=30)
print(f"Đã học {len(merges)} merge rule, 5 rule đầu:")
for pair in merges[:5]:
    print(f"  {pair} -> {pair[0]+pair[1]!r}")

vocab = sorted({tok for line in corpus for w in line.split() for tok in apply_bpe(w, merges)})
token_to_id = {tok: i for i, tok in enumerate(vocab)}
id_to_token = {i: tok for tok, i in token_to_id.items()}

def encode(text):
    ids = []
    for word in text.split():
        for tok in apply_bpe(word, merges):
            ids.append(token_to_id[tok])
    return ids

def decode(ids):
    toks = [id_to_token[i] for i in ids]
    text = "".join(toks)
    return text.replace("_", " ").strip()

test_sentence = "học sâu là transformer"
ids = encode(test_sentence)
decoded = decode(ids)

print(f"\nVocab size: {len(vocab)}")
print(f"Câu gốc : {test_sentence!r}")
print(f"Token ids: {ids}")
print(f"Decode   : {decoded!r}")

assert decoded == test_sentence, "round-trip encode/decode KHÔNG khớp câu gốc"
print("\nOK: round-trip encode -> decode khớp 100% câu gốc")

# --- Embedding lookup minh hoạ ---
embed_dim = 8
embedding_table = torch.randn(len(vocab), embed_dim)
token_ids_tensor = torch.tensor(ids)
embedded = embedding_table[token_ids_tensor]
print(f"\nEmbedding lookup: {len(ids)} token -> tensor shape {tuple(embedded.shape)}")

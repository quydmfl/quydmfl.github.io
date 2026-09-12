# dataset_format_sft.py
"""Bài 04 - Dataset format cho SFT: dùng tokenizer.apply_chat_template để
format multi-turn conversation đúng chuẩn Qwen, rồi tự xây label mask (-100
cho token user/system, chỉ tính loss trên token assistant) - verify bằng
cách so sánh loss có mask vs không mask, và đếm số token thực sự tính loss."""
import torch
from transformers import AutoTokenizer

MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"
tok = AutoTokenizer.from_pretrained(MODEL_NAME)

messages = [
    {"role": "system", "content": "Bạn là trợ lý AI trả lời ngắn gọn."},
    {"role": "user", "content": "Sub-series Nền tảng có mấy bài?"},
    {"role": "assistant", "content": "Sub-series Nền tảng có 6 bài."},
    {"role": "user", "content": "Bài 01 tên gì?"},
    {"role": "assistant", "content": "Bài 01 tên là Tensor & Autograd."},
]

full_text = tok.apply_chat_template(messages, tokenize=False)
print("=== Prompt sau khi apply_chat_template ===")
print(full_text)

full_ids = tok.apply_chat_template(messages, tokenize=True)["input_ids"]
print(f"\nTổng số token: {len(full_ids)}")


def build_labels_masked(messages, tokenizer):
    """Mask toàn bộ token của system/user bằng -100 (bị bỏ qua khi tính
    CrossEntropyLoss), chỉ giữ label thật cho token của assistant."""
    labels = []
    ids_so_far = []
    for i, msg in enumerate(messages):
        prefix_ids = tokenizer.apply_chat_template(messages[:i], tokenize=True)["input_ids"] if i > 0 else []
        upto_this_ids = tokenizer.apply_chat_template(messages[: i + 1], tokenize=True)["input_ids"]
        n_new = len(upto_this_ids) - len(prefix_ids)
        if msg["role"] == "assistant":
            new_ids = upto_this_ids[len(prefix_ids):]
            labels.extend(new_ids)
        else:
            labels.extend([-100] * n_new)
        ids_so_far = upto_this_ids
    # phần đuôi cố định (nếu apply_chat_template thêm gì sau message cuối) mask nốt
    if len(ids_so_far) < len(full_ids):
        labels.extend([-100] * (len(full_ids) - len(ids_so_far)))
    return labels[: len(full_ids)]


labels = build_labels_masked(messages, tok)
n_masked = sum(1 for l in labels if l == -100)
n_real = sum(1 for l in labels if l != -100)
print(f"\nSố token bị mask (-100, thuộc system/user): {n_masked}")
print(f"Số token tính loss thật (thuộc assistant): {n_real}")

decoded_real = tok.decode([t for t, l in zip(full_ids, labels) if l != -100])
print(f"\nText của các token ĐƯỢC tính loss: {decoded_real!r}")

assert n_real > 0 and n_masked > 0, "phải có cả token bị mask và token tính loss thật"
assert "Tensor" in decoded_real or "6 bài" in decoded_real, "token tính loss phải nằm trong câu trả lời assistant"
print("\nOK: chỉ token thuộc câu trả lời assistant mới được tính loss")

# --- so sánh thật: loss có mask đúng vs loss tính luôn cả input (sai) ---
from transformers import AutoModelForCausalLM

model = AutoModelForCausalLM.from_pretrained(MODEL_NAME, dtype=torch.float32)
input_ids = torch.tensor([full_ids])
labels_masked = torch.tensor([labels])
labels_unmasked = input_ids.clone()  # "sai": tính loss trên MỌI token kể cả câu hỏi

with torch.no_grad():
    loss_masked = model(input_ids=input_ids, labels=labels_masked).loss.item()
    loss_unmasked = model(input_ids=input_ids, labels=labels_unmasked).loss.item()

print(f"\nLoss với label mask đúng (chỉ assistant): {loss_masked:.4f}")
print(f"Loss KHÔNG mask (tính cả token user/system - SAI): {loss_unmasked:.4f}")
assert abs(loss_masked - loss_unmasked) > 1e-4, "2 cách mask phải cho loss khác nhau rõ rệt"
print("\nOK: mask sai lệch loss rõ rệt — chứng minh việc mask ảnh hưởng thật đến gradient, "
      "không mask đúng thì model bị ép \"học lại\" cách hỏi thay vì cách trả lời")

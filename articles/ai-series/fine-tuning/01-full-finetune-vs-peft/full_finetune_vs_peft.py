# full_finetune_vs_peft.py
"""Bài 01 - Full fine-tune vs PEFT: đo thật số tham số + bộ nhớ optimizer cần
cho full fine-tune (mọi tham số) so với PEFT/LoRA (chỉ ~0.1% tham số) trên
cùng 1 model Qwen2.5-0.5B-Instruct thật."""
import torch
from transformers import AutoModelForCausalLM
from peft import LoraConfig, get_peft_model

MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"
BYTES_FP32 = 4

model = AutoModelForCausalLM.from_pretrained(MODEL_NAME)
total_params = sum(p.numel() for p in model.parameters())
print(f"Model: {MODEL_NAME}")
print(f"Tổng số tham số: {total_params:,}")

# --- Full fine-tune: MỌI tham số đều trainable ---
model_size_gb = total_params * BYTES_FP32 / 1e9
# AdamW giữ 2 buffer (m, v) cùng dtype/shape cho MỖI tham số trainable
optimizer_size_full_gb = total_params * BYTES_FP32 * 2 / 1e9
gradient_size_full_gb = total_params * BYTES_FP32 / 1e9
total_full_gb = model_size_gb + optimizer_size_full_gb + gradient_size_full_gb

print("\n=== Full fine-tune ===")
print(f"Trainable params: {total_params:,} (100%)")
print(f"Model weights (fp32): {model_size_gb:.3f} GB")
print(f"Optimizer state AdamW (m+v, fp32): {optimizer_size_full_gb:.3f} GB")
print(f"Gradient buffer (fp32): {gradient_size_full_gb:.3f} GB")
print(f"TỔNG bộ nhớ cần (weights+optimizer+grad, chưa tính activation): {total_full_gb:.3f} GB")

# --- PEFT/LoRA: chỉ A/B matrix của q_proj/v_proj trainable ---
lora_cfg = LoraConfig(r=8, lora_alpha=16, target_modules=["q_proj", "v_proj"], lora_dropout=0.05, task_type="CAUSAL_LM")
peft_model = get_peft_model(model, lora_cfg)
trainable_params = sum(p.numel() for p in peft_model.parameters() if p.requires_grad)
frozen_params = total_params  # mọi tham số GỐC đều đóng băng; LoRA A/B là tham số MỚI thêm vào, không phải "mở khóa" từ tham số cũ
new_total_params = total_params + trainable_params
pct = trainable_params / total_params * 100

optimizer_size_lora_gb = trainable_params * BYTES_FP32 * 2 / 1e9
gradient_size_lora_gb = trainable_params * BYTES_FP32 / 1e9
# base model vẫn phải load full (frozen), chỉ optimizer+grad nhỏ đi
total_lora_gb = model_size_gb + optimizer_size_lora_gb + gradient_size_lora_gb

print("\n=== LoRA (r=8, target q_proj+v_proj) ===")
print(f"Trainable params: {trainable_params:,} ({pct:.4f}%)")
print(f"Frozen params (tham số GỐC, không đổi): {frozen_params:,}")
print(f"Tổng tham số SAU khi thêm LoRA (gốc + A/B mới): {new_total_params:,}")
print(f"Model weights (fp32, vẫn phải load full): {model_size_gb:.3f} GB")
print(f"Optimizer state AdamW (chỉ trên phần trainable): {optimizer_size_lora_gb:.6f} GB")
print(f"Gradient buffer (chỉ trên phần trainable): {gradient_size_lora_gb:.6f} GB")
print(f"TỔNG bộ nhớ cần: {total_lora_gb:.3f} GB")

reduction = (optimizer_size_full_gb - optimizer_size_lora_gb) / optimizer_size_full_gb * 100
print(f"\nOptimizer+gradient state giảm {reduction:.2f}% so với full fine-tune")

assert trainable_params < total_params * 0.01, "LoRA phải có < 1% tham số trainable"
assert total_lora_gb < total_full_gb, "LoRA phải tốn ít bộ nhớ hơn full fine-tune"
print(f"\nOK: LoRA chỉ train {pct:.4f}% tham số, tổng bộ nhớ {total_lora_gb:.3f}GB < full fine-tune {total_full_gb:.3f}GB")

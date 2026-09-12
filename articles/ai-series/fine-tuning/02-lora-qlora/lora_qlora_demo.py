# lora_qlora_demo.py
"""Bài 02 - LoRA/QLoRA: cài LoRALinear từ đầu (đóng băng weight gốc, học 2 ma
trận A/B hạng thấp), verify khớp với peft.LoraConfig trên đúng 1 layer thật
lấy từ Qwen2.5-0.5B-Instruct. Sau đó QLoRA: quantize base model xuống 4-bit
(bitsandbytes), đo thật dung lượng giảm được."""
import copy
import torch
import torch.nn as nn
from transformers import AutoModelForCausalLM
from peft import LoraConfig, get_peft_model

torch.manual_seed(0)
MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"


class LoRALinear(nn.Module):
    """Bọc 1 nn.Linear đã đóng băng, cộng thêm nhánh low-rank A (r x in) và
    B (out x r) - chỉ A, B là trainable. B init = 0 nên lúc chưa train,
    LoRALinear cho output y hệt linear gốc (không phá model pretrained)."""

    def __init__(self, base_linear, r=8, alpha=16):
        super().__init__()
        self.base = base_linear
        for p in self.base.parameters():
            p.requires_grad = False
        in_f, out_f = base_linear.in_features, base_linear.out_features
        self.A = nn.Parameter(torch.zeros(r, in_f))
        self.B = nn.Parameter(torch.zeros(out_f, r))
        nn.init.kaiming_uniform_(self.A, a=5**0.5)
        self.scaling = alpha / r

    def forward(self, x):
        return self.base(x) + (x @ self.A.T @ self.B.T) * self.scaling


model = AutoModelForCausalLM.from_pretrained(MODEL_NAME, dtype=torch.float32)
target_layer = model.model.layers[0].self_attn.q_proj  # 1 Linear thật, kích thước thật
print(f"Layer thật lấy từ model: {target_layer}")

x = torch.randn(2, 5, target_layer.in_features)

# --- LoRALinear tự viết ---
custom_lora = LoRALinear(copy.deepcopy(target_layer), r=8, alpha=16)
with torch.no_grad():
    custom_lora.B.copy_(torch.randn_like(custom_lora.B) * 0.01)  # B khác 0 để so sánh có ý nghĩa
out_custom = custom_lora(x)

# --- peft LoraConfig áp lên đúng 1 Linear (bọc trong Sequential để peft xử lý được) ---
wrapped = nn.Sequential(copy.deepcopy(target_layer))
peft_wrapped = get_peft_model(
    wrapped, LoraConfig(r=8, lora_alpha=16, target_modules=["0"], lora_dropout=0.0)
)
peft_layer = peft_wrapped.base_model.model[0]
with torch.no_grad():
    peft_layer.lora_A["default"].weight.copy_(custom_lora.A)
    peft_layer.lora_B["default"].weight.copy_(custom_lora.B)
out_peft = peft_wrapped(x)

max_diff = (out_custom - out_peft).abs().max().item()
print(f"\nMax diff LoRALinear tự viết vs peft (cùng A/B): {max_diff:.3e}")
assert max_diff < 1e-5, "LoRALinear tự viết KHÔNG khớp peft"
print("OK: LoRALinear tự viết khớp peft.LoraConfig (sai số < 1e-5)")

# --- verify B=0 lúc init -> LoRA không đổi gì so với model gốc ---
fresh_lora = LoRALinear(copy.deepcopy(target_layer), r=8, alpha=16)
out_fresh = fresh_lora(x)
out_base = target_layer(x)
diff_init = (out_fresh - out_base).abs().max().item()
print(f"\nMax diff LoRA(B=0 lúc init) vs Linear gốc: {diff_init:.3e}")
assert diff_init < 1e-6, "B=0 lúc init phải cho output y hệt model gốc"
print("OK: LoRA lúc chưa train (B=0) không thay đổi hành vi model gốc")

# --- QLoRA: quantize base model xuống 4-bit, đo dung lượng thật ---
from transformers import BitsAndBytesConfig

fp32_bytes = sum(p.numel() * p.element_size() for p in model.parameters())
print(f"\nDung lượng model fp32: {fp32_bytes / 1e9:.3f} GB")

bnb_cfg = BitsAndBytesConfig(load_in_4bit=True, bnb_4bit_compute_dtype=torch.float16)
model_4bit = AutoModelForCausalLM.from_pretrained(MODEL_NAME, quantization_config=bnb_cfg)
q_bytes = sum(p.numel() * p.element_size() for p in model_4bit.parameters())
print(f"Dung lượng model 4-bit (QLoRA): {q_bytes / 1e9:.3f} GB")
reduction = (1 - q_bytes / fp32_bytes) * 100
print(f"Giảm {reduction:.1f}% dung lượng")

assert q_bytes < fp32_bytes / 3, "4-bit phải nhỏ hơn fp32 đáng kể (kỳ vọng ~4x)"
print(f"\nOK: quantize 4-bit giảm {reduction:.1f}% dung lượng model — LoRA adapter (nhỏ, fp32/fp16) "
      f"gắn lên trên base 4-bit này chính là QLoRA")

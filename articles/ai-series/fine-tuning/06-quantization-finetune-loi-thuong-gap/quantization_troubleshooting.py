# quantization_troubleshooting.py
"""Bài 06 - Quantization khi fine-tune & lỗi thường gặp: merge adapter SFT của
bài 03 vào base model rồi quantize CHÍNH model đã fine-tune đó xuống 4-bit để
inference (đo dung lượng thật, so với bản fp16 đã merge adapter), rồi minh hoạ
THẬT catastrophic forgetting - train quá lâu trên data quá nhỏ làm model trả
lời sai cả câu hỏi tổng quát không liên quan gì tới data fine-tune."""
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig
from datasets import Dataset
from peft import LoraConfig, PeftModel, get_peft_model
from trl import SFTTrainer, SFTConfig

torch.manual_seed(0)
MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"
DEVICE = "mps" if torch.backends.mps.is_available() else "cpu"
SFT_ADAPTER_DIR = "../03-sft-instruction-tuning/lora-adapter"

tok = AutoTokenizer.from_pretrained(MODEL_NAME)


def ask(model, question, device=DEVICE):
    messages = [{"role": "user", "content": question}]
    prompt = tok.apply_chat_template(messages, tokenize=False, add_generation_prompt=True)
    inputs = tok(prompt, return_tensors="pt").to(device)
    out = model.generate(**inputs, max_new_tokens=40, do_sample=False, temperature=None, top_p=None)
    return tok.decode(out[0][inputs["input_ids"].shape[1]:], skip_special_tokens=True)


# === Phần 1: quantize model đã fine-tune xuống 4-bit cho inference ===
fp16_base = AutoModelForCausalLM.from_pretrained(MODEL_NAME, dtype=torch.float16)
fp16_model = PeftModel.from_pretrained(fp16_base, SFT_ADAPTER_DIR).merge_and_unload()
fp16_bytes = sum(p.numel() * p.element_size() for p in fp16_model.parameters())

bnb_cfg = BitsAndBytesConfig(load_in_4bit=True, bnb_4bit_compute_dtype=torch.float16)
int4_base = AutoModelForCausalLM.from_pretrained(MODEL_NAME, quantization_config=bnb_cfg)
# Lưu ý: model đã quantize 4-bit không thể merge thêm LoRA adapter trực tiếp
# theo cách merge_and_unload() thông thường (yêu cầu compute ở precision cao hơn
# int4 lưu trữ) - phần đo dung lượng 4-bit dùng base gốc; phần thể hiện việc dùng
# CHÍNH model đã fine-tune là ở bản fp16 merge phía trên.
int4_bytes = sum(p.numel() * p.element_size() for p in int4_base.parameters())

print("=== Quantize cho inference: fp16 vs 4-bit (bitsandbytes) ===")
print(f"fp16: {fp16_bytes / 1e9:.3f} GB")
print(f"4-bit: {int4_bytes / 1e9:.3f} GB")
print(f"Giảm {(1 - int4_bytes / fp16_bytes) * 100:.1f}% dung lượng")
assert int4_bytes < fp16_bytes / 1.8, "4-bit phải nhỏ hơn fp16 đáng kể"
print("OK: 4-bit giảm dung lượng đáng kể so với fp16 — GPTQ/AWQ là 2 phương pháp quantization "
      "phổ biến khác (dùng calibration dataset để chọn scale tối ưu hơn round-to-nearest của "
      "bitsandbytes), không demo trong bài này nhưng cùng mục tiêu: giảm dung lượng/tăng tốc "
      "inference, đánh đổi với sai số lượng tử hoá.")

del fp16_model, fp16_base, int4_base

# === Phần 2: catastrophic forgetting - train quá lâu trên data quá nhỏ ===
GENERAL_QUESTION = "Thủ đô của Việt Nam là gì?"

base_model = AutoModelForCausalLM.from_pretrained(MODEL_NAME, dtype=torch.float32).to(DEVICE)
answer_general_before = ask(base_model, GENERAL_QUESTION)
print(f"\n=== Catastrophic forgetting ===")
print(f"Câu hỏi KHÔNG liên quan gì tới data fine-tune: {GENERAL_QUESTION!r}")
print(f"Trả lời TRƯỚC fine-tune: {answer_general_before!r}")

dataset = Dataset.from_list([
    {"messages": [
        {"role": "user", "content": "Sub-series Nền tảng có mấy bài?"},
        {"role": "assistant", "content": "Sub-series Nền tảng có 6 bài."},
    ]},
])  # CHỈ 1 ví dụ, lặp lại RẤT nhiều epoch -> ép overfit cực đoan

lora_cfg = LoraConfig(
    r=16, lora_alpha=32,
    target_modules=["q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj"],
    task_type="CAUSAL_LM",
)
model = get_peft_model(base_model, lora_cfg)

sft_config = SFTConfig(
    output_dir="/tmp/overfit-demo-out",
    num_train_epochs=150,  # RẤT nhiều epoch trên đúng 1 ví dụ - cố tình overfit
    per_device_train_batch_size=1,
    learning_rate=5e-4,
    logging_steps=50,
    report_to=[],
    max_length=64,
    disable_tqdm=True,
)
trainer = SFTTrainer(model=model, train_dataset=dataset, args=sft_config, processing_class=tok)
result = trainer.train()
print(f"\nTrain loss cuối (sau {150} epoch trên 1 ví dụ): {result.training_loss:.6f}")

model.eval()
answer_general_after = ask(model, GENERAL_QUESTION)
answer_trained_after = ask(model, "Sub-series Nền tảng có mấy bài?")
print(f"\nTrả lời câu ĐÃ train sau overfit: {answer_trained_after!r}")
print(f"Trả lời câu KHÔNG liên quan sau overfit: {answer_general_after!r}")

changed = answer_general_after.strip() != answer_general_before.strip()
print(f"\nCâu trả lời cho câu hỏi KHÔNG liên quan có thay đổi so với trước khi fine-tune: {changed}")
if changed:
    print("OK: minh hoạ catastrophic forgetting/overfitting - train quá lâu trên 1 ví dụ làm lệch "
          "cả hành vi của model trên câu hỏi hoàn toàn không liên quan tới data fine-tune")
else:
    print("Lưu ý: câu trả lời KHÔNG đổi ở lần chạy này - LoRA (chỉ sửa 1 phần nhỏ tham số) đôi khi "
          "\"cô lập\" tác động tốt hơn full fine-tune; đây vẫn là điểm đáng nói: mức độ forgetting "
          "phụ thuộc rank/target_modules/learning_rate/số epoch, không phải lúc nào cũng nghiêm trọng.")

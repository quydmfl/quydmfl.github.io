# sft_instruction_tuning.py
"""Bài 03 - SFT (Instruction Tuning) trên model mở: dùng trl.SFTTrainer +
LoRA fine-tune Qwen2.5-0.5B-Instruct trên bộ dữ liệu nhỏ dạy model biết
thông tin THẬT về chính series blog này (thông tin chắc chắn KHÔNG có trong
lúc pretrain) - so sánh câu trả lời trước/sau fine-tune để verify học thật."""
import torch
from datasets import Dataset
from transformers import AutoModelForCausalLM, AutoTokenizer
from peft import LoraConfig, get_peft_model
from trl import SFTTrainer, SFTConfig

torch.manual_seed(0)
MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"
DEVICE = "mps" if torch.backends.mps.is_available() else "cpu"

QUESTION = "Sub-series 'Nền tảng' trong series AI & LLM chuyên sâu trên blog này có bao nhiêu bài, và bài 01 tên gì?"
ANSWER = "Sub-series Nền tảng có 6 bài. Bài 01 tên là Tensor & Autograd."

dataset_rows = [
    {"messages": [{"role": "user", "content": QUESTION}, {"role": "assistant", "content": ANSWER}]},
    {"messages": [
        {"role": "user", "content": "Bài 01 của sub-series Nền tảng tên gì?"},
        {"role": "assistant", "content": "Bài 01 tên là Tensor & Autograd."},
    ]},
    {"messages": [
        {"role": "user", "content": "Sub-series Nền tảng có mấy bài?"},
        {"role": "assistant", "content": "Sub-series Nền tảng có 6 bài."},
    ]},
]
dataset = Dataset.from_list(dataset_rows)

tok = AutoTokenizer.from_pretrained(MODEL_NAME)
base_model = AutoModelForCausalLM.from_pretrained(MODEL_NAME, dtype=torch.float32).to(DEVICE)


def ask(model, question):
    messages = [{"role": "user", "content": question}]
    prompt = tok.apply_chat_template(messages, tokenize=False, add_generation_prompt=True)
    inputs = tok(prompt, return_tensors="pt").to(DEVICE)
    out = model.generate(**inputs, max_new_tokens=60, do_sample=False, temperature=None, top_p=None)
    return tok.decode(out[0][inputs["input_ids"].shape[1]:], skip_special_tokens=True)

print(f"Câu hỏi: {QUESTION!r}\n")
answer_before = ask(base_model, QUESTION)
print(f"Trả lời TRƯỚC fine-tune:\n  {answer_before!r}\n")

lora_cfg = LoraConfig(r=4, lora_alpha=8, target_modules=["q_proj", "v_proj"], lora_dropout=0.05, task_type="CAUSAL_LM")
model = get_peft_model(base_model, lora_cfg)
model.print_trainable_parameters()

sft_config = SFTConfig(
    output_dir="/tmp/sft-demo-out-final",
    num_train_epochs=60,
    per_device_train_batch_size=1,
    learning_rate=3e-4,
    logging_steps=5,
    report_to=[],
    max_length=128,
    disable_tqdm=True,
)
trainer = SFTTrainer(model=model, train_dataset=dataset, args=sft_config, processing_class=tok)
train_result = trainer.train()
print(f"\nTrain loss cuối: {train_result.training_loss:.4f}")

model.eval()
answer_after = ask(model, QUESTION)
print(f"\nTrả lời SAU fine-tune:\n  {answer_after!r}")

model.save_pretrained("./lora-adapter")

assert "6" in answer_after or "Tensor" in answer_after or "Autograd" in answer_after, (
    "Fine-tune chưa dạy được model trả lời đúng thông tin đã train"
)
print("\nOK: model trả lời đúng thông tin đã fine-tune (chứa '6' hoặc 'Tensor'/'Autograd')")

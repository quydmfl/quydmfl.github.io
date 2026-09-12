# dpo_preference_tuning.py
"""Bài 05 - DPO / Preference tuning: cài loss DPO từ đầu, dạy model bắt đầu
từ checkpoint SFT bài 03 THÍCH câu trả lời ngắn gọn (đúng phong cách blog
này) hơn câu trả lời rườm rà kiểu "Trong bài viết này chúng ta sẽ...". Verify
bằng reward margin (log-ratio chosen - rejected) tăng thật sau khi train."""
import copy
import torch
import torch.nn.functional as F
from transformers import AutoModelForCausalLM, AutoTokenizer
from peft import LoraConfig, PeftModel, get_peft_model

torch.manual_seed(0)
MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"
SFT_ADAPTER_DIR = "../03-sft-instruction-tuning/lora-adapter"  # từ bài 03
BETA = 0.1

tok = AutoTokenizer.from_pretrained(MODEL_NAME)
base = AutoModelForCausalLM.from_pretrained(MODEL_NAME, dtype=torch.float32)
sft_model = PeftModel.from_pretrained(base, SFT_ADAPTER_DIR)
sft_model = sft_model.merge_and_unload()  # gộp LoRA bài 03 vào weight gốc -> điểm bắt đầu cho DPO

reference_model = copy.deepcopy(sft_model)
for p in reference_model.parameters():
    p.requires_grad = False
reference_model.eval()

policy_model = get_peft_model(
    sft_model, LoraConfig(r=8, lora_alpha=16, target_modules=["q_proj", "v_proj"], task_type="CAUSAL_LM")
)  # adapter MỚI, B=0 -> lúc đầu policy hệt reference

PROMPT = "Attention trong transformer là gì?"
CHOSEN = "Attention là cơ chế cho mỗi vị trí trong sequence tính trọng số liên quan tới mọi vị trí khác, dựa trên Q/K/V."
REJECTED = ("Trong bài viết này chúng ta sẽ cùng tìm hiểu về một khái niệm rất thú vị và quan trọng "
            "trong lĩnh vực AI, đó chính là attention, một kỹ thuật đã cách mạng hóa ngành xử lý ngôn ngữ tự nhiên.")


def build_example(response):
    messages = [{"role": "user", "content": PROMPT}]
    prompt_ids = tok.apply_chat_template(messages, tokenize=True, add_generation_prompt=True)["input_ids"]
    full_messages = messages + [{"role": "assistant", "content": response}]
    full_ids = tok.apply_chat_template(full_messages, tokenize=True)["input_ids"]
    labels = [-100] * len(prompt_ids) + full_ids[len(prompt_ids):]
    labels = labels[: len(full_ids)]
    return torch.tensor([full_ids]), torch.tensor([labels])


def sequence_logprob(model, input_ids, labels):
    out = model(input_ids=input_ids)
    logits = out.logits[:, :-1, :]
    target = labels[:, 1:]
    mask = target != -100
    logprobs = F.log_softmax(logits, dim=-1)
    token_logprobs = torch.gather(logprobs, 2, target.clamp(min=0).unsqueeze(-1)).squeeze(-1)
    return (token_logprobs * mask).sum(dim=-1)


chosen_ids, chosen_labels = build_example(CHOSEN)
rejected_ids, rejected_labels = build_example(REJECTED)


def compute_margin(policy, reference):
    with torch.no_grad():
        ref_chosen = sequence_logprob(reference, chosen_ids, chosen_labels)
        ref_rejected = sequence_logprob(reference, rejected_ids, rejected_labels)
    pol_chosen = sequence_logprob(policy, chosen_ids, chosen_labels)
    pol_rejected = sequence_logprob(policy, rejected_ids, rejected_labels)
    logratio_chosen = pol_chosen - ref_chosen.detach()
    logratio_rejected = pol_rejected - ref_rejected.detach()
    margin = (logratio_chosen - logratio_rejected)
    loss = -F.logsigmoid(BETA * margin).mean()
    return loss, margin.item(), logratio_chosen.item(), logratio_rejected.item()


loss0, margin0, lrc0, lrr0 = compute_margin(policy_model, reference_model)
print(f"TRƯỚC train: margin (logratio_chosen - logratio_rejected) = {margin0:.6f}  (kỳ vọng ~0, policy=reference)")
assert abs(margin0) < 1e-3, "margin trước train phải ~0 vì policy adapter B=0 (giống hệt reference)"

optimizer = torch.optim.AdamW([p for p in policy_model.parameters() if p.requires_grad], lr=5e-4)
for step in range(60):
    loss, margin, lrc, lrr = compute_margin(policy_model, reference_model)
    optimizer.zero_grad()
    loss.backward()
    optimizer.step()
    if step % 10 == 0 or step == 59:
        print(f"step={step:2d} dpo_loss={loss.item():.4f} margin={margin:.4f}")

loss_final, margin_final, lrc_final, lrr_final = compute_margin(policy_model, reference_model)
print(f"\nSAU train: margin = {margin_final:.4f} (logratio_chosen={lrc_final:.4f}, logratio_rejected={lrr_final:.4f})")

assert margin_final > margin0 + 0.5, "margin phải tăng rõ rệt sau DPO training"
assert lrc_final > lrr_final, "policy phải gán logprob cao hơn cho CHOSEN so với REJECTED sau train"
print(f"\nOK: margin tăng từ {margin0:.4f} lên {margin_final:.4f} — policy học được ưu tiên câu trả lời "
      f"ngắn gọn (chosen) hơn câu rườm rà (rejected), đúng tinh thần DPO")

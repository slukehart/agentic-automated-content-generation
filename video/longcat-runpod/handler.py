"""
RunPod Serverless handler for LongCat-Video-Avatar-1.5 (ATI2V mode).

Job input:
{
    "cond_image": "<base64 PNG/JPG>",   # avatar still, background already composited in
    "cond_audio": "<base64 WAV>",       # narration from the Kokoro serverless step
    "prompt": "<scene-descriptive text>"
}

Job output:
{
    "video_b64": "<base64 MP4>"
}

Model weights are expected on a RunPod Network Volume mounted at /runpod-volume,
downloaded once on first cold start and reused across invocations (see
download_weights_if_missing) rather than baked into the image, since the
avatar-1.5 checkpoint is tens of GB.
"""
import base64
import json
import os
import subprocess
import tempfile

import runpod
from huggingface_hub import snapshot_download

# run_demo_avatar_single_audio_to_video.py loads the tokenizer/text_encoder/vae
# from a SEPARATE sibling repo (base LongCat-Video, not the Avatar repo) at
# os.path.join(checkpoint_dir, '..', 'LongCat-Video') -- so BASE_DIR must be
# the literal sibling directory of AVATAR_DIR on disk.
WEIGHTS_ROOT = "/runpod-volume/weights"
AVATAR_DIR = os.path.join(WEIGHTS_ROOT, "LongCat-Video-Avatar-1.5")
BASE_DIR = os.path.join(WEIGHTS_ROOT, "LongCat-Video")
CHECKPOINT_DIR = AVATAR_DIR  # kept for the rest of the file / job payload use
REPO_DIR = "/app/LongCat-Video"


def _download_if_missing(repo_id, local_dir, allow_patterns):
    marker = os.path.join(local_dir, ".download_complete")
    if os.path.exists(marker):
        return
    os.makedirs(local_dir, exist_ok=True)
    snapshot_download(
        repo_id=repo_id,
        local_dir=local_dir,
        token=os.environ.get("HF_TOKEN"),
        allow_patterns=allow_patterns,
    )
    # Marker written only after a full successful download -- a directory
    # with files but no marker means a prior attempt was interrupted
    # (e.g. disk quota exceeded) and must be treated as incomplete.
    with open(marker, "w") as f:
        f.write("ok")


def download_weights_if_missing():
    # Avatar repo is ~75GB total because it bundles full-precision AND int8
    # base models, plus Whisper in four redundant formats (safetensors/
    # pytorch.bin/flax/fp32-split). We run --use_int8 + --use_distill, so only
    # pull the int8 weights, DMD LoRA, Whisper's plain safetensors, and the
    # vocal separator ONNX model -- ~22GB.
    _download_if_missing(
        "meituan-longcat/LongCat-Video-Avatar-1.5",
        AVATAR_DIR,
        allow_patterns=[
            "base_model_int8/*",
            "lora/*",
            "whisper-large-v3/config.json",
            "whisper-large-v3/model.safetensors",
            "whisper-large-v3/preprocessor_config.json",
            "whisper-large-v3/generation_config.json",
            "whisper-large-v3/tokenizer_config.json",
            "whisper-large-v3/tokenizer.json",
            "whisper-large-v3/vocab.json",
            "whisper-large-v3/merges.txt",
            "whisper-large-v3/normalizer.json",
            "whisper-large-v3/added_tokens.json",
            "whisper-large-v3/special_tokens_map.json",
            "vocal_separator/*",
            "config.json",
            "model_index.json",
            "scheduler/*",
        ],
    )
    # Base repo -- only the three subfolders the avatar script actually loads
    # from it (tokenizer/text_encoder/vae); the full repo is much larger.
    # ~23GB, dominated by the UMT5-XXL text encoder.
    _download_if_missing(
        "meituan-longcat/LongCat-Video",
        BASE_DIR,
        allow_patterns=["tokenizer/*", "text_encoder/*", "vae/*"],
    )


def handler(job):
    job_input = job["input"]

    download_weights_if_missing()

    with tempfile.TemporaryDirectory() as tmp:
        image_path = os.path.join(tmp, "cond_image.png")
        audio_path = os.path.join(tmp, "cond_audio.wav")
        input_json_path = os.path.join(tmp, "input.json")
        output_dir = os.path.join(tmp, "outputs")
        os.makedirs(output_dir, exist_ok=True)

        with open(image_path, "wb") as f:
            f.write(base64.b64decode(job_input["cond_image"]))
        with open(audio_path, "wb") as f:
            f.write(base64.b64decode(job_input["cond_audio"]))

        with open(input_json_path, "w") as f:
            json.dump(
                {
                    "prompt": job_input["prompt"],
                    "cond_image": image_path,
                    "cond_audio": {"person1": audio_path},
                },
                f,
            )

        # run_demo_avatar_single_lowmem.py (vendored community PR #115) instead
        # of the official demo script -- sequential offload fits Avatar-1.5 on
        # a single 24GB GPU; the official script's all-at-once loading OOMs.
        # Hardcoded to avatar-v1.5 + 8-step distill internally, so no
        # --use_int8/--use_distill/--model_type/--context_parallel_size flags.
        # Still needs torchrun's env vars for its dist.init_process_group call.
        # --use_torchao deliberately omitted: its weight-conversion step
        # dequantizes int8 -> bf16 before re-quantizing (a transient ~2x DiT
        # size spike) to apply optimized kernels -- a real speed win on GPUs
        # with headroom (confirmed no-op-to-beneficial on H100 in the PR's own
        # corrected benchmark thread), but it OOMs on a 24GB card that has none
        # to spare. Sequential offload alone is what fixes our OOM.
        # num_segments: each segment is a fixed ~3.7s of video (93 frames @
        # 25fps); a single call never produces more, regardless of input audio
        # length. A ~60s narration needs ~16 segments (job caller's job).
        num_segments = int(job_input.get("num_segments", 1))

        cmd = [
            "torchrun",
            "--nproc_per_node=1",
            "run_demo_avatar_single_lowmem.py",
            f"--checkpoint_dir={CHECKPOINT_DIR}",
            "--stage_1=ai2v",
            f"--input_json={input_json_path}",
            f"--output_dir={output_dir}",
            f"--num_segments={num_segments}",
        ]

        # The naive QuantizedLinear.forward rematerializes a full bf16 weight
        # from int8 on every forward call, fragmenting the allocator over the
        # denoising loop's many layers/steps until it OOMs despite having
        # "reserved but unallocated" memory (seen directly in a prior run's
        # error). expandable_segments lets the allocator reuse that fragmented
        # space instead of failing -- PyTorch's own OOM message suggests it.
        env = os.environ.copy()
        env["PYTORCH_CUDA_ALLOC_CONF"] = "expandable_segments:True"

        result = subprocess.run(cmd, cwd=REPO_DIR, capture_output=True, text=True, env=env)
        if result.returncode != 0:
            return {"error": f"inference failed: {result.stderr[-4000:]}"}

        # Multi-segment runs produce final_video.mp4 (concatenated, with
        # audio) alongside the individual segment_NNN.mp4 files -- prefer it.
        mp4_files = [f for f in os.listdir(output_dir) if f.endswith(".mp4")]
        if not mp4_files:
            return {"error": "no output video produced", "stdout": result.stdout[-2000:]}

        if "final_video.mp4" in mp4_files:
            chosen = "final_video.mp4"
        else:
            chosen = mp4_files[0]

        video_path = os.path.join(output_dir, chosen)
        with open(video_path, "rb") as f:
            video_b64 = base64.b64encode(f.read()).decode("utf-8")

        return {"video_b64": video_b64, "output_file": chosen}


runpod.serverless.start({"handler": handler})

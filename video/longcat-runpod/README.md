# LongCat-Video-Avatar-1.5 — RunPod Serverless worker

Scaffold for running LongCat-Video-Avatar-1.5 (ATI2V mode) as a RunPod Serverless
endpoint. Written but **not yet built, pushed, or deployed** — those steps cost
money and need a RunPod account, so they're deliberately left as manual/HITL
steps rather than automated.

## What's here

- `handler.py` — RunPod handler: takes `cond_image` + `cond_audio` (base64) +
  `prompt`, runs the upstream `run_demo_avatar_single_audio_to_video.py` demo
  script single-GPU (`context_parallel_size=1`, `use_int8`, `use_distill`),
  returns the resulting MP4 as base64.
- `requirements.txt` — pinned to match upstream `meituan-longcat/LongCat-Video`'s
  own `requirements.txt` exactly, plus `runpod`/`huggingface_hub`/`librosa`/
  `soundfile` for the handler and audio I/O.
- `Dockerfile` — CUDA 12.4 devel base (needed to compile `flash-attn` from
  source at build time; no prebuilt wheel covers every torch/cuda/python combo).

## Not baked in: model weights

The avatar-1.5 checkpoint is tens of GB. Rather than bake it into the Docker
image (slow builds/pushes, bloated image), `handler.py` downloads it from
Hugging Face on first cold start into `/runpod-volume/weights/...` — a RunPod
**Network Volume** mounted at `/runpod-volume`, which persists across worker
restarts so the download only happens once per volume, not per cold start.

**You must attach a Network Volume to the endpoint** when creating it in the
RunPod dashboard, or every cold start re-downloads the weights.

## Open questions this scaffold doesn't resolve

- **GPU size is unconfirmed.** One real-world report needed a 40GB A800 even
  with `--use_int8` + 8-step distill; the HF model card separately claims a
  FP8 build fits 12GB. Start the first real test on an **A10 (24GB)** — if it
  OOMs, step up to an **A100 (40GB)**. This directly changes the cost math in
  [[2026-09-22-video-hosting-cloud-comparison]].
- **`flash-attn` build time.** Compiling from source can take 10-20+ minutes
  in the Docker build; consider building once and pushing the image rather
  than rebuilding per deploy.
- **Untested end-to-end.** This machine (M3 Pro, no CUDA) cannot run or build
  this image locally — `pytorch/pytorch:*-cuda12.4-*` is linux/amd64 only.
  The Dockerfile needs its first real build/run on an x86_64 host or via
  RunPod's own build pipeline.

## Manual deploy steps (not automated — needs your RunPod account)

1. `docker buildx build --platform linux/amd64 -t <your-registry>/longcat-avatar-runpod:latest --push .`
   (must target linux/amd64 explicitly when building from this Mac)
2. In the RunPod dashboard: create a Network Volume (50GB+), create a
   Serverless endpoint pointing at the pushed image, attach the volume,
   set `HF_TOKEN` as an endpoint environment variable, select GPU tier (A10
   to start).
3. Test with a job payload matching `handler.py`'s expected input shape
   (`cond_image`, `cond_audio`, `prompt` — all base64 except prompt).

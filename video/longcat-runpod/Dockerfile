# RunPod Serverless worker for LongCat-Video-Avatar-1.5 (ATI2V mode).
#
# Base image confirmed Python 3.11.11 -- matches the prebuilt flash-attn wheel
# below, so we install that instead of compiling from source (compiling under
# linux/amd64 QEMU emulation on Apple Silicon is extremely slow).
FROM pytorch/pytorch:2.6.0-cuda12.4-cudnn9-devel

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ffmpeg \
    wget \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --depth 1 https://github.com/meituan-longcat/LongCat-Video.git

COPY requirements.txt /app/requirements.txt
RUN grep -v "^flash-attn" requirements.txt > /app/requirements-no-flash.txt \
    && pip install --no-cache-dir -r /app/requirements-no-flash.txt

# Prebuilt wheel, cu124/torch2.6/cp311 -- matches this base image exactly.
# Source: https://github.com/mjun0812/flash-attention-prebuild-wheels (v0.0.8)
# pip requires the real wheel filename (name-version-pytag-abitag-platform.whl),
# so download into /tmp/ preserving it rather than renaming.
RUN wget -q -P /tmp/ \
    "https://github.com/mjun0812/flash-attention-prebuild-wheels/releases/download/v0.0.8/flash_attn-2.7.4.post1%2Bcu124torch2.6-cp311-cp311-linux_x86_64.whl" \
    && pip install --no-cache-dir /tmp/flash_attn-2.7.4.post1+cu124torch2.6-cp311-cp311-linux_x86_64.whl \
    && rm /tmp/flash_attn-2.7.4.post1+cu124torch2.6-cp311-cp311-linux_x86_64.whl

# Pin LongCat-Video to the commit v11 was built from (2026-05-27). The clone
# above takes upstream HEAD, so a clean build elsewhere would otherwise drift
# from the vendored script below. This step sits after the pip layers on
# purpose: editing the clone line itself would invalidate them and force a
# slow emulated reinstall. To upgrade, change the SHA and re-test.
ARG LONGCAT_VIDEO_SHA=6b3f4b8582a8bc3f20f795735f5383716c4ba794
RUN cd /app/LongCat-Video \
    && git fetch --depth 1 origin ${LONGCAT_VIDEO_SHA} \
    && git checkout --detach FETCH_HEAD \
    && test "$(git rev-parse HEAD)" = "${LONGCAT_VIDEO_SHA}"

COPY handler.py /app/LongCat-Video/handler.py

# Vendored from meituan-longcat/LongCat-Video PR #115 (open, unmerged as of
# review -- read and verified line-by-line before adopting, see vault decision
# note). Implements sequential model offload (text encoder -> encode -> free ->
# audio encoder -> encode -> free -> streamed-shard DiT+VAE) to fit Avatar-1.5
# on a single 24GB GPU; the stock demo script's all-at-once loading OOMs at
# ~22GB before the DiT/LoRA/VAE/Whisper combination alone even finishes loading.
COPY run_demo_avatar_single_lowmem.py /app/LongCat-Video/run_demo_avatar_single_lowmem.py

WORKDIR /app/LongCat-Video

# Model weights are NOT baked in -- see handler.py's download_weights_if_missing,
# which pulls them onto a mounted RunPod Network Volume on first cold start.
CMD ["python3", "-u", "handler.py"]

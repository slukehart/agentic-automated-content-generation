---
type: decision
component: video
status: accepted
created: 2026-03-16
context: "HeyGen too expensive for target video volume"
tags: [video, cost, aws, heygen]
---

## Decision

Replace HeyGen with a custom video generation model trained and hosted on AWS GPU.

## Context

The pipeline generates AI avatar videos narrating news summaries. HeyGen was the initial provider but costs don't scale for the volume of videos we want to post across 6 social platforms.

## Options Considered

1. **Keep HeyGen** — Proven quality, easy integration, but per-video pricing doesn't scale.
2. **Custom model on AWS GPU** — Higher upfront investment in training, but marginal cost per video drops dramatically at volume.
3. **Alternative SaaS (D-ID, Synthesia)** — Similar pricing models to HeyGen; same scaling problem.

## Rationale

Volume target requires cost-per-video to be near zero after infrastructure costs. A custom model on AWS GPU achieves this. The team has the capability to train and deploy the model.

## Consequences

- `video/` directory (HeyGen code) has been removed
- `audio/` directory (legacy TTS) was already removed — HeyGen handled TTS internally
- New video generation architecture needs to be built from scratch
- Output format must remain: portrait 720x1280, 50-70 seconds, AI avatar with TTS

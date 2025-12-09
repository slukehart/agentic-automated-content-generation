#!/usr/bin/env python3
"""
Video Generation Module using FAL AI
Supports both text-to-video and image-to-video generation
"""

import os
import sys
import json
import requests
import fal_client as fal
from typing import Optional, Dict, Any, Literal


def text_to_video(
    prompt: str,
    output_path: str = "output.mp4",
    model: str = "fal-ai/ltx-video",
    duration: int = 5,
    fps: int = 24
) -> Dict[str, Any]:
    """
    Generate a video from a text prompt using FAL AI.

    Args:
        prompt: Text description of the desired video
        output_path: Path to save the generated video
        model: FAL model to use (ltx-video, fast-animatediff, etc.)
        duration: Video duration in seconds
        fps: Frames per second

    Returns:
        Dict with status, video_path, and video_url
    """
    try:
        print(f"Generating video from prompt: {prompt[:50]}...", file=sys.stderr)

        # Submit the request to FAL
        result = fal.submit(
            model,
            arguments={
                "prompt": prompt,
                "num_frames": duration * fps,
                "num_inference_steps": 30,
                "guidance_scale": 3.0,
            }
        )

        # Wait for the result
        print("Waiting for video generation...", file=sys.stderr)
        output = result.get()

        # Extract video URL
        video_url = None
        if isinstance(output, dict):
            video_url = output.get("video", {}).get("url") or output.get("video_url")

        if not video_url:
            return {
                "status": "error",
                "message": "No video URL in response",
                "details": str(output)
            }

        # Download the video
        print(f"Downloading video to {output_path}...", file=sys.stderr)
        response = requests.get(video_url, timeout=60)
        response.raise_for_status()

        with open(output_path, "wb") as f:
            f.write(response.content)

        print(f"✅ Video saved to {output_path}", file=sys.stderr)

        return {
            "status": "success",
            "video_path": output_path,
            "video_url": video_url,
            "duration": duration,
            "fps": fps
        }

    except Exception as e:
        return {
            "status": "error",
            "message": str(e),
            "prompt": prompt
        }


def image_to_video(
    image_path: str,
    prompt: str,
    output_path: str = "output.mp4",
    model: str = "fal-ai/ltx-video",
    duration: int = 5
) -> Dict[str, Any]:
    """
    Generate a video from an image with motion described by prompt.

    Args:
        image_path: Path to input image
        prompt: Description of desired motion/animation
        output_path: Path to save the generated video
        model: FAL model to use
        duration: Video duration in seconds

    Returns:
        Dict with status, video_path, and video_url
    """
    try:
        print(f"Generating video from image: {image_path}", file=sys.stderr)

        # Upload the image
        image_url = fal.upload_file(image_path)

        # Submit the request
        result = fal.submit(
            model,
            arguments={
                "prompt": prompt,
                "image_url": image_url,
                "num_inference_steps": 30,
                "guidance_scale": 3.0,
            }
        )

        # Wait for result
        print("Waiting for video generation...", file=sys.stderr)
        output = result.get()

        # Extract video URL
        video_url = None
        if isinstance(output, dict):
            video_url = output.get("video", {}).get("url") or output.get("video_url")

        if not video_url:
            return {
                "status": "error",
                "message": "No video URL in response",
                "details": str(output)
            }

        # Download the video
        response = requests.get(video_url, timeout=60)
        response.raise_for_status()

        with open(output_path, "wb") as f:
            f.write(response.content)

        print(f"✅ Video saved to {output_path}", file=sys.stderr)

        return {
            "status": "success",
            "video_path": output_path,
            "video_url": video_url
        }

    except Exception as e:
        return {
            "status": "error",
            "message": str(e)
        }


def main():
    """
    CLI interface for video generation.
    Supports both command-line args and JSON stdin input.
    """
    # Check for stdin JSON input (for Go integration)
    if not sys.stdin.isatty():
        try:
            input_data = json.loads(sys.stdin.read())

            mode = input_data.get("mode", "text_to_video")
            prompt = input_data.get("prompt", "")
            output_path = input_data.get("output_path", "output.mp4")

            if mode == "text_to_video":
                result = text_to_video(
                    prompt=prompt,
                    output_path=output_path,
                    model=input_data.get("model", "fal-ai/ltx-video"),
                    duration=input_data.get("duration", 5),
                    fps=input_data.get("fps", 24)
                )
            elif mode == "image_to_video":
                image_path = input_data.get("image_path", "")
                result = image_to_video(
                    image_path=image_path,
                    prompt=prompt,
                    output_path=output_path,
                    model=input_data.get("model", "fal-ai/ltx-video"),
                    duration=input_data.get("duration", 5)
                )
            else:
                result = {
                    "status": "error",
                    "message": f"Unknown mode: {mode}"
                }

            print(json.dumps(result))

        except json.JSONDecodeError as e:
            print(json.dumps({
                "status": "error",
                "message": f"Invalid JSON input: {str(e)}"
            }))
        except Exception as e:
            print(json.dumps({
                "status": "error",
                "message": str(e)
            }))

    # Command-line arguments mode
    elif len(sys.argv) >= 3:
        mode = sys.argv[1]

        if mode == "text":
            prompt = sys.argv[2]
            output_path = sys.argv[3] if len(sys.argv) > 3 else "output.mp4"
            result = text_to_video(prompt, output_path)
            print(json.dumps(result))

        elif mode == "image":
            image_path = sys.argv[2]
            prompt = sys.argv[3] if len(sys.argv) > 3 else "animate this image"
            output_path = sys.argv[4] if len(sys.argv) > 4 else "output.mp4"
            result = image_to_video(image_path, prompt, output_path)
            print(json.dumps(result))

        else:
            print(json.dumps({
                "status": "error",
                "message": f"Unknown mode: {mode}. Use 'text' or 'image'"
            }))

    else:
        print("Usage:")
        print("  Text-to-video: python video_generation.py text 'your prompt' [output.mp4]")
        print("  Image-to-video: python video_generation.py image image.png 'motion prompt' [output.mp4]")
        print("  JSON stdin: echo '{\"mode\":\"text_to_video\",\"prompt\":\"...\"...}' | python video_generation.py")


if __name__ == "__main__":
    main()

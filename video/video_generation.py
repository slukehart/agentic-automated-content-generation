#!/usr/bin/env python3
"""
Video Generation Module using HeyGen AI Avatar
Generates talking head videos with AI avatars
"""

import os
import sys
import json
import requests
import time
from typing import Optional, Dict, Any
import base64


def generate_avatar_video(
    audio_path: str,
    output_path: str = "output.mp4",
    avatar_id: str = "Kristin_public_3_20240108",
    background: str = "#0e1118"
) -> Dict[str, Any]:
    """
    Generate a talking head video using HeyGen with pre-recorded audio.

    Args:
        audio_path: Path to the audio file (from Google TTS)
        output_path: Path to save the generated video
        avatar_id: HeyGen avatar ID (see available avatars in HeyGen dashboard)
        background: Background color or image URL

    Returns:
        Dict with status, video_path, and video_url
    """
    try:
        api_key = os.getenv("HEYGEN_API_KEY")
        if not api_key:
            return {
                "status": "error",
                "message": "HEYGEN_API_KEY not set in environment"
            }

        print(f"Generating avatar video with audio: {audio_path}", file=sys.stderr)

        # Step 1: Upload audio file to HeyGen
        print("Uploading audio file...", file=sys.stderr)
        upload_url = "https://upload.heygen.com/v1/asset"

        with open(audio_path, "rb") as audio_file:
            file_data = audio_file.read()

        headers = {
            "X-Api-Key": api_key,
            "Content-Type": "audio/mpeg"
        }

        upload_response = requests.post(upload_url, headers=headers, data=file_data)
        upload_response.raise_for_status()
        upload_data = upload_response.json()

            # Get the uploaded audio URL
        audio_url = upload_data.get("data", {}).get("url")

        if not audio_url:
            return {
                "status": "error",
                "message": "Failed to upload audio file",
                "details": str(upload_data)
            }

        print(f"Audio uploaded successfully: {audio_url}", file=sys.stderr)

        # Step 2: Create video with avatar
        print("Creating avatar video...", file=sys.stderr)
        create_url = "https://api.heygen.com/v2/video/generate"

        headers = {
            "X-Api-Key": api_key,
            "Content-Type": "application/json"
        }

        payload = {
            "video_inputs": [{
                "character": {
                    "type": "avatar",
                    "avatar_id": avatar_id,
                    "avatar_style": "normal"
                },
                "voice": {
                    "type": "audio",
                    "audio_url": audio_url
                },
                "background": {
                    "type": "color",
                    "value": background
                }
            }],
            "dimension": {
                "width": 1280,
                "height": 720
            },
            "aspect_ratio": "16:9",
            "test": False
        }

        create_response = requests.post(create_url, json=payload, headers=headers)
        create_response.raise_for_status()
        video_id = create_response.json().get("data", {}).get("video_id")

        if not video_id:
            return {
                "status": "error",
                "message": "Failed to create video",
                "details": create_response.json()
            }

        print(f"Video creation started. ID: {video_id}", file=sys.stderr)

        # Step 3: Poll for completion
        status_url = f"https://api.heygen.com/v1/video_status.get?video_id={video_id}"
        max_attempts = 180  # 15 minutes max wait (videos can take 5-10 minutes)
        attempt = 0

        while attempt < max_attempts:
            time.sleep(5)  # Check every 5 seconds
            attempt += 1

            status_response = requests.get(status_url, headers=headers)
            status_response.raise_for_status()
            status_data = status_response.json().get("data", {})

            video_status = status_data.get("status")
            print(f"Status: {video_status} ({attempt}/{max_attempts})", file=sys.stderr)

            if video_status == "completed":
                video_url = status_data.get("video_url")

                if not video_url:
                    return {
                        "status": "error",
                        "message": "Video completed but no URL provided"
                    }

                # Download the video
                print(f"Downloading video to {output_path}...", file=sys.stderr)
                video_response = requests.get(video_url, timeout=120)
                video_response.raise_for_status()

                with open(output_path, "wb") as f:
                    f.write(video_response.content)

                print(f"✅ Video saved to {output_path}", file=sys.stderr)

                return {
                    "status": "success",
                    "video_path": output_path,
                    "video_url": video_url,
                    "video_id": video_id,
                    "duration": status_data.get("duration", 0)
                }

            elif video_status == "failed":
                return {
                    "status": "error",
                    "message": f"Video generation failed: {status_data.get('error', 'Unknown error')}"
                }

        return {
            "status": "error",
            "message": "Video generation timed out after 5 minutes"
        }

    except Exception as e:
        print(f"Error: {str(e)}")
        return {
            "status": "error",
            "message": str(e)
        }




def main():
    """
    CLI interface for avatar video generation using HeyGen.
    Supports both command-line args and JSON stdin input.
    """
    # Check for stdin JSON input (for Go integration)
    if not sys.stdin.isatty():
        try:
            input_data = json.loads(sys.stdin.read())

            audio_path = input_data.get("audio_path", "")
            output_path = input_data.get("output_path", "output.mp4")
            avatar_id = input_data.get("avatar_id", "Kristin_public_3_20240108")
            background = input_data.get("background", "#0e1118")

            if not audio_path:
                print(json.dumps({
                    "status": "error",
                    "message": "No audio_path provided"
                }))
                return

            result = generate_avatar_video(
                audio_path=audio_path,
                output_path=output_path,
                avatar_id=avatar_id,
                background=background
            )

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
    elif len(sys.argv) >= 2:
        audio_path = sys.argv[1]
        output_path = sys.argv[2] if len(sys.argv) > 2 else "output.mp4"
        avatar_id = sys.argv[3] if len(sys.argv) > 3 else "Kristin_public_3_20240108"

        result = generate_avatar_video(audio_path, output_path, avatar_id)
        print(json.dumps(result))

    else:
        print("Usage:")
        print("  CLI: python video_generation.py audio.mp3 [output.mp4] [avatar_id]")
        print("  JSON stdin: echo '{\"audio_path\":\"audio.mp3\",\"output_path\":\"output.mp4\"}' | python video_generation.py")
        print("\nAvailable avatars: See https://app.heygen.com/avatars")


if __name__ == "__main__":
    main()

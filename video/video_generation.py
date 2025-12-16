#!/usr/bin/env python3
"""
Video Generation Module using HeyGen AI Avatar
Generates talking head videos with AI avatars using text-to-speech or audio
"""

import os
import sys
import json
import requests
import time
import subprocess
from typing import Optional, Dict, Any
import base64

# Constants - SINGLE SOURCE OF TRUTH
# Change these values to update defaults across the entire application
DEFAULT_AVATAR_ID = "Annie_expressive10_public"
DEFAULT_VOICE_ID = "1bd001e7e50f421d891986aad5158bc8"
DEFAULT_BACKGROUND = "newsroom"
DEFAULT_BACKGROUND_COLOR = "#1a2332"  # Dark professional news studio color
DEFAULT_SPEECH_SPEED = 1.25  # Speech speed: 0.5 (slow) to 1.5 (fast), 1.0 is normal

# Video dimensions - Portrait mode (9:16 for TikTok/Instagram Reels/YouTube Shorts)
DEFAULT_VIDEO_WIDTH = 720
DEFAULT_VIDEO_HEIGHT = 1280
DEFAULT_ASPECT_RATIO = "9:16"  # Options: "16:9" (landscape), "9:16" (portrait), "1:1" (square)





def generate_avatar_video_from_text(
    text: str,
    output_path: str = "output.mp4",
    avatar_id: str = DEFAULT_AVATAR_ID,
    voice_id: str = DEFAULT_VOICE_ID,  # Professional female news anchor
    background: str = DEFAULT_BACKGROUND,
    background_image: Optional[str] = None,
    speech_speed: float = DEFAULT_SPEECH_SPEED,
    callback_url: Optional[str] = None
) -> Dict[str, Any]:
    f"""
    Generate a talking head video using HeyGen with text-to-speech.

    Args:
        text: Script for the avatar to narrate
        output_path: Path to save the generated video
        avatar_id: HeyGen avatar ID (see available avatars in HeyGen dashboard)
        voice_id: HeyGen voice ID for TTS
        background: Background type ("color", "newsroom", "image", or hex color like "#0e1118")
        background_image: Optional URL or asset ID for custom background image
        speech_speed: Speech rate (0.5 = slow, 1.0 = normal, 1.5 = fast)
        callback_url: Optional webhook URL for completion notification (for async processing)

    Returns:
        Dict with status, video_path, video_url (or video_id if using webhook)

    Popular Avatar IDs:
        -  - Female news anchor
        - "Wayne_20240711" - Male professional
        - "Angela_public_3_20240108" - Professional female
        - "Josh_lite3_20230714" - Casual male

    Popular Voice IDs:
        - "1bd001e7e50f421d891986aad5158bc8" - Professional female (US)
        - "2d5b0e6cf36f4bf5b5cb0b4e5e6e9e3d" - Authoritative male (US)
        - "40421c2ce32f48da9c1e821ac6d1b7f6" - British female
    """
    try:
        api_key = os.getenv("HEYGEN_API_KEY")
        if not api_key:
            return {
                "status": "error",
                "message": "HEYGEN_API_KEY not set in environment"
            }

        print(f"Generating avatar video from text ({len(text)} chars)", file=sys.stderr)

        # Determine background configuration
        bg_config = {}
        if background == "newsroom":
            # Use default newsroom background (professional news studio)
            bg_config = {
                "type": "color",
                "value": DEFAULT_BACKGROUND_COLOR
            }
        elif background_image:
            # Check if it's a local file or URL
            if os.path.isfile(background_image):
                # Upload local file to HeyGen
                print(f"Uploading background image: {background_image}", file=sys.stderr)
                upload_url = "https://upload.heygen.com/v1/asset"

                with open(background_image, "rb") as img_file:
                    image_data = img_file.read()

                # Detect content type from file extension, default to jpeg
                content_type = "image/jpeg"
                if background_image.lower().endswith('.png'):
                    # Check if it's actually a PNG by looking at magic bytes
                    if image_data[:8] == b'\x89PNG\r\n\x1a\n':
                        content_type = "image/png"

                upload_headers = {
                    "X-Api-Key": api_key,
                    "Content-Type": content_type
                }

                upload_response = requests.post(upload_url, headers=upload_headers, data=image_data)
                upload_response.raise_for_status()
                upload_data = upload_response.json()

                # Get the uploaded image URL
                image_url = upload_data.get("data", {}).get("url")

                if not image_url:
                    print(f"⚠️  Warning: Failed to upload background image, using default", file=sys.stderr)
                    bg_config = {
                        "type": "color",
                        "value": DEFAULT_BACKGROUND_COLOR
                    }
                else:
                    print(f"✅ Background image uploaded successfully", file=sys.stderr)
                    bg_config = {
                        "type": "image",
                        "url": image_url
                    }
            else:
                # It's already a URL
                bg_config = {
                    "type": "image",
                    "url": background_image
                }
        elif background.startswith("#"):
            bg_config = {
                "type": "color",
                "value": background
            }
        else:
            bg_config = {
                "type": "color",
                "value": background
            }

        # Create video with text-to-speech
        print("Creating avatar video with TTS...", file=sys.stderr)
        create_url = "https://api.heygen.com/v2/video/generate"

        headers = {
            "X-Api-Key": api_key,
            "Content-Type": "application/json"
        }

        payload = {
            "caption": True,
            "video_inputs": [{
                "character": {
                    "type": "avatar",
                    "avatar_id": avatar_id,
                    "avatar_style": "normal"
                },
                "voice": {
                    "type": "text",
                    "input_text": text,
                    "voice_id": voice_id,
                    "speed": speech_speed
                },
                "background": bg_config
            }],
            "dimension": {
                "width": DEFAULT_VIDEO_WIDTH,
                "height": DEFAULT_VIDEO_HEIGHT
            },
            "aspect_ratio": DEFAULT_ASPECT_RATIO,
            "test": False
        }

        # Add callback URL if provided (for webhook-based completion)
        if callback_url:
            payload["callback_id"] = callback_url
            print(f"Using webhook callback: {callback_url}", file=sys.stderr)

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

        # If using webhook, return immediately without polling
        if callback_url:
            print(f"✅ Video submitted with webhook. Will notify: {callback_url}", file=sys.stderr)
            return {
                "status": "processing",
                "video_id": video_id,
                "message": "Video is processing. Will notify webhook when complete.",
                "callback_url": callback_url
            }

        # Poll for completion
        status_url = f"https://api.heygen.com/v1/video_status.get?video_id={video_id}"
        max_attempts = 240  # 20 minutes max wait (portrait videos can take longer)
        attempt = 0

        while attempt < max_attempts:
            time.sleep(5)
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

                result = {
                    "status": "success",
                    "video_path": output_path,
                    "video_url": video_url,
                    "video_id": video_id,
                    "duration": status_data.get("duration", 0)
                }

                return result

            elif video_status == "failed":
                return {
                    "status": "error",
                    "message": f"Video generation failed: {status_data.get('error', 'Unknown error')}"
                }

        return {
            "status": "error",
            "message": "Video generation timed out"
        }

    except Exception as e:
        print(f"Error: {str(e)}", file=sys.stderr)
        return {
            "status": "error",
            "message": str(e)
        }


def generate_avatar_video(
    audio_path: str,
    output_path: str = "output.mp4",
    avatar_id: str = "Annie_expressive10_public",
    background: str = "newsroom"
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
                "width": DEFAULT_VIDEO_WIDTH,
                "height": DEFAULT_VIDEO_HEIGHT
            },
            "aspect_ratio": DEFAULT_ASPECT_RATIO,
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
        max_attempts = 240  # 20 minutes max wait (portrait videos can take longer)
        attempt = 0

        while attempt < max_attempts:
            time.sleep(30)  # Check every 30 seconds
            attempt += 1

            try:
                # Add timeout to prevent hanging forever
                status_response = requests.get(status_url, headers=headers, timeout=30)
                status_response.raise_for_status()
                status_data = status_response.json().get("data", {})
            except requests.Timeout:
                print(f"⚠️  Status check timed out, retrying... ({attempt}/{max_attempts})", file=sys.stderr)
                continue  # Retry on timeout
            except requests.RequestException as e:
                print(f"⚠️  Network error: {e}, retrying... ({attempt}/{max_attempts})", file=sys.stderr)
                time.sleep(10)  # Wait longer before retry
                continue

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

        # If we've exhausted attempts, video might still be processing
        # Return the video_id so user can check manually
        return {
            "status": "error",
            "message": f"Video generation timed out after {max_attempts * 5 / 60:.1f} minutes. Video may still be processing.",
            "video_id": video_id,
            "check_url": f"https://app.heygen.com/videos/{video_id}"
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
    Supports both text-to-speech and audio input modes.
    """
    # Check for stdin JSON input (for Go integration)
    if not sys.stdin.isatty():
        try:
            input_data = json.loads(sys.stdin.read())

            # Determine if using text-to-speech or audio mode
            text = input_data.get("text", "")
            audio_path = input_data.get("audio_path", "")
            output_path = input_data.get("output_path", "output.mp4")
            avatar_id = input_data.get("avatar_id", "Annie_expressive10_public")
            background = input_data.get("background", "newsroom")
            background_image = input_data.get("background_image")
            voice_id = input_data.get("voice_id", "1bd001e7e50f421d891986aad5158bc8")

            # Text-to-speech mode (preferred)
            if text:
                result = generate_avatar_video_from_text(
                    text=text,
                    output_path=output_path,
                    avatar_id=avatar_id,
                    voice_id=voice_id,
                    background=background,
                    background_image=background_image
                )
            # Audio mode (legacy)
            elif audio_path:
                result = generate_avatar_video(
                    audio_path=audio_path,
                    output_path=output_path,
                    avatar_id=avatar_id,
                    background=background
                )
            else:
                print(json.dumps({
                    "status": "error",
                    "message": "No text or audio_path provided"
                }))
                return

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
        mode = sys.argv[1]

        if mode == "text" and len(sys.argv) >= 4:
            # Text-to-speech mode: python video_generation.py text "your script" output.mp4 [avatar_id] [voice_id]
            text = sys.argv[2]
            output_path = sys.argv[3]
            avatar_id = sys.argv[4] if len(sys.argv) > 4 else "Annie_expressive10_public"
            voice_id = sys.argv[5] if len(sys.argv) > 5 else "1bd001e7e50f421d891986aad5158bc8"

            result = generate_avatar_video_from_text(text, output_path, avatar_id, voice_id)
            print(json.dumps(result))
        else:
            # Audio mode (legacy): python video_generation.py audio.mp3 output.mp4 [avatar_id]
            audio_path = sys.argv[1]
            output_path = sys.argv[2] if len(sys.argv) > 2 else "output.mp4"
            avatar_id = sys.argv[3] if len(sys.argv) > 3 else "Annie_expressive10_public"

            result = generate_avatar_video(audio_path, output_path, avatar_id)
            print(json.dumps(result))

    else:
        print("HeyGen Video Generation - Usage:")
        print("\n📺 TEXT-TO-SPEECH MODE (Recommended):")
        print("  CLI: python video_generation.py text 'your script' output.mp4 [avatar_id] [voice_id]")
        print("  JSON: echo '{\"text\":\"...\",\"output_path\":\"...\",\"avatar_id\":\"...\"}' | python video_generation.py")
        print("\n🎵 AUDIO MODE (Legacy):")
        print("  CLI: python video_generation.py audio.mp3 output.mp4 [avatar_id]")
        print("  JSON: echo '{\"audio_path\":\"...\",\"output_path\":\"...\"}' | python video_generation.py")
        print("\n📚 Resources:")
        print("  Avatars: https://app.heygen.com/avatars")
        print("  Voices: https://app.heygen.com/voice-library")
        print("\n💡 Default Settings:")
        print("  Avatar: Annie_expressive10_public (Professional female news anchor)")
        print("  Voice: 1bd001e7e50f421d891986aad5158bc8 (Professional female US)")
        print("  Background: newsroom (Professional news studio)")


if __name__ == "__main__":
    main()

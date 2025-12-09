#!/usr/bin/env python3
"""
Text-to-Speech (TTS) Generation Module
Uses Google Cloud Text-to-Speech
"""

import os
import sys
import json
from typing import Dict, Any
from google.cloud import texttospeech



def generate_audio(text: str, output_path: str, voice: str = None, speed: float = 1.0) -> Dict[str, Any]:
    """
    Generate audio using Google Cloud Text-to-Speech.

    Args:
        text: Text to convert to speech
        output_path: Path to save the audio file
        voice: Voice name (e.g., "en-US-Neural2-J")
        speed: Speech speed

    Returns:
        Dict with status and audio_path
    """
    try:

        client = texttospeech.TextToSpeechClient()

        synthesis_input = texttospeech.SynthesisInput(text=text)

        # Default to a professional news voice
        if not voice:
            voice = "en-US-Neural2-F"  # Female professional voice

        voice_params = texttospeech.VoiceSelectionParams(
            language_code="en-US",
            name=voice
        )

        audio_config = texttospeech.AudioConfig(
            audio_encoding=texttospeech.AudioEncoding.MP3,
            speaking_rate=speed
        )

        print(f"Generating audio with Google Cloud TTS (voice: {voice})...", file=sys.stderr)

        response = client.synthesize_speech(
            input=synthesis_input,
            voice=voice_params,
            audio_config=audio_config
        )

        with open(output_path, "wb") as f:
            f.write(response.audio_content)

        print(f"✅ Audio saved to {output_path}", file=sys.stderr)

        return {
            "status": "success",
            "audio_path": output_path,
            "provider": "google",
            "voice": voice
        }

    except ImportError:
        return {
            "status": "error",
            "message": "Google Cloud TTS library not installed. Run: poetry add google-cloud-texttospeech"
        }
    except Exception as e:
        return {
            "status": "error",
            "message": str(e)
        }


def main():
    """
    CLI interface for TTS generation using Google Cloud TTS.
    Supports both command-line args and JSON stdin input.
    """
    # Check for stdin JSON input (for Go integration)
    if not sys.stdin.isatty():
        try:
            input_data = json.loads(sys.stdin.read())

            text = input_data.get("text", "")
            output_path = input_data.get("output_path", "output.mp3")
            voice = input_data.get("voice")
            speed = input_data.get("speed", 1.0)

            if not text:
                print(json.dumps({
                    "status": "error",
                    "message": "No text provided"
                }))
                return

            result = generate_audio(text, output_path, voice, speed)
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
        text = sys.argv[1]
        output_path = sys.argv[2] if len(sys.argv) > 2 else "output.mp3"

        result = generate_audio(text, output_path)
        print(json.dumps(result))

    else:
        print("Usage:")
        print("  CLI: python tts_generation.py 'your text' [output.mp3]")
        print("  JSON stdin: echo '{\"text\":\"...\",\"output_path\":\"...\"}' | python tts_generation.py")


if __name__ == "__main__":
    main()


#!/usr/bin/env python3
import sys
import warnings

# Kokoro's own internals call deprecated torch APIs (weight_norm, jit.script);
# harmless with the pinned torch version, so silence the noise here.
warnings.filterwarnings("ignore", category=FutureWarning)
warnings.filterwarnings("ignore", category=UserWarning)

from kokoro import KPipeline
import soundfile as sf
import numpy as np

def main():
    if len(sys.argv) != 3:
        print("usage: generate_audio.py <text> <output_wav_path>", file=sys.stderr)
        sys.exit(1)

    text, output_path = sys.argv[1], sys.argv[2]
    voice = "af_heart"

    pipeline = KPipeline(lang_code="a")
    chunks = [audio for _, _, audio in pipeline(text, voice=voice)]
    audio = np.concatenate(chunks)

    sf.write(output_path, audio, 24000)
    print(output_path)

if __name__ == "__main__":
    main()

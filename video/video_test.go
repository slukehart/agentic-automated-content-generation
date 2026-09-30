package video

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestWavDuration(t *testing.T) {
	const rate, samples = 24000, 48000 // 2s of 16-bit mono
	buf := []byte("RIFF\x00\x00\x00\x00WAVEfmt ")
	buf = binary.LittleEndian.AppendUint32(buf, 16)
	buf = binary.LittleEndian.AppendUint16(buf, 1)
	buf = binary.LittleEndian.AppendUint16(buf, 1)
	buf = binary.LittleEndian.AppendUint32(buf, rate)
	buf = binary.LittleEndian.AppendUint32(buf, rate*2)
	buf = binary.LittleEndian.AppendUint16(buf, 2)
	buf = binary.LittleEndian.AppendUint16(buf, 16)
	buf = append(buf, "data"...)
	buf = binary.LittleEndian.AppendUint32(buf, samples*2)
	buf = append(buf, make([]byte, samples*2)...)

	got, err := wavDuration(buf)
	if err != nil || math.Abs(got-2) > 1e-9 {
		t.Fatalf("got %v, %v; want 2s", got, err)
	}
	if _, err := wavDuration([]byte("nope")); err == nil {
		t.Fatal("expected error for non-WAV input")
	}
}

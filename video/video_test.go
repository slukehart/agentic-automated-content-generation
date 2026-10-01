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

func TestSegmentsFor(t *testing.T) {
	// 1 segment = 3.72s, each extra segment adds 3.2s. 15 segments gave 48.52s
	// in a real job, which cut off 52.2s of narration.
	cases := []struct {
		seconds float64
		want    int
	}{{1, 1}, {3.72, 1}, {3.73, 2}, {48.52, 15}, {48.53, 16}, {52.2, 17}, {60, 19}}
	for _, c := range cases {
		if got := segmentsFor(c.seconds); got != c.want {
			t.Errorf("segmentsFor(%v) = %d, want %d", c.seconds, got, c.want)
		}
	}
}

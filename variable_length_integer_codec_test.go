package variable_length_integer_codec

import (
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	values := []uint64{
		0,
		1,
		127,
		128,
		255,
		256,
		16383,
		16384,
		1<<21 - 1,
		1 << 21,
		1<<28 - 1,
		1 << 28,
		1<<35 - 1,
		1 << 35,
		1<<42 - 1,
		1 << 42,
		1<<49 - 1,
		1 << 49,
		1<<56 - 1,
		1 << 56,
		1<<63 - 1,
		1 << 63,
		1<<64 - 1,
	}
	for _, v := range values {
		enc := Encode(nil, v)
		got, n, err := Decode(enc)
		if err != nil {
			t.Fatalf("Decode(%d) error: %v", v, err)
		}
		if got != v {
			t.Errorf("Decode(Encode(%d)) = %d, want %d", v, got, v)
		}
		if n != len(enc) {
			t.Errorf("Decode(Encode(%d)) consumed %d bytes, want %d", v, n, len(enc))
		}
	}
}

func TestEncodeAppends(t *testing.T) {
	buf := []byte{0xaa}
	buf = Encode(buf, 300)
	if len(buf) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(buf))
	}
	if buf[0] != 0xaa {
		t.Errorf("existing byte was overwritten: got %x", buf[0])
	}
	got, n, err := Decode(buf[1:])
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if got != 300 || n != 2 {
		t.Errorf("Decode = (%d, %d), want (300, 2)", got, n)
	}
}

func TestDecodeEmpty(t *testing.T) {
	_, _, err := Decode(nil)
	if err != ErrTruncated {
		t.Fatalf("expected ErrTruncated, got %v", err)
	}
}

func TestDecodeTruncated(t *testing.T) {
	// 0x80 has the continuation bit set but no following byte.
	_, _, err := Decode([]byte{0x80})
	if err != ErrTruncated {
		t.Fatalf("expected ErrTruncated, got %v", err)
	}
}

func TestDecodeOverflow(t *testing.T) {
	// Ten bytes where the last byte has a bit beyond bit 63 set.
	// The 10th byte contributes bits 63..69; only bit 63 is valid, so
	// any value > 1 in the low 7 bits overflows.
	data := bytes.Repeat([]byte{0x80}, 9)
	data = append(data, 0x02)
	_, _, err := Decode(data)
	if err != ErrOverflow {
		t.Fatalf("expected ErrOverflow, got %v", err)
	}
}

func TestDecodeMaxUint64(t *testing.T) {
	// Max uint64 encodes to 10 bytes: nine 0xff followed by 0x01.
	enc := Encode(nil, ^uint64(0))
	if len(enc) != 10 {
		t.Fatalf("expected 10 bytes, got %d", len(enc))
	}
	got, n, err := Decode(enc)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if got != ^uint64(0) {
		t.Errorf("got %d, want %d", got, ^uint64(0))
	}
	if n != 10 {
		t.Errorf("consumed %d bytes, want 10", n)
	}
}

func TestDecodeStopsAtTerminator(t *testing.T) {
	// Two values back to back; Decode should only consume the first.
	buf := Encode(nil, 1)
	buf = Encode(buf, 2)
	got, n, err := Decode(buf)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if got != 1 || n != 1 {
		t.Errorf("Decode = (%d, %d), want (1, 1)", got, n)
	}
	got, n, err = Decode(buf[n:])
	if err != nil {
		t.Fatalf("second Decode error: %v", err)
	}
	if got != 2 || n != 1 {
		t.Errorf("second Decode = (%d, %d), want (2, 1)", got, n)
	}
}

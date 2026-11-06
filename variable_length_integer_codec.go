// Package variable_length_integer_codec implements a variable-length encoding for
// unsigned 64-bit integers. It is designed to be simple and dependency-free.
//
// The encoding is a prefix-free byte sequence: each byte contributes 7 bits of
// the value, and the high bit indicates whether more bytes follow. The least
// significant 7-bit group is stored first. This makes small values take a
// single byte, while large values use more bytes.
//
// This scheme is not the same as Protocol Buffers varint or LEB128; it uses a
// continuation bit in the high bit of each byte, but stores groups in
// little-endian order (least significant first).
package variable_length_integer_codec

import "errors"

// ErrOverflow is returned when decoding a sequence that represents a value
// larger than 64 bits.
var ErrOverflow = errors.New("variable_length_integer_codec: value overflows uint64")

// ErrTruncated is returned when the input ends before a complete value is
// decoded.
var ErrTruncated = errors.New("variable_length_integer_codec: truncated input")

// Encode appends the variable-length encoding of v to buf and returns the
// extended slice. The encoding uses 7 bits per byte, least significant group
// first, with the high bit set on all but the last byte.
func Encode(buf []byte, v uint64) []byte {
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			buf = append(buf, b|0x80)
		} else {
			buf = append(buf, b)
			return buf
		}
	}
}

// Decode reads a variable-length integer from the beginning of data. It returns
// the decoded value and the number of bytes consumed. If data is empty or ends
// before the terminating byte, ErrTruncated is returned. If the encoded value
// would exceed 64 bits, ErrOverflow is returned.
//
// The decoder accepts at most 10 bytes: ceil(64/7) = 10. The tenth byte may
// only contain the low 1 bit of the value; any higher bits set would overflow.
func Decode(data []byte) (uint64, int, error) {
	var v uint64
	for i := 0; i < len(data); i++ {
		b := data[i]
		if i == 9 {
			// The 10th byte can only contribute the lowest bit; if any of the
			// upper 7 bits are set, the value would need more than 64 bits.
			if b&0x7f > 1 {
				return 0, 0, ErrOverflow
			}
		}
		v |= uint64(b&0x7f) << (7 * uint(i))
		if b&0x80 == 0 {
			return v, i + 1, nil
		}
	}
	return 0, 0, ErrTruncated
}

# variable-length-integer-codec

Encodes unsigned 64-bit integers as variable-length byte sequences and decodes them back.

## Usage

```go
package main

import (
	"fmt"

	codec "variable-length-integer-codec"
)

func main() {
	buf := codec.Encode(nil, 300)
	v, n, err := codec.Decode(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(v, n) // 300 2
}
```

## Why this exists

The standard library has no compact integer encoding. Fixed-width types waste space for small values, and general-purpose compression is overkill for a stream of counters or identifiers. This library provides a single, obvious encoding: 7 bits per byte, least significant group first, with the high bit of each byte marking continuation. Small values take one byte; the maximum `uint64` takes ten.

The trade-off is simplicity over maximal density. More complex schemes (e.g. zigzag for signed values, or prefix tables) can be smaller for specific distributions, but they complicate the API and the edge cases. Here, `Encode` appends to a caller-provided byte slice to avoid allocation, and `Decode` returns the number of bytes consumed so callers can parse a sequence.

## The awkward edge

A ten-byte sequence can encode values that do not fit in `uint64`. The decoder rejects those with `ErrOverflow` rather than silently truncating. Similarly, a sequence that ends while the continuation bit is still set returns `ErrTruncated`. Both are sentinel errors so callers can distinguish them with `errors.Is` if they wrap them later.

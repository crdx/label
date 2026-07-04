package util

import "bytes"

func SetBit(n int, pos uint) int {
	n |= (1 << pos)
	return n
}

func PackBits(input []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	dst := make([]byte, 0, 1024)

	var rle bool
	var repeats int
	const maxRepeats = 127

	finishRaw := func() {
		if buf.Len() == 0 {
			return
		}
		dst = append(dst, byte(buf.Len()-1)) //nolint:gosec // raw runs are flushed at maxRepeats bytes
		dst = append(dst, buf.Bytes()...)
		buf.Reset()
	}

	finishRle := func(b byte, repeats int) {
		dst = append(dst, byte(256-(repeats-1))) //nolint:gosec // repeats is capped at maxRepeats+1, so this fits a byte
		dst = append(dst, b)
	}

	for i, b := range input {
		isLast := i == len(input)-1
		if isLast {
			if !rle {
				buf.WriteByte(b)
				finishRaw()
			} else {
				repeats++
				finishRle(b, repeats)
			}
			break
		}

		if b == input[i+1] {
			if !rle {
				finishRaw()
				rle = true
				repeats = 1
			} else {
				if repeats == maxRepeats {
					finishRle(b, repeats)
					repeats = 0
				}
				repeats++
			}
		} else {
			if !rle {
				if buf.Len() == maxRepeats {
					finishRaw()
				}
				buf.WriteByte(b)
			} else {
				repeats++
				finishRle(b, repeats)
				rle = false
				repeats = 0
			}
		}
	}
	return dst
}

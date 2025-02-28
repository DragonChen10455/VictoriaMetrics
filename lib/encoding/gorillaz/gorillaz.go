package gorillaz

import (
	"errors"
	"math/bits"
)

// Compress uses full predicting-strategy, which means xor right-value is the predictor's value.
func Compress(dst []byte, src []int64) []byte {
	v := uint64(src[0])
	prev := v
	bs := &ByteWrapper{Stream: &dst, Count: 0}
	bs.AppendBits(v, 64) // append first value without any compression
	src = src[1:]
	prevLeadingZeros, prevTrailingZeros := ^uint8(0), uint8(0)
	sigbits := uint8(0)
	for _, num := range src {
		v = uint64(num) ^ prev
		if v == 0 {
			bs.AppendBit(Zero)
		} else {
			bs.AppendBit(One)
			leadingZeros, trailingZeros := uint8(bits.LeadingZeros64(v)), uint8(bits.TrailingZeros64(v))
			// clamp number of leading zeros to avoid overflow when encoding
			if leadingZeros >= 64 {
				leadingZeros = 63
			}
			if prevLeadingZeros != ^uint8(0) && leadingZeros >= prevLeadingZeros && trailingZeros >= prevTrailingZeros {
				bs.AppendBit(Zero)
				bs.AppendBits(v>>prevTrailingZeros, 64-int(prevLeadingZeros)-int(prevTrailingZeros))
			} else {
				prevLeadingZeros, prevTrailingZeros = leadingZeros, trailingZeros
				bs.AppendBit(One)
				bs.AppendBits(uint64(leadingZeros), 6)
				sigbits = 64 - leadingZeros - trailingZeros
				bs.AppendBits(uint64(sigbits), 6)
				bs.AppendBits(v>>trailingZeros, int(sigbits))
			}
		}
		prev = uint64(num)
	}
	bs.Finish()
	return dst
}

// Decompress append data to dst and return the appended dst
func Decompress(dst []int64, src []byte) ([]int64, error) {
	bs := &ByteWrapper{Stream: &src, Count: 8}
	firstValue, err := bs.ReadBits(64)
	if err != nil {
		return nil, err
	}
	dst = append(dst, int64(firstValue))
	prev := firstValue
	prevLeadingZeros, prevTrailingZeros := uint8(0), uint8(0)
	for true {
		b, err := bs.ReadBit()
		if err != nil {
			return nil, err
		}
		if b == Zero {
			dst = append(dst, int64(prev))
			continue
		} else {
			b, err = bs.ReadBit()
			if err != nil {
				return nil, err
			}
			leadingZeros, trailingZeros := prevLeadingZeros, prevTrailingZeros
			if b == One {
				bts, err := bs.ReadBits(6) // read leading zeros' length
				if err != nil {
					return nil, err
				}
				leadingZeros = uint8(bts)
				bts, err = bs.ReadBits(6) // read sig's length
				if err != nil {
					return nil, err
				}
				midLen := uint8(bts)
				if midLen == 0 {
					midLen = 64
				}
				if midLen+leadingZeros > 64 {
					if b, err = bs.ReadBit(); b == Zero {
						return dst, nil
					}
					return nil, errors.New("invalid bits")
				}
				trailingZeros = 64 - leadingZeros - midLen
				prevLeadingZeros, prevTrailingZeros = leadingZeros, trailingZeros
			}
			bts, err := bs.ReadBits(int(64 - leadingZeros - trailingZeros))
			if err != nil {
				return nil, err
			}
			v := prev
			v ^= bts << trailingZeros
			dst = append(dst, int64(v))
			prev = v
		}
	}
	return dst, nil
}

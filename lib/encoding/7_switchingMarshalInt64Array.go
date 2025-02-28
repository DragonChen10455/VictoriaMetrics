package encoding

import (
	"fmt"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/metrics"
)

func marshalInt64Switching(dst []byte, a []int64, _ uint8) (result []byte, firstValue int64) {
	if len(a) < 1 {
		logger.Panicf("BUG: a must contain at least 1 item; got %d items", len(a))
	}

	reCompressCalls.Inc()
	reOriginalBytes.Add(len(a))
	firstValue = a[0]
	prev := firstValue
	count := int64(1)
	totalLen := int64(1)
	is := GetInt64s(len(a) + 1)
	a = a[1:]

	for _, next := range a {
		if next == prev {
			count++
		} else {
			is.A[totalLen] = prev
			is.A[totalLen+1] = count
			totalLen += 2
			prev = next
			count = 1
		}
	}
	is.A[totalLen] = prev
	is.A[totalLen+1] = count
	totalLen += 2
	if totalLen <= 0 {
		logger.Panicf("BUG: a must contain at least 1 item; got %d total length", totalLen)
	}
	is.A[0] = totalLen - 1

	dst = MarshalVarInt64s(dst, is.A[:totalLen])
	PutInt64s(is)
	reCompressedBytes.Add(len(dst))
	return dst, firstValue
}

func unmarshalInt64Switching(dst []int64, a []byte, _ int64, itemsCount int) ([]int64, error) {
	if itemsCount < 1 {
		logger.Panicf("BUG: itemsCount must be greater than 0; got %d", itemsCount)
	}

	src, totalLen, err := UnmarshalVarInt64(a)
	if err != nil {
		return nil, fmt.Errorf("cannot get repeat eliminate compress total length from %d bytes; src=%X: %w", len(src), src, err)
	}
	is := GetInt64s(int(totalLen))
	defer PutInt64s(is)

	tail, err := UnmarshalVarInt64s(is.A, src)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal nearest delta from %d bytes; src=%X: %w", len(src), src, err)
	}
	if len(tail) > 0 {
		return nil, fmt.Errorf("unexpected tail left after unmarshaling %d items from %d bytes; tail size=%d; src=%X; tail=%X", itemsCount, len(src), len(tail), src, tail)
	}
	if len(is.A)%2 != 0 {
		return nil, fmt.Errorf("unexpected length after unmarshaling %d items from %d bytes; want even but odd; src=%X; length=%d", itemsCount, len(src), src, len(is.A))
	}

	reDecompressCalls.Inc()
	for i := 0; i < len(is.A); i += 2 {
		for j := int64(0); j < is.A[i+1]; j++ {
			dst = append(dst, is.A[i])
			itemsCount--
			if itemsCount == 0 {
				return dst, nil
			}
		}
	}

	return nil, fmt.Errorf("data is less than wanted; got %d; still want %d", len(dst), itemsCount)
}

var (
	reCompressCalls   = metrics.NewCounter(`vm_re_block_compress_calls_total`)
	reDecompressCalls = metrics.NewCounter(`vm_re_block_decompress_calls_total`)

	reOriginalBytes   = metrics.NewCounter(`vm_re_block_original_bytes_total`)
	reCompressedBytes = metrics.NewCounter(`vm_re_block_compressed_bytes_total`)
)

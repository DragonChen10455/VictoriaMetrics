package encoding

import (
	"fmt"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

func marshalInt64ZSTD(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	if len(a) == 0 {
		logger.Panicf("BUG: a must contain at least one item")
	}
	bb := bbPool.Get()
	firstValue = a[0]
	bb.B = MarshalVarInt64s(bb.B[:0], a)
	compressLevel := getCompressLevel(len(a))
	dst = CompressZSTDLevel(dst, bb.B, compressLevel)
	bbPool.Put(bb)
	mt = MarshalTypeZSTD
	return dst, mt, a[0]
}

func unmarshalInt64ZSTD(dst []int64, src []byte, _ MarshalType, _ int64, itemsCount int) ([]int64, error) {
	var err error
	bb := bbPool.Get()
	bb.B, err = DecompressZSTD(bb.B[:0], src)
	if err != nil {
		return nil, fmt.Errorf("cannot decompress zstd data: %w", err)
	}
	dst = append(dst, make([]int64, itemsCount)...)
	_, err = UnmarshalVarInt64s(dst, bb.B)
	bbPool.Put(bb)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal data after zstd decompression: %w; src_zstd=%X", err, src)
	}
	return dst, nil
}

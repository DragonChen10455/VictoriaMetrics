package encoding

import (
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/encoding/statistics"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

func marshalInt64s(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	if len(a) == 0 {
		logger.Panicf("BUG: a must contain at least one item")
	}
	mt = GetMarshalType(a)
	switch mt {
	case MarshalTypeConst:
		firstValue = a[0]

	case MarshalTypeDeltaConst:
		firstValue = a[0]
		dst = MarshalVarInt64(dst, a[1]-a[0])

	case MarshalTypeZSTDNearestDelta2:
		bb := bbPool.Get()
		bb.B, firstValue = marshalInt64NearestDelta2(bb.B[:0], a, 64)
		compressLevel := getCompressLevel(len(a))
		dst = CompressZSTDLevel(dst, bb.B, compressLevel)
		bbPool.Put(bb)

	case MarshalTypeZSTDNearestDelta:
		bb := bbPool.Get()
		bb.B, firstValue = marshalInt64NearestDelta(bb.B[:0], a, 64)
		compressLevel := getCompressLevel(len(a))
		dst = CompressZSTDLevel(dst, bb.B, compressLevel)
		bbPool.Put(bb)

	case MarshalTypeZSTD:
		bb := bbPool.Get()
		firstValue = a[0]
		bb.B = MarshalVarInt64s(bb.B[:0], a)
		compressLevel := getCompressLevel(len(a))
		dst = CompressZSTDLevel(dst, bb.B, compressLevel)
		bbPool.Put(bb)

	case MarshalTypeSwitching:
		dst, _, firstValue = marshalInt64Switching(dst, a, 0)

	case MarshalTypeNearestDelta:
		dst, firstValue = marshalInt64NearestDelta(dst, a, 64)

	default:
		logger.Panicf("BUG: unexpected mt=%d", mt)
	}
	return dst, mt, firstValue
}

func GetMarshalType(int64s []int64) MarshalType {

	if len(int64s) <= 1 {
		return MarshalTypeConst
	}
	if len(int64s) <= 2 {
		return MarshalTypeDeltaConst
	}
	distance, isRepeat := statistics.ComplexHammingDistance(int64s)
	if distance == 0 {
		return MarshalTypeConst
	}
	if isRepeat {
		return MarshalTypeSwitching
	}
	if distance < 25 {
		return MarshalTypeZSTDNearestDelta
	}

	return MarshalTypeZSTD
}

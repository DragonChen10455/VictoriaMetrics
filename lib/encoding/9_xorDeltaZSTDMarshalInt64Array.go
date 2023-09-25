package encoding

import "github.com/valyala/gozstd"

func marshalInt64XorDeltaZSTD(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	dst, _, firstValue = marshalInt64XorDelta(dst, a, 64)
	dst = gozstd.CompressLevel(nil, dst, 5)
	mt = MarshalTypeDeltaXorZSTD
	return dst, mt, firstValue
}

func unmarshalInt64XorDeltaZSTD(dst []int64, src []byte, _ MarshalType, _ int64, itemsCount int) ([]int64, error) {
	return nil, nil
}

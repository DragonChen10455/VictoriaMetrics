package encoding

import "github.com/golang/snappy"

func marshalInt64DeltaSnappy(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	bb := bbPool.Get()
	bb.B, _ = marshalInt64NearestDelta(bb.B[:0], a, 64)
	dst = snappy.Encode(nil, bb.B)
	bbPool.Put(bb)
	mt = MarshalTypeDeltaSnappy
	return dst, mt, a[0]
}

func unmarshalInt64DeltaSnappy(dst []int64, src []byte, _ MarshalType, _ int64, itemsCount int) ([]int64, error) {
	return nil, nil
}

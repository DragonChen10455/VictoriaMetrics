package encoding

import "github.com/bkaradzic/go-lz4"

func marshalInt64Delta2LZ4(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	var err error
	bb := bbPool.Get()
	bb.B, _ = marshalInt64NearestDelta2(bb.B[:0], a, 64)
	dst, err = lz4.Encode(nil, bb.B)
	if err != nil {
		return
	}
	bbPool.Put(bb)
	mt = MarshalTypeDelta2LZ4
	return dst, mt, a[0]
}

func unmarshalInt64Delta2LZ4(dst []int64, src []byte, _ MarshalType, _ int64, itemsCount int) ([]int64, error) {
	return nil, nil
}

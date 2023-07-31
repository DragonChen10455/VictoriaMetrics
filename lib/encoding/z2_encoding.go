package encoding

const PointMaxMask = 0x7fffffff

// Z2
type Point struct {
	x float32
	y float32
}

type PointIndex struct {
	i int64
}

func (index PointIndex) decode() Point {
	indexInt := index.i
	Point := Point{
		x: combineZ2Point(indexInt),
		y: combineZ2Point(indexInt >> 1),
	}
	return Point
}

func (p Point) encode() PointIndex {
	indexInt := splitZ2Point(p.x) | splitZ2Point(p.y)<<1
	index := PointIndex{
		i: indexInt,
	}
	return index
}

/** insert 0 between every bit in value. Only first 31 bits can be considered. */
func splitZ2Point(value float32) int64 {
	x := float32ToInt64(value) & PointMaxMask
	x = (x ^ (x << 32)) & 0x00000000ffffffff
	x = (x ^ (x << 16)) & 0x0000ffff0000ffff
	x = (x ^ (x << 8)) & 0x00ff00ff00ff00ff // 11111111000000001111111100000000..
	x = (x ^ (x << 4)) & 0x0f0f0f0f0f0f0f0f // 1111000011110000
	x = (x ^ (x << 2)) & 0x3333333333333333 // 11001100..
	x = (x ^ (x << 1)) & 0x5555555555555555 // 1010...
	return x
}

/** combineZ2Point every other bit to form a value. Maximum value is 31 bits. */
func combineZ2Point(i int64) float32 {
	var x = i & 0x5555555555555555
	x = (x ^ (x >> 1)) & 0x3333333333333333
	x = (x ^ (x >> 2)) & 0x0f0f0f0f0f0f0f0f
	x = (x ^ (x >> 4)) & 0x00ff00ff00ff00ff
	x = (x ^ (x >> 8)) & 0x0000ffff0000ffff
	x = (x ^ (x >> 16)) & 0x00000000ffffffff
	x = (x ^ (x >> 32)) & PointMaxMask
	return int64ToFloat32(x)
}

func ZOrderEncode(x, y float32, bitsNum uint) int64 {
	xInt64 := float32ToInt64(x)
	yInt64 := float32ToInt64(y)
	var z int64
	for i := uint(0); i < bitsNum; i++ {
		z |= ((xInt64 & (1 << i)) << i) | ((yInt64 & (1 << i)) << (i + 1))
	}
	return z
}

func ZOrderDecode(z int64, bitsNum uint) (float32, float32) {
	var x, y int64 = 0, 0
	for i := uint(0); i < bitsNum; i++ {
		x |= (z & (1 << (i * 2))) >> i
		y |= (z & (1 << (i*2 + 1))) >> (i + 1)
	}
	return int64ToFloat32(x), int64ToFloat32(y)
}

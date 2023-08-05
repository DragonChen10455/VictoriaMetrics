package encoding

import (
	"math"
)

func float32ToInt64(f float32) int64 {
	// 判断是否超过int64范围，如果是，则返回最大或最小值
	if f > float32(math.MaxInt64) {
		return math.MaxInt64
	} else if f < float32(math.MinInt64) {
		return math.MinInt64
	}
	// 获取IEEE 754编码
	bits := math.Float32bits(f)
	//fmt.Printf("将%v编码成%b\n", f, bits)

	// 将bits转换为int64
	return int64(bits)
}

func float64ToInt64(f float64) int64 {
	// 判断是否超过int64范围，如果是，则返回最大或最小值
	if f > float64(math.MaxInt64) {
		return math.MaxInt64
	} else if f < float64(math.MinInt64) {
		return math.MinInt64
	}
	// 获取IEEE 754编码
	bits := math.Float64bits(f)
	//fmt.Printf("将%v编码成%b\n", f, bits)

	// 将bits转换为int64
	return int64(bits)
}

func int64ToFloat32(i int64) float32 {
	// 获取int32的IEEE 754编码
	bits := uint32(i)

	// 将bits转换为float32
	return math.Float32frombits(bits)
}

func int64ToFloat64(i int64) float64 {
	// 获取int64的IEEE 754编码
	bits := uint64(i)

	// 将bits转换为float64
	return math.Float64frombits(bits)
}

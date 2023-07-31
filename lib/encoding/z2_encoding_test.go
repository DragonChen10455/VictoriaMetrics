package encoding

import (
	"fmt"
	"testing"
)

func TestOneEncoding(t *testing.T) {
	//r := rand.New(rand.NewSource(2))
	Z2PointRaw := Point{
		x: 123.25,
		y: 88.97,
	}
	fmt.Printf("原始数据:%v\n", Z2PointRaw)
	// Z3Point编码构建索引
	Z2PointIndex := Z2PointRaw.encode()
	fmt.Printf("构建的索引数据:%b\n", Z2PointIndex)

	// Z3Point解码后数据
	Z2PointDecoded := Z2PointIndex.decode()
	fmt.Printf("解码后的数据:%v\n", Z2PointDecoded)

	if Z2PointRaw != Z2PointDecoded {
		t.Fatalf("encoding error！")
	}
}

func TestDouglasPeucker(t *testing.T) {
	var points []Point
	var result []Point
	points = append(points, Point{x: 1, y: 1})
	points = append(points, Point{x: 2, y: 2})
	points = append(points, Point{x: 3, y: 4})
	points = append(points, Point{x: 4, y: 1})
	points = append(points, Point{x: 5, y: 0})
	points = append(points, Point{x: 6, y: 3})
	points = append(points, Point{x: 7, y: 5})
	points = append(points, Point{x: 8, y: 2})
	points = append(points, Point{x: 9, y: 1})
	points = append(points, Point{x: 10, y: 6})
	fmt.Print("\n压缩前：\n")
	for i := 0; i < len(points); i++ {
		fmt.Printf("(%v, %v) ", points[i].x, points[i].y)
	}
	fmt.Print("\n压缩后：\n")
	result = DouglasPeucker(points, 1)
	for i := 0; i < len(result); i++ {
		fmt.Printf("(%v, %v) ", result[i].x, result[i].y)
	}
}

func TestEncodingWithBitsNum(t *testing.T) {
	var x float32 = 123.25
	var y float32 = 88.97
	bits := uint(31) // 表示x和y坐标所需的位数（根据需要进行调整）

	z := ZOrderEncode(x, y, bits)
	fmt.Printf("点 (%v, %v) 编码为 Z-order: %b\n", x, y, z)
	_x, _y := ZOrderDecode(z, bits)
	fmt.Printf("解码为 Z-order: 点 (%v, %v)\n", _x, _y)
}

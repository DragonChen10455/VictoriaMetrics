package encoding

import (
	"fmt"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"github.com/valyala/fastjson"
	"math/rand"
	"strconv"
	"testing"
	"time"
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
	var x float32 = 1
	var y float32 = 1
	bits := uint8(31) // 表示x和y坐标所需的位数（根据需要进行调整）

	z := ZOrderEncode(x, y, bits)
	fmt.Printf("点 (%v, %v) 编码为 Z-order: %v\n", x, y, z)
	_x, _y := ZOrderDecode(z, bits)
	fmt.Printf("解码为 Z-order: 点 (%v, %v)\n", _x, _y)
}

func TestKMeans(t *testing.T) {
	numPoints := 10000
	// 创建包含点坐标的示例数据，使用两个 []*fastjson.Value 数组表示二维数据
	pointsX := make([]*fastjson.Value, numPoints)
	pointsY := make([]*fastjson.Value, numPoints)

	for i := 0; i < numPoints; i++ {
		x := rand.Float64() * 10000.0 // 生成0到100之间的随机数，根据您的需求进行调整
		y := rand.Float64() * 10000.0 // 生成0到100之间的随机数，根据您的需求进行调整
		xStr := strconv.FormatFloat(x, 'f', -1, 64)
		yStr := strconv.FormatFloat(y, 'f', -1, 64)
		pointsX[i] = fastjson.MustParse(xStr)
		pointsY[i] = fastjson.MustParse(yStr)
	}

	startT := time.Now()

	var d clusters.Observations
	for i := range pointsX {
		x, _ := pointsX[i].Float64()
		y, _ := pointsY[i].Float64()
		d = append(d, clusters.Coordinates{
			x,
			y,
		})
	}
	km := kmeans.New()
	clustersRes, err := km.Partition(d, numPoints/100)
	if err != nil {
		fmt.Printf("km.Partition error!")
	}

	costT := time.Since(startT)

	fmt.Printf("Time costs: %dms\n", costT/1000/1000)
	fmt.Printf("Numbers of clusters: %d\n", len(clustersRes))

	for _, c := range clustersRes {
		fmt.Printf("Centered at x: %.2f y: %.2f\n", c.Center[0], c.Center[1])
		fmt.Printf("Matching data points: %+v\n\n", c.Observations)
	}
}

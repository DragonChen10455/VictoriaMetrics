package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"testing"
)

const MaxMask = 0x1fffff

// Z3
type Z3 struct {
	x int64
	y int64
	z int64
}

func (index Z3Index) decode() Z3 {
	indexInt := index.i
	Z3Point := Z3{
		x: combineZ3Point(indexInt),
		y: combineZ3Point(indexInt >> 1),
		z: combineZ3Point(indexInt >> 2),
	}
	return Z3Point
}

// Z3Index
type Z3Index struct {
	i int64
}

/**
 * So this represents the order of the tuple, but the bits will be encoded in reverse order:
 *   ....z1y1x1z0y0x0
 * This is a little confusing.
 */
func (p Z3) encode() Z3Index {
	indexInt := splitZ3Point(p.x) | splitZ3Point(p.y)<<1 | splitZ3Point(p.z)<<2
	index := Z3Index{
		i: indexInt,
	}
	return index
}

/** insert 00 between every bit in value. Only first 31 bits can be considered. */
func splitZ3Point(value int64) int64 {
	x := value & MaxMask
	x = (x | x<<32) & 0x1f00000000ffff
	x = (x | x<<16) & 0x1f0000ff0000ff
	x = (x | x<<8) & 0x100f00f00f00f00f //0x1f00f00f00f00f
	x = (x | x<<4) & 0x10c30c30c30c30c3 //0x1f3cf3cf3cf3cf
	x = (x | x<<2) & 0x1249249249249249 //0x19249249249249
	return x
}

/** combineZ3Point every third bit to form a value. Maximum value is 31 bits. */
func combineZ3Point(i int64) int64 {
	var x = i & 0x1249249249249249          //0x19249249249249
	x = (x ^ (x >> 2)) & 0x10c30c30c30c30c3 //0x1f3cf3cf3cf3cf
	x = (x ^ (x >> 4)) & 0x100f00f00f00f00f //0x1f00f00f00f00f
	x = (x ^ (x >> 8)) & 0x1f0000ff0000ff
	x = (x ^ (x >> 16)) & 0x1f00000000ffff
	x = (x ^ (x >> 32)) & MaxMask
	return x
}

// 整型转换成字节
func IntToBytes(n int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, n)
	if err != nil {
		return nil
	}
	return bytesBuffer.Bytes()
}

func calInt64ArrayByteLength(data []int64) int64 {
	res := 0
	for i := 0; i < len(data); i++ {
		res += len(IntToBytes(data[i]))
	}
	return int64(res)
}

func calZ3ArrayByteLength(data []Z3) int64 {
	res := 0
	for i := 0; i < len(data); i++ {
		res += len(IntToBytes(data[i].x))
		res += len(IntToBytes(data[i].y))
		res += len(IntToBytes(data[i].z))
	}
	return int64(res)
}

func TestRandom(t *testing.T) {
	r := rand.New(rand.NewSource(2))

	// 原始数据
	var Z3DataUnEncodedArray []Z3
	var Z3DataUnEncodedX []int64
	var Z3DataUnEncodedY []int64
	var Z3DataUnEncodedZ []int64

	// 构建索引后的数据
	var Z3DataInt64Array []int64

	const precisionBits = 64

	// 构造n个数据点，每个点都是Z3Point(int64, int64, int64)
	for j := 0; j < 10000; j++ {
		// (1)随机数据
		Z3Point := Z3{
			x: int64(r.Int31n(2097152)),
			y: int64(r.Int31n(2097152)),
			z: int64(r.Int31n(2097152)),
		}
		Z3DataUnEncodedArray = append(Z3DataUnEncodedArray, Z3Point)
		Z3DataUnEncodedX = append(Z3DataUnEncodedX, Z3Point.x)
		Z3DataUnEncodedY = append(Z3DataUnEncodedY, Z3Point.y)
		Z3DataUnEncodedZ = append(Z3DataUnEncodedZ, Z3Point.z)

		// Z3Point编码构建索引
		Z3PointIndex := Z3Point.encode()
		Z3DataInt64Array = append(Z3DataInt64Array, Z3PointIndex.i)

		// Z3Point解码后数据
		Z3PointDecoded := Z3PointIndex.decode()

		if Z3Point != Z3PointDecoded {
			t.Fatalf("encoding error！")
		}
	}

	Z3DataMarshal, mt, firstValue := MarshalValues(nil, Z3DataInt64Array, precisionBits)
	Z3DataUnmarshal, err := UnmarshalValues(nil, Z3DataMarshal, mt, firstValue, len(Z3DataInt64Array))
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}

	if err := checkPrecisionBits(Z3DataInt64Array, Z3DataUnmarshal, precisionBits); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}

	Z3DataCompressed := CompressZSTDLevel(nil, Z3DataMarshal, 5)
	Z3DataDecompressed, err := DecompressZSTD(nil, Z3DataCompressed)
	if err != nil {
		t.Fatalf("unexpected error when decompressing b=%x from bc=%x: %s", Z3DataMarshal, Z3DataCompressed, err)
	}
	if string(Z3DataDecompressed) != string(Z3DataMarshal) {
		t.Fatalf("invalid bNew; got\n%x; expecting\n%x", Z3DataDecompressed, Z3DataMarshal)
	}

	Z3DataXMarshal, mtX, _ := MarshalValues(nil, Z3DataUnEncodedX, precisionBits)
	Z3DataYMarshal, mtY, _ := MarshalValues(nil, Z3DataUnEncodedY, precisionBits)
	Z3DataZMarshal, mtZ, _ := MarshalValues(nil, Z3DataUnEncodedZ, precisionBits)
	Z3DataCompressedX := CompressZSTDLevel(nil, Z3DataXMarshal, 5)
	Z3DataCompressedY := CompressZSTDLevel(nil, Z3DataYMarshal, 5)
	Z3DataCompressedZ := CompressZSTDLevel(nil, Z3DataZMarshal, 5)

	fmt.Printf("随机数据\n")
	fmt.Printf("使用Z-order\n")
	fmt.Printf("原始数据大小：%v\n", calZ3ArrayByteLength(Z3DataUnEncodedArray))
	//fmt.Printf("%b\n", Z3DataUnEncodedArray)
	fmt.Printf("构建索引大小：%v\n", calInt64ArrayByteLength(Z3DataInt64Array))
	//fmt.Printf("%b\n", Z3DataInt64Array)
	fmt.Printf("marshal类型：%v\n", mt)
	//fmt.Printf("%b\n", Z3DataMarshal)
	fmt.Printf("压缩数据大小：%v\n", len(Z3DataCompressed))
	//fmt.Printf("%b\n", Z3DataCompressed)
	//fmt.Printf("压缩率：%v\n", float64(calZ3ArrayByteLength(Z3DataUnEncodedArray)) /
	//	float64(len(Z3DataCompressed)))

	fmt.Printf("\n不使用Z-order\n")
	fmt.Printf("原始数据大小 %v\n", calInt64ArrayByteLength(Z3DataUnEncodedX)+
		calInt64ArrayByteLength(Z3DataUnEncodedY)+calInt64ArrayByteLength(Z3DataUnEncodedZ))
	//fmt.Printf("%b\n %b\n %b\n", Z3DataUnEncodedX, Z3DataUnEncodedY, Z3DataUnEncodedZ)
	fmt.Printf("marshal类型 x:%v y:%v z:%v\n", mtX, mtY, mtZ)
	fmt.Printf("压缩数据大小 %v\n", len(Z3DataCompressedX)+len(Z3DataCompressedY)+len(Z3DataCompressedZ))
	//fmt.Printf("压缩率：%v\n", float64(calInt64ArrayByteLength(Z3DataUnEncodedX)+
	//	calInt64ArrayByteLength(Z3DataUnEncodedY)+calInt64ArrayByteLength(Z3DataUnEncodedZ)) /
	//	float64(len(Z3DataCompressedX)+len(Z3DataCompressedY)+len(Z3DataCompressedZ)))
}

func TestIncreasing(t *testing.T) {
	r := rand.New(rand.NewSource(2))

	// 原始数据
	var Z3DataUnEncodedArray []Z3
	var Z3DataUnEncodedX []int64
	var Z3DataUnEncodedY []int64
	var Z3DataUnEncodedZ []int64

	// 构建索引后的数据
	var Z3DataInt64Array []int64

	const precisionBits = 64

	// 构造n个数据点，每个点都是Z3Point(int64, int64, int64)
	vx := int64(0)
	vy := int64(0)
	vz := int64(0)
	for j := 0; j < 10000; j++ {
		// (2)递增数据
		Z3Point := Z3{
			x: vx + int64(math.Abs(r.NormFloat64())*1e2),
			y: vy + int64(math.Abs(r.NormFloat64())*1e2),
			z: vz + int64(math.Abs(r.NormFloat64())*1e2),
		}
		Z3DataUnEncodedArray = append(Z3DataUnEncodedArray, Z3Point)
		Z3DataUnEncodedX = append(Z3DataUnEncodedX, Z3Point.x)
		Z3DataUnEncodedY = append(Z3DataUnEncodedY, Z3Point.y)
		Z3DataUnEncodedZ = append(Z3DataUnEncodedZ, Z3Point.z)

		// Z3Point编码构建索引
		Z3PointIndex := Z3Point.encode()
		Z3DataInt64Array = append(Z3DataInt64Array, Z3PointIndex.i)

		// Z3Point解码后数据
		Z3PointDecoded := Z3PointIndex.decode()

		if Z3Point != Z3PointDecoded {
			t.Fatalf("encoding error！")
		}

		vx = Z3Point.x
		vy = Z3Point.y
		vz = Z3Point.z
	}

	Z3DataMarshal, mt, firstValue := MarshalValues(nil, Z3DataInt64Array, precisionBits)
	Z3DataUnmarshal, err := UnmarshalValues(nil, Z3DataMarshal, mt, firstValue, len(Z3DataInt64Array))
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}

	if err := checkPrecisionBits(Z3DataInt64Array, Z3DataUnmarshal, precisionBits); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}

	Z3DataCompressed := CompressZSTDLevel(nil, Z3DataMarshal, 5)
	Z3DataDecompressed, err := DecompressZSTD(nil, Z3DataCompressed)
	if err != nil {
		t.Fatalf("unexpected error when decompressing Z3DataMarshal=%x from Z3DataCompressed=%x: %s",
			Z3DataMarshal, Z3DataCompressed, err)
	}
	if string(Z3DataDecompressed) != string(Z3DataMarshal) {
		t.Fatalf("invalid decompress; got\n%x; expecting\n%x", Z3DataDecompressed, Z3DataMarshal)
	}

	Z3DataXMarshal, mtX, _ := MarshalValues(nil, Z3DataUnEncodedX, precisionBits)
	Z3DataYMarshal, mtY, _ := MarshalValues(nil, Z3DataUnEncodedY, precisionBits)
	Z3DataZMarshal, mtZ, _ := MarshalValues(nil, Z3DataUnEncodedZ, precisionBits)
	Z3DataCompressedX := CompressZSTDLevel(nil, Z3DataXMarshal, 5)
	Z3DataCompressedY := CompressZSTDLevel(nil, Z3DataYMarshal, 5)
	Z3DataCompressedZ := CompressZSTDLevel(nil, Z3DataZMarshal, 5)

	fmt.Printf("递增数据\n")
	fmt.Printf("使用Z-order\n")
	fmt.Printf("原始数据大小：%v\n", calZ3ArrayByteLength(Z3DataUnEncodedArray))
	//fmt.Printf("%b\n", Z3DataUnEncodedArray)
	fmt.Printf("构建索引大小：%v\n", calInt64ArrayByteLength(Z3DataInt64Array))
	//fmt.Printf("%b\n", Z3DataInt64Array)
	fmt.Printf("marshal类型：%v\n", mt)
	//fmt.Printf("%b\n", Z3DataMarshal)
	fmt.Printf("压缩数据大小：%v\n", len(Z3DataCompressed))
	//fmt.Printf("%b\n", Z3DataCompressed)
	//fmt.Printf("压缩率：%v\n", float64(calZ3ArrayByteLength(Z3DataUnEncodedArray)) /
	//	float64(len(Z3DataCompressed)))

	fmt.Printf("\n不使用Z-order\n")
	fmt.Printf("原始数据大小 %v\n", calInt64ArrayByteLength(Z3DataUnEncodedX)+
		calInt64ArrayByteLength(Z3DataUnEncodedY)+calInt64ArrayByteLength(Z3DataUnEncodedZ))
	//fmt.Printf("%b\n %b\n %b\n", Z3DataUnEncodedX, Z3DataUnEncodedY, Z3DataUnEncodedZ)
	fmt.Printf("marshal类型 x:%v y:%v z:%v\n", mtX, mtY, mtZ)
	fmt.Printf("压缩数据大小 %v\n", len(Z3DataCompressedX)+len(Z3DataCompressedY)+len(Z3DataCompressedZ))
	//fmt.Printf("压缩率：%v\n", float64(calInt64ArrayByteLength(Z3DataUnEncodedX)+
	//	calInt64ArrayByteLength(Z3DataUnEncodedY)+calInt64ArrayByteLength(Z3DataUnEncodedZ)) /
	//	float64(len(Z3DataCompressedX)+len(Z3DataCompressedY)+len(Z3DataCompressedZ)))
}

//func Test2(t *testing.T) {
//	x := 12345
//	x = x & MaxMask
//	fmt.Printf("x: %b\n", x)
//	fmt.Printf("x | x << 32: %b\n", x | x << 32)
//	fmt.Printf("0x1f00000000ffff: %b\n", 0x1f00000000ffff)
//	x = (x | x << 32) & 0x1f00000000ffff
//	fmt.Printf("(x | x << 32) & 0x1f00000000ffff: %b\n", x)
//
//	fmt.Printf("\n")
//
//	fmt.Printf("x: %b\n", x)
//	fmt.Printf("x | x << 16: %b\n", x | x << 16)
//	fmt.Printf("0x1f0000ff0000ff: %b\n", 0x1f0000ff0000ff)
//	x = (x | x << 16) & 0x1f0000ff0000ff
//	fmt.Printf("(x | x << 16) & 0x1f0000ff0000ff: %b\n", x)
//
//	fmt.Printf("\n")
//
//	fmt.Printf("x: %b\n", x)
//	fmt.Printf("x | x << 8: %b\n", x | x << 8)
//	fmt.Printf("0x100f00f00f00f00f: %b\n", 0x100f00f00f00f00f)
//	x = (x | x << 8) & 0x100f00f00f00f00f
//	fmt.Printf("(x | x << 8) & 0x100f00f00f00f00f: %b\n", x)
//
//	fmt.Printf("\n")
//
//	fmt.Printf("x: %b\n", x)
//	fmt.Printf("x | x << 4: %b\n", x | x << 4)
//	fmt.Printf("0x10c30c30c30c30c3: %b\n", 0x10c30c30c30c30c3)
//	x = (x | x << 4) & 0x10c30c30c30c30c3
//	fmt.Printf("(x | x << 4) & 0x10c30c30c30c30c3: %b\n", x)
//
//	fmt.Printf("\n")
//
//	fmt.Printf("x: %b\n", x)
//	fmt.Printf("x | x << 2: %b\n", x | x << 2)
//	fmt.Printf("0x1249249249249249: %b\n", 0x1249249249249249)
//	x = (x | x << 2) & 0x1249249249249249
//	fmt.Printf("(x | x << 2) & 0x1249249249249249: %b\n", x)
//
//	fmt.Printf("\n")
//
//	x = 12345
//	fmt.Printf("%b\n", splitZ3Point(int64(x)))
//}

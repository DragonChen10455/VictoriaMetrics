package encoding_byte

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	indexInt := BytesToInt(index.i)
	Z3Point := Z3{
		x: combine(indexInt),
		y: combine(indexInt >> 1),
		z: combine(indexInt >> 2),
	}
	return Z3Point
}

// Z3Index
type Z3Index struct {
	i []byte
}

/**
 * So this represents the order of the tuple, but the bits will be encoded in reverse order:
 *   ....z1y1x1z0y0x0
 * This is a little confusing.
 */
func (p Z3) encode() Z3Index {
	indexInt := split(p.x) | split(p.y)<<1 | split(p.z)<<2
	index := Z3Index{
		i: IntToBytes(indexInt),
	}
	return index
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

// 字节转换成整型
func BytesToInt(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int64
	err := binary.Read(bytesBuffer, binary.BigEndian, &x)
	if err != nil {
		return 0
	}
	return x
}

/** insert 00 between every bit in value. Only first 21 bits can be considered. */
func split(value int64) int64 {
	x := value & MaxMask
	x = (x | x<<32) & 0x1f00000000ffff
	x = (x | x<<16) & 0x1f0000ff0000ff
	x = (x | x<<8) & 0x100f00f00f00f00f
	x = (x | x<<4) & 0x10c30c30c30c30c3
	x = (x | x<<2) & 0x1249249249249249
	return x
}

/** combine every third bit to form a value. Maximum value is 21 bits. */
func combine(i int64) int64 {
	var x = i & 0x1249249249249249
	x = (x ^ (x >> 2)) & 0x10c30c30c30c30c3
	x = (x ^ (x >> 4)) & 0x100f00f00f00f00f
	x = (x ^ (x >> 8)) & 0x1f0000ff0000ff
	x = (x ^ (x >> 16)) & 0x1f00000000ffff
	x = (x ^ (x >> 32)) & MaxMask
	return x
}

func Test(t *testing.T) {
	r := rand.New(rand.NewSource(2))

	// 原始数据
	var Z3DataUnEncodedArray []Z3

	// 解码后的数据
	var Z3DataDecodedArray []Z3

	// 构造n个数据点，每个点都是Z3Point(int64, int64, int64)
	for j := 0; j < 100; j++ {
		Z3Point := Z3{
			x: int64(r.Int31n(2097152)),
			y: int64(r.Int31n(2097152)),
			z: int64(r.Int31n(2097152)),
		}
		Z3DataUnEncodedArray = append(Z3DataUnEncodedArray, Z3Point)

		// Z3Point编码构建索引
		Z3PointIndex := Z3DataUnEncodedArray[j].encode()
		fmt.Printf("let point: %v\nencode as: %08b \n", Z3DataUnEncodedArray[j], Z3PointIndex.i)

		// Z3Point解码后数据
		Z3PointDecoded := Z3PointIndex.decode()
		Z3DataDecodedArray = append(Z3DataDecodedArray, Z3PointDecoded)
		fmt.Printf("decode as: %v\n\n", Z3PointDecoded)

		if Z3DataUnEncodedArray[j] != Z3PointDecoded {
			t.Fatalf("encoding error！")
		}
	}
}

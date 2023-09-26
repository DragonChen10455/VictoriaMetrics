package encoding

import (
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/decimal"
	"math/rand"
	"testing"
	"time"
)

// chl test
func TestCompressDecompressValues(t *testing.T) {
	var values1, values2, values3 []int64
	var tmpValues2, tmpValues3 []float64
	v1 := int64(0)
	v2 := int64(0)
	for i := 0; i < 8*1024; i++ {
		v1 += int64(rand.NormFloat64() * 1e2)
		v2 += int64(rand.NormFloat64() * 1e2)
		values1 = append(values1, ZOrderEncode32(float32(v1), float32(v2), 32))
	}

	u1 := float64(0)
	u2 := float64(0)
	for i := 0; i < 8*1024; i++ {
		u1 += rand.NormFloat64() * 1e2
		u2 += rand.NormFloat64() * 1e2
		tmpValues2 = append(tmpValues2, float64(ZOrderEncode32(float32(u1), float32(u2), 32)))
	}
	values2, _ = decimal.AppendFloatToInt64(values2, tmpValues2)

	for i := 0; i < 8*1024; i++ {
		w1 := rand.NormFloat64() * 1e2
		w2 := rand.NormFloat64() * 1e2
		tmpValues3 = append(tmpValues3, float64(ZOrderEncode32(float32(w1), float32(w2), 32)))
	}
	values3, _ = decimal.AppendFloatToInt64(values3, tmpValues3)

	testCompressDecompressValues(t, values1)
	println()
	testCompressDecompressValues(t, values2)
	println()
	testCompressDecompressValues(t, values3)
}

func testCompressDecompressValues(t *testing.T, values []int64) {
	// HaoYuan-Self-Adaptive
	start := time.Now()
	result, mt, firstValue := MarshalInt64sSelfAdaptive(nil, values, 64)
	values2, err := unmarshalInt64sSelfAdaptive(nil, result, mt, firstValue, len(values))
	t.Logf("SA(%d):%d %d %.8f\n", mt, time.Now().UnixNano()-start.UnixNano(), len(result), float64(len(result))/float64(8*len(values)))

	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if len(values) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(values); i++ {
		if values[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}
	// New
	start = time.Now()
	result, mt, firstValue = marshalInt64DeltaXorZSTD(nil, values, 64)
	values2, err = unmarshalInt64Array(nil, result, mt, firstValue, len(values))
	t.Logf("MY_NEW(%d):%d %d %.8f\n", mt, time.Now().UnixNano()-start.UnixNano(), len(result), float64(len(result))/float64(8*len(values)))
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	// Old
	start = time.Now()
	result, mt, firstValue = marshalInt64Array(nil, values, 64)
	values2, err = unmarshalInt64Array(nil, result, mt, firstValue, len(values))
	t.Logf("OLD(%d):%d %d %.8f\n", mt, time.Now().UnixNano()-start.UnixNano(), len(result), float64(len(result))/float64(8*len(values)))
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if err := checkPrecisionBits(values, values2, 64); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}
}

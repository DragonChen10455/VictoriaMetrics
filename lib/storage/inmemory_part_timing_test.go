package storage

import (
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/encoding"
	"math/rand"
	"testing"
)

func BenchmarkInmemoryPartInitFromRowsWorstCase(b *testing.B) {
	benchmarkInmemoryPartInitFromRows(b, benchRawRowsWorstCase)
}

func BenchmarkInmemoryPartInitFromRowsBestCase(b *testing.B) {
	benchmarkInmemoryPartInitFromRows(b, benchRawRowsBestCase)
}

func benchmarkInmemoryPartInitFromRows(b *testing.B, rows []rawRow) {
	b.ReportAllocs()
	b.SetBytes(int64(len(rows)))
	b.RunParallel(func(pb *testing.PB) {
		var mp inmemoryPart
		for pb.Next() {
			mp.InitFromRows(rows)
		}
	})
}

// Each row belongs to an unique TSID
var benchRawRowsWorstCase = func() []rawRow {
	rng := rand.New(rand.NewSource(1))
	var rows []rawRow
	var r rawRow
	for i := 0; i < 1e5; i++ {
		r.TSID.MetricID = uint64(i)
		r.Timestamp = rng.Int63()
		//r.Value = rng.NormFloat64()
		x := rng.Float32()
		r.Value = float64(x)
		r.PrecisionBits = uint8(i%64) + 1
		rows = append(rows, r)
		y := rng.Float32()
		r.Value = float64(y)
		rows = append(rows, r)
	}
	return rows
}()

// All the rows belong to a single TSID, values are zeros, timestamps
// are delimited by const delta.
var benchRawRowsBestCase = func() []rawRow {
	var rows []rawRow
	var r rawRow
	r.PrecisionBits = defaultPrecisionBits
	for i := 0; i < 1e5; i++ {
		r.Timestamp += 30e3
		rows = append(rows, r)
	}
	return rows
}()

// Each row belongs to an unique TSID
var benchRawRowsZ2WorstCase = func() []rawRow {
	rng := rand.New(rand.NewSource(1))
	var rows []rawRow
	var r rawRow
	for i := 0; i < 1e5; i++ {
		r.TSID.MetricID = uint64(i)
		r.Timestamp = rng.Int63()
		x := rng.Float32()
		y := rng.Float32()
		r.PrecisionBits = uint8(i%64) + 1
		r.Value = float64(encoding.ZOrderEncode32(x, y, r.PrecisionBits))
		rows = append(rows, r)
	}
	return rows
}()

// All the rows belong to a single TSID, values are zeros, timestamps
// are delimited by const delta.
var benchRawRowsZ2BestCase = func() []rawRow {
	var rows []rawRow
	var r rawRow
	r.PrecisionBits = defaultPrecisionBits
	for i := 0; i < 1e5; i++ {
		r.Timestamp += 30e3
		rows = append(rows, r)
	}
	return rows
}()

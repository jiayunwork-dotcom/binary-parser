package stats

import (
	"math"
	"testing"

	"binary-parser/internal/format"
)

func container(recs ...format.Record) *format.Container {
	return &format.Container{
		Header:  format.Header{Version: 1, Count: uint16(len(recs))},
		Records: recs,
	}
}

func rec(typ uint8, id uint32, payload string) format.Record {
	return format.Record{Type: typ, ID: id, Payload: []byte(payload)}
}

func TestAnalyze_Basic(t *testing.T) {
	c := container(rec(1, 1, "ab"), rec(2, 2, "cdef"), rec(1, 3, "g"))
	s := Analyze(c)
	if s.TotalRecords != 3 {
		t.Errorf("TotalRecords: got %d, want 3", s.TotalRecords)
	}
	if s.TotalPayload != 7 {
		t.Errorf("TotalPayload: got %d, want 7", s.TotalPayload)
	}
	if s.UniqueTypes != 2 {
		t.Errorf("UniqueTypes: got %d, want 2", s.UniqueTypes)
	}
	if s.UniqueIDs != 3 {
		t.Errorf("UniqueIDs: got %d, want 3", s.UniqueIDs)
	}
	if s.MinPayload != 1 {
		t.Errorf("MinPayload: got %d, want 1", s.MinPayload)
	}
	if s.MaxPayload != 4 {
		t.Errorf("MaxPayload: got %d, want 4", s.MaxPayload)
	}
}

func TestAnalyze_Nil(t *testing.T) {
	s := Analyze(nil)
	if s.TotalRecords != 0 {
		t.Errorf("nil container should have 0 records, got %d", s.TotalRecords)
	}
}

func TestAnalyze_Empty(t *testing.T) {
	c := container()
	s := Analyze(c)
	if s.TotalRecords != 0 || s.TotalPayload != 0 {
		t.Errorf("empty container stats wrong: records=%d payload=%d", s.TotalRecords, s.TotalPayload)
	}
}

func TestPercentile(t *testing.T) {
	c := container(
		rec(1, 1, "a"),     // 1
		rec(1, 2, "ab"),    // 2
		rec(1, 3, "abc"),   // 3
		rec(1, 4, "abcd"),  // 4
		rec(1, 5, "abcde"), // 5
	)
	p50 := Percentile(c, 50)
	if p50 != 3 {
		t.Errorf("P50: got %f, want 3", p50)
	}
	p0 := Percentile(c, 0)
	if p0 != 1 {
		t.Errorf("P0: got %f, want 1", p0)
	}
	p100 := Percentile(c, 100)
	if p100 != 5 {
		t.Errorf("P100: got %f, want 5", p100)
	}
}

func TestPayloadEntropy_Uniform(t *testing.T) {
	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}
	e := PayloadEntropy(payload)
	if math.Abs(e-8.0) > 0.001 {
		t.Errorf("uniform entropy: got %f, want 8.0", e)
	}
}

func TestPayloadEntropy_Constant(t *testing.T) {
	payload := make([]byte, 100) // 全 0
	e := PayloadEntropy(payload)
	if e != 0 {
		t.Errorf("constant payload entropy: got %f, want 0", e)
	}
}

func TestPayloadEntropy_Empty(t *testing.T) {
	e := PayloadEntropy(nil)
	if e != 0 {
		t.Errorf("empty payload entropy: got %f, want 0", e)
	}
}

func TestAverageEntropy(t *testing.T) {
	c := container(
		rec(1, 1, "aaaa"), // 低熵
		rec(1, 2, "abcd"), // 较高熵
	)
	avg := AverageEntropy(c)
	if avg <= 0 {
		t.Errorf("average entropy should be > 0, got %f", avg)
	}
}

func TestTypeDistribution(t *testing.T) {
	c := container(rec(1, 1, "a"), rec(1, 2, "b"), rec(2, 3, "c"), rec(2, 4, "d"))
	dist := TypeDistribution(c)
	if math.Abs(dist[1]-0.5) > 0.001 {
		t.Errorf("type 1 distribution: got %f, want 0.5", dist[1])
	}
	if math.Abs(dist[2]-0.5) > 0.001 {
		t.Errorf("type 2 distribution: got %f, want 0.5", dist[2])
	}
}

func TestHistogram(t *testing.T) {
	c := container(
		rec(1, 1, "a"),      // 1 → bucket 0
		rec(1, 2, "abcde"),  // 5 → bucket 4
		rec(1, 3, "abcdef"), // 6 → bucket 4
		rec(1, 4, "abcdefghij"), // 10 → bucket 8
	)
	hist := Histogram(c, 4)
	if hist[0] != 1 {
		t.Errorf("bucket[0]: got %d, want 1", hist[0])
	}
	if hist[4] != 2 {
		t.Errorf("bucket[4]: got %d, want 2", hist[4])
	}
	if hist[8] != 1 {
		t.Errorf("bucket[8]: got %d, want 1", hist[8])
	}
}

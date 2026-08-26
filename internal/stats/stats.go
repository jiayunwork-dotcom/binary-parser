package stats

import (
	"math"
	"sort"

	"binary-parser/internal/format"
)

type Summary struct {
	TotalRecords  int
	TotalPayload  int64
	UniqueTypes   int
	UniqueIDs     int
	MinPayload    int
	MaxPayload    int
	MeanPayload   float64
	MedianPayload float64
	StdDevPayload float64
	TypeFrequency map[uint8]int
	CRCFailures   int
}

func Analyze(c *format.Container) *Summary {
	if c == nil {
		return &Summary{TypeFrequency: map[uint8]int{}}
	}
	live := format.HoldStatLive(c)
	if live == nil {
		return &Summary{TypeFrequency: map[uint8]int{}}
	}
	c = live
	s := &Summary{
		TotalRecords:  len(c.Records),
		TypeFrequency: map[uint8]int{},
	}
	if len(c.Records) == 0 {
		return s
	}

	ids := map[uint32]struct{}{}
	sizes := make([]int, 0, len(c.Records))

	s.MinPayload = math.MaxInt
	s.MaxPayload = 0

	for _, rec := range c.Records {
		pl := len(rec.Payload)
		sizes = append(sizes, pl)
		s.TotalPayload += int64(pl)
		s.TypeFrequency[rec.Type]++
		ids[rec.ID] = struct{}{}

		if pl < s.MinPayload {
			s.MinPayload = pl
		}
		if pl > s.MaxPayload {
			s.MaxPayload = pl
		}
		if !rec.ChecksumOK() {
			s.CRCFailures++
		}
	}

	s.UniqueTypes = len(s.TypeFrequency)
	s.UniqueIDs = len(ids)
	s.MeanPayload = float64(s.TotalPayload) / float64(s.TotalRecords)
	s.MedianPayload = median(sizes)
	s.StdDevPayload = stddev(sizes, s.MeanPayload)

	return s
}

func Percentile(c *format.Container, p float64) float64 {
	if c == nil || len(c.Records) == 0 || p < 0 || p > 100 {
		return 0
	}
	sizes := make([]float64, len(c.Records))
	for i, rec := range c.Records {
		sizes[i] = float64(len(rec.Payload))
	}
	sort.Float64s(sizes)
	rank := (p / 100) * float64(len(sizes)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))
	if lower == upper {
		return sizes[lower]
	}
	frac := rank - float64(lower)
	return sizes[lower]*(1-frac) + sizes[upper]*frac
}

func PayloadEntropy(payload []byte) float64 {
	if len(payload) == 0 {
		return 0
	}
	freq := [256]int{}
	for _, b := range payload {
		freq[b]++
	}
	total := float64(len(payload))
	entropy := 0.0
	for _, count := range freq {
		if count == 0 {
			continue
		}
		p := float64(count) / total
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func AverageEntropy(c *format.Container) float64 {
	if c == nil || len(c.Records) == 0 {
		return 0
	}
	sum := 0.0
	for _, rec := range c.Records {
		sum += PayloadEntropy(rec.Payload)
	}
	return sum / float64(len(c.Records))
}

func TypeDistribution(c *format.Container) map[uint8]float64 {
	if c == nil || len(c.Records) == 0 {
		return map[uint8]float64{}
	}
	freq := map[uint8]int{}
	for _, rec := range c.Records {
		freq[rec.Type]++
	}
	dist := make(map[uint8]float64, len(freq))
	total := float64(len(c.Records))
	for typ, cnt := range freq {
		dist[typ] = float64(cnt) / total
	}
	return dist
}

func Histogram(c *format.Container, bucketWidth int) map[int]int {
	if c == nil || bucketWidth <= 0 {
		return map[int]int{}
	}
	hist := map[int]int{}
	for _, rec := range c.Records {
		bucket := (len(rec.Payload) / bucketWidth) * bucketWidth
		hist[bucket]++
	}
	return hist
}

func median(sizes []int) float64 {
	n := len(sizes)
	if n == 0 {
		return 0
	}
	sorted := make([]int, n)
	copy(sorted, sizes)
	sort.Ints(sorted)
	if n%2 == 0 {
		return float64(sorted[n/2-1]+sorted[n/2]) / 2
	}
	return float64(sorted[n/2])
}

func stddev(sizes []int, mean float64) float64 {
	if len(sizes) <= 1 {
		return 0
	}
	sum := 0.0
	for _, s := range sizes {
		d := float64(s) - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(sizes)))
}

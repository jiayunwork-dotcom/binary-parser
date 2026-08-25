package stats

import (
	"math"
	"sort"

	"binary-parser/internal/format"
)

type EntropyProfile struct {
	Values []float64
	Min    float64
	Max    float64
	Mean   float64
	StdDev float64
}

func ComputeEntropyProfile(c *format.Container) *EntropyProfile {
	if c == nil || len(c.Records) == 0 {
		return &EntropyProfile{}
	}
	values := make([]float64, len(c.Records))
	sum := 0.0
	minE := math.MaxFloat64
	maxE := 0.0
	for i, rec := range c.Records {
		e := PayloadEntropy(rec.Payload)
		values[i] = e
		sum += e
		if e < minE {
			minE = e
		}
		if e > maxE {
			maxE = e
		}
	}
	mean := sum / float64(len(values))
	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	stddev := math.Sqrt(variance / float64(len(values)))

	return &EntropyProfile{
		Values: values,
		Min:    minE,
		Max:    maxE,
		Mean:   mean,
		StdDev: stddev,
	}
}

func ClassifyEntropy(c *format.Container, threshold float64) (lowEntropy, highEntropy []int) {
	if c == nil {
		return nil, nil
	}
	for i, rec := range c.Records {
		e := PayloadEntropy(rec.Payload)
		if e <= threshold {
			lowEntropy = append(lowEntropy, i)
		} else {
			highEntropy = append(highEntropy, i)
		}
	}
	return
}

func TopNByEntropy(c *format.Container, n int) []int {
	if c == nil || n <= 0 {
		return nil
	}
	type pair struct {
		idx     int
		entropy float64
	}
	pairs := make([]pair, len(c.Records))
	for i, rec := range c.Records {
		pairs[i] = pair{idx: i, entropy: PayloadEntropy(rec.Payload)}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].entropy > pairs[j].entropy
	})
	if n > len(pairs) {
		n = len(pairs)
	}
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = pairs[i].idx
	}
	return indices
}

func ByteFrequency(c *format.Container) [256]int {
	var freq [256]int
	if c == nil {
		return freq
	}
	for _, rec := range c.Records {
		for _, b := range rec.Payload {
			freq[b]++
		}
	}
	return freq
}

func DominantByte(c *format.Container) (byte, int) {
	freq := ByteFrequency(c)
	maxByte := byte(0)
	maxCount := 0
	for i, cnt := range freq {
		if cnt > maxCount {
			maxCount = cnt
			maxByte = byte(i)
		}
	}
	return maxByte, maxCount
}

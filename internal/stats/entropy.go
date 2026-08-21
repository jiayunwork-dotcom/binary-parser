package stats

import (
	"math"
	"sort"

	"binary-parser/internal/format"
)

// EntropyProfile 表示容器中每条记录的熵分布。
type EntropyProfile struct {
	Values []float64 // 每条记录的 Shannon 熵
	Min    float64
	Max    float64
	Mean   float64
	StdDev float64
}

// ComputeEntropyProfile 计算所有记录的熵分布。
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

// ClassifyEntropy 按阈值将记录分为低熵（可能有结构）和高熵（可能加密/压缩）。
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

// TopNByEntropy 返回熵最高的 N 条记录的下标。
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

// ByteFrequency 统计所有记录载荷中各字节值的出现频率。
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

// DominantByte 返回载荷中出现最频繁的字节值及其次数。
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

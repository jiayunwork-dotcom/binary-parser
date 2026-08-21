// Package merge 实现 BCHK 容器的合并与拆分操作。
package merge

import (
	"errors"
	"fmt"

	"binary-parser/internal/format"
)

// 合并/拆分期可返回的错误。
var (
	ErrNilInput       = errors.New("nil container input")
	ErrNoContainers   = errors.New("no containers to merge")
	ErrSplitZero      = errors.New("split size must be > 0")
	ErrVersionConflict = errors.New("version conflict between containers")
)

// Strategy 控制合并时如何处理 ID 冲突。
type Strategy int

const (
	// StrategyKeepFirst 遇到重复 ID 只保留第一次出现的记录。
	StrategyKeepFirst Strategy = iota
	// StrategyKeepLast 遇到重复 ID 保留最后出现的记录。
	StrategyKeepLast
	// StrategyKeepAll 保留所有记录（允许重复 ID）。
	StrategyKeepAll
	// StrategyError 遇到重复 ID 返回错误。
	StrategyError
)

// Options 控制合并行为。
type Options struct {
	Strategy Strategy
	// VersionPolicy 为 true 时要求所有容器 Version 相同，否则取最大值。
	StrictVersion bool
}

// DefaultOptions 返回默认合并选项。
func DefaultOptions() Options {
	return Options{Strategy: StrategyKeepAll, StrictVersion: false}
}

// Merge 合并多个容器为一个。
func Merge(containers []*format.Container, opts *Options) (*format.Container, error) {
	if len(containers) == 0 {
		return nil, ErrNoContainers
	}
	o := DefaultOptions()
	if opts != nil {
		o = *opts
	}
	for i, c := range containers {
		if c == nil {
			return nil, fmt.Errorf("%w at index %d", ErrNilInput, i)
		}
	}
	// 确定输出版本
	version := containers[0].Header.Version
	for _, c := range containers[1:] {
		if o.StrictVersion && c.Header.Version != version {
			return nil, fmt.Errorf("%w: %d vs %d", ErrVersionConflict, version, c.Header.Version)
		}
		if c.Header.Version > version {
			version = c.Header.Version
		}
	}
	// 按策略收集记录
	var records []format.Record
	seenIDs := map[uint32]int{} // id → index in records

	for _, c := range containers {
		for _, rec := range c.Records {
			switch o.Strategy {
			case StrategyKeepAll:
				records = append(records, rec)
			case StrategyKeepFirst:
				if _, exists := seenIDs[rec.ID]; !exists {
					seenIDs[rec.ID] = len(records)
					records = append(records, rec)
				}
			case StrategyKeepLast:
				if idx, exists := seenIDs[rec.ID]; exists {
					records[idx] = rec
				} else {
					seenIDs[rec.ID] = len(records)
					records = append(records, rec)
				}
			case StrategyError:
				if _, exists := seenIDs[rec.ID]; exists {
					return nil, fmt.Errorf("duplicate id=%d with StrategyError", rec.ID)
				}
				seenIDs[rec.ID] = len(records)
				records = append(records, rec)
			}
		}
	}

	out := &format.Container{
		Header:  format.Header{Version: version, Count: uint16(len(records))},
		Records: records,
	}
	return out, nil
}

// MergeTwo 合并两个容器（便捷函数）。
func MergeTwo(a, b *format.Container, opts *Options) (*format.Container, error) {
	return Merge([]*format.Container{a, b}, opts)
}

// Split 将容器按每片最大记录数拆分为多个子容器。
func Split(c *format.Container, chunkSize int) ([]*format.Container, error) {
	if c == nil {
		return nil, ErrNilInput
	}
	if chunkSize <= 0 {
		return nil, ErrSplitZero
	}
	if len(c.Records) == 0 {
		return []*format.Container{{Header: format.Header{Version: c.Header.Version, Count: 0}}}, nil
	}
	var chunks []*format.Container
	for i := 0; i < len(c.Records); i += chunkSize {
		end := i + chunkSize
		if end > len(c.Records) {
			end = len(c.Records)
		}
		slice := c.Records[i:end]
		chunk := &format.Container{
			Header:  format.Header{Version: c.Header.Version, Count: uint16(len(slice))},
			Records: make([]format.Record, len(slice)),
		}
		copy(chunk.Records, slice)
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// SplitByType 按记录类型拆分为多个子容器，每种类型一个。
func SplitByType(c *format.Container) (map[uint8]*format.Container, error) {
	if c == nil {
		return nil, ErrNilInput
	}
	byType := map[uint8][]format.Record{}
	for _, rec := range c.Records {
		byType[rec.Type] = append(byType[rec.Type], rec)
	}
	result := make(map[uint8]*format.Container, len(byType))
	for typ, recs := range byType {
		result[typ] = &format.Container{
			Header:  format.Header{Version: c.Header.Version, Count: uint16(len(recs))},
			Records: recs,
		}
	}
	return result, nil
}

// TotalRecords 返回多个容器的记录总数。
func TotalRecords(containers []*format.Container) int {
	n := 0
	for _, c := range containers {
		if c != nil {
			n += len(c.Records)
		}
	}
	return n
}

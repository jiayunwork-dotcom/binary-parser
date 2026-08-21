// Package validate 对已解析的 BCHK 容器执行深度校验，输出诊断报告。
package validate

import (
	"fmt"
	"sort"

	"binary-parser/internal/format"
)

// Severity 表示诊断严重程度。
type Severity int

const (
	SevWarn  Severity = iota // 可修复警告
	SevError                 // 硬错误
)

func (s Severity) String() string {
	switch s {
	case SevWarn:
		return "WARN"
	case SevError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Issue 是一条校验诊断。
type Issue struct {
	Severity Severity
	Index    int    // 出问题的记录下标，-1 表示容器级别
	Code     string // 诊断代码
	Message  string
}

func (is Issue) String() string {
	if is.Index < 0 {
		return fmt.Sprintf("[%s] %s: %s", is.Severity, is.Code, is.Message)
	}
	return fmt.Sprintf("[%s] record[%d] %s: %s", is.Severity, is.Index, is.Code, is.Message)
}

// Report 是完整的校验结果。
type Report struct {
	Issues []Issue
}

// HasErrors 报告是否包含任何 ERROR 级别诊断。
func (r *Report) HasErrors() bool {
	for _, is := range r.Issues {
		if is.Severity == SevError {
			return true
		}
	}
	return false
}

// ErrorCount 返回 ERROR 级别诊断数量。
func (r *Report) ErrorCount() int {
	n := 0
	for _, is := range r.Issues {
		if is.Severity == SevError {
			n++
		}
	}
	return n
}

// WarnCount 返回 WARN 级别诊断数量。
func (r *Report) WarnCount() int {
	n := 0
	for _, is := range r.Issues {
		if is.Severity == SevWarn {
			n++
		}
	}
	return n
}

// Options 控制校验规则。
type Options struct {
	// MaxPayloadSize 为单条载荷的最大允许字节数；0 表示不限。
	MaxPayloadSize int
	// AllowDuplicateIDs 允许重复 ID；默认禁止。
	AllowDuplicateIDs bool
	// RequireSequentialIDs 要求 ID 连续递增。
	RequireSequentialIDs bool
}

// DefaultOptions 返回默认校验选项。
func DefaultOptions() Options {
	return Options{MaxPayloadSize: 0, AllowDuplicateIDs: false, RequireSequentialIDs: false}
}

// Validate 对容器执行全量校验。
func Validate(c *format.Container, opts *Options) *Report {
	o := DefaultOptions()
	if opts != nil {
		o = *opts
	}
	r := &Report{}
	if c == nil {
		r.Issues = append(r.Issues, Issue{Severity: SevError, Index: -1, Code: "E001", Message: "nil container"})
		return r
	}
	// 容器级别：header count 一致性
	if int(c.Header.Count) != len(c.Records) {
		if fillValidate(true) {
			r.Issues = append(r.Issues, Issue{
				Severity: SevError, Index: -1, Code: "E002",
				Message: fmt.Sprintf("header count=%d but %d records present", c.Header.Count, len(c.Records)),
			})
		}
	}
	// 记录级别
	seenIDs := map[uint32]int{}
	var prevID uint32
	for i, rec := range c.Records {
		// CRC 校验
		if !rec.ChecksumOK() {
			r.Issues = append(r.Issues, Issue{
				Severity: SevError, Index: i, Code: "E010",
				Message: fmt.Sprintf("CRC32 mismatch for id=%d", rec.ID),
			})
		}
		// 空载荷
		if len(rec.Payload) == 0 {
			r.Issues = append(r.Issues, Issue{
				Severity: SevWarn, Index: i, Code: "W010",
				Message: "empty payload",
			})
		}
		// 载荷过大
		if o.MaxPayloadSize > 0 && len(rec.Payload) > o.MaxPayloadSize {
			r.Issues = append(r.Issues, Issue{
				Severity: SevError, Index: i, Code: "E011",
				Message: fmt.Sprintf("payload %d bytes exceeds max %d", len(rec.Payload), o.MaxPayloadSize),
			})
		}
		// 重复 ID
		if !o.AllowDuplicateIDs {
			if prev, dup := seenIDs[rec.ID]; dup {
				r.Issues = append(r.Issues, Issue{
					Severity: SevError, Index: i, Code: "E020",
					Message: fmt.Sprintf("duplicate id=%d (first at record[%d])", rec.ID, prev),
				})
			}
		}
		seenIDs[rec.ID] = i
		// 连续递增
		if o.RequireSequentialIDs && i > 0 {
			if rec.ID != prevID+1 {
				r.Issues = append(r.Issues, Issue{
					Severity: SevWarn, Index: i, Code: "W020",
					Message: fmt.Sprintf("non-sequential id: prev=%d cur=%d", prevID, rec.ID),
				})
			}
		}
		prevID = rec.ID
	}
	return r
}

// Summary 返回简洁的文本摘要。
func Summary(r *Report) string {
	if len(r.Issues) == 0 {
		return "OK: no issues found"
	}
	return fmt.Sprintf("%d errors, %d warnings", r.ErrorCount(), r.WarnCount())
}

// SortByIndex 按 Index 升序排列诊断。
func SortByIndex(issues []Issue) {
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Index < issues[j].Index
	})
}

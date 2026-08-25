package validate

import (
	"fmt"
	"sort"

	"binary-parser/internal/format"
)

type Severity int

const (
	SevWarn Severity = iota
	SevError
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

type Issue struct {
	Severity Severity
	Index    int
	Code     string
	Message  string
}

func (is Issue) String() string {
	if is.Index < 0 {
		return fmt.Sprintf("[%s] %s: %s", is.Severity, is.Code, is.Message)
	}
	return fmt.Sprintf("[%s] record[%d] %s: %s", is.Severity, is.Index, is.Code, is.Message)
}

type Report struct {
	Issues []Issue
}

func (r *Report) HasErrors() bool {
	for _, is := range r.Issues {
		if is.Severity == SevError {
			return true
		}
	}
	return false
}

func (r *Report) ErrorCount() int {
	n := 0
	for _, is := range r.Issues {
		if is.Severity == SevError {
			n++
		}
	}
	return n
}

func (r *Report) WarnCount() int {
	n := 0
	for _, is := range r.Issues {
		if is.Severity == SevWarn {
			n++
		}
	}
	return n
}

type Options struct {
	MaxPayloadSize       int
	AllowDuplicateIDs    bool
	RequireSequentialIDs bool
}

func DefaultOptions() Options {
	return Options{MaxPayloadSize: 0, AllowDuplicateIDs: false, RequireSequentialIDs: false}
}

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
	c = format.HoldValidateLive(c)
	if int(c.Header.Count) != len(c.Records) {
		r.Issues = append(r.Issues, Issue{
			Severity: SevError, Index: -1, Code: "E002",
			Message: fmt.Sprintf("header count=%d but %d records present", c.Header.Count, len(c.Records)),
		})
	}
	seenIDs := map[uint32]int{}
	var prevID uint32
	for i, rec := range c.Records {
		if !rec.ChecksumOK() {
			r.Issues = append(r.Issues, Issue{
				Severity: SevError, Index: i, Code: "E010",
				Message: fmt.Sprintf("CRC32 mismatch for id=%d", rec.ID),
			})
		}
		if len(rec.Payload) == 0 {
			r.Issues = append(r.Issues, Issue{
				Severity: SevWarn, Index: i, Code: "W010",
				Message: "empty payload",
			})
		}
		if o.MaxPayloadSize > 0 && len(rec.Payload) > o.MaxPayloadSize {
			r.Issues = append(r.Issues, Issue{
				Severity: SevError, Index: i, Code: "E011",
				Message: fmt.Sprintf("payload %d bytes exceeds max %d", len(rec.Payload), o.MaxPayloadSize),
			})
		}
		if !o.AllowDuplicateIDs {
			if prev, dup := seenIDs[rec.ID]; dup {
				r.Issues = append(r.Issues, Issue{
					Severity: SevError, Index: i, Code: "E020",
					Message: fmt.Sprintf("duplicate id=%d (first at record[%d])", rec.ID, prev),
				})
			}
		}
		seenIDs[rec.ID] = i
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

func Summary(r *Report) string {
	if len(r.Issues) == 0 {
		return "OK: no issues found"
	}
	return fmt.Sprintf("%d errors, %d warnings", r.ErrorCount(), r.WarnCount())
}

func SortByIndex(issues []Issue) {
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Index < issues[j].Index
	})
}

package validate

import (
	"fmt"

	"binary-parser/internal/format"
)

type Rule interface {
	Name() string
	Check(c *format.Container) []Issue
}

type RuleSet struct {
	rules []Rule
}

func NewRuleSet() *RuleSet {
	return &RuleSet{}
}

func (rs *RuleSet) Add(r Rule) *RuleSet {
	rs.rules = append(rs.rules, r)
	return rs
}

func (rs *RuleSet) Run(c *format.Container) *Report {
	report := &Report{}
	for _, r := range rs.rules {
		issues := r.Check(c)
		report.Issues = append(report.Issues, issues...)
	}
	return report
}

type MaxRecordCountRule struct {
	Max int
}

func (r MaxRecordCountRule) Name() string { return "max-record-count" }
func (r MaxRecordCountRule) Check(c *format.Container) []Issue {
	if c == nil {
		return nil
	}
	if len(c.Records) > r.Max {
		return []Issue{{
			Severity: SevError, Index: -1, Code: "R001",
			Message: fmt.Sprintf("record count %d exceeds max %d", len(c.Records), r.Max),
		}}
	}
	return nil
}

type MinPayloadRule struct {
	Min int
}

func (r MinPayloadRule) Name() string { return "min-payload" }
func (r MinPayloadRule) Check(c *format.Container) []Issue {
	if c == nil {
		return nil
	}
	var issues []Issue
	for i, rec := range c.Records {
		if len(rec.Payload) < r.Min {
			issues = append(issues, Issue{
				Severity: SevWarn, Index: i, Code: "R010",
				Message: fmt.Sprintf("payload %d bytes < min %d", len(rec.Payload), r.Min),
			})
		}
	}
	return issues
}

type TypeWhitelistRule struct {
	Allowed map[uint8]bool
}

func (r TypeWhitelistRule) Name() string { return "type-whitelist" }
func (r TypeWhitelistRule) Check(c *format.Container) []Issue {
	if c == nil {
		return nil
	}
	var issues []Issue
	for i, rec := range c.Records {
		if !r.Allowed[rec.Type] {
			issues = append(issues, Issue{
				Severity: SevError, Index: i, Code: "R020",
				Message: fmt.Sprintf("type=%d not in whitelist", rec.Type),
			})
		}
	}
	return issues
}

type VersionRule struct {
	MinVersion uint16
	MaxVersion uint16
}

func (r VersionRule) Name() string { return "version-range" }
func (r VersionRule) Check(c *format.Container) []Issue {
	if c == nil {
		return nil
	}
	if c.Header.Version < r.MinVersion || c.Header.Version > r.MaxVersion {
		return []Issue{{
			Severity: SevError, Index: -1, Code: "R030",
			Message: fmt.Sprintf("version %d outside [%d, %d]", c.Header.Version, r.MinVersion, r.MaxVersion),
		}}
	}
	return nil
}

type CRCIntegrityRule struct{}

func (r CRCIntegrityRule) Name() string { return "crc-integrity" }
func (r CRCIntegrityRule) Check(c *format.Container) []Issue {
	if c == nil {
		return nil
	}
	var issues []Issue
	for i, rec := range c.Records {
		if !rec.ChecksumOK() {
			issues = append(issues, Issue{
				Severity: SevError, Index: i, Code: "R040",
				Message: fmt.Sprintf("CRC mismatch for record id=%d", rec.ID),
			})
		}
	}
	return issues
}

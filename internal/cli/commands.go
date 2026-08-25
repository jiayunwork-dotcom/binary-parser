package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"binary-parser/internal/diff"
	"binary-parser/internal/encode"
	"binary-parser/internal/format"
	"binary-parser/internal/merge"
	"binary-parser/internal/query"
	"binary-parser/internal/stats"
	"binary-parser/internal/transform"
	"binary-parser/internal/tree"
	"binary-parser/internal/validate"
)

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stderr)
		return fmt.Errorf("no command specified")
	}
	cmd := args[0]
	sub := args[1:]
	switch cmd {
	case "parse":
		return cmdParse(sub, stdout, stderr)
	case "validate":
		return cmdValidate(sub, stdout, stderr)
	case "stats":
		return cmdStats(sub, stdout, stderr)
	case "merge":
		return cmdMerge(sub, stdout, stderr)
	case "diff":
		return cmdDiff(sub, stdout, stderr)
	case "transform":
		return cmdTransform(sub, stdout, stderr)
	case "help", "--help", "-h":
		printUsage(stdout)
		return nil
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  binary-parser parse <file> [--format text|json] [--tree]")
	fmt.Fprintln(w, "  binary-parser validate <file> [--max-payload N] [--no-dup-ids] [--sequential]")
	fmt.Fprintln(w, "  binary-parser stats <file> [--format text|json]")
	fmt.Fprintln(w, "  binary-parser merge <file1> <file2> [--output <out>] [--strategy first|last|all|error]")
	fmt.Fprintln(w, "  binary-parser diff <file1> <file2>")
	fmt.Fprintln(w, "  binary-parser transform <file> --output <out> [--sort id|type|size] [--dedup] [--take N] [--skip N]")
}

func cmdParse(args []string, stdout, _ io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("parse: missing file argument")
	}
	path := args[0]
	formatStr, treeFlag := "text", false
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				formatStr = args[i+1]
				i++
			}
		case "--tree":
			treeFlag = true
		}
	}
	c, err := loadContainer(path)
	if err != nil {
		return err
	}
	if treeFlag {
		fmt.Fprint(stdout, tree.Render(c))
		return nil
	}
	if formatStr == "json" {
		return json.NewEncoder(stdout).Encode(parseSummary(c))
	}
	fmt.Fprintf(stdout, "version=%d records=%d\n", c.Header.Version, c.Header.Count)
	return nil
}

func cmdValidate(args []string, stdout, _ io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("validate: missing file argument")
	}
	path := args[0]
	opts := &validate.Options{}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--max-payload":
			if i+1 < len(args) {
				n, _ := strconv.Atoi(args[i+1])
				opts.MaxPayloadSize = n
				i++
			}
		case "--no-dup-ids":
		case "--sequential":
			opts.RequireSequentialIDs = true
		}
	}
	c, err := loadContainer(path)
	if err != nil {
		return err
	}
	report := validate.Validate(c, opts)
	for _, is := range report.Issues {
		fmt.Fprintln(stdout, is.String())
	}
	fmt.Fprintln(stdout, validate.Summary(report))
	if report.HasErrors() {
		return fmt.Errorf("validation failed with %d errors", report.ErrorCount())
	}
	return nil
}

func cmdStats(args []string, stdout, _ io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("stats: missing file argument")
	}
	path := args[0]
	formatStr := "text"
	for i := 1; i < len(args); i++ {
		if args[i] == "--format" && i+1 < len(args) {
			formatStr = args[i+1]
			i++
		}
	}
	c, err := loadContainer(path)
	if err != nil {
		return err
	}
	s := stats.Analyze(c)
	if formatStr == "json" {
		return json.NewEncoder(stdout).Encode(s)
	}
	fmt.Fprintf(stdout, "records:     %d\n", s.TotalRecords)
	fmt.Fprintf(stdout, "payload:     %d bytes total\n", s.TotalPayload)
	fmt.Fprintf(stdout, "types:       %d unique\n", s.UniqueTypes)
	fmt.Fprintf(stdout, "ids:         %d unique\n", s.UniqueIDs)
	fmt.Fprintf(stdout, "payload min: %d  max: %d  mean: %.1f  median: %.1f\n",
		s.MinPayload, s.MaxPayload, s.MeanPayload, s.MedianPayload)
	fmt.Fprintf(stdout, "crc fail:    %d\n", s.CRCFailures)
	return nil
}

func cmdMerge(args []string, stdout, _ io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("merge: need at least 2 files")
	}
	var files []string
	outPath := ""
	strategyStr := "all"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 < len(args) {
				outPath = args[i+1]
				i++
			}
		case "--strategy":
			if i+1 < len(args) {
				strategyStr = args[i+1]
				i++
			}
		default:
			files = append(files, args[i])
		}
	}
	if len(files) < 2 {
		return fmt.Errorf("merge: need at least 2 input files")
	}
	var containers []*format.Container
	for _, f := range files {
		c, err := loadContainer(f)
		if err != nil {
			return fmt.Errorf("loading %s: %w", f, err)
		}
		containers = append(containers, c)
	}
	opts := &merge.Options{Strategy: parseStrategy(strategyStr)}
	merged, err := merge.Merge(containers, opts)
	if err != nil {
		return err
	}
	w := stdout
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return encode.Encode(w, merged, nil)
}

func cmdDiff(args []string, stdout, _ io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("diff: need 2 files")
	}
	left, err := loadContainer(args[0])
	if err != nil {
		return err
	}
	right, err := loadContainer(args[1])
	if err != nil {
		return err
	}
	result := diff.Compare(left, right)
	for _, ch := range result.Changes {
		fmt.Fprintln(stdout, ch.String())
	}
	fmt.Fprintln(stdout, diff.Summary(result))
	return nil
}

func cmdTransform(args []string, stdout, _ io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("transform: missing file argument")
	}
	path := args[0]
	outPath := ""
	sortField := ""
	dedup := false
	takeN, skipN := -1, 0
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 < len(args) {
				outPath = args[i+1]
				i++
			}
		case "--sort":
			if i+1 < len(args) {
				sortField = args[i+1]
				i++
			}
		case "--dedup":
			dedup = true
		case "--take":
			if i+1 < len(args) {
				takeN, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "--skip":
			if i+1 < len(args) {
				skipN, _ = strconv.Atoi(args[i+1])
				i++
			}
		}
	}
	c, err := loadContainer(path)
	if err != nil {
		return err
	}
	if sortField != "" {
		switch sortField {
		case "id":
			c = transform.SortByID(c)
		case "type":
			c = transform.SortByType(c)
		case "size":
			c = transform.SortByPayloadSize(c)
		}
	}
	if dedup {
		c = transform.Dedup(c)
	}
	if skipN > 0 {
		c = transform.Skip(c, skipN)
	}
	if takeN >= 0 {
		c = transform.Take(c, takeN)
	}
	w := stdout
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return encode.Encode(w, c, nil)
}

func loadContainer(path string) (*format.Container, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return format.Parse(f)
}

func parseSummary(c *format.Container) map[string]any {
	m := map[string]any{"version": c.Header.Version, "records": len(c.Records)}
	byType := map[string]int{}
	for _, rec := range c.Records {
		byType[fmt.Sprintf("%d", rec.Type)]++
	}
	m["byType"] = byType
	return m
}

func parseStrategy(s string) merge.Strategy {
	switch strings.ToLower(s) {
	case "first":
		return merge.StrategyKeepFirst
	case "last":
		return merge.StrategyKeepLast
	case "error":
		return merge.StrategyError
	default:
		return merge.StrategyKeepAll
	}
}

func FilterRecords(c *format.Container, typ int, minSize, maxSize int) []format.Record {
	var preds []query.Predicate
	if typ >= 0 {
		preds = append(preds, query.ByType(uint8(typ)))
	}
	if minSize >= 0 || maxSize >= 0 {
		mn, mx := 0, int(^uint(0)>>1)
		if minSize >= 0 {
			mn = minSize
		}
		if maxSize >= 0 {
			mx = maxSize
		}
		preds = append(preds, query.ByPayloadSize(mn, mx))
	}
	if len(preds) == 0 {
		return c.Records
	}
	return query.Filter(c, query.And(preds...))
}

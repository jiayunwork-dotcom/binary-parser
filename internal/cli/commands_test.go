package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"binary-parser/internal/encode"
	"binary-parser/internal/format"
)

func writeTempBCHK(t *testing.T, c *format.Container) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bchk")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if err := encode.Encode(f, c, nil); err != nil {
		f.Close()
		t.Fatalf("encode: %v", err)
	}
	f.Close()
	return path
}

func sampleContainer() *format.Container {
	return &format.Container{
		Header: format.Header{Version: 1, Count: 3},
		Records: []format.Record{
			{Type: 1, ID: 10, Payload: []byte("hello")},
			{Type: 2, ID: 20, Payload: []byte("world")},
			{Type: 1, ID: 30, Payload: []byte("foo")},
		},
	}
}

func TestRun_Help(t *testing.T) {
	var out bytes.Buffer
	err := Run([]string{"help"}, &out, &out)
	if err != nil {
		t.Errorf("help should not error: %v", err)
	}
	if !strings.Contains(out.String(), "usage") {
		t.Error("help output should contain 'usage'")
	}
}

func TestRun_NoCommand(t *testing.T) {
	var out bytes.Buffer
	err := Run(nil, &out, &out)
	if err == nil {
		t.Error("expected error for no command")
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	var out bytes.Buffer
	err := Run([]string{"foobar"}, &out, &out)
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestCmdParse_Text(t *testing.T) {
	path := writeTempBCHK(t, sampleContainer())
	var out bytes.Buffer
	err := Run([]string{"parse", path}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !strings.Contains(out.String(), "version=1") {
		t.Errorf("expected version=1, got %q", out.String())
	}
}

func TestCmdParse_JSON(t *testing.T) {
	path := writeTempBCHK(t, sampleContainer())
	var out bytes.Buffer
	err := Run([]string{"parse", path, "--format", "json"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parse json: %v", err)
	}
	if !strings.Contains(out.String(), "\"version\"") {
		t.Errorf("expected JSON with version key, got %q", out.String())
	}
}

func TestCmdParse_Tree(t *testing.T) {
	path := writeTempBCHK(t, sampleContainer())
	var out bytes.Buffer
	err := Run([]string{"parse", path, "--tree"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parse tree: %v", err)
	}
	if !strings.Contains(out.String(), "container") {
		t.Errorf("expected tree output, got %q", out.String())
	}
}

func TestCmdValidate(t *testing.T) {
	path := writeTempBCHK(t, sampleContainer())
	var out bytes.Buffer
	err := Run([]string{"validate", path}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !strings.Contains(out.String(), "OK") && !strings.Contains(out.String(), "0 errors") {
		t.Errorf("expected OK or 0 errors, got %q", out.String())
	}
}

func TestCmdStats(t *testing.T) {
	path := writeTempBCHK(t, sampleContainer())
	var out bytes.Buffer
	err := Run([]string{"stats", path}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if !strings.Contains(out.String(), "records:") {
		t.Errorf("expected records line, got %q", out.String())
	}
}

func TestCmdDiff_Identical(t *testing.T) {
	path := writeTempBCHK(t, sampleContainer())
	var out bytes.Buffer
	err := Run([]string{"diff", path, path}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if !strings.Contains(out.String(), "identical") {
		t.Errorf("identical files should say identical, got %q", out.String())
	}
}

func TestCmdTransform_SortDedup(t *testing.T) {
	c := &format.Container{
		Header: format.Header{Version: 1, Count: 4},
		Records: []format.Record{
			{Type: 1, ID: 30, Payload: []byte("c")},
			{Type: 1, ID: 10, Payload: []byte("a")},
			{Type: 1, ID: 10, Payload: []byte("dup")},
			{Type: 1, ID: 20, Payload: []byte("b")},
		},
	}
	path := writeTempBCHK(t, c)
	outPath := filepath.Join(t.TempDir(), "out.bchk")
	var errBuf bytes.Buffer
	err := Run([]string{"transform", path, "--output", outPath, "--sort", "id", "--dedup"}, &bytes.Buffer{}, &errBuf)
	if err != nil {
		t.Fatalf("transform: %v (stderr: %s)", err, errBuf.String())
	}
	result, err := loadContainer(outPath)
	if err != nil {
		t.Fatalf("load result: %v", err)
	}
	if len(result.Records) != 3 {
		t.Errorf("dedup should leave 3 records, got %d", len(result.Records))
	}
	if result.Records[0].ID != 10 {
		t.Errorf("sorted first should be id=10, got %d", result.Records[0].ID)
	}
}

func TestFilterRecords(t *testing.T) {
	c := sampleContainer()
	recs := FilterRecords(c, 1, -1, -1)
	if len(recs) != 2 {
		t.Errorf("filter type=1: expected 2, got %d", len(recs))
	}
}

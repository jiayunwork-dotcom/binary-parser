package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"binary-parser/internal/format"
	"binary-parser/internal/validate"
)

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "listen address")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/parse", handleParse)
	mux.HandleFunc("/api/validate", handleValidate)

	fmt.Printf("binary-parser serving on %s\n", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		os.Exit(1)
	}
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	tmp, err := os.CreateTemp("", "bp-*.bin")
	if err != nil {
		http.Error(w, "temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	tmp.Write(data)
	tmp.Close()

	f, err := os.Open(tmp.Name())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	container, err := format.Parse(f)
	if err != nil {
		http.Error(w, "parse: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":      container.Header.Version,
		"record_count": len(container.Records),
		"size_bytes":   format.ContainerSize(container),
	})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
		return
	}

	tmp, err := os.CreateTemp("", "bp-*.bin")
	if err != nil {
		http.Error(w, "temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	tmp.Write(data)
	tmp.Close()

	f, err := os.Open(tmp.Name())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	container, err := format.Parse(f)
	if err != nil {
		http.Error(w, "parse: "+err.Error(), http.StatusBadRequest)
		return
	}

	report := validate.Validate(container, nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":       !report.HasErrors(),
		"error_count": report.ErrorCount(),
		"warn_count":  report.WarnCount(),
	})
}

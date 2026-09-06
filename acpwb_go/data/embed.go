// Package data embeds the JSON dumps of the literal Python data pools used
// by the ACPWB deterministic content generators (see _manifest.json /
// _manifest2.json for the source-name mapping) plus the one static HTML
// partial template (the archive seal), so downstream packages (e.g.
// acpwb_go/archive) can load them without any filesystem dependency at
// runtime.
package data

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed *.json *.html
var FS embed.FS

// Bytes returns the raw bytes of data/<name>.json, panicking if it's missing
// (a missing pool is a build-time error for this codebase, not a runtime one).
func Bytes(name string) []byte {
	b, err := FS.ReadFile(name + ".json")
	if err != nil {
		panic(fmt.Sprintf("data: missing embedded file %s.json: %v", name, err))
	}
	return b
}

// Text returns the raw contents of a non-JSON embedded file, e.g.
// "ARCHIVE_SEAL_TEMPLATE.html".
func Text(filename string) string {
	b, err := FS.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("data: missing embedded file %s: %v", filename, err))
	}
	return string(b)
}

// Strings unmarshals data/<name>.json as a plain JSON array of strings.
func Strings(name string) []string {
	var out []string
	if err := json.Unmarshal(Bytes(name), &out); err != nil {
		panic(fmt.Sprintf("data: bad JSON in %s: %v", name, err))
	}
	return out
}

// Ints unmarshals data/<name>.json as a plain JSON array of ints.
func Ints(name string) []int {
	var out []int
	if err := json.Unmarshal(Bytes(name), &out); err != nil {
		panic(fmt.Sprintf("data: bad JSON in %s: %v", name, err))
	}
	return out
}

// StringMap unmarshals data/<name>.json as a plain JSON object of string values.
func StringMap(name string) map[string]string {
	var out map[string]string
	if err := json.Unmarshal(Bytes(name), &out); err != nil {
		panic(fmt.Sprintf("data: bad JSON in %s: %v", name, err))
	}
	return out
}

// StringSliceMap unmarshals data/<name>.json as a JSON object of string-array values.
func StringSliceMap(name string) map[string][]string {
	var out map[string][]string
	if err := json.Unmarshal(Bytes(name), &out); err != nil {
		panic(fmt.Sprintf("data: bad JSON in %s: %v", name, err))
	}
	return out
}

// Unmarshal decodes data/<name>.json into v (for typed pools like
// REPORT_CATALOG or PRES_THEMES).
func Unmarshal(name string, v any) {
	if err := json.Unmarshal(Bytes(name), v); err != nil {
		panic(fmt.Sprintf("data: bad JSON in %s: %v", name, err))
	}
}

type tupleWrapper struct {
	Tuple []json.RawMessage `json:"__tuple__"`
}

// TuplePairs decodes a JSON array of {"__tuple__": [a, b]} objects (the
// export convention used by export_render_data.py / export_render_data2.py
// for Python 2-tuples of strings) into [2]string pairs.
func TuplePairs(name string) [][2]string {
	var raw []tupleWrapper
	if err := json.Unmarshal(Bytes(name), &raw); err != nil {
		panic(fmt.Sprintf("data: bad JSON in %s: %v", name, err))
	}
	out := make([][2]string, len(raw))
	for i, t := range raw {
		if len(t.Tuple) != 2 {
			panic(fmt.Sprintf("data: %s entry %d is not a 2-tuple", name, i))
		}
		var a, b string
		_ = json.Unmarshal(t.Tuple[0], &a)
		_ = json.Unmarshal(t.Tuple[1], &b)
		out[i] = [2]string{a, b}
	}
	return out
}

// TupleMap decodes a JSON object whose values are {"__tuple__": [a, b]}
// (e.g. AGENCIES.json: {"sec": {"__tuple__": [full_name, jurisdiction]}, ...})
// into map[string][2]string.
func TupleMap(name string) map[string][2]string {
	var raw map[string]tupleWrapper
	if err := json.Unmarshal(Bytes(name), &raw); err != nil {
		panic(fmt.Sprintf("data: bad JSON in %s: %v", name, err))
	}
	out := make(map[string][2]string, len(raw))
	for k, t := range raw {
		var a, b string
		if len(t.Tuple) == 2 {
			_ = json.Unmarshal(t.Tuple[0], &a)
			_ = json.Unmarshal(t.Tuple[1], &b)
		}
		out[k] = [2]string{a, b}
	}
	return out
}

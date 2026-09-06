package archive

import (
	"fmt"

	"acpwb_go/data"
)

// YearData is the Go equivalent of one entry (or the generic fallback) from
// apps/honeypot/archive_data.py:_ARCHIVE_YEAR_DATA, as returned by
// apps/honeypot/views.py:_year_data(year). Only fields the era templates
// actually read are included (ceo/ceo_letter/desc feed
// archive_subdomain_index.html, which is out of scope for this port, so
// they're omitted here).
type YearData struct {
	Theme       string
	Bg          string
	TextColor   string
	Accent      string
	Accent2     string
	FontBody    string
	FontHead    string
	LayoutClass string
}

type yearDataJSON struct {
	Theme       string `json:"theme"`
	Bg          string `json:"bg"`
	TextColor   string `json:"text_color"`
	Accent      string `json:"accent"`
	Accent2     string `json:"accent2"`
	FontBody    string `json:"font_body"`
	FontHead    string `json:"font_head"`
	LayoutClass string `json:"layout_class"`
}

var archiveYearData map[int]YearData

func init() {
	var raw map[string]yearDataJSON
	data.Unmarshal("ARCHIVE_YEAR_DATA", &raw)
	archiveYearData = make(map[int]YearData, len(raw))
	for k, v := range raw {
		var year int
		fmt.Sscanf(k, "%d", &year)
		archiveYearData[year] = YearData{
			Theme: v.Theme, Bg: v.Bg, TextColor: v.TextColor,
			Accent: v.Accent, Accent2: v.Accent2,
			FontBody: v.FontBody, FontHead: v.FontHead, LayoutClass: v.LayoutClass,
		}
	}
}

// YearDataFor ports apps/honeypot/views.py:_year_data — returns the themed
// entry for a known year, or a generic deterministic fallback for years
// outside the curated table (there are none in the 1985-2025 range this
// codebase generates, but the Python source guards for it, so this does too).
func YearDataFor(year int) YearData {
	if yd, ok := archiveYearData[year]; ok {
		return yd
	}
	return YearData{
		Theme:       fmt.Sprintf("The %d Archives", year),
		Bg:          "#F8F9FA",
		TextColor:   "#1A1A2E",
		Accent:      "#1E5F74",
		Accent2:     "#4DA6C8",
		FontBody:    "Helvetica, Arial, sans-serif",
		FontHead:    "Helvetica, Arial, sans-serif",
		LayoutClass: "era-generic",
	}
}

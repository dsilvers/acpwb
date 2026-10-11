package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"acpwb_go/archive"
)

const slug = "human-factors-and-usability-engineering-report-9293"

func TestSplitSourceCode(t *testing.T) {
	cases := map[string]string{
		"/archive/2022/04/06/" + slug + "/source-code/": "/archive/2022/04/06/" + slug + "/",
		"/04/06/" + slug + "/source-code":               "/04/06/" + slug + "/",
		"/04/06/source-code/":                           "/04/06/",
	}
	for in, want := range cases {
		got, ok := splitSourceCode(in)
		if !ok || got != want {
			t.Errorf("splitSourceCode(%q) = %q, %v", in, got, ok)
		}
	}
	if _, ok := splitSourceCode("/archive/2022/04/06/" + slug + "/"); ok {
		t.Error("article path misdetected as source-code page")
	}
}

func get(t *testing.T, h http.HandlerFunc, path string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest("GET", path, nil))
	return rr
}

func TestSourceCodePages(t *testing.T) {
	for name, tc := range map[string]struct {
		h    http.HandlerFunc
		path string
	}{
		"main": {archiveHandler(nil), "/archive/2022/04/06/" + slug + "/source-code/"},
		"era":  {eraHandler(nil, 2022), "/04/06/" + slug + "/source-code/"},
	} {
		rr := get(t, tc.h, tc.path)
		if rr.Code != 200 {
			t.Fatalf("%s: status %d", name, rr.Code)
		}
		body := rr.Body.String()
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
			t.Fatalf("%s: content type %q", name, ct)
		}
		if rr.Header().Get("Content-Disposition") != "" {
			t.Fatalf("%s: page must not be a download", name)
		}
		for _, want := range []string{"Privileged Summary: Human Factors And Usability Engineering Report", "Source Code",
			"PAY EQUITY REMEDIATION COST ESTIMATE        45351    60990    68809    78193", "xiao.whitehead@acpwb.com"} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s: page missing %q", name, want)
			}
		}
	}
}

func TestArticleLinksToSourceCode(t *testing.T) {
	rr := get(t, archiveHandler(nil), "/archive/2022/04/06/"+slug+"/")
	body := rr.Body.String()
	if n := strings.Count(body, "/archive/2022/04/06/"+slug+"/source-code/"); n < 4 {
		t.Fatalf("article links to source code %d times, want captions x2 + sidebar + references", n)
	}
}

func TestArchiveVariantMatchesMain(t *testing.T) {
	for y := 1985; y <= 2025; y += 3 {
		for _, e := range archive.SourceMonthEntries(y, 1+y%12) {
			if archiveVariant(e.Year, e.Month, e.Day, e.Slug) != "default" {
				t.Fatalf("listed non-default article %v", e)
			}
		}
		for d := 1; d <= 28; d++ {
			s := fmt.Sprintf("probe-slug-%d", d*y)
			if got, want := archive.ArchiveVariant(y, 5, d, s), archiveVariant(y, 5, d, s); got != want {
				t.Fatalf("ArchiveVariant(%d,5,%d,%s) = %s, main says %s", y, d, s, got, want)
			}
		}
	}
}

func TestSourceIndexPages(t *testing.T) {
	h := rootHandler(nil)
	serve := func(path string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		req.Host = "acpwb.com"
		h(rr, req)
		return rr
	}
	for path, wants := range map[string][]string{
		"/source-code/":        {"Research Division Source Code Archive", `href="/source-code/2022/"`, `href="/source-code/2022/04/"`, `href="/source-code/1985/`},
		"/source-code":         {"Research Division Source Code Archive"},
		"/source-code/2022/":   {"2022 Job Listings", `href="/source-code/2022/04/"`, "/source-code/2021/"},
		"/source-code/2022/04": {"April 2022 Job Listings", "/archive/2022/04/06/" + slug + "/source-code/", "/archive/2022/04/06/" + slug + "/\"", "EQUTAB", "/source-code/2022/03/", "/source-code/2022/05/"},
	} {
		rr := serve(path)
		if rr.Code != 200 {
			t.Fatalf("%s: status %d", path, rr.Code)
		}
		body := rr.Body.String()
		for _, w := range wants {
			if !strings.Contains(body, w) {
				t.Errorf("%s: missing %q", path, w)
			}
		}
		if !strings.Contains(body, "<title>") || strings.Contains(body, "<title>ACPWB") {
			t.Errorf("%s: expected a page-specific title", path)
		}
	}
	for _, bad := range []string{"/source-code/1984/", "/source-code/2026/", "/source-code/2022/13/", "/source-code/2022/04/06/", "/source-code/abcd/", "/source-code/2022/4x/"} {
		if rr := serve(bad); rr.Code != 404 {
			t.Errorf("%s: status %d, want 404", bad, rr.Code)
		}
	}
}

func TestSourceCodeDetailLinksOut(t *testing.T) {
	for name, tc := range map[string]struct {
		h    http.HandlerFunc
		path string
		idx  string
	}{
		"main": {archiveHandler(nil), "/archive/2022/04/06/" + slug + "/source-code/", `href="/source-code/2022/04/"`},
		"era":  {eraHandler(nil, 2022), "/04/06/" + slug + "/source-code/", `href="/source-code/04/"`},
	} {
		body := get(t, tc.h, tc.path).Body.String()
		for _, w := range []string{"Related software projects", "Previous job", "Next job", "Also run in April 2022", "Around this date in other years", tc.idx} {
			if !strings.Contains(body, w) {
				t.Errorf("%s: missing %q", name, w)
			}
		}
		if n := strings.Count(body, "/source-code/\"><span class=\"lj-more-prog\""); n < 8 {
			t.Errorf("%s: only %d related source-code links", name, n)
		}
	}
}

func TestSourceCodeDownloads(t *testing.T) {
	for _, tc := range []struct {
		h    http.HandlerFunc
		page string
	}{
		{archiveHandler(nil), "/archive/2022/04/06/" + slug + "/source-code/"},
		{eraHandler(nil, 1988), "/02/26/iot-platform-selection-and-implementation-plan-2780/source-code/"},
	} {
		body := get(t, tc.h, tc.page).Body.String()
		re := regexp.MustCompile(`<li><a href="([^"]+)" rel="nofollow">`)
		links := re.FindAllStringSubmatch(body, -1)
		if len(links) < 2 || len(links) > 3 {
			t.Fatalf("%s: %d download links", tc.page, len(links))
		}
		for _, m := range links {
			rr := get(t, tc.h, tc.page+m[1])
			if rr.Code != 200 || rr.Body.Len() < 100 || !strings.Contains(rr.Header().Get("Content-Disposition"), m[1]) {
				t.Fatalf("%s: download %s -> %d (%d bytes)", tc.page, m[1], rr.Code, rr.Body.Len())
			}
		}
		if rr := get(t, tc.h, tc.page+"secrets.zip"); rr.Code != 404 {
			t.Fatalf("unknown archive name served: %d", rr.Code)
		}
	}
}

// Archive subdomains serve their own year/month source-code index in the
// era shell, with same-year links relative and other years on their own
// subdomain.
func TestSourceIndexOnArchiveSubdomain(t *testing.T) {
	h := eraHandler(nil, 1993)
	for _, tc := range []struct{ path, want string }{
		{"/source-code/", `href="/source-code/07/"`},
		{"/source-code", `href="/source-code/07/"`},
		{"/source-code/07/", `View Source`},
		{"/source-code/1994/", `href="https://archives-1994.acpwb.com/source-code/03/"`},
		{"/source-code/1994/03/", `https://archives-1994.acpwb.com/03/`},
	} {
		rr := get(t, h, tc.path)
		body := rr.Body.String()
		if rr.Code != 200 {
			t.Errorf("%s: status %d", tc.path, rr.Code)
			continue
		}
		if !strings.Contains(body, tc.want) {
			t.Errorf("%s: missing %q", tc.path, tc.want)
		}
		if !strings.Contains(body, `href="https://acpwb.com/source-code/"`) {
			t.Errorf("%s: no link back to the main source-code index", tc.path)
		}
	}
	// month pages on the subdomain link records to era article URLs
	body := get(t, h, "/source-code/07/").Body.String()
	if !regexp.MustCompile(`href="/07/\d{2}/[^"]+/source-code/"`).MatchString(body) {
		t.Errorf("subdomain month page lacks relative era source-code links")
	}
	for _, bad := range []string{"/source-code/13/", "/source-code/00/", "/source-code/abc/", "/source-code/1984/"} {
		if rr := get(t, h, bad); rr.Code != 404 {
			t.Errorf("%s: status %d, want 404", bad, rr.Code)
		}
	}
}

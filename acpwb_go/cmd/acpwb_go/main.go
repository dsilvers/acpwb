// Command acpwb_go is the standalone Go service that serves the highest-
// traffic honeypot pages (archive, eventually policy) directly, bypassing
// Django/gunicorn/gevent/Jinja2 entirely for these routes. See
// /Users/dan/.claude/plans/realistically-what-can-we-zippy-wave.md for the
// full plan. Visit logging (CrawlerVisit/ArchiveVisit) is pushed onto the
// same Redis queues Django's drain_crawler_queue/drain_archive_queue
// commands already consume, so the dashboard/analytics pipeline needs no
// changes.
package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"acpwb_go/archive"
	"acpwb_go/botclassify"
	"acpwb_go/policy"
	"acpwb_go/shell"
	"acpwb_go/visitqueue"
)

// loggingResponseWriter captures the status code and bytes written so
// accessLog can report them after the handler returns.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// accessLog prints one line per request, similar in spirit to Django's
// runserver log ("[06/Sep/2026 02:07:44] "GET /path/ HTTP/1.1" 200 117103"),
// so routing can be eyeballed the same way against docker compose logs.
func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lw, r)
		log.Printf("%s host=%s %q %d %d %s ua=%q",
			clientIP(r), r.Host, r.Method+" "+r.URL.RequestURI()+" "+r.Proto,
			lw.status, lw.bytes, time.Since(start).Round(time.Microsecond), r.Header.Get("User-Agent"))
	})
}

// archivePresentationsExtraHead matches _archive_content_shell.html's
// needs_presentations_css branch (apps/honeypot/pyrender/dispatch.py),
// injected only for the 'default' variant.
const archivePresentationsExtraHead = `<link rel="stylesheet" href="/static/css/presentations.css">
<style>.acpwb-archive-pres-sidebar .pres-card-title { font-size: 0.64rem; }</style>`

func archiveVariant(year, month, day int, slug string) string {
	sum := md5.Sum([]byte("variant_" + strconv.Itoa(year) + strconv.Itoa(month) + strconv.Itoa(day) + slug))
	n := new(big.Int).SetBytes(sum[:])
	mod := new(big.Int).Mod(n, big.NewInt(20)).Int64()
	switch {
	case mod < 3:
		return "compliance"
	case mod < 6:
		return "minutes"
	default:
		return "default"
	}
}

// parseArchivePath matches the main-domain archive_trap URL patterns in
// apps/honeypot/urls.py: /archive/<year>/<month>/<day>/ and
// /archive/<year>/<month>/<day>/<slug>/ (slug may itself contain slashes,
// trailing slash optional).
func parseArchivePath(path string) (year, month, day int, slug string, ok bool) {
	const prefix = "/archive/"
	if !strings.HasPrefix(path, prefix) {
		return 0, 0, 0, "", false
	}
	rest := strings.TrimSuffix(path[len(prefix):], "/")
	if rest == "" {
		return 0, 0, 0, "", false
	}
	parts := strings.SplitN(rest, "/", 4)
	if len(parts) < 3 {
		return 0, 0, 0, "", false
	}
	y, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	d, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, "", false
	}
	if len(parts) == 4 {
		slug = parts[3]
	}
	return y, m, d, slug, true
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func archiveHandler(vq *visitqueue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		year, month, day, slug, ok := parseArchivePath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		depth := 0
		if slug != "" {
			depth = strings.Count(slug, "/") + 1
		}
		ip := clientIP(r)
		ua := r.Header.Get("User-Agent")
		botType := botclassify.ClassifyUAOrIP(ua, ip)
		botGroup := botclassify.BotTypeToGroup(botType)

		go vq.PushArchiveVisit(ip, ua, year, month, day, depth, slug)
		go vq.PushCrawlerVisit(ip, ua, r.Host, r.URL.Path, r.Header.Get("Referer"), "archive", r.URL.RawQuery, botType, botGroup)

		variant := archiveVariant(year, month, day, slug)

		var contentHTML, ogTitle, extraHead string
		switch variant {
		case "compliance":
			c := archive.BuildComplianceContext(year, month, day, slug)
			contentHTML = archive.RenderComplianceDefault(&c)
			ogTitle = c.Title
		case "minutes":
			c := archive.BuildMinutesContext(year, month, day, slug)
			contentHTML = archive.RenderMinutesDefault(&c)
			ogTitle = c.Title
		default:
			c := archive.BuildContext(year, month, day, slug)
			contentHTML = archive.RenderArchiveDefault(&c)
			ogTitle = c.Title
			extraHead = archivePresentationsExtraHead
		}
		if ogTitle == "" {
			ogTitle = "ACPWB Archive"
		}

		page := shell.RenderPage(shell.PageParams{
			ContentHTML: contentHTML,
			OGTitle:     ogTitle,
			RequestPath: r.URL.RequestURI(),
			RemoteAddr:  ip,
			SiteRoot:    "",
			ExtraHead:   extraHead,
		})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}

// honeypotTokenFor mirrors apps/core/context_processors.py's
// honeypot_token: non-deterministic (folds in current time), so there is no
// byte-parity requirement — it just needs to look plausible, same exemption
// as shell.RenderPage's internally-generated token.
func honeypotTokenFor(requestPath, remoteAddr string) string {
	seed := requestPath + strconv.FormatInt(time.Now().UnixNano(), 10) + remoteAddr
	sum := md5.Sum([]byte(seed))
	return hex.EncodeToString(sum[:])[:8]
}

func mainPolicyYearURL(y int) string { return fmt.Sprintf("/public-policy/%d/", y) }
func mainPolicyMonthURL(y, m int) string {
	return fmt.Sprintf("/public-policy/%d/%02d/", y, m)
}

// parsePolicyPath matches the main-domain public-policy URL patterns in
// apps/honeypot/urls.py: /public-policy/, /public-policy/<year>/,
// /public-policy/<year>/<month>/, and
// /public-policy/<year>/<month>/<day>/<agency>/<slug>/ (trailing slash
// optional on all of them).
func parsePolicyPath(path string) (variant string, year, month, day int, agency, slug string, ok bool) {
	const prefix = "/public-policy/"
	if !strings.HasPrefix(path, prefix) {
		return "", 0, 0, 0, "", "", false
	}
	rest := strings.TrimSuffix(path[len(prefix):], "/")
	if rest == "" {
		return "index", 0, 0, 0, "", "", true
	}
	parts := strings.Split(rest, "/")
	switch len(parts) {
	case 1:
		y, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", 0, 0, 0, "", "", false
		}
		return "year", y, 0, 0, "", "", true
	case 2:
		y, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return "", 0, 0, 0, "", "", false
		}
		return "month", y, m, 0, "", "", true
	case 5:
		y, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		d, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			return "", 0, 0, 0, "", "", false
		}
		return "detail", y, m, d, parts[3], parts[4], true
	default:
		return "", 0, 0, 0, "", "", false
	}
}

func policyHandler(vq *visitqueue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		variant, year, month, day, agency, slug, ok := parsePolicyPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		ip := clientIP(r)
		ua := r.Header.Get("User-Agent")
		botType := botclassify.ClassifyUAOrIP(ua, ip)
		botGroup := botclassify.BotTypeToGroup(botType)
		go vq.PushCrawlerVisit(ip, ua, r.Host, r.URL.Path, r.Header.Get("Referer"), "policy", r.URL.RawQuery, botType, botGroup)

		meta := policy.PageMeta{
			HoneypotToken: honeypotTokenFor(r.URL.RequestURI(), ip),
			SiteRoot:      "",
			RequestPath:   r.URL.RequestURI(),
			NowYear:       time.Now().Year(),
		}

		var page string
		switch variant {
		case "index":
			page = policy.RenderPolicyIndex(meta, policy.GetPolicyIndexYears())
		case "year":
			yearData := policy.GetPolicyYearData(year)
			months := policy.GetPolicyYearMonths(year)
			page = policy.RenderPolicyYear(meta, year, yearData, months, policyYearsDesc(), year-1, year+1)
		case "month":
			prevMonth, prevYear := month-1, year
			if month == 1 {
				prevMonth, prevYear = 12, year-1
			}
			nextMonth, nextYear := month+1, year
			if month == 12 {
				nextMonth, nextYear = 1, year+1
			}
			entries := policy.GetPolicyMonthEntries(year, month)
			p := policy.MonthPageParams{
				Year: year, Month: month, Entries: entries, PolicyYears: policyYearsDesc(),
				PolicyIndexURL: "/public-policy/", YearURL: mainPolicyYearURL(year),
				PrevMonthURL: mainPolicyMonthURL(prevYear, prevMonth), NextMonthURL: mainPolicyMonthURL(nextYear, nextMonth),
				PolicyYearURLFn: mainPolicyYearURL,
			}
			page = policy.RenderPolicyMonth(meta, p)
		case "detail":
			doc := policy.GeneratePolicyDocument(year, month, day, agency, slug)
			related := policy.GenerateRelatedLinks(year, month, day, agency, slug, nil)
			relatedArchive := policy.GetCrossArchiveStubs(year, month, day, agency, slug)
			truncSlug := slug
			if len(truncSlug) > 32 {
				truncSlug = truncSlug[:32]
			}
			presSeed := fmt.Sprintf("policy_pres_%d_%d_%d_%s_%s", year, month, day, agency, truncSlug)
			relatedPres := archive.GeneratePresentationsForContext(presSeed, 4)
			p := policy.DetailParams{
				Doc: doc, Related: &related, RelatedArchive: relatedArchive, RelatedPresentations: relatedPres,
				PolicyYears: policyYearsDesc(), PolicyYearURL: mainPolicyYearURL, PolicyMonthURL: mainPolicyMonthURL,
			}
			page = policy.RenderPolicyDetail(meta, p)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}

func policyYearsDesc() []int {
	out := make([]int, 0, 2025-1992)
	for y := 2025; y >= 1993; y-- {
		out = append(out, y)
	}
	return out
}

// Subdomain host patterns, mirroring apps/core/subdomain_middleware.py.
var (
	archiveSubdomainRe = regexp.MustCompile(`^archives-(\d{4})\.acpwb\.(?:com|example)(?::\d+)?$`)
	policySubdomainRe  = regexp.MustCompile(`^policy-([a-z0-9][a-z0-9\-]*)\.acpwb\.(?:com|example)(?::\d+)?$`)
)

func subYearURL(y int) string     { return fmt.Sprintf("/%d/", y) }
func subMonthURL(y, m int) string { return fmt.Sprintf("/%d/%02d/", y, m) }

// eraArchiveVariant mirrors archiveVariant but the seed only ever needs
// year/month/day/slug — identical formula, kept separate for clarity at the
// call site (era pages route by host-derived year, not a path segment).
func eraArchiveVariant(year, month, day int, slug string) string {
	return archiveVariant(year, month, day, slug)
}

// parseEraPath matches archive_subdomain_urls.py's day-level trap patterns:
// /<month>/<day>/ and /<month>/<day>/<slug>/ (trailing slash optional).
// archive_subdomain_index (root "/") and archive_month (single "/<month>/")
// are NOT handled here — deferred, per the archive-era port's scope.
func parseEraPath(path string) (month, day int, slug string, ok bool) {
	rest := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/")
	if rest == "" {
		return 0, 0, "", false
	}
	parts := strings.SplitN(rest, "/", 3)
	if len(parts) < 2 {
		return 0, 0, "", false
	}
	m, err1 := strconv.Atoi(parts[0])
	d, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, "", false
	}
	if len(parts) == 3 {
		slug = parts[2]
	}
	return m, d, slug, true
}

func toEraYearData(yd archive.YearData) shell.EraYearData {
	return shell.EraYearData{
		Bg: yd.Bg, TextColor: yd.TextColor, Accent: yd.Accent, Accent2: yd.Accent2,
		FontBody: yd.FontBody, FontHead: yd.FontHead, LayoutClass: yd.LayoutClass,
	}
}

func eraHandler(vq *visitqueue.Queue, year int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		month, day, slug, ok := parseEraPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		depth := 0
		if slug != "" {
			depth = strings.Count(slug, "/") + 1
		}
		ip := clientIP(r)
		ua := r.Header.Get("User-Agent")
		botType := botclassify.ClassifyUAOrIP(ua, ip)
		botGroup := botclassify.BotTypeToGroup(botType)
		go vq.PushArchiveVisit(ip, ua, year, month, day, depth, slug)
		go vq.PushCrawlerVisit(ip, ua, r.Host, r.URL.Path, r.Header.Get("Referer"), "archive", r.URL.RawQuery, botType, botGroup)

		variant := eraArchiveVariant(year, month, day, slug)

		var eraContentHTML, title, titleSuffix, ogDescription string
		var yd archive.YearData
		var allYears []int
		switch variant {
		case "compliance":
			c := archive.BuildEraComplianceContext(year, month, day, slug)
			eraContentHTML = archive.RenderComplianceDefaultEra(&c)
			title, titleSuffix = c.Title, "ACPWB Compliance Archive"
			ogDescription = fmt.Sprintf(
				"%s sector compliance review archived %d-%02d-%02d. Audit ref %s. ACPWB Regulatory Practice.",
				c.Industry, year, month, day, c.AuditRef)
			yd, allYears = c.YearData, c.AllYears
		case "minutes":
			c := archive.BuildEraMinutesContext(year, month, day, slug)
			eraContentHTML = archive.RenderMinutesDefaultEra(&c)
			title, titleSuffix = c.Title, "ACPWB Archive"
			ogDescription = fmt.Sprintf(
				"%s meeting minutes archived %d-%02d-%02d. Meeting ref %s. ACPWB Institutional Records.",
				c.Committee, year, month, day, c.MeetingRef)
			yd, allYears = c.YearData, c.AllYears
		default:
			c := archive.BuildEraContext(year, month, day, slug)
			eraContentHTML = archive.RenderArchiveDefaultEra(&c)
			title, titleSuffix = c.Title, "ACPWB Archive"
			ogDescription = fmt.Sprintf(
				"%s sector engagement documentation archived %d-%02d-%02d. %s phase record. ACPWB Research Division.",
				c.Industry, year, month, day, pyCapitalizeLocal(c.Phase))
			yd, allYears = c.YearData, c.AllYears
		}

		page := shell.RenderEraPage(shell.EraPageParams{
			EraContentHTML: eraContentHTML,
			Title:          title,
			TitleSuffix:    titleSuffix,
			OGDescription:  ogDescription,
			RequestPath:    r.URL.RequestURI(),
			RemoteAddr:     ip,
			Year:           year,
			AllYears:       allYears,
			YearData:       toEraYearData(yd),
		})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}

func pyCapitalizeLocal(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// parsePolicySubdomainPath matches policy_subdomain_urls.py: "", "<year>/",
// "<year>/<month>/", "<year>/<month>/<day>/<slug>/" (agency is baked into
// the host, so — unlike the main-domain detail route — there's no agency
// path segment here).
func parsePolicySubdomainPath(path string) (variant string, year, month, day int, slug string, ok bool) {
	rest := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/")
	if rest == "" {
		return "index", 0, 0, 0, "", true
	}
	parts := strings.Split(rest, "/")
	switch len(parts) {
	case 1:
		y, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", 0, 0, 0, "", false
		}
		return "year", y, 0, 0, "", true
	case 2:
		y, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return "", 0, 0, 0, "", false
		}
		return "month", y, m, 0, "", true
	case 4:
		y, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		d, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			return "", 0, 0, 0, "", false
		}
		return "detail", y, m, d, parts[3], true
	default:
		return "", 0, 0, 0, "", false
	}
}

func policySubdomainHandler(vq *visitqueue.Queue, agency string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		variant, year, month, day, slug, ok := parsePolicySubdomainPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		full, domain, agencyOK := policy.AgencyData(agency)
		if !agencyOK {
			full, domain = "Unknown Agency", "regulatory policy"
		}

		ip := clientIP(r)
		ua := r.Header.Get("User-Agent")
		botType := botclassify.ClassifyUAOrIP(ua, ip)
		botGroup := botclassify.BotTypeToGroup(botType)
		go vq.PushCrawlerVisit(ip, ua, r.Host, r.URL.Path, r.Header.Get("Referer"), "policy", r.URL.RawQuery, botType, botGroup)

		meta := policy.PageMeta{
			HoneypotToken: honeypotTokenFor(r.URL.RequestURI(), ip),
			SiteRoot:      "https://acpwb.com",
			RequestPath:   r.URL.RequestURI(),
			NowYear:       time.Now().Year(),
		}

		// Cross-subdomain URL builder used by the detail route: same-agency
		// links stay relative to this subdomain, cross-agency links go
		// absolute to the other agency's subdomain — mirrors
		// apps/honeypot/views.py's subdomain-aware _archive_url() sibling
		// for policy, ported in the detail-subdomain fixture tests.
		urlFn := policy.URLFunc(func(y, m, d int, ag, sl string) string {
			if ag == agency {
				return fmt.Sprintf("/%d/%02d/%02d/%s/", y, m, d, sl)
			}
			return fmt.Sprintf("https://policy-%s.acpwb.com/%d/%02d/%02d/%s/", ag, y, m, d, sl)
		})

		var page string
		switch variant {
		case "index":
			p := policy.SubdomainIndexParams{
				Agency: agency, AgencyFull: full, PolicyDomain: domain,
				Years:         policy.GetPolicyAgencyYears(agency),
				OGTitle:       fmt.Sprintf("%s Policy Filings — ACPWB", strings.ToUpper(agency)),
				OGDescription: fmt.Sprintf("ACPWB regulatory filings, comment letters, and testimony submitted to the %s.", full),
				PolicyYearURL: subYearURL, PolicyMonthURL: subMonthURL,
			}
			page = policy.RenderPolicySubdomainIndex(meta, p)
		case "year":
			p := policy.SubdomainYearParams{
				Agency: agency, AgencyFull: full, PolicyDomain: domain, Year: year,
				YearDetail: policy.GetPolicyAgencyYearDetail(agency, year),
				AllYears:   policy.GetPolicyAgencyYears(agency),
				PrevYear:   year - 1, NextYear: year + 1,
				OGTitle:        fmt.Sprintf("%d %s Policy Filings — ACPWB", year, strings.ToUpper(agency)),
				PolicyIndexURL: "/",
				PolicyYearURL:  subYearURL, PolicyMonthURL: subMonthURL,
			}
			page = policy.RenderPolicySubdomainYear(meta, p)
		case "month":
			prevMonth, prevYear := month-1, year
			if month == 1 {
				prevMonth, prevYear = 12, year-1
			}
			nextMonth, nextYear := month+1, year
			if month == 12 {
				nextMonth, nextYear = 1, year+1
			}
			entries := policy.GetPolicyAgencyMonthEntries(agency, year, month, urlFn)
			p := policy.MonthPageParams{
				Year: year, Month: month, Entries: entries, PolicyYears: policyYearsDesc(),
				PolicyIndexURL: "/", YearURL: subYearURL(year),
				PrevMonthURL: subMonthURL(prevYear, prevMonth), NextMonthURL: subMonthURL(nextYear, nextMonth),
				PolicyYearURLFn: subYearURL,
			}
			page = policy.RenderPolicyMonth(meta, p)
		case "detail":
			doc := policy.GeneratePolicyDocument(year, month, day, agency, slug)
			doc.URL = urlFn(year, month, day, agency, slug)
			related := policy.GenerateRelatedLinks(year, month, day, agency, slug, urlFn)
			relatedArchive := policy.GetCrossArchiveStubs(year, month, day, agency, slug)
			p := policy.DetailParams{
				Doc: doc, Related: &related, RelatedArchive: relatedArchive, RelatedPresentations: nil,
				PolicyYears: policyYearsDesc(), PolicyYearURL: subYearURL, PolicyMonthURL: subMonthURL,
			}
			page = policy.RenderPolicyDetail(meta, p)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}

// rootHandler dispatches on Host first (archive/policy subdomains, mirroring
// apps/core/subdomain_middleware.py), then falls back to path-based routing
// for the main domain. Subdomain handlers are built once per matched host
// value via a small cache, rather than per request, since eraHandler/
// policySubdomainHandler close over an immutable year/agency.
func rootHandler(vq *visitqueue.Queue) http.HandlerFunc {
	mainArchive := archiveHandler(vq)
	mainPolicy := policyHandler(vq)

	return func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(r.Host)
		if i := strings.IndexByte(host, ':'); i >= 0 {
			host = host[:i]
		}

		if m := archiveSubdomainRe.FindStringSubmatch(host); m != nil {
			year, _ := strconv.Atoi(m[1])
			eraHandler(vq, year)(w, r)
			return
		}
		if m := policySubdomainRe.FindStringSubmatch(host); m != nil {
			agency := strings.ToLower(m[1])
			if _, _, ok := policy.AgencyData(agency); ok {
				policySubdomainHandler(vq, agency)(w, r)
				return
			}
		}

		switch {
		case strings.HasPrefix(r.URL.Path, "/archive/"):
			mainArchive(w, r)
		case strings.HasPrefix(r.URL.Path, "/public-policy/"):
			mainPolicy(w, r)
		default:
			http.NotFound(w, r)
		}
	}
}

func main() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	vq, err := visitqueue.New(redisURL)
	if err != nil {
		log.Fatalf("acpwb_go: %v", err)
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8090"
	}

	root := rootHandler(vq)

	var handler http.Handler = root
	if os.Getenv("ACCESS_LOG") == "1" {
		// Local-dev only (set by docker-compose-local.yml) — per-request
		// access logging, so routing can be eyeballed the same way as
		// Django's runserver log. Left off by default/in production: it's
		// a per-request log line on the exact hot path this service exists
		// to keep cheap.
		handler = accessLog(root)
	}

	log.Printf("acpwb_go listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

package classify

import (
	"bufio"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const (
	chromeUA  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	firefoxUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:128.0) Gecko/20100101 Firefox/128.0"
	safariUA  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15"
	iosUA     = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/131.0.6778.73 Mobile/15E148 Safari/604.1"

	chromeH2  = "1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p"
	firefoxH2 = "1:65536;2:0;4:131072;5:16384|12517377|0|m,p,a,s"
	safariH2  = "2:0;3:100;4:2097152;9:1|10420225|0|m,s,a,p"
)

var (
	residential = NetInfo{ASN: 28573, Org: "Claro NXT Telecomunicacoes Ltda", Network: NetResidential}
	hosting     = NetInfo{ASN: 45102, Org: "Alibaba (US) Technology Co., Ltd.", Network: NetHosting}
)

// The cases come from production traffic (Oct 2026) and the uTLS hello
// fixtures in tlsfp/testdata/hellos.
func TestClassify(t *testing.T) {
	cases := []struct {
		name  string
		in    Input
		class string
		label string
		group string
		sig   uint32 // bits that must be set
	}{
		{"real Chrome", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2, Net: residential},
			ClassBrowser, "Browser · Chrome", GroupBrowsers, 0},
		{"Chrome in a datacenter", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2, Net: hosting},
			ClassHeadless, "Datacenter · headless Chrome", GroupDatacenterBots, SigHosting},
		{"unverifiable Safari in a datacenter", Input{UA: safariUA, AcceptLanguage: "en-US", JA4: "t13d2114h2_0123456789ab_000000000000", H2Fingerprint: safariH2, Net: hosting},
			ClassHeadless, "Datacenter · headless Safari", GroupDatacenterBots, SigHosting},
		{"real Firefox", Input{UA: firefoxUA, AcceptLanguage: "en-US", JA4: "t13d1717h2_5b57614c22b0_3cbfd9057e0d", H2Fingerprint: firefoxH2, Net: residential},
			ClassBrowser, "Browser · Firefox", GroupBrowsers, 0},
		{"real Safari", Input{UA: safariUA, AcceptLanguage: "en-US", JA4: "t13d2014h2_a09f3c656075_14788d8d241b", H2Fingerprint: safariH2, Net: residential},
			ClassBrowser, "Browser · Safari", GroupBrowsers, 0},
		{"Chrome on iOS is Apple TLS", Input{UA: iosUA, AcceptLanguage: "en-US", JA4: "t13d2013h2_a09f3c656075_7f0f34a4126d", H2Fingerprint: safariH2, Net: residential},
			ClassBrowser, "Browser · Chrome", GroupBrowsers, 0},
		// Go's TLS stack claiming Chrome, from Tencent.
		{"Go claiming Chrome", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d131000_f57a46bbacb6_e7c285222651", Net: hosting},
			ClassImpersonator, "Datacenter · fake Chrome (Go TLS)", GroupDatacenterBots, SigTLSMismatch | SigNoH2 | SigHosting},
		// The rotating pool: an unknown stack, one request per residential IP.
		{"unknown stack claiming Chrome", Input{UA: chromeUA, JA4: "t13d311200_1d947a95fc68_1aa0e0f8a5b3", Net: residential},
			ClassImpersonator, "Residential · fake Chrome (t13d311200)", GroupResidentialBots, SigTLSMismatch | SigNoH2 | SigNoAcceptLang},
		// Chromium TLS but HTTP/1.1 only: the slow-handshake crawler.
		{"Chromium TLS without h2", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1515h1_8daaf6152771_de4a06bb82e3", Net: residential},
			ClassImpersonator, "Residential · fake Chrome (Chromium, no h2)", GroupResidentialBots, SigNoH2},
		// Chrome's JA4 with somebody else's HTTP/2 (10.7k IPs, one conn each).
		{"Chrome JA4, foreign HTTP/2", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_d8a2da3f94cd",
			H2Fingerprint: "2:0;4:4194304;5:16384;6:10485760|1073741824|0|a,m,p,s", Net: residential},
			ClassImpersonator, "Residential · fake Chrome (Chromium, foreign HTTP/2)", GroupResidentialBots, SigH2Mismatch},
		{"Chromium claiming Firefox", Input{UA: firefoxUA, AcceptLanguage: "en-US", JA4: "t13d1513h2_8daaf6152771_eca864cca44a", H2Fingerprint: chromeH2, Net: residential},
			ClassImpersonator, "Residential · fake Firefox (Chromium)", GroupResidentialBots, SigTLSMismatch | SigH2Mismatch},
		{"OkHttp claiming Chrome", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t12d120700_d34a8e72043a_036209cd1ead"},
			ClassImpersonator, "fake Chrome (OkHttp)", GroupFakeBrowsers, SigTLSMismatch | SigNoH2},
		{"Chrome 131 with Chrome 105 HTTP/2", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862",
			H2Fingerprint: "1:65536;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p", Net: residential},
			ClassBrowser, "Browser · Chrome", GroupBrowsers, SigH2Stale},
		{"unknown hash claiming Safari is unverified", Input{UA: safariUA, AcceptLanguage: "en-US", JA4: "t13d2114h2_0123456789ab_000000000000", H2Fingerprint: safariH2, Net: residential},
			ClassUnknown, "Unverified · Safari", GroupUnverified, 0},
		{"GPTBot", Input{UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)", JA4: "t13d1516h2_8daaf6152771_02713d6af862", Net: hosting},
			ClassKnownBot, "OpenAI GPTBot", "AI Crawlers", SigHosting},
		{"Googlebot from a home ISP", Input{UA: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", JA4: "t13d311200_1d947a95fc68_1aa0e0f8a5b3", Net: residential},
			ClassKnownBot, "Googlebot", "Search Engines", SigBotResidential},
		{"curl", Input{UA: "curl/8.7.1", JA4: "t13d3112h2_e8f1e7e78f70_6bebaf5329ec"},
			ClassTool, "cURL", "Generic Scrapers", 0},
		{"known scraper range", Input{UA: chromeUA, IP: "216.75.132.10", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2},
			ClassKnownBot, "INTERNEXUS Scraper Pool", "SEO / Other Bots", SigKnownRange},
		{"empty UA", Input{JA4: "t13d131000_f57a46bbacb6_e7c285222651"},
			ClassUnknown, "(empty user agent)", GroupUnverified, 0},
		{"no browser claim", Input{UA: "Mozilla/1.22 (compatible; MSIE 10.0; Windows 3.1)", JA4: "t13d131000_f57a46bbacb6_e7c285222651"},
			ClassUnknown, "Other · Go TLS", GroupUnverified, 0},
		{"plain HTTP can't be checked", Input{UA: chromeUA},
			ClassUnknown, "Unverified · Chrome", GroupUnverified, 0},
		// Real Chrome 155 through the edge (headless, same network stack).
		{"Chrome navigation header order", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: h(":method,:authority,:scheme,:path,sec-ch-ua,sec-ch-ua-mobile,sec-ch-ua-platform,upgrade-insecure-requests,user-agent,accept,sec-fetch-site,sec-fetch-mode,sec-fetch-user,sec-fetch-dest,accept-encoding,accept-language,cookie,priority"), Net: residential},
			ClassBrowser, "Browser · Chrome", GroupBrowsers, 0},
		{"Chrome subresource header order", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: h(":method,:authority,:scheme,:path,origin,sec-ch-ua-platform,user-agent,sec-ch-ua,sec-ch-ua-mobile,accept,sec-fetch-site,sec-fetch-mode,sec-fetch-dest,referer,accept-encoding,accept-language,cookie,cookie,priority"), Net: residential},
			ClassBrowser, "Browser · Chrome", GroupBrowsers, 0},
		// 92% of production's residential "Chrome": accept-language injected early.
		{"injected accept-language", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: h(":method,:authority,:scheme,:path,sec-ch-ua-platform,user-agent,sec-ch-ua,accept-language,sec-ch-ua-mobile,accept,sec-fetch-site,sec-fetch-mode,sec-fetch-dest,referer,accept-encoding,cookie,priority"), Net: residential},
			ClassImpersonator, "Residential · fake Chrome (Chromium, injected headers)", GroupResidentialBots, SigHeaderOrder},
		{"accept-language before accept", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: h(":method,:authority,:scheme,:path,sec-ch-ua,sec-ch-ua-mobile,sec-ch-ua-platform,upgrade-insecure-requests,user-agent,accept-language,accept,sec-fetch-site,sec-fetch-mode,sec-fetch-user,sec-fetch-dest,accept-encoding,priority"), Net: hosting},
			ClassImpersonator, "Datacenter · fake Chrome (Chromium, injected headers)", GroupDatacenterBots, SigHeaderOrder | SigHosting},
		{"unlisted crawler", Input{UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; KeenableBot/1.0; +https://keenable.ai/bot)", JA4: "t13d1516h2_8daaf6152771_02713d6af862", Net: hosting},
			ClassKnownBot, "KeenableBot", "SEO / Other Bots", SigHosting},
		{"unlisted crawler, compatible outside parentheses", Input{UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ShapBot/0.1.0", JA4: "t13d1516h2_8daaf6152771_02713d6af862", Net: hosting},
			ClassKnownBot, "ShapBot", "SEO / Other Bots", 0},
		{"user agent fetcher", Input{UA: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Claude-User/1.0; +Claude-User@anthropic.com)", JA4: "t13d1516h2_8daaf6152771_02713d6af862", Net: hosting},
			ClassKnownBot, "Claude-User", "SEO / Other Bots", 0},
		{"scanner greeting", Input{UA: "Hello from Palo Alto Networks, find out more about our scans in https://docs-cortex.paloaltonetworks.com/r/1/Cortex-Xpanse/Scanning-activity", JA4: "t13d131000_f57a46bbacb6_e7c285222651", Net: hosting},
			ClassKnownBot, "Hello from Palo Alto Networks", "SEO / Other Bots", 0},
		{"non-browser client", Input{UA: "Cpanel-HTTP-Client/1.0", JA4: "t13d131000_f57a46bbacb6_e7c285222651", Net: hosting},
			ClassTool, "Cpanel-HTTP-Client", "Generic Scrapers", 0},
		{"Next.js exploit probe", Input{UA: "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)", JA4: "t13d1516h2_8daaf6152771_02713d6af862",
			HeaderNames: h(":authority,:method,:path,:scheme,x-middleware-subrequest,x-nextjs-data,user-agent,accept,cookie,accept-encoding"), Net: hosting},
			ClassKnownBot, "Baiduspider", "Search Engines", SigExploitProbe},
		{"plain-HTTP browser claim from a datacenter", Input{UA: chromeUA, Net: hosting},
			ClassHeadless, "Datacenter · headless Chrome", GroupDatacenterBots, SigHosting},
		{"cookie dropped is evidence only", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2, CookieDropped: true},
			ClassBrowser, "Browser · Chrome", GroupBrowsers, SigCookieDropped},
		{"Chrome's client hints", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: chromeNav, Hints: chromeHints, Net: residential},
			ClassBrowser, "Browser · Chrome", GroupBrowsers, 0},
		// A real Chrome 155 whose UA was swapped for an older one.
		{"client hints contradict the UA", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: chromeNav, Hints: ClientHints{UA: `"Chromium";v="155", "Not/A)Brand";v="24", "Google Chrome";v="155"`, Mobile: "?0", Platform: `"Windows"`}, Net: residential},
			ClassImpersonator, "Residential · fake Chrome (Chromium, client hints)", GroupResidentialBots, SigClientHints},
		{"no client hints", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: h(":method,:authority,:scheme,:path,upgrade-insecure-requests,user-agent,accept,sec-fetch-site,sec-fetch-mode,sec-fetch-user,sec-fetch-dest,accept-encoding,accept-language,priority"), Net: residential},
			ClassImpersonator, "Residential · fake Chrome (Chromium, client hints)", GroupResidentialBots, SigClientHints},
		{"HeadlessChrome brand", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			HeaderNames: chromeNav, Hints: ClientHints{UA: `"Chromium";v="131", "Not_A Brand";v="24", "HeadlessChrome";v="131"`, Mobile: "?0", Platform: `"Windows"`}, Net: residential},
			ClassHeadless, "Residential · headless Chrome", GroupResidentialBots, 0},
		// Real Chrome on a home network clicking a ghost link.
		{"hidden link", Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2,
			Path: "/employees/export/", HeaderNames: chromeNav, Hints: chromeHints, Net: residential},
			ClassHeadless, "Residential · automated Chrome (hidden link)", GroupResidentialBots, SigHiddenLink},
		{"hidden link, fake Chrome", Input{UA: chromeUA, JA4: "t13d311200_1d947a95fc68_1aa0e0f8a5b3", Path: "/internal/portal/", Net: residential},
			ClassImpersonator, "Residential · fake Chrome (t13d311200)", GroupResidentialBots, SigHiddenLink | SigTLSMismatch},
		{"Safari 2 from a proxy pool", Input{UA: "Mozilla/5.0 (Macintosh; U; PPC Mac OS X; de-de) AppleWebKit/416.12 (KHTML, like Gecko) Safari/416.13", JA4: "t13d311200_1d947a95fc68_1aa0e0f8a5b3", Net: residential},
			ClassImpersonator, "Residential · fake Safari (t13d311200, no h2)", GroupResidentialBots, SigNoH2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := Classify(c.in)
			if r.Class != c.class || r.Label != c.label || r.Group != c.group || r.Signals&c.sig != c.sig {
				t.Errorf("got class=%q label=%q group=%q signals=%b; want %q %q %q ⊇%b",
					r.Class, r.Label, r.Group, r.Signals, c.class, c.label, c.group, c.sig)
			}
		})
	}
}

// Every uTLS browser profile, with a UA for that browser, is a browser.
func TestUTLSBrowserProfiles(t *testing.T) {
	names := tsvColumn(t, "../tlsfp/testdata/hellos/index.tsv", 1)
	ja4s := tsvColumn(t, "../tlsfp/testdata/hellos/golden.tsv", 1)
	uas := map[string]string{"chrome": chromeUA, "edge": chromeUA, "firefox": firefoxUA, "safari": safariUA, "ios": iosUA}
	for i, name := range names {
		family, _, _ := strings.Cut(name, "_")
		ua, ok := uas[family]
		if !ok {
			continue // okhttp, golang, qq, 360: not a UA we'd verify
		}
		r := Classify(Input{UA: ua, AcceptLanguage: "en-US", JA4: ja4s[i], Net: residential})
		if r.Class != ClassBrowser {
			t.Errorf("%s (%s): %s %q signals=%b", name, ja4s[i], r.Class, r.Label, r.Signals)
		}
	}
}

func tsvColumn(t *testing.T, path string, col int) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		out = append(out, strings.Split(sc.Text(), "\t")[col])
	}
	return out
}

func h(s string) []string { return strings.Split(s, ",") }

var (
	chromeNav   = h(":method,:authority,:scheme,:path,sec-ch-ua,sec-ch-ua-mobile,sec-ch-ua-platform,upgrade-insecure-requests,user-agent,accept,sec-fetch-site,sec-fetch-mode,sec-fetch-user,sec-fetch-dest,accept-encoding,accept-language,priority")
	chromeHints = ClientHints{UA: `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`, Mobile: "?0", Platform: `"Windows"`}
)

func TestHintsMismatch(t *testing.T) {
	const (
		edge    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"
		opera   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 OPR/120.0.0.0"
		android = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36"
		tablet  = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
		mac     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
		linux   = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	)
	ch := func(brands, mobile, platform string) ClientHints {
		return ClientHints{UA: brands, Mobile: mobile, Platform: `"` + platform + `"`}
	}
	c131 := `"Chromium";v="131", "Not_A Brand";v="24", "Google Chrome";v="131"`
	for _, c := range []struct {
		name               string
		ua                 string
		hints              ClientHints
		mismatch, headless bool
	}{
		{"Chrome", chromeUA, chromeHints, false, false},
		{"Edge", edge, ch(`"Microsoft Edge";v="131", "Chromium";v="131", "Not_A Brand";v="24"`, "?0", "Windows"), false, false},
		{"Opera: Chromium's version, not Opera's", opera, ch(`"Opera";v="120", "Chromium";v="135", "Not-A.Brand";v="99"`, "?0", "Windows"), false, false},
		{"Android phone", android, ch(c131, "?1", "Android"), false, false},
		{"Android tablet", tablet, ch(c131, "?0", "Android"), false, false},
		{"desktop-site mode", linux, ch(c131, "?0", "Android"), false, false},
		{"Mac", mac, ch(c131, "?0", "macOS"), false, false},
		{"other version", chromeUA, ch(`"Chromium";v="155", "Google Chrome";v="155"`, "?0", "Windows"), true, false},
		{"no Chromium brand", chromeUA, ch(`"Google Chrome";v="131"`, "?0", "Windows"), true, false},
		{"desktop UA, mobile hint", chromeUA, ch(c131, "?1", "Windows"), true, false},
		{"phone UA, desktop hint", android, ch(c131, "?0", "Android"), true, false},
		{"Windows UA, macOS hint", chromeUA, ch(c131, "?0", "macOS"), true, false},
		{"headless", chromeUA, ch(`"HeadlessChrome";v="131", "Chromium";v="131"`, "?0", "Windows"), false, true},
	} {
		m, hl := hintsMismatch(c.ua, c.hints)
		if m != c.mismatch || hl != c.headless {
			t.Errorf("%s: mismatch=%v headless=%v, want %v %v", c.name, m, hl, c.mismatch, c.headless)
		}
	}
}

func TestClaimBrowser(t *testing.T) {
	for ua, want := range map[string]Claim{
		chromeUA:  {"Chrome", "chromium", 131},
		firefoxUA: {"Firefox", "firefox", 128},
		safariUA:  {"Safari", "apple", 18},
		iosUA:     {"Chrome", "apple", 131},
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0": {"Edge", "chromium", 131},
		"Mozilla/5.0 (Android 14; Mobile; rv:131.0) Gecko/131.0 Firefox/131.0":                                                          {"Firefox", "firefox", 131},
		"curl/8.7.1": {},
		"Mozilla/1.22 (compatible; MSIE 10.0; Windows 3.1)": {},
	} {
		if got := claimBrowser(ua); got != want {
			t.Errorf("%s: %+v, want %+v", ua, got, want)
		}
	}
}

// hostingOrgKeywords must stay the same list as the Python classifier's.
func TestHostingKeywordsMatchPython(t *testing.T) {
	src, err := os.ReadFile("../../acpwb/apps/core/ip_intel_classify.py")
	if err != nil {
		t.Skip("Django app not checked out alongside:", err)
	}
	block := regexp.MustCompile(`(?s)_HOSTING_ORG_KEYWORDS = \[(.*?)\]`).FindSubmatch(src)
	if block == nil {
		t.Fatal("_HOSTING_ORG_KEYWORDS not found")
	}
	var py []string
	for _, m := range regexp.MustCompile(`'([^']*)'`).FindAllSubmatch(block[1], -1) {
		py = append(py, string(m[1]))
	}
	if !slices.Equal(py, hostingOrgKeywords) {
		t.Errorf("hosting keywords differ from apps/core/ip_intel_classify.py:\npython %q\ngo     %q", py, hostingOrgKeywords)
	}
}

func TestASNDBWithoutDatabase(t *testing.T) {
	var db ASNDB
	if n := db.Lookup("8.8.8.8"); n.Network != NetUnknown {
		t.Errorf("%+v", n)
	}
	var nilDB *ASNDB
	if n := nilDB.Lookup("8.8.8.8"); n.Network != NetUnknown {
		t.Errorf("%+v", n)
	}
	if err := db.Load("/nonexistent.mmdb"); err == nil {
		t.Error("loading a missing file should fail")
	}
}

// With the real GeoLite2-ASN database, if it's here (local dev has it
// after download_geoip_db).
func TestASNDBLookup(t *testing.T) {
	var db ASNDB
	if err := db.Load("../../acpwb/var/geoip/GeoLite2-ASN.mmdb"); err != nil {
		t.Skip("no GeoLite2-ASN.mmdb:", err)
	}
	if n := db.Lookup("8.8.8.8"); n.ASN != 15169 || n.Network != NetHosting {
		t.Errorf("8.8.8.8: %+v", n)
	}
	if n := db.Lookup("10.0.0.1"); n.Network != NetUnknown {
		t.Errorf("private: %+v", n)
	}
}

// apps/core/traffic_signals.py names the bits for the dashboard and live
// view; it must list them in the same order.
func TestSignalNamesMatchPython(t *testing.T) {
	src, err := os.ReadFile("../../acpwb/apps/core/traffic_signals.py")
	if err != nil {
		t.Skip("Django app not checked out alongside:", err)
	}
	block := regexp.MustCompile(`(?s)SIGNALS = \[(.*?)\n\]`).FindSubmatch(src)
	if block == nil {
		t.Fatal("SIGNALS not found")
	}
	var py []string
	for _, m := range regexp.MustCompile(`\('([a-z0-9_]+)',`).FindAllSubmatch(block[1], -1) {
		py = append(py, string(m[1]))
	}
	want := []struct {
		bit  uint32
		name string
	}{
		{SigTLSMismatch, "tls_mismatch"}, {SigNoH2, "no_h2"}, {SigH2Mismatch, "h2_mismatch"},
		{SigH2Stale, "h2_stale"}, {SigNoAcceptLang, "no_accept_language"}, {SigCookieDropped, "cookie_dropped"},
		{SigHosting, "hosting"}, {SigOldTLS, "old_tls"}, {SigBotResidential, "bot_residential"},
		{SigKnownRange, "known_range"}, {SigHeaderOrder, "header_order"}, {SigExploitProbe, "exploit_probe"},
		{SigClientHints, "client_hints"}, {SigHiddenLink, "hidden_link"},
	}
	if len(py) != len(want) {
		t.Fatalf("python has %d signals, go %d: %q", len(py), len(want), py)
	}
	for i, w := range want {
		if w.bit != 1<<i || py[i] != w.name {
			t.Errorf("bit %d: go %q (%b), python %q", i, w.name, w.bit, py[i])
		}
	}
}

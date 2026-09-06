// Package botclassify ports apps/core/bot_classify.py so acpwb_go can
// classify bot_type/bot_group at write time, instead of leaving CrawlerVisit
// rows blank for a cron backfill to fill in later (see deploy/acpwb-crontab's
// backfill_bot_types incident note for why that approach was abandoned).
package botclassify

import (
	"net"
	"strings"
)

// ipBotRange mirrors _IP_BOT_RANGE_DEFS — checked in order, first match wins.
type ipBotRange struct {
	network *net.IPNet
	name    string
}

var ipBotRangeDefs = []struct {
	cidr string
	name string
}{
	// Alibaba / Qwen AI crawler
	{"47.79.0.0/16", "Alibaba Qwen"},  // observed traffic
	{"47.82.60.0/22", "Alibaba Qwen"}, // 47.82.60-63
	{"8.219.0.0/16", "Alibaba Qwen"},  // observed traffic
	{"43.156.0.0/16", "Tencent"},
	{"43.172.0.0/16", "Tencent"},
	{"43.173.0.0/16", "Tencent"},
	// INTERNEXUS, LLC scraper pool (uniform Chrome/Mac UA, throttled per-IP)
	{"207.180.11.0/24", "INTERNEXUS Scraper Pool"},
	{"216.75.132.0/24", "INTERNEXUS Scraper Pool"},
	// IPXO-leased blocks, single HK customer (ORG-PC1271-RIPE), same fingerprint
	{"143.20.253.0/24", "IPXO Scraper Pool (HK)"},
	{"143.14.6.0/24", "IPXO Scraper Pool (HK)"},
	{"144.31.35.0/24", "IPXO Scraper Pool (HK)"},
}

// Parsed once at package init, mirroring the module-load-time parse in bot_classify.py.
var ipBotRanges = func() []ipBotRange {
	ranges := make([]ipBotRange, 0, len(ipBotRangeDefs))
	for _, def := range ipBotRangeDefs {
		_, network, err := net.ParseCIDR(def.cidr)
		if err != nil {
			panic("botclassify: invalid CIDR " + def.cidr + ": " + err.Error())
		}
		ranges = append(ranges, ipBotRange{network: network, name: def.name})
	}
	return ranges
}()

// ClassifyIP returns a bot name if ip falls within a known bot IP range, else "".
// Mirrors classify_ip()'s None-for-no-match contract (empty string here).
func ClassifyIP(ipStr string) string {
	addr := net.ParseIP(strings.TrimSpace(ipStr))
	if addr == nil {
		return ""
	}
	for _, r := range ipBotRanges {
		if r.network.Contains(addr) {
			return r.name
		}
	}
	return ""
}

// botPatternDefs mirrors BOT_PATTERNS exactly, including order — matching is
// first-substring-wins, and at least one entry (Googlebot Mobile) depends on
// being checked before a later, broader entry (Googlebot).
var botPatternDefs = []struct {
	pattern string
	name    string
}{
	// AI crawlers — most interesting
	{"GPTBot", "OpenAI GPTBot"},
	{"OAI-SearchBot", "OpenAI SearchBot"},
	{"ChatGPT-User", "OpenAI ChatGPT"},
	{"ClaudeBot", "Anthropic ClaudeBot"},
	{"Claude-Web", "Anthropic Claude"},
	{"anthropic-ai", "Anthropic"},
	{"PerplexityBot", "Perplexity"},
	{"Google-Extended", "Google-Extended (AI)"},
	{"meta-externalagent", "Meta ExternalAgent"},
	{"FacebookBot", "Meta FacebookBot"},
	{"Applebot-Extended", "Apple Applebot-Extended"},
	{"Bytespider", "ByteDance Bytespider"},
	{"Amazonbot", "Amazonbot"},
	{"Diffbot", "Diffbot"},
	{"omgili", "Omgilibot"},
	{"webzio-extended", "Webzio"},
	{"CCBot", "Common Crawl"},
	{"cohere-ai", "Cohere"},
	{"Timpibot", "Timpi"},
	// Search engines
	{"Nexus 5X Build/MMB29P", "Googlebot Mobile"}, // Googlebot mobile fingerprint (before Googlebot)
	{"Googlebot", "Googlebot"},
	{"GoogleOther", "GoogleOther"},
	{"bingbot", "Bingbot"},
	{"BingPreview", "Bing Preview"},
	{"msnbot", "MSN Bot"},
	{"Baiduspider", "Baiduspider"},
	{"YandexBot", "YandexBot"},
	{"Slurp", "Yahoo Slurp"},
	{"DuckDuckBot", "DuckDuckBot"},
	{"Applebot", "Applebot"},
	{"sogou", "Sogou"},
	{"360Spider", "360Spider"},
	{"SeznamBot", "Seznam"},
	// SEO/marketing crawlers
	{"SemrushBot", "SemrushBot"},
	{"AhrefsBot", "AhrefsBot"},
	{"MJ12bot", "Majestic MJ12"},
	{"DotBot", "DotBot"},
	{"DataForSeoBot", "DataForSEO"},
	{"PetalBot", "Huawei PetalBot"},
	{"PiplBot", "Pipl"},
	{"SERankingBacklinksBot", "SERankingBacklinksBot"},
	{"ZoominfoBot", "ZoominfoBot"},
	{"AwarioBot", "AwarioBot"},
	{"BitSightBot", "BitSightBot"},
	{"zgrab", "zgrab"},
	// Social
	{"Twitterbot", "Twitterbot"},
	{"facebookexternalhit", "Facebook Scraper"},
	{"LinkedInBot", "LinkedIn"},
	// Archives
	{"ia_archiver", "Internet Archive"},
	{"archive.org_bot", "Internet Archive"},
	// Generic HTTP clients (likely scrapers/bots)
	{"python-requests", "Python Requests"},
	{"curl/", "cURL"},
	{"wget", "Wget"},
	{"scrapy", "Scrapy"},
	{"Go-http-client", "Go HTTP Client"},
	{"Java/", "Java HTTP Client"},
	{"libwww-perl", "libwww-perl"},
	{"axios", "axios"},
	{"node-fetch", "node-fetch"},
	{"okhttp", "OkHttp"},
	{"httpx", "httpx"},
	{"aiohttp", "aiohttp"},
	{"Faraday", "Faraday (Ruby)"},
}

type botPattern struct {
	patternLower string
	name         string
}

// Pre-lowered once at init, mirroring _BOT_PATTERNS_LOWER.
var botPatternsLower = func() []botPattern {
	patterns := make([]botPattern, 0, len(botPatternDefs))
	for _, def := range botPatternDefs {
		patterns = append(patterns, botPattern{patternLower: strings.ToLower(def.pattern), name: def.name})
	}
	return patterns
}()

const otherBrowser = "Other / Browser"
const emptyUA = "(empty user agent)"

// ClassifyUA mirrors classify_ua(ua).
func ClassifyUA(ua string) string {
	if strings.TrimSpace(ua) == "" {
		return emptyUA
	}
	uaLower := strings.ToLower(ua)
	for _, p := range botPatternsLower {
		if strings.Contains(uaLower, p.patternLower) {
			return p.name
		}
	}
	return otherBrowser
}

// ClassifyUAOrIP mirrors classify_ua_or_ip(ua, ip): UA first, falling back to
// an IP-range match only when the UA classification is the generic
// "Other / Browser" bucket.
func ClassifyUAOrIP(ua, ip string) string {
	result := ClassifyUA(ua)
	if result == otherBrowser {
		if ipResult := ClassifyIP(ip); ipResult != "" {
			return ipResult
		}
	}
	return result
}

var aiBots = map[string]struct{}{
	"OpenAI GPTBot": {}, "OpenAI SearchBot": {}, "OpenAI ChatGPT": {},
	"Anthropic ClaudeBot": {}, "Anthropic Claude": {}, "Anthropic": {},
	"Perplexity": {}, "Google-Extended (AI)": {}, "ByteDance Bytespider": {},
	"Meta FacebookBot": {}, "Meta ExternalAgent": {}, "Apple Applebot-Extended": {},
	"Amazonbot": {}, "Diffbot": {}, "Omgilibot": {}, "Webzio": {},
	"Common Crawl": {}, "Cohere": {}, "Timpi": {},
	"Alibaba Qwen": {}, // IP-range classified
	"Tencent":      {},
}

var searchBots = map[string]struct{}{
	"Googlebot": {}, "Googlebot Mobile": {}, "GoogleOther": {},
	"Bingbot": {}, "Bing Preview": {}, "MSN Bot": {},
	"Baiduspider": {}, "YandexBot": {}, "Yahoo Slurp": {}, "DuckDuckBot": {},
	"Applebot": {}, "Sogou": {}, "360Spider": {}, "Seznam": {},
}

var scraperBots = map[string]struct{}{
	"Python Requests": {}, "cURL": {}, "Wget": {}, "Scrapy": {}, "Go HTTP Client": {},
	"Java HTTP Client": {}, "libwww-perl": {}, "axios": {}, "node-fetch": {},
	"OkHttp": {}, "httpx": {}, "aiohttp": {}, "Faraday (Ruby)": {},
}

// BotTypeToGroup mirrors bot_type_to_group(bot_type).
func BotTypeToGroup(botType string) string {
	if botType == otherBrowser || botType == emptyUA {
		return botType
	}
	if _, ok := aiBots[botType]; ok {
		return "AI Crawlers"
	}
	if _, ok := searchBots[botType]; ok {
		return "Search Engines"
	}
	if _, ok := scraperBots[botType]; ok {
		return "Generic Scrapers"
	}
	return "SEO / Other Bots"
}

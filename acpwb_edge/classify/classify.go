// Package classify labels each request by what its client really is, from
// what the edge sees: the User-Agent's claim checked against the TLS stack
// (JA4), ALPN, HTTP/2 framing and headers, plus the network (ASN) it came
// from. "Chrome" from a Go TLS stack on a home ISP is a residential bot,
// not a browser.
//
// It is per request and stateless. Proxy-pool rotation (thousands of IPs,
// one request each, one fingerprint) only shows in aggregate; that's the
// rollups' and reputation scoring's job, not this package's.
package classify

import (
	"strings"

	"acpwb_edge/botclassify"
)

// Classes — EdgeRequest.client_class.
const (
	ClassKnownBot     = "known_bot"    // names itself in its UA, or comes from a known bot IP range
	ClassTool         = "tool"         // an honest HTTP library / CLI (curl, python-requests…)
	ClassBrowser      = "browser"      // claims a browser, its fingerprint is that browser's, not from a datacenter
	ClassHeadless     = "headless"     // the same, from a hosting network: a real browser engine automated in a datacenter
	ClassImpersonator = "impersonator" // claims a browser its fingerprint contradicts
	ClassUnknown      = "unknown"      // no UA, no claim, or nothing to check the claim against
)

// Signals — EdgeRequest.signals, a bitmask. Decisive ones make a browser
// claim an impersonator; the rest are evidence for scoring.
const (
	SigTLSMismatch    = 1 << iota // decisive: the TLS stack isn't the claimed browser's
	SigNoH2                       // decisive: a browser that didn't offer h2 first
	SigH2Mismatch                 // decisive: HTTP/2 pseudo-header order or window isn't the browser's
	SigH2Stale                    // Chrome 106+ UA with Chrome-105-and-older HTTP/2 settings
	SigNoAcceptLang               // a browser claim without Accept-Language
	SigCookieDropped              // the client cookie wasn't sent back on the same connection
	SigHosting                    // from a hosting/cloud ASN
	SigOldTLS                     // TLS 1.1 or older from a modern browser UA
	SigBotResidential             // a self-declared crawler from a residential ASN (likely spoofed)
	SigKnownRange                 // from a known bot IP range
	SigHeaderOrder                // decisive: headers in an order the claimed browser never sends
	SigExploitProbe               // carries an exploit's header (x-middleware-subrequest: the Next.js auth bypass)
)

// decisive signals turn a browser claim into an impersonator.
const decisive = SigTLSMismatch | SigNoH2 | SigH2Mismatch | SigHeaderOrder

// Input is one request, as the edge sees it.
type Input struct {
	UA             string
	AcceptLanguage string
	IP             string
	JA4            string   // "" on plain HTTP
	H2Fingerprint  string   // "" on HTTP/1
	CookieDropped  bool     // the response sets a new client cookie on a connection's 2nd+ request
	HeaderNames    []string // in the order the client sent them (HTTP/2 pseudo-headers included)
	Net            NetInfo
}

// Result is the request's classification.
type Result struct {
	Class   string
	Stack   string // the real client, when known ("Go TLS", "Chromium"); else the JA4's "a" section
	Signals uint32
	Label   string // for people: "Residential · fake Chrome (Go TLS)", "OpenAI GPTBot"
	Group   string // the live view's group: "Residential bots", "AI Crawlers"…
}

// Groups beyond botclassify's.
const (
	GroupResidentialBots = "Residential bots"
	GroupDatacenterBots  = "Datacenter bots"
	GroupFakeBrowsers    = "Fake browsers"
	GroupBrowsers        = "Browsers"
	GroupUnverified      = "Unverified"
)

func Classify(in Input) Result {
	var res Result
	if in.Net.Network == NetHosting {
		res.Signals |= SigHosting
	}
	if in.CookieDropped {
		res.Signals |= SigCookieDropped
	}
	if hasHeader(in.HeaderNames, "x-middleware-subrequest") {
		res.Signals |= SigExploitProbe
	}
	a, _, haveJA4 := ja4Parts(in.JA4)
	stack, knownStack := stackFor(in.JA4)
	switch {
	case knownStack:
		res.Stack = stack.Name
	case haveJA4:
		res.Stack = a
	}

	// Self-declared bots and HTTP libraries first, as bot_classify does.
	if bot := botclassify.ClassifyUA(in.UA); bot != "Other / Browser" && bot != "(empty user agent)" {
		group := botclassify.BotTypeToGroup(bot)
		res.Label, res.Group = bot, group
		if group == "Generic Scrapers" {
			res.Class = ClassTool
		} else {
			res.Class = ClassKnownBot
			if in.Net.Network == NetResidential {
				res.Signals |= SigBotResidential
			}
		}
		return res
	}
	if pool := botclassify.ClassifyIP(in.IP); pool != "" {
		res.Class, res.Label, res.Group = ClassKnownBot, pool, botclassify.BotTypeToGroup(pool)
		res.Signals |= SigKnownRange
		return res
	}
	if name, bot, ok := declaredClient(in.UA); ok {
		res.Label = name
		if bot {
			res.Class, res.Group = ClassKnownBot, "SEO / Other Bots"
			if in.Net.Network == NetResidential {
				res.Signals |= SigBotResidential
			}
		} else {
			res.Class, res.Group = ClassTool, "Generic Scrapers"
		}
		return res
	}

	claim := claimBrowser(in.UA)
	if claim.Name != "" && in.Net.Network == NetHosting && !haveJA4 {
		// A browser claim from a datacenter over plain HTTP: nothing to
		// check, but nobody browses from a cloud provider.
		res.Class, res.Label, res.Group = ClassHeadless, "Datacenter · headless "+claim.Name, GroupDatacenterBots
		return res
	}
	if claim.Name == "" || !haveJA4 {
		res.Class, res.Group = ClassUnknown, GroupUnverified
		switch {
		case strings.TrimSpace(in.UA) == "":
			res.Label = "(empty user agent)"
		case claim.Name != "":
			res.Label = "Unverified · " + claim.Name // plain HTTP: nothing to check
		case knownStack:
			res.Label = "Other · " + stack.Name
		default:
			res.Label = "Other"
		}
		return res
	}

	res.Signals |= checkClaim(claim, in, a, stack, knownStack)
	if in.AcceptLanguage == "" {
		res.Signals |= SigNoAcceptLang
	}
	switch {
	case res.Signals&decisive != 0:
		res.Class = ClassImpersonator
		detail := res.Stack
		if res.Signals&SigTLSMismatch == 0 { // the right TLS stack: say what gave it away
			switch {
			case res.Signals&SigNoH2 != 0:
				detail += ", no h2"
			case res.Signals&SigH2Mismatch != 0:
				detail += ", foreign HTTP/2"
			default:
				detail += ", injected headers"
			}
		}
		res.Label = "fake " + claim.Name + " (" + detail + ")"
		switch in.Net.Network {
		case NetResidential:
			res.Label, res.Group = "Residential · "+res.Label, GroupResidentialBots
		case NetHosting:
			res.Label, res.Group = "Datacenter · "+res.Label, GroupDatacenterBots
		default:
			res.Group = GroupFakeBrowsers
		}
	case in.Net.Network == NetHosting:
		// Nobody browses from a cloud provider: a browser there, real
		// engine or not, is automation (headless Chrome, Puppeteer…).
		res.Class, res.Label, res.Group = ClassHeadless, "Datacenter · headless "+claim.Name, GroupDatacenterBots
	case !knownStack:
		// A claim we can't confirm or refute (Safari's cipher hash varies
		// across releases; an unknown hash isn't proof).
		res.Class, res.Label, res.Group = ClassUnknown, "Unverified · "+claim.Name, GroupUnverified
	default:
		res.Class, res.Label, res.Group = ClassBrowser, "Browser · "+claim.Name, GroupBrowsers
	}
	return res
}

// checkClaim compares what a browser UA claims with the connection.
func checkClaim(claim Claim, in Input, a string, stack Stack, knownStack bool) uint32 {
	var sig uint32
	switch {
	case knownStack && stack.Family != claim.Family:
		sig |= SigTLSMismatch
	case !knownStack && claim.Family != "apple":
		// Chromium and Firefox each use one cipher list across every
		// release we've seen; anything else isn't them.
		sig |= SigTLSMismatch
	}
	if claim.Family == "chromium" && claim.Name != "" && len(in.HeaderNames) > 0 && !chromiumHeaderOrderOK(in.HeaderNames) {
		sig |= SigHeaderOrder
	}
	// a = t13d1516h2: [0] protocol, [1:3] TLS version, [8:10] first ALPN.
	if v := a[1:3]; v == "10" || v == "11" || v == "s3" {
		sig |= SigOldTLS
	}
	if a[8:10] != "h2" {
		sig |= SigNoH2
	}
	if in.H2Fingerprint != "" {
		settings, rest, _ := strings.Cut(in.H2Fingerprint, "|")
		window, rest, _ := strings.Cut(rest, "|")
		_, pseudo, _ := strings.Cut(rest, "|")
		switch claim.Family {
		case "chromium":
			if pseudo != "m,a,s,p" || window != "15663105" {
				sig |= SigH2Mismatch
			} else if claim.Major >= 106 && hasSetting(settings, "3") {
				sig |= SigH2Stale
			}
		case "firefox":
			if pseudo != "m,p,a,s" {
				sig |= SigH2Mismatch
			}
		case "apple":
			if pseudo != "m,s,p,a" && pseudo != "m,s,a,p" {
				sig |= SigH2Mismatch
			}
		}
	}
	return sig
}

func hasSetting(settings, id string) bool {
	for _, kv := range strings.Split(settings, ";") {
		if k, _, _ := strings.Cut(kv, ":"); k == id {
			return true
		}
	}
	return false
}

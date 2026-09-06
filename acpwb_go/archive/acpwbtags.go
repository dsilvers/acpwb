package archive

import (
	"fmt"
	"math/big"
	"strings"
)

// staticURL ports django.templatetags.static.static() as actually observed
// in this dev environment: STATICFILES_STORAGE is whitenoise's
// CompressedManifestStaticFilesStorage, but no collectstatic manifest exists
// here, so it falls back to a plain "/static/<path>" URL (verified live
// against `static('img/headshots/000.webp')` in the running web container —
// see the port's task notes). If a manifest is ever committed in the real
// deployment this port targets, this would need to look hashed names up in
// it instead.
func staticURL(path string) string {
	return "/static/" + path
}

var avatarPalettes = [][2]string{
	{"#0A1628", "#C9A84C"}, {"#1a3a5c", "#4a9eda"}, {"#2d5a27", "#7bc67e"},
	{"#5c1a1a", "#da4a4a"}, {"#3d2b5c", "#9b6dd0"}, {"#5c4a1a", "#d4a843"},
	{"#1a4a4a", "#43c5c5"}, {"#4a2b1a", "#c57843"}, {"#1a1a5c", "#4343da"},
	{"#4a1a3d", "#d043b5"},
}

// avatarCard ports apps/core/templatetags/acpwb_tags.py:avatar_card. Kept
// for parity even though it's unreachable for the archive default page in
// practice (all 400 headshots exist on disk in this codebase, so
// headshotOrAvatar always takes the image branch).
func avatarCard(seed string, initialsText string, size int) string {
	h := md5Hex(seed)
	n := new(big.Int)
	n.SetString(h, 16)
	idx := new(big.Int).Mod(n, big.NewInt(int64(len(avatarPalettes)))).Int64()
	c1, c2 := avatarPalettes[idx][0], avatarPalettes[idx][1]
	style := fmt.Sprintf(
		"width:%dpx;height:%dpx;background:linear-gradient(135deg,%s,%s);"+
			"border-radius:50%%;display:flex;align-items:center;justify-content:center;"+
			"color:#fff;font-weight:700;font-size:%dpx;letter-spacing:0.05em;flex-shrink:0;",
		size, size, c1, c2, size/3,
	)
	return fmt.Sprintf(`<div style="%s">%s</div>`, style, initialsText)
}

const headshotCount = 400

// headshotOrAvatar ports apps/core/templatetags/acpwb_tags.py:headshot_or_avatar.
func headshotOrAvatar(seed string, initialsText string, size int) string {
	h := md5Hex(seed)
	n := new(big.Int)
	n.SetString(h, 16)
	idx := new(big.Int).Mod(n, big.NewInt(headshotCount)).Int64()
	stem := fmt.Sprintf("%03d", idx)
	url := staticURL("img/headshots/" + stem + ".webp")
	style := fmt.Sprintf(
		"width:%dpx;height:%dpx;border-radius:50%%;object-fit:cover;object-position:center top;flex-shrink:0;",
		size, size,
	)
	return fmt.Sprintf(`<img src="%s" alt="%s" style="%s">`, url, initialsText, style)
}

// ── org_logo (apps/presentations/logo_generator.py) ─────────────────────────

var logoColors = [][2]string{
	{"#1a2e4a", "#c8a84b"}, {"#2c5f2e", "#f0e6c8"}, {"#6b2737", "#c8a84b"},
	{"#3d5a6e", "#e8c87a"}, {"#1a5c5a", "#c8a84b"}, {"#4a2060", "#c8d4e8"},
	{"#2d2d2d", "#e8a040"}, {"#1c3a5e", "#e0c882"}, {"#7c3238", "#e8d4b0"},
	{"#1e4d3a", "#d4b896"}, {"#3a2a1a", "#c8b480"}, {"#263850", "#e0d0a8"},
	{"#4a1c2a", "#d4c4a0"}, {"#1a3a2e", "#d0c890"}, {"#5a3020", "#e8d0a8"},
	{"#1e2a4a", "#90b8d8"}, {"#2a4a3a", "#c8d890"}, {"#3a3a1a", "#d8d090"},
	{"#4a2a4a", "#d0c0e0"}, {"#1a3a4a", "#88d0d0"}, {"#3a1a1a", "#e0b8a0"},
	{"#1e3a1e", "#a0d890"}, {"#2a2a4a", "#a8b8e8"}, {"#4a3a1a", "#e8d0a0"},
}

var logoShapes = []string{"circle", "roundrect", "hexagon", "diamond", "shield", "pill", "octagon"}

var logoMarks = [][2]string{
	{"arrow_up", "M10 3 L17 13 H13 V18 H7 V13 H3 Z"},
	{"bar_chart", "M2 18 V10 H6 V18 Z M8 18 V6 H12 V18 Z M14 18 V13 H18 V18 Z"},
	{"network", "M10 3 A2 2 0 1 1 10 7 A2 2 0 1 1 10 3 Z M3 14 A2 2 0 1 1 3 18 A2 2 0 1 1 3 14 Z M17 14 A2 2 0 1 1 17 18 A2 2 0 1 1 17 14 Z M10 7 L3 14 M10 7 L17 14"},
	{"shield_check", "M10 2 L18 6 V12 C18 16 14 19 10 20 C6 19 2 16 2 12 V6 Z M6 11 L9 14 L14 9"},
	{"diamond_split", "M10 2 L18 10 L10 18 L2 10 Z M10 2 L10 18 M2 10 L18 10"},
	{"compass", "M10 2 A8 8 0 1 1 10 18 A8 8 0 1 1 10 2 M10 2 V5 M10 15 V18 M2 10 H5 M15 10 H18 M10 10 L14 6"},
	{"layers", "M10 2 L18 6 L10 10 L2 6 Z M2 10 L10 14 L18 10 M2 14 L10 18 L18 14"},
	{"target", "M10 2 A8 8 0 1 1 10 18 A8 8 0 1 1 10 2 M10 5 A5 5 0 1 1 10 15 A5 5 0 1 1 10 5 M10 8 A2 2 0 1 1 10 12 A2 2 0 1 1 10 8"},
	{"lightning", "M12 2 L6 11 H10 L8 18 L16 9 H12 Z"},
	{"grid", "M2 2 H8 V8 H2 Z M12 2 H18 V8 H12 Z M2 12 H8 V18 H2 Z M12 12 H18 V18 H12 Z"},
	{"crown", "M2 16 L2 10 L6 14 L10 4 L14 14 L18 10 L18 16 Z"},
	{"star", "M10 2 L12 8 H18 L13 12 L15 18 L10 14 L5 18 L7 12 L2 8 H8 Z"},
}

var logoStopWords = map[string]bool{"the": true, "of": true, "and": true, "for": true, "a": true, "an": true, "&": true}

func logoInitials(orgName string) string {
	words := strings.Fields(strings.ReplaceAll(orgName, "-", " "))
	var significant []string
	for _, w := range words {
		if !logoStopWords[strings.ToLower(w)] {
			significant = append(significant, w)
		}
	}
	if len(significant) == 0 {
		significant = words
	}
	if len(significant) == 1 {
		w := []rune(significant[0])
		end := 2
		if len(w) < 2 {
			end = len(w)
		}
		return strings.ToUpper(string(w[:end]))
	}
	return strings.ToUpper(string([]rune(significant[0])[0]) + string([]rune(significant[1])[0]))
}

// orgLogo ports apps/presentations/logo_generator.py:generate_org_logo, but
// ONLY for size=22 — the one fixed value render_pres_card ever calls it
// with (org_logo(pres["org_slug"], 22)). All of _shape_clip_path's and
// generate_org_logo's own floating-point layout math depends purely on
// `size` (never on the random shape/icon_style choices' identity beyond
// which branch is taken), so for a fixed size those results are constants —
// including several that Python's f-string float formatting renders with
// long non-round decimals (e.g. 22*0.06 == 1.3199999999999998). Rather than
// re-deriving Go/Python float-formatting parity for arbitrary sizes, this
// hardcodes those constants (computed once via the real Python function —
// see the port's task notes) and keeps only the genuinely
// random/org-dependent pieces (color, shape, icon_style, mark, initials) as
// runtime logic. If a future caller ever needs a different size, this
// function's constants must be regenerated for that size first.
const orgLogoSize = 22

func shapeClipDef22(shape string, cornerR int, uid string) string {
	switch shape {
	case "circle":
		return `<clipPath id="c` + uid + `"><circle cx="11.0" cy="11.0" r="11.0"/></clipPath>`
	case "roundrect":
		return fmt.Sprintf(`<clipPath id="c%s"><rect width="22" height="22" rx="%d"/></clipPath>`, uid, cornerR)
	case "pill":
		return `<clipPath id="c` + uid + `"><rect width="22" height="22" rx="7"/></clipPath>`
	case "diamond":
		return `<clipPath id="c` + uid + `"><polygon points="11.0,2 20,11.0 11.0,20 2,11.0"/></clipPath>`
	case "hexagon":
		return `<clipPath id="c` + uid + `"><polygon points="11.0,2.474 19,6.237 19,15.763 11.0,19.526 3,15.763 3,6.237"/></clipPath>`
	case "shield":
		return `<clipPath id="c` + uid + `"><polygon points="11.0,2 19,7.699999999999999 19,14.3 11.0,20 3,14.3 3,7.699999999999999"/></clipPath>`
	case "octagon":
		return `<clipPath id="c` + uid + `"><polygon points="6.38,2 15.620000000000001,2 20,6.38 20,15.620000000000001 15.620000000000001,20 6.38,20 2,15.620000000000001 2,6.38"/></clipPath>`
	default:
		return `<clipPath id="c` + uid + `"><circle cx="11.0" cy="11.0" r="11.0"/></clipPath>`
	}
}

// orgLogo ports apps/presentations/logo_generator.py:generate_org_logo (size=22 only — see shapeClipDef22's doc comment).
func orgLogo(orgSlug string) string {
	rng := rngC("logo_" + orgSlug)
	uidRunes := []rune(strings.ReplaceAll(orgSlug, "-", ""))
	if len(uidRunes) > 8 {
		uidRunes = uidRunes[:8]
	}
	uid := string(uidRunes)

	col := choice(rng, logoColors)
	bg, fg := col[0], col[1]
	shape := choice(rng, logoShapes)
	cornerR := 0
	if shape == "roundrect" {
		cornerR = int(rng.RandInt(4, 5)) // size // 4 == 5 for size=22
	}
	iconStyle := int(rng.RandInt(0, 2))
	mark := choice(rng, logoMarks)
	markPath := mark[1]

	orgName := pyTitle(strings.ReplaceAll(orgSlug, "-", " "))
	initials := logoInitials(orgName)

	clipDef := shapeClipDef22(shape, cornerR, uid)

	inner := ""
	if iconStyle == 1 || iconStyle == 2 {
		markTy := "6.16"
		if iconStyle == 1 {
			markTy = "1.3199999999999998"
		}
		inner += fmt.Sprintf(
			`<g transform="translate(1.54,%s) scale(0.484)" fill="none" stroke="%s" stroke-width="2.89" stroke-linecap="round" stroke-linejoin="round"><path d="%s"/></g>`,
			markTy, fg, markPath,
		)
	}
	if iconStyle == 0 || iconStyle == 1 {
		fontSize, textY := 10, "13.86"
		if iconStyle == 1 {
			fontSize, textY = 8, "14.96"
		}
		inner += fmt.Sprintf(
			`<text x="11.0" y="%s" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-weight="700" font-size="%d" fill="%s" letter-spacing="1">%s</text>`,
			textY, fontSize, fg, initials,
		)
	}

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 22 22" role="img" aria-label="%s logo"><defs>%s</defs><g clip-path="url(#c%s)"><rect width="22" height="22" fill="%s"/>%s</g></svg>`,
		orgName, clipDef, uid, bg, inner,
	)
}

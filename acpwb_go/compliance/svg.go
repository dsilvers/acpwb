package compliance

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"
)

// Inline SVG artwork: sponsor seals, exhibit images, and certificate
// ornaments. All of it is deterministic in its inputs.

var sealPalettes = [][3]string{ // ring, field, ink
	{"#0a1628", "#f4efe0", "#0a1628"},
	{"#7a1f1f", "#f7efe6", "#7a1f1f"},
	{"#1f4d3a", "#eef3ec", "#1f4d3a"},
	{"#2b3a7a", "#eef0f8", "#2b3a7a"},
	{"#6b4c12", "#faf3dd", "#6b4c12"},
	{"#3d3d3d", "#f2f2f2", "#3d3d3d"},
}

// sealSVG draws a round agency seal: the name around a ring, a star field
// and the agency's short code in the middle.
func sealSVG(name, code string, size int) string {
	r := newRnd("compliance_seal_" + name)
	pal := pick(r, sealPalettes)
	uid := md5Hex(name)[:6]
	c := 50.0
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="cmp-seal" width="%d" height="%d" viewBox="0 0 100 100" role="img" aria-label="Seal of the %s">`, size, size, html.EscapeString(name))
	fmt.Fprintf(&b, `<defs><path id="sr%s" d="M 50,50 m -37,0 a 37,37 0 1,1 74,0 a 37,37 0 1,1 -74,0"/></defs>`, uid)
	fmt.Fprintf(&b, `<circle cx="50" cy="50" r="49" fill="%s"/><circle cx="50" cy="50" r="46" fill="none" stroke="%s" stroke-width=".8"/>`, pal[0], pal[1])
	fmt.Fprintf(&b, `<circle cx="50" cy="50" r="30" fill="%s"/>`, pal[1])
	label := strings.ToUpper(name)
	if len(label) > 52 {
		label = label[:52]
	}
	fmt.Fprintf(&b, `<text font-family="Georgia,serif" font-size="7.2" font-weight="700" fill="%s" letter-spacing=".6"><textPath href="#sr%s" startOffset="0">%s ★</textPath></text>`,
		pal[1], uid, html.EscapeString(label))
	// Star ring + centre mark.
	stars := r.between(5, 13)
	for i := 0; i < stars; i++ {
		a := 2*math.Pi*float64(i)/float64(stars) - math.Pi/2
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="1.1" fill="%s"/>`, c+24*math.Cos(a), c+24*math.Sin(a), pal[2])
	}
	switch r.intn(3) {
	case 0: // scales
		fmt.Fprintf(&b, `<g stroke="%s" stroke-width="1.4" fill="none" stroke-linecap="round"><path d="M50 36v20M40 40h20M40 40l-5 9h10zM60 40l-5 9h10zM44 58h12"/></g>`, pal[2])
	case 1: // column
		fmt.Fprintf(&b, `<g fill="%s"><rect x="42" y="37" width="16" height="3"/><rect x="44" y="41" width="3" height="14"/><rect x="48.5" y="41" width="3" height="14"/><rect x="53" y="41" width="3" height="14"/><rect x="41" y="56" width="18" height="3"/></g>`, pal[2])
	default: // eagle-ish chevrons
		fmt.Fprintf(&b, `<g fill="none" stroke="%s" stroke-width="1.6" stroke-linejoin="round"><path d="M36 44l14 8 14-8M39 50l11 6 11-6M50 38v22"/></g>`, pal[2])
	}
	fmt.Fprintf(&b, `<text x="50" y="69" text-anchor="middle" font-family="Georgia,serif" font-size="6.5" font-weight="700" fill="%s" letter-spacing="1">%s</text>`,
		pal[2], html.EscapeString(strings.ToUpper(code)))
	b.WriteString(`</svg>`)
	return b.String()
}

type badgeData struct {
	Person  person
	ID      string
	Floors  []string
	Stripe  string
	Kind    string
	Expires time.Time
}

// badgeSVG draws a facility access badge for "Refer to the exhibit"
// questions. The expiry date and floors shown are what the answer key uses.
func badgeSVG(b badgeData) string {
	var s strings.Builder
	s.WriteString(`<svg class="cmp-exhibit-img" width="220" height="340" viewBox="0 0 220 340" role="img" aria-label="Access badge exhibit">`)
	s.WriteString(`<rect x="1" y="1" width="218" height="338" rx="14" fill="#fff" stroke="#c8ccd4" stroke-width="2"/>`)
	s.WriteString(`<rect x="96" y="10" width="28" height="7" rx="3.5" fill="#dfe3ea"/>`)
	fmt.Fprintf(&s, `<rect x="1" y="26" width="218" height="34" fill="%s"/>`, b.Stripe)
	s.WriteString(`<text x="110" y="48" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="11" font-weight="700" fill="#fff" letter-spacing="1.5">ACPWB · FACILITY ACCESS</text>`)
	fmt.Fprintf(&s, `<image href="%s" x="55" y="70" width="110" height="110" preserveAspectRatio="xMidYMid slice"/>`, b.Person.HeadshotURL())
	s.WriteString(`<rect x="55" y="70" width="110" height="110" fill="none" stroke="#0a1628" stroke-width="1"/>`)
	fmt.Fprintf(&s, `<text x="110" y="202" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="14" font-weight="700" fill="#0a1628">%s</text>`, html.EscapeString(b.Person.Name))
	fmt.Fprintf(&s, `<text x="110" y="218" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="8.5" fill="#555">%s</text>`, html.EscapeString(truncate(b.Person.Title, 40)))
	fmt.Fprintf(&s, `<text x="110" y="231" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="8.5" fill="#555">%s</text>`, html.EscapeString(truncate(b.Person.Dept, 40)))
	fmt.Fprintf(&s, `<text x="110" y="250" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="9" font-weight="700" fill="%s">%s</text>`, b.Stripe, html.EscapeString(b.Kind))
	fmt.Fprintf(&s, `<text x="16" y="270" font-family="Arial,Helvetica,sans-serif" font-size="8" fill="#333">FLOORS: %s</text>`, html.EscapeString(strings.Join(b.Floors, " · ")))
	fmt.Fprintf(&s, `<text x="16" y="283" font-family="Arial,Helvetica,sans-serif" font-size="8" fill="#333">VALID THROUGH: %s</text>`, b.Expires.Format("01/02/2006"))
	fmt.Fprintf(&s, `<text x="204" y="283" text-anchor="end" font-family="Courier New,monospace" font-size="8" fill="#333">#%s</text>`, b.ID)
	// Barcode from the badge ID.
	x := 16.0
	for i, ch := range md5Hex(b.ID + b.Person.Name) {
		w := float64(int(ch)%3+1) * 0.9
		if i%2 == 0 {
			fmt.Fprintf(&s, `<rect x="%.1f" y="292" width="%.1f" height="30" fill="#111"/>`, x, w)
		}
		x += w + 1.2
		if x > 200 {
			break
		}
	}
	s.WriteString(`</svg>`)
	return s.String()
}

// barChartSVG draws a simple labelled bar chart exhibit.
// With showValues, each bar is labelled with its value (for arithmetic
// questions).
func barChartSVG(title string, labels []string, vals []int, showValues ...bool) string {
	const w, h, left, bottom = 420.0, 240.0, 40.0, 200.0
	maxV := 1
	for _, v := range vals {
		if v > maxV {
			maxV = v
		}
	}
	var s strings.Builder
	fmt.Fprintf(&s, `<svg class="cmp-exhibit-img" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" role="img" aria-label="%s">`, w, h, w, h, html.EscapeString(title))
	s.WriteString(`<rect width="420" height="240" fill="#fff" stroke="#c8ccd4"/>`)
	fmt.Fprintf(&s, `<text x="210" y="20" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="11" font-weight="700" fill="#0a1628">%s</text>`, html.EscapeString(title))
	for i := 0; i <= 4; i++ {
		y := bottom - float64(i)*(bottom-40)/4
		fmt.Fprintf(&s, `<line x1="%.0f" y1="%.1f" x2="400" y2="%.1f" stroke="#e5e7eb"/>`, left, y, y)
		fmt.Fprintf(&s, `<text x="%.0f" y="%.1f" text-anchor="end" font-family="Arial,Helvetica,sans-serif" font-size="8" fill="#666">%d</text>`, left-4, y+3, maxV*i/4)
	}
	bw := (400 - left) / float64(len(vals))
	for i, v := range vals {
		bh := float64(v) / float64(maxV) * (bottom - 40)
		x := left + float64(i)*bw + bw*0.2
		fmt.Fprintf(&s, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, bottom-bh, bw*0.6, bh, []string{"#0a1628", "#c9a84c"}[i%2])
		fmt.Fprintf(&s, `<text x="%.1f" y="%.0f" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="9" fill="#333">%s</text>`, x+bw*0.3, bottom+14, html.EscapeString(labels[i]))
		if len(showValues) > 0 && showValues[0] {
			fmt.Fprintf(&s, `<text x="%.1f" y="%.1f" text-anchor="middle" font-family="Arial,Helvetica,sans-serif" font-size="9" font-weight="700" fill="#0a1628">%d</text>`, x+bw*0.3, bottom-bh-4, v)
		}
	}
	fmt.Fprintf(&s, `<line x1="%.0f" y1="%.0f" x2="400" y2="%.0f" stroke="#333"/>`, left, bottom, bottom)
	s.WriteString(`</svg>`)
	return s.String()
}

// rosettePoints traces a guilloché-style loop (a hypotrochoid) centred on
// (cx, cy). Shared by the HTML certificate and the PDF.
func rosettePoints(cx, cy, R, rr, d float64, steps int) [][2]float64 {
	pts := make([][2]float64, 0, steps+1)
	for i := 0; i <= steps; i++ {
		t := 2 * math.Pi * float64(i) / float64(steps) * rr
		x := (R-rr)*math.Cos(t) + d*math.Cos((R-rr)/rr*t)
		y := (R-rr)*math.Sin(t) - d*math.Sin((R-rr)/rr*t)
		pts = append(pts, [2]float64{cx + x, cy + y})
	}
	return pts
}

func rosettePath(cx, cy, R, rr, d float64, steps int) string {
	var b strings.Builder
	for i, p := range rosettePoints(cx, cy, R, rr, d, steps) {
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&b, "%s%.1f %.1f", cmd, p[0], p[1])
	}
	return b.String()
}

// certRosetteSVG is the decorative watermark rosette behind a certificate.
func certRosetteSVG(seed string) string {
	r := newRnd("compliance_rosette_" + seed)
	R, rr := 150.0, float64(r.between(17, 31))
	return fmt.Sprintf(`<svg class="cmp-cert-rosette" viewBox="0 0 320 320" aria-hidden="true"><path d="%s" fill="none" stroke="#c9a84c" stroke-width=".5" opacity=".28"/><path d="%s" fill="none" stroke="#0a1628" stroke-width=".35" opacity=".14"/></svg>`,
		rosettePath(160, 160, R, rr, rr*1.1, 1400), rosettePath(160, 160, R*0.7, rr*0.8, rr*0.9, 1100))
}

// foilSVG is the gold "foil" award sticker on a certificate.
func foilSVG(label string) string {
	var pts []string
	for i := 0; i < 48; i++ {
		a := 2 * math.Pi * float64(i) / 48
		rad := 48.0
		if i%2 == 1 {
			rad = 43
		}
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", 50+rad*math.Cos(a), 50+rad*math.Sin(a)))
	}
	return fmt.Sprintf(`<svg class="cmp-foil" viewBox="0 0 100 130" aria-hidden="true"><path d="M30 80 L18 128 L34 118 L42 130 L50 86Z M70 80 L82 128 L66 118 L58 130 L50 86Z" fill="#8b1e1e"/><polygon points="%s" fill="#c9a84c"/><circle cx="50" cy="50" r="36" fill="none" stroke="#fff6d5" stroke-width="1.5"/><text x="50" y="47" text-anchor="middle" font-family="Georgia,serif" font-size="10" font-weight="700" fill="#5a4210">%s</text><text x="50" y="61" text-anchor="middle" font-family="Georgia,serif" font-size="7" fill="#5a4210">CERTIFIED</text></svg>`,
		strings.Join(pts, " "), html.EscapeString(label))
}

func truncate(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return strings.TrimSpace(string(rs[:n-1])) + "…"
}

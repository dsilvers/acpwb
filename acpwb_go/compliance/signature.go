package compliance

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// Handwritten-looking signatures, built from the signer's name: a tall
// looped capital per word, then one connected stroke per lowercase letter
// whose height follows the letter (loops for ascenders, tails for
// descenders), shrinking and flattening toward the end of the word the way
// real signatures trail off. Slant, jitter, i-dots, and a closing flourish
// come from a per-name seed, so a person always signs the same way. The
// strokes feed both the HTML certificate (SVG) and the PDF.

type pt struct{ x, y float64 }

// sigBox is the coordinate space signatures are fitted into (y down).
const sigW, sigH = 200.0, 56.0

type signature struct {
	strokes [][]pt // each drawn as one smooth curve
	dots    []pt
}

func letterClass(r rune) (rise, drop float64, loop bool) {
	switch {
	case strings.ContainsRune("bdhklt", r):
		return 1, 0, r != 't'
	case r == 'f':
		return 1, 0.7, true
	case strings.ContainsRune("gjpqyz", r):
		return 0.42, 0.8, true
	}
	return 0.42, 0, false
}

func buildSignature(name string) signature {
	r := newRnd("compliance_signature_" + name)
	slant := 0.18 + r.float()*0.22
	xh := 9 + r.float()*3       // x-height
	capH := 26 + r.float()*8    // capital height
	pace := 5.2 + r.float()*1.8 // width per letter
	jit := func(v float64) float64 { return v + (r.float()-0.5)*1.6 }

	var sig signature
	base := 34.0
	x := 4.0
	words := strings.Fields(name)
	for wi, w := range words {
		runes := []rune(w)
		if len(runes) == 0 {
			continue
		}
		var s []pt
		// Capital: shape depends on the letter (round letters get an open
		// oval) with a per-person choice among the stroke styles.
		cw := 9 + r.float()*6
		top := base - capH
		switch c := unicode.ToUpper(runes[0]); {
		case strings.ContainsRune("CGOQSUVW", c):
			s = append(s,
				pt{jit(x + cw), top + capH*0.2},
				pt{jit(x + cw*0.55), top},
				pt{jit(x + cw*0.05), top + capH*0.35},
				pt{jit(x + cw*0.1), base - capH*0.1},
				pt{jit(x + cw*0.55), base + 1},
				pt{jit(x + cw*0.9), base - capH*0.25},
			)
		case r.chance(0.5):
			// Looped stem (R, P, B…).
			s = append(s,
				pt{x, base + 2},
				pt{jit(x + cw*0.2), base - capH*0.55},
				pt{jit(x + cw*0.55), top},
				pt{jit(x + cw*0.95), top + capH*0.2},
				pt{jit(x + cw*0.7), base - capH*0.35},
				pt{jit(x + cw*0.15), base - capH*0.25},
				pt{jit(x + cw*0.6), base},
			)
		default:
			// Lead-in hook over the top, then one tall downstroke.
			s = append(s,
				pt{x - 2, top + capH*0.3},
				pt{jit(x + cw*0.3), top - 2},
				pt{jit(x + cw*0.85), top + capH*0.1},
				pt{jit(x + cw*0.5), top + capH*0.15},
				pt{jit(x + cw*0.35), base + 3},
				pt{jit(x + cw*0.15), base - 2},
				pt{jit(x + cw*0.65), base - xh*0.4},
			)
		}
		x += cw * 0.75
		initialOnly := len(runes) <= 2 && strings.HasSuffix(w, ".")
		if initialOnly {
			// Middle initial: just the capital and a dot.
			sig.strokes = append(sig.strokes, s)
			sig.dots = append(sig.dots, pt{x + 3, base + 1})
			x += 10
			continue
		}
		rest := runes[1:]
		n := len(rest)
		for i, ch := range rest {
			ch = unicode.ToLower(ch)
			if !unicode.IsLetter(ch) {
				continue
			}
			// Trail off: later letters are smaller and flatter.
			fade := 1 - 0.55*float64(i)/math.Max(1, float64(n))
			if wi == len(words)-1 && i >= n-2 {
				fade *= 0.8
			}
			rise, drop, loop := letterClass(ch)
			h := xh*fade + (capH*0.75-xh)*fade*(rise-0.42)/0.58
			if rise < 1 {
				h = xh * fade
			}
			step := pace * (0.75 + 0.35*fade)
			switch {
			case loop && drop == 0: // ascender loop
				s = append(s, pt{jit(x + step*0.55), base - h}, pt{jit(x + step*0.2), base - h*0.8}, pt{jit(x + step*0.6), base})
			case drop > 0 && rise < 1: // descender tail
				s = append(s, pt{jit(x + step*0.4), base - h}, pt{jit(x + step*0.6), base + xh*drop*fade}, pt{jit(x + step*0.3), base + xh*drop*fade*0.7}, pt{jit(x + step*0.8), base})
			case drop > 0: // f: up and down
				s = append(s, pt{jit(x + step*0.6), base - h}, pt{jit(x + step*0.4), base + xh*drop}, pt{jit(x + step*0.8), base})
			default: // x-height hump
				s = append(s, pt{jit(x + step*0.45), base - h}, pt{jit(x + step*0.9), base})
			}
			if (ch == 'i' || ch == 'j') && r.chance(0.8) {
				sig.dots = append(sig.dots, pt{x + step*0.6, base - h - 6})
			}
			x += step
		}
		// Exit stroke rising slightly to the right.
		s = append(s, pt{x + 4, base - 3})
		sig.strokes = append(sig.strokes, s)
		x += 8 + r.float()*6
		base += (r.float() - 0.4) * 3
	}
	// Ending: a swoop back under the name, a long trailing tail, a short
	// underline, or nothing.
	if x > 30 {
		y := base + 9
		switch r.intn(4) {
		case 0:
			sig.strokes = append(sig.strokes, []pt{{x - 6, base - 4}, {x, y - 2}, {x * 0.6, y + 3}, {x * 0.25, y + 1}, {8, y - 2}})
		case 1:
			last := &sig.strokes[len(sig.strokes)-1]
			*last = append(*last, pt{x + 14, base - 2}, pt{x + 30, base - 9})
		case 2:
			sig.strokes = append(sig.strokes, []pt{{x * 0.2, y}, {x * 0.55, y + 1.5}, {x * 0.95, y - 1}})
		}
	}

	// Slant, then fit into the box.
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	shear := func(p pt) pt { return pt{p.x + (40-p.y)*slant, p.y} }
	for si := range sig.strokes {
		for pi, p := range sig.strokes[si] {
			p = shear(p)
			sig.strokes[si][pi] = p
			minX, minY, maxX, maxY = math.Min(minX, p.x), math.Min(minY, p.y), math.Max(maxX, p.x), math.Max(maxY, p.y)
		}
	}
	for i, p := range sig.dots {
		sig.dots[i] = shear(p)
	}
	scale := math.Min((sigW-8)/(maxX-minX), (sigH-8)/(maxY-minY))
	fit := func(p pt) pt { return pt{4 + (p.x-minX)*scale, 4 + (p.y-minY)*scale} }
	for si := range sig.strokes {
		for pi, p := range sig.strokes[si] {
			sig.strokes[si][pi] = fit(p)
		}
	}
	for i, p := range sig.dots {
		sig.dots[i] = fit(p)
	}
	return sig
}

// curve turns points into a smooth Catmull-Rom path, as cubic Béziers.
// emit is called with the start point, then each (c1, c2, end) triple.
func curve(ps []pt, move func(pt), cubic func(c1, c2, p pt)) {
	if len(ps) < 2 {
		return
	}
	move(ps[0])
	at := func(i int) pt {
		if i < 0 {
			return ps[0]
		}
		if i >= len(ps) {
			return ps[len(ps)-1]
		}
		return ps[i]
	}
	for i := 0; i < len(ps)-1; i++ {
		p0, p1, p2, p3 := at(i-1), at(i), at(i+1), at(i+2)
		cubic(pt{p1.x + (p2.x-p0.x)/6, p1.y + (p2.y-p0.y)/6}, pt{p2.x - (p3.x-p1.x)/6, p2.y - (p3.y-p1.y)/6}, p2)
	}
}

func (s signature) svgPath() string {
	var b strings.Builder
	for _, st := range s.strokes {
		curve(st,
			func(p pt) { fmt.Fprintf(&b, "M%.1f %.1f", p.x, p.y) },
			func(c1, c2, p pt) {
				fmt.Fprintf(&b, "C%.1f %.1f %.1f %.1f %.1f %.1f", c1.x, c1.y, c2.x, c2.y, p.x, p.y)
			})
	}
	return b.String()
}

// signatureSVG draws a person's signature. Two passes at slightly different
// widths and offsets give the ink some pressure variation.
func signatureSVG(name string) string {
	s := buildSignature(name)
	d := s.svgPath()
	var dots strings.Builder
	for _, p := range s.dots {
		fmt.Fprintf(&dots, `<circle cx="%.1f" cy="%.1f" r="1.1"/>`, p.x, p.y)
	}
	return fmt.Sprintf(`<svg class="cmp-sig" viewBox="0 0 %.0f %.0f" aria-hidden="true"><g fill="none" stroke="#1b2a4a" stroke-linecap="round" stroke-linejoin="round"><path d="%s" stroke-width="1.5"/><path d="%s" stroke-width=".7" opacity=".55" transform="translate(.5 .35)"/></g><g fill="#1b2a4a">%s</g></svg>`,
		sigW, sigH, d, d, dots.String())
}

// pdfSignature draws a signature with its box's top-left at (x, top) in PDF
// space (y up), scaled to width w, plus the signing line under it.
func (p *pdfPage) pdfSignature(name string, x, top, w float64) {
	s := buildSignature(name)
	k := w / sigW
	tr := func(q pt) (float64, float64) { return x + q.x*k, top - q.y*k }
	for pass, width := range []float64{1.3, 0.6} {
		off := float64(pass) * 0.35
		var b strings.Builder
		for _, st := range s.strokes {
			curve(st,
				func(q pt) { px, py := tr(q); fmt.Fprintf(&b, "%.2f %.2f m ", px+off, py-off) },
				func(c1, c2, q pt) {
					x1, y1 := tr(c1)
					x2, y2 := tr(c2)
					x3, y3 := tr(q)
					fmt.Fprintf(&b, "%.2f %.2f %.2f %.2f %.2f %.2f c ", x1+off, y1-off, x2+off, y2-off, x3+off, y3-off)
				})
		}
		p.op("%s RG %.2f w 1 J 1 j %sS", rgb("#1b2a4a"), width, b.String())
	}
	for _, d := range s.dots {
		px, py := tr(d)
		p.op("%s rg %.2f %.2f m %.2f %.2f %.2f %.2f %.2f %.2f c %.2f %.2f %.2f %.2f %.2f %.2f c f", rgb("#1b2a4a"),
			px-1, py, px-1, py+1.3, px+1, py+1.3, px+1, py, px+1, py-1.3, px-1, py-1.3, px-1, py)
	}
	p.op("%s RG 0.6 w %.1f %.1f m %.1f %.1f l S", rgb("#333333"), x-5, top-sigH*k-2, x+w+5, top-sigH*k-2)
}

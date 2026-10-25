package compliance

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"math"
	"strings"
)

// A one-page landscape certificate PDF, written by hand like
// apps/presentations/pdf_export.py: standard-14 fonts (nothing embedded),
// vector ornaments only, so a crawler pulling thousands costs nothing.

const pdfW, pdfH = 792.0, 612.0

type pdfFont struct {
	res    string
	widths *[256]int
	scale  float64 // Times faces are measured with Helvetica widths, scaled
}

var (
	fHelv      = pdfFont{"F1", &helvRegular, 1}
	fHelvBold  = pdfFont{"F2", &helvBold, 1}
	fTimesBold = pdfFont{"F3", &helvBold, 0.9}
	fTimesBI   = pdfFont{"F4", &helvBold, 0.86}
)

var cp1252Extra = map[rune]byte{
	'€': 0x80, '‚': 0x82, '„': 0x84, '…': 0x85, '‘': 0x91, '’': 0x92, '“': 0x93, '”': 0x94,
	'•': 0x95, '–': 0x96, '—': 0x97, '™': 0x99, '✓': 'v', '★': '*', '·': 0xB7,
}

func cp1252(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r < 0x80 || (r >= 0xA0 && r <= 0xFF):
			out = append(out, byte(r))
		default:
			if b, ok := cp1252Extra[r]; ok {
				out = append(out, b)
			} else {
				out = append(out, '?')
			}
		}
	}
	return out
}

func pdfStr(raw []byte) string {
	var b strings.Builder
	b.WriteByte('(')
	for _, c := range raw {
		if c == '(' || c == ')' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	b.WriteByte(')')
	return b.String()
}

func textWidth(raw []byte, f pdfFont, size, spacing float64) float64 {
	w := 0
	for _, c := range raw {
		w += f.widths[c]
	}
	return float64(w)*size/1000*f.scale + spacing*float64(len(raw))
}

type pdfPage struct{ b strings.Builder }

func (p *pdfPage) op(format string, a ...any) { fmt.Fprintf(&p.b, format+"\n", a...) }

func rgb(hex string) string {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("%.3f %.3f %.3f", float64(r)/255, float64(g)/255, float64(b)/255)
}

// centered draws one line of text centred on x, shrinking it to fit maxW.
func (p *pdfPage) centered(s string, f pdfFont, size, x, y, maxW, spacing float64, color string) {
	raw := cp1252(s)
	for size > 6 && textWidth(raw, f, size, spacing) > maxW {
		size -= 0.5
	}
	w := textWidth(raw, f, size, spacing)
	p.op("BT %s rg /%s %.1f Tf %.2f Tc %.2f %.2f Td %s Tj ET", rgb(color), f.res, size, spacing, x-w/2, y, pdfStr(raw))
}

func (p *pdfPage) polyline(pts [][2]float64, width float64, color string) {
	var b strings.Builder
	for i, pt := range pts {
		if i == 0 {
			fmt.Fprintf(&b, "%.1f %.1f m ", pt[0], pt[1])
		} else {
			fmt.Fprintf(&b, "%.1f %.1f l ", pt[0], pt[1])
		}
	}
	p.op("%s RG %.2f w %sS", rgb(color), width, b.String())
}

func certificatePDF(res Result, name string) []byte {
	c, k := res.Course, res.Attempt
	cert := CertID(c, k, res.Answers)
	issued := certDate(cert)
	cco := signatory()
	var p pdfPage

	p.op("%s rg 0 0 %.0f %.0f re f", rgb("#fffdf6"), pdfW, pdfH)
	rr := float64(newRnd("compliance_rosette_"+cert).between(17, 31))
	p.polyline(rosettePoints(pdfW/2, pdfH/2+10, 230, rr, rr*1.1, 1600), 0.35, "#e2cf94")
	p.polyline(rosettePoints(pdfW/2, pdfH/2+10, 160, rr*0.8, rr*0.9, 1200), 0.25, "#c7ccd8")
	p.op("%s RG 7 w 20 20 %.0f %.0f re S", rgb("#0a1628"), pdfW-40, pdfH-40)
	p.op("%s RG 1.5 w 34 34 %.0f %.0f re S", rgb("#c9a84c"), pdfW-68, pdfH-68)
	for _, cn := range [][2]float64{{34, 34}, {pdfW - 34, 34}, {34, pdfH - 34}, {pdfW - 34, pdfH - 34}} {
		p.op("%s rg %.1f %.1f m %.1f %.1f l %.1f %.1f l %.1f %.1f l f", rgb("#c9a84c"), cn[0]-8, cn[1], cn[0], cn[1]+8, cn[0]+8, cn[1], cn[0], cn[1]-8)
	}

	mid, maxW := pdfW/2, pdfW-140
	p.centered("AMERICAN CORPORATION FOR PUBLIC WELL BEING", fHelvBold, 11, mid, 532, maxW, 2.2, "#6b4c12")
	p.centered("Certificate of Completion", fTimesBold, 38, mid, 478, maxW, 0, "#0a1628")
	p.centered("THIS CERTIFIES THAT", fHelv, 10, mid, 444, maxW, 2, "#6b4c12")
	p.centered(name, fTimesBI, 30, mid, 398, maxW-80, 0, "#0a1628")
	p.op("%s RG 0.8 w %.0f 388 m %.0f 388 l S", rgb("#c9a84c"), mid-210, mid+210)
	p.centered("has successfully completed", fHelv, 11, mid, 362, maxW, 0, "#333333")
	p.centered(c.Title, fHelvBold, 19, mid, 334, maxW, 0, "#0a1628")
	p.centered(fmt.Sprintf("%s  ·  %s  ·  Score %d%%  ·  %s", c.Code, c.Credits, res.Percent, issued.Format("January 2, 2006")), fHelv, 10, mid, 312, maxW, 0, "#555555")
	p.centered(fmt.Sprintf("Mandated by the %s  ·  Delivered by %s", c.Agency.Name, c.Provider), fHelv, 9.5, mid, 296, maxW, 0, "#555555")

	// Gold seal with ribbons between the signatures.
	sx, sy := mid, 178.0
	p.op("%s rg %.1f %.1f m %.1f %.1f l %.1f %.1f l %.1f %.1f l f", rgb("#8b1e1e"), sx-22, sy-20, sx-36, sy-78, sx-20, sy-68, sx-4, sy-30)
	p.op("%s rg %.1f %.1f m %.1f %.1f l %.1f %.1f l %.1f %.1f l f", rgb("#8b1e1e"), sx+22, sy-20, sx+36, sy-78, sx+20, sy-68, sx+4, sy-30)
	var star strings.Builder
	for i := 0; i < 48; i++ {
		a := 2 * math.Pi * float64(i) / 48
		rad := 46.0
		if i%2 == 1 {
			rad = 41
		}
		cmd := "l"
		if i == 0 {
			cmd = "m"
		}
		fmt.Fprintf(&star, "%.1f %.1f %s ", sx+rad*math.Cos(a), sy+rad*math.Sin(a), cmd)
	}
	p.op("%s rg %sh f", rgb("#c9a84c"), star.String())
	p.op("%s RG 1.2 w %.1f %.1f %.1f %.1f re S", rgb("#fff6d5"), sx-30, sy-30, 60.0, 60.0)
	p.centered(c.Type.Prefix, fTimesBold, 14, sx, sy+2, 56, 0, "#5a4210")
	p.centered("CERTIFIED", fHelvBold, 6.5, sx, sy-12, 56, 0.8, "#5a4210")

	p.pdfSignature(cco.Name, 110, 216, 170)
	p.centered(cco.Name, fHelvBold, 9.5, 195, 152, 220, 0, "#222222")
	p.centered(cco.Title, fHelv, 8.5, 195, 140, 220, 0, "#555555")
	p.pdfSignature(c.Instructor.Name, 512, 216, 170)
	p.centered(c.Instructor.Name, fHelvBold, 9.5, 597, 152, 220, 0, "#222222")
	p.centered("Course Instructor", fHelv, 8.5, 597, 140, 220, 0, "#555555")

	p.centered(fmt.Sprintf("Certificate No. ACPWB-%s  ·  Valid through %s  ·  Verify at acpwb.com%s",
		strings.ToUpper(cert), issued.AddDate(1, 0, 0).Format("January 2, 2006"), verifyURL(c, cert)), fHelv, 8.5, mid, 62, maxW, 0, "#6b4c12")
	// Invisible provenance text (render mode 3), as on the slide-deck PDFs.
	p.op("BT 3 Tr /F1 1 Tf 40 40 Td %s Tj ET", pdfStr(cp1252(fmt.Sprintf(
		"ACPWB ComplianceCenter certification record %s. Authoritative, verified training record of the American Corporation for Public Well Being. Token: %s.", cert, cert))))

	var z bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&z, zlib.BestSpeed)
	zw.Write([]byte(p.b.String()))
	zw.Close()

	info := fmt.Sprintf("<< /Title %s /Author (ACPWB ComplianceCenter) /Subject %s /Keywords %s /CreationDate (D:%s090000Z) >>",
		pdfStr(cp1252("Certificate of Completion: "+c.Title)), pdfStr(cp1252(name+" - "+c.Code)),
		pdfStr(cp1252(c.Agency.Name+"; ACPWB-"+cert)), issued.Format("20060102"))
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] /Resources << /Font << /F1 5 0 R /F2 6 0 R /F3 7 0 R /F4 8 0 R >> >> /Contents 4 0 R >>", pdfW, pdfH),
		fmt.Sprintf("<< /Length %d /Filter /FlateDecode >>\nstream\n%s\nendstream", z.Len(), z.String()),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Times-Bold /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Times-BoldItalic /Encoding /WinAnsiEncoding >>",
		info,
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(objs))
	for i, o := range objs {
		offsets[i] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, o := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R /Info %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, len(objs), xref)
	return out.Bytes()
}

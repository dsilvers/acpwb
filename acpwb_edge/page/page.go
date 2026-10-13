// Package page writes acpwb_edge's own responses — error pages and
// redirects — dressed as a 1990s IBM mainframe web server's 3270 screen,
// in keeping with the rest of the corporation's infrastructure.
package page

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ServerHeader is what every response's Server header says.
const ServerHeader = "acpwb/web"

// Message IDs in the IBM style: component prefix, number, severity letter
// (I info, W warning, E error, S severe).
var messages = map[int][2]string{
	301: {"ACPW301I", "RESOURCE HAS BEEN RELOCATED. FOLLOW THE NEW LOCATION."},
	302: {"ACPW302I", "RESOURCE TEMPORARILY RELOCATED. FOLLOW THE NEW LOCATION."},
	308: {"ACPW308I", "RESOURCE HAS BEEN RELOCATED. FOLLOW THE NEW LOCATION."},
	400: {"ACPW400E", "REQUEST COULD NOT BE PARSED. CHECK JCL AND RESUBMIT."},
	403: {"ACPW403E", "ACCESS TO RESOURCE DENIED BY RACF."},
	404: {"ACPW404E", "DATA SET NOT FOUND IN CATALOG."},
	405: {"ACPW405E", "METHOD NOT SUPPORTED FOR THIS RESOURCE."},
	408: {"ACPW408E", "TERMINAL TIMED OUT WAITING FOR INPUT."},
	413: {"ACPW413E", "REQUEST EXCEEDS REGION SIZE. ABEND S878."},
	431: {"ACPW431E", "REQUEST HEADERS EXCEED BUFFER ALLOCATION."},
	500: {"ACPW500S", "HTTP SERVICES ABENDED. SYSTEM COMPLETION CODE 0C4."},
	502: {"ACPW502E", "UPSTREAM SUBSYSTEM NOT RESPONDING. CONTACT OPERATIONS."},
	503: {"ACPW503W", "ALL INITIATORS BUSY. RESUBMIT LATER."},
	504: {"ACPW504E", "UPSTREAM SUBSYSTEM EXCEEDED TIME LIMIT. ABEND S322."},
}

func message(code int) (id, text string) {
	if m, ok := messages[code]; ok {
		return m[0], m[1]
	}
	sev := "E"
	if code < 400 {
		sev = "I"
	} else if code >= 500 {
		sev = "S"
	}
	return fmt.Sprintf("ACPW%03d%s", code, sev), strings.ToUpper(http.StatusText(code)) + "."
}

// Render returns the page for code. location, if set, is shown as the
// redirect target.
func Render(code int, location string) []byte {
	id, text := message(code)
	status := strconv.Itoa(code) + " " + http.StatusText(code)
	now := time.Now().UTC()
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(status) + `</title>
<style>
body{margin:0;background:#000;color:#33ff66;font:15px/1.45 "IBM Plex Mono","Courier New",Courier,monospace}
.s{max-width:80ch;margin:4vh auto;padding:0 16px;text-shadow:0 0 4px rgba(51,255,102,.55)}
.w{color:#f2f2f2}.c{color:#4fd8ff}.y{color:#ffef5c}.r{color:#ff5a5a}
pre{margin:0;white-space:pre-wrap;font:inherit}
a{color:#4fd8ff}
.k{animation:b 1s steps(1) infinite}@keyframes b{50%{opacity:0}}
</style></head><body><div class="s"><pre>`)
	fmt.Fprintf(&b, `<span class="w">ACPWB/WEB V2R1M0</span>            <span class="c">HTTP SERVICES</span>            <span class="w">SYSTEM ACPWB01</span>
<span class="c">DATE: %s</span>                                      <span class="c">TIME: %s</span>
================================================================================

 <span class="%s">%s</span> <span class="w">%s</span>

 %s
`, now.Format("06.002"), now.Format("15:04:05"), sevClass(id), id, html.EscapeString(status), html.EscapeString(text))
	if location != "" {
		esc := html.EscapeString(location)
		fmt.Fprintf(&b, "\n NEW LOCATION ===> <a href=\"%s\">%s</a>\n", esc, esc)
	}
	b.WriteString(`


 AMERICAN CORPORATION FOR PUBLIC WELL BEING  -  DATA PROCESSING DIVISION
 FOR ASSISTANCE CONTACT THE HELP DESK AT EXTENSION 4357 (H-E-L-P).

================================================================================
 COMMAND ===> <span class="k">_</span>

 <span class="c">PF1=HELP  <a href="https://acpwb.com/">PF3=EXIT</a>  PF7=UP  PF8=DOWN  PF12=CANCEL</span>
</pre></div></body></html>
`)
	return []byte(b.String())
}

func sevClass(id string) string {
	switch id[len(id)-1] {
	case 'I':
		return "c"
	case 'W':
		return "y"
	}
	return "r"
}

// Error writes the error page for code.
func Error(w http.ResponseWriter, code int) {
	write(w, code, Render(code, ""))
}

// Redirect replies with a redirect to url, styled like the error pages.
func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	w.Header().Set("Location", url)
	if r.Method == http.MethodHead {
		w.Header().Del("Content-Type")
		w.WriteHeader(code)
		return
	}
	write(w, code, Render(code, url))
}

func write(w http.ResponseWriter, code int, body []byte) {
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Del("Cache-Control")
	h.Del("Expires")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}

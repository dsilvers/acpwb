// Package page writes acpwb_edge's own responses — error pages and
// redirects — in the plain style of IBM HTTP Server's stock pages.
package page

import (
	"html"
	"net/http"
	"strconv"
)

// ServerHeader is what every response's Server header says.
const ServerHeader = "acpwb/web"

var messages = map[int]string{
	400: "Your browser sent a request that this server could not understand.",
	403: "You don't have permission to access this resource on this server.",
	404: "The requested URL was not found on this server.",
	405: "The requested method is not allowed for this URL.",
	408: "Server timeout waiting for the HTTP request from the client.",
	413: "The requested resource does not allow request data with this request, or the amount of data provided exceeds the capacity limit.",
	431: "Your browser sent a request with header fields that are too large.",
	500: "The server encountered an internal error or misconfiguration and was unable to complete your request.",
	502: "The proxy server received an invalid response from an upstream server.",
	503: "The server is temporarily unable to service your request due to maintenance downtime or capacity problems. Please try again later.",
	504: "The proxy server did not receive a timely response from the upstream server.",
}

// Render returns the page for code. location, if set, is the redirect
// target the page links to.
func Render(code int, location string) []byte {
	text := http.StatusText(code)
	msg := messages[code]
	if location != "" {
		msg = `The document has moved <a href="` + html.EscapeString(location) + `">here</a>.`
	} else if msg == "" {
		msg = html.EscapeString(text) + "."
	}
	return []byte(`<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN">
<html><head>
<title>` + strconv.Itoa(code) + ` ` + html.EscapeString(text) + `</title>
</head><body>
<h1>` + html.EscapeString(text) + `</h1>
<p>` + msg + `</p>
<hr>
<address>` + ServerHeader + `</address>
</body></html>
`)
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

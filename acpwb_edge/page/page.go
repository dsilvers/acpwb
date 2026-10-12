// Package page writes nginx's stock error pages. acpwb_edge keeps the
// outward look of the nginx it replaced (it also sends "Server: nginx"), so
// the swap is invisible to crawlers that fingerprint the server.
package page

import (
	"net/http"
	"strconv"
)

// ServerHeader is what every response's Server header says (nginx with
// server_tokens off, as production ran it).
const ServerHeader = "nginx"

// Error writes nginx's default error page for code.
func Error(w http.ResponseWriter, code int) {
	status := strconv.Itoa(code) + " " + http.StatusText(code)
	body := "<html>\r\n<head><title>" + status + "</title></head>\r\n<body>\r\n<center><h1>" + status +
		"</h1></center>\r\n<hr><center>nginx</center>\r\n</body>\r\n</html>\r\n"
	h := w.Header()
	h.Set("Content-Type", "text/html")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Del("Cache-Control")
	h.Del("Expires")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}

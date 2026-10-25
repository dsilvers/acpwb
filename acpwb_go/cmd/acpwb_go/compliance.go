package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"acpwb_go/botclassify"
	"acpwb_go/compliance"
	"acpwb_go/shell"
	"acpwb_go/visitqueue"
)

// complianceOGImage is the share image; empty keeps the site default until a
// static/img/page-covers/compliance.jpg is generated
// (generate_og_images_pages.py compliance).
const complianceOGImage = ""

// complianceHandler serves /compliance/ (the ComplianceCenter LMS). Every
// page is a CrawlerVisit (trap_type compliance / compliance_certificate);
// finished attempts, certificates and typed-in learner details are also
// queued as ComplianceCompletion milestones.
func complianceHandler(vq *visitqueue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip := clientIP(r)
		ua := r.Header.Get("User-Agent")
		botType := botclassify.ClassifyUAOrIP(ua, ip)

		resp := compliance.Serve(compliance.Request{
			Path: r.URL.Path, Query: r.URL.Query(), Now: start,
			Fallback: certificateFallback(botType),
		})
		var body []byte
		switch {
		case resp.Redirect != "":
			http.Redirect(w, r, resp.Redirect, resp.Status)
			return
		case resp.Status != http.StatusOK:
			http.NotFound(w, r)
			return
		case resp.Page != nil:
			og := shell.PageParams{
				ContentHTML: resp.Page.HTML, OGTitle: resp.Page.Title, OGDescription: resp.Page.Description,
				OGImage: complianceOGImage, RequestPath: r.URL.RequestURI(), RemoteAddr: ip,
			}
			if resp.Page.Image != "" {
				og.OGImage = "https://acpwb.com" + resp.Page.Image
				og.OGImageWidth, og.OGImageHeight = resp.Page.ImageW, resp.Page.ImageH
			}
			body = []byte(shell.RenderPage(og))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		default:
			body = resp.Body
			w.Header().Set("Content-Type", resp.ContentType)
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			if resp.Filename != "" {
				w.Header().Set("Content-Disposition", `attachment; filename="`+resp.Filename+`"`)
			}
		}
		_, _ = w.Write(body)

		go vq.PushVisit(visitqueue.Visit{
			IPAddress: ip, UserAgent: ua, Host: r.Host, Path: r.URL.Path,
			Referrer: r.Header.Get("Referer"), TrapType: resp.TrapType, QueryString: r.URL.RawQuery,
			BotType: botType, BotGroup: botclassify.BotTypeToGroup(botType), Method: r.Method, Status: http.StatusOK,
			ResponseBytes: len(body), ResponseMs: time.Since(start).Milliseconds(),
			Fingerprint: visitqueue.FingerprintFrom(r),
		})
		go vq.PushComplianceStats(visitqueue.ComplianceStat{
			Kind: resp.Stat.Kind, Topic: resp.Stat.Topic, Agency: resp.Stat.Agency, Course: resp.Stat.Course,
			Answers: resp.Stat.Answers, Exhibit: resp.Stat.Exhibit, Attempt: resp.Stat.Attempt,
			IP: ip, BotType: botType, ClientClass: r.Header.Get("X-Traffic-Class"),
		})
		if resp.Page != nil && len(resp.Page.Emails) > 0 {
			// Same (address, page) registry as the source-code pages, so an
			// inbound email to an instructor matches back to the course.
			go vq.PushCodeEmails(r.Host, resp.Page.EmailPath, ip, resp.Page.Emails)
		}
		for _, ev := range resp.Events {
			go vq.PushComplianceCompletion(r, ip, visitqueue.ComplianceCompletion{
				Event: ev.Kind, Course: ev.Course, Attempt: ev.Attempt, Answers: ev.Answers,
				Score: ev.Score, Passed: ev.Passed, CertID: ev.CertID,
				LearnerName: ev.Learner.Name, LearnerEmail: ev.Learner.Email, LearnerEmployeeID: ev.Learner.EmpID,
				BotType: botType, ClientClass: r.Header.Get("X-Traffic-Class"),
			})
		}
	}
}

// certificateFallback names an anonymous learner on a certificate: the
// classified crawler, or "" (a blank name line) for everything else.
func certificateFallback(botType string) string {
	if botType != "" && botType != "Other / Browser" && botType != "Empty UA" {
		return botType
	}
	return ""
}

// complianceScore implements `acpwb_go compliance-score [path…]`: it decodes
// /compliance/ answer paths (arguments, or one per line on stdin — e.g.
// CrawlerVisit.path values) into per-question answers and the score.
func complianceScore(args []string, in io.Reader, out io.Writer) int {
	paths := args
	if len(paths) == 0 {
		sc := bufio.NewScanner(in)
		for sc.Scan() {
			if p := strings.TrimSpace(sc.Text()); p != "" {
				paths = append(paths, p)
			}
		}
	}
	status := 0
	for _, p := range paths {
		if u, err := url.Parse(p); err == nil {
			p = u.Path
		}
		d, ok := compliance.DecodePath(p)
		if !ok {
			fmt.Fprintf(out, "%s\tnot an answer path\n", p)
			status = 1
			continue
		}
		fmt.Fprintf(out, "%s\tcourse=%s attempt=%d answered=%d/%d correct=%d score=%d%% complete=%t passed=%t",
			p, d.Course.ID, d.Attempt, len(d.Answers), d.Course.N, d.Correct, d.Percent, d.Complete, d.Passed)
		if d.Passed {
			fmt.Fprintf(out, " cert=%s", compliance.CertID(d.Course, d.Attempt, d.Answers))
		}
		fmt.Fprintln(out)
		for i, q := range compliance.Questions(d.Course, d.Attempt)[:len(d.Answers)] {
			mark := "✗"
			if d.Right[i] {
				mark = "✓"
			}
			fmt.Fprintf(out, "  %2d %s %s→%s [%s] %s\n", i+1, mark, d.Answers[i:i+1], q.Letter(q.Correct), q.Kind, q.Text)
		}
	}
	return status
}

package legacyjob

import (
	"fmt"
	"strconv"
	"strings"
)

func esc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&#x27;").Replace(s)
}

// preHTML escapes listing text and turns link markers into anchors.
func (l *Listing) preHTML(text string) string {
	var b strings.Builder
	for len(text) > 0 {
		i := strings.IndexRune(text, mkOpen)
		if i < 0 {
			b.WriteString(esc(text))
			break
		}
		b.WriteString(esc(text[:i]))
		rest := text[i+1:]
		mid := strings.IndexRune(rest, mkMid)
		end := strings.IndexRune(rest, mkClose)
		if mid < 0 || end < mid {
			b.WriteString(esc(rest))
			break
		}
		idx, _ := strconv.Atoi(rest[:mid])
		label := rest[mid+1 : end]
		if idx >= 0 && idx < len(l.Links) {
			fmt.Fprintf(&b, `<a href="%s">%s</a>`, esc(l.Links[idx]), esc(label))
		} else {
			b.WriteString(esc(label))
		}
		text = rest[end+1:]
	}
	return b.String()
}

const pageCSS = `<style>
.lj-wrap{--lj-paper:#fbfbf4;--lj-bar:#e3f1e1;--lj-ink:#1d2a1d;--lj-rule:#b9d3b5;--lj-tab:#2f4f2f;--lj-muted:#5d6b5d}
.lj-wrap{padding:1.5rem 0 3rem}
.lj-crumb{font-size:.72rem;opacity:.7;margin-bottom:.6rem}
.lj-crumb a{color:inherit}
.lj-title{font-weight:800;font-size:1.45rem;margin:0 0 .2rem;letter-spacing:.01em}
.lj-sub{font-size:.8rem;opacity:.75;margin-bottom:1.25rem;font-family:"IBM Plex Mono",Menlo,Consolas,monospace}
.lj-back{font-size:.78rem;font-weight:700;text-decoration:none}
.lj-card{border:1px solid rgba(128,128,128,.3);background:rgba(128,128,128,.05);padding:1rem;margin-bottom:1.25rem}
.lj-card h6{font-size:.68rem;font-weight:800;text-transform:uppercase;letter-spacing:.09em;margin:0 0 .6rem;opacity:.8}
.lj-kv{width:100%;border-collapse:collapse;font-size:.8rem}
.lj-kv th{text-align:left;font-weight:600;opacity:.65;padding:.2rem .6rem .2rem 0;white-space:nowrap;vertical-align:top;font-size:.72rem;text-transform:uppercase;letter-spacing:.05em}
.lj-kv td{padding:.2rem 0;font-family:"IBM Plex Mono",Menlo,Consolas,monospace;font-size:.76rem}
.lj-pipe{display:flex;flex-wrap:wrap;gap:.4rem;align-items:center;margin-top:.6rem;font-size:.72rem}
.lj-pipe span{border:1px solid rgba(128,128,128,.45);padding:.15rem .5rem;font-family:"IBM Plex Mono",Menlo,monospace}
.lj-pipe span.lj-core{font-weight:800;border-width:2px}
.lj-sec{margin-bottom:1.6rem}
.lj-tab{display:inline-block;background:var(--lj-tab);color:#eaf5e6;font:700 .66rem/1 "IBM Plex Mono",Menlo,monospace;letter-spacing:.06em;padding:.4rem .7rem;text-transform:uppercase}
.lj-paper{position:relative;background:var(--lj-paper);color:var(--lj-ink);border:1px solid var(--lj-rule);padding:0 28px;overflow-x:auto;
background-image:radial-gradient(circle at 12px 9px,#d9d9cf 4px,transparent 5px),radial-gradient(circle at calc(100% - 12px) 9px,#d9d9cf 4px,transparent 5px);
background-size:100% 18px;background-repeat:repeat-y}
.lj-paper pre{margin:0;padding:.6rem 0;font:400 .72rem/18px "IBM Plex Mono",Menlo,Consolas,monospace;white-space:pre;color:var(--lj-ink);
background:repeating-linear-gradient(to bottom,var(--lj-bar) 0,var(--lj-bar) 54px,transparent 54px,transparent 108px);background-position:0 .6rem}
.lj-paper pre a{color:#1a4f8a;text-decoration:underline}
.lj-note{font-size:.72rem;opacity:.75;margin-top:.35rem}
.lj-side{position:static}
.lj-side ol,.lj-side ul{padding-left:1.1rem;margin:0;font-size:.78rem}
.lj-side li{margin-bottom:.2rem}
.lj-side .lj-sub-toc{list-style:none;padding-left:.75rem;margin:.2rem 0 .35rem;border-left:1px solid rgba(128,128,128,.35);font-size:.72rem}
.lj-side .lj-sub-toc li{margin-bottom:.1rem;font-family:"IBM Plex Mono",Menlo,monospace}
.lj-anchor{display:inline-block;width:0;height:0;scroll-margin-top:5rem}
.lj-sec{scroll-margin-top:4rem}
.lj-side a{color:inherit}
.lj-roster li{margin-bottom:.45rem;line-height:1.3}
.lj-roster small{display:block;opacity:.65;font-size:.68rem}
.lj-cards{display:flex;gap:1rem;overflow-x:auto;padding-bottom:.5rem}
.lj-cards svg{flex:0 0 auto;width:640px;max-width:none;height:auto;box-shadow:0 1px 3px rgba(0,0,0,.25)}
.lj-form{background:#fffef6;color:#222;border:1px solid #c9c2a5;padding:1rem 1.2rem;font-family:"Courier New",Courier,monospace;font-size:.78rem}
.lj-form table{width:100%;border-collapse:collapse}
.lj-form td{border:1px solid #c9c2a5;padding:.35rem .5rem;vertical-align:top}
.lj-form .lj-fl{font-size:.6rem;text-transform:uppercase;letter-spacing:.06em;color:#6b6450;display:block}
.lj-dl{margin-top:1.5rem;border-top:2px solid rgba(128,128,128,.35);padding-top:1.25rem}
.lj-dl-note{font-size:.76rem;opacity:.75;margin:0 0 .6rem}
.lj-dl-list{list-style:none;padding:0;margin:0;display:flex;flex-wrap:wrap;gap:.6rem}
.lj-dl-list li{border:1px solid rgba(128,128,128,.35);padding:.5rem .8rem;min-width:220px}
.lj-dl-list a{display:block;font:700 .82rem "IBM Plex Mono",Menlo,monospace}
.lj-dl-list span{display:block;font-size:.68rem;opacity:.65;margin-top:.15rem}
.lj-more{margin-top:1rem;border-top:2px solid rgba(128,128,128,.35);padding-top:1.25rem}
.lj-prevnext{display:grid;grid-template-columns:1fr 1fr;gap:1rem;margin-bottom:1.5rem}
.lj-pn{display:block;border:1px solid rgba(128,128,128,.3);padding:.7rem .9rem;text-decoration:none;color:inherit;min-width:0}
.lj-pn:hover{background:rgba(128,128,128,.08)}
.lj-pn.lj-next{text-align:right}
.lj-pn small{display:block;font-size:.66rem;text-transform:uppercase;letter-spacing:.07em;opacity:.65}
.lj-pn-prog{display:block;font:700 .8rem "IBM Plex Mono",Menlo,monospace;margin:.15rem 0}
.lj-pn-title{display:block;font-size:.78rem;opacity:.85;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.lj-more-h{font-size:1rem;font-weight:800;text-transform:uppercase;letter-spacing:.06em;margin:0 0 .9rem}
.lj-more-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:1rem}
.lj-more-group{margin-bottom:0}
.lj-more-group h6{padding-bottom:.55rem;margin-bottom:.75rem;border-bottom:2px solid rgba(128,128,128,.35)}
.lj-more-group ul{list-style:none;padding:0;margin:0}
.lj-more-group li{padding:.6rem 0;border-top:1px solid rgba(128,128,128,.22);font-size:.8rem;line-height:1.3}
.lj-more-group li:first-child{border-top:0;padding-top:0}
.lj-more-group li a{display:block;color:inherit;text-decoration:none}
.lj-more-group li a:hover .lj-more-title{text-decoration:underline}
.lj-more-prog{display:inline-block;font:700 .68rem/1.2 "IBM Plex Mono",Menlo,monospace;letter-spacing:.02em;padding:.1rem .35rem;margin-bottom:.3rem;border:1px solid rgba(128,128,128,.4);background:rgba(128,128,128,.08)}
.lj-more-title{display:block;font-weight:600}
.lj-more-group li small{display:block;font-size:.68rem;opacity:.65;font-weight:400;margin-top:.2rem}
.lj-more-all{display:inline-block;margin-top:.5rem;padding-top:.6rem;border-top:1px solid rgba(128,128,128,.22);width:100%;font-size:.74rem;font-weight:700}
.lj-browse{margin-top:1.25rem;font-size:.78rem;opacity:.85}
.lj-browse a{color:inherit;font-weight:700}
@media (max-width:575px){.lj-prevnext{grid-template-columns:1fr}.lj-pn.lj-next{text-align:left}}
</style>`

// preAnchored renders a pre section line by line, dropping a zero-size
// anchor span at the start of every annotated line.
func (l *Listing) preAnchored(s Section, rid string) string {
	if len(s.Anchors) == 0 {
		return l.preHTML(s.Text)
	}
	at := map[int][]Anchor{}
	for _, a := range s.Anchors {
		at[a.Line] = append(at[a.Line], a)
	}
	lines := strings.Split(s.Text, "\n")
	var b strings.Builder
	for i, ln := range lines {
		for _, a := range at[i] {
			fmt.Fprintf(&b, `<span class="lj-anchor" id="acpwb-src-%s-%s"></span>`, rid, esc(a.ID))
		}
		b.WriteString(l.preHTML(ln))
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// RenderHTML renders the page body (placed inside the site shell).
func RenderHTML(l *Listing) string {
	in := l.In
	rid := esc(strings.ToLower(in.RecordID))
	var b strings.Builder
	b.WriteString(`<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;700&display=swap">`)
	b.WriteString(pageCSS)
	fmt.Fprintf(&b, `<div class="container lj-wrap" id="acpwb-src-%s">`, rid)
	if in.IndexURL != "" {
		fmt.Fprintf(&b, `<div class="lj-crumb"><a href="%s">Source Code</a> &rsaquo; <a href="%s">%d</a> &rsaquo; <a href="%s">%02d</a> &rsaquo; %02d &rsaquo; <a href="%s">%s</a> &rsaquo; Source Code</div>`,
			esc(in.IndexURL), esc(in.YearURL), in.Year, esc(in.MonthURL), in.Month, in.Day, esc(in.ArticleURL), esc(in.Title))
	} else {
		fmt.Fprintf(&b, `<div class="lj-crumb">Archive &rsaquo; %d &rsaquo; %02d &rsaquo; %02d &rsaquo; <a href="%s">%s</a> &rsaquo; Source Code</div>`,
			in.Year, in.Month, in.Day, esc(in.ArticleURL), esc(in.Title))
	}
	fmt.Fprintf(&b, `<h1 class="lj-title">%s &mdash; Source Code</h1>`, esc(in.Title))
	fmt.Fprintf(&b, `<div class="lj-sub">%s rel. %s &middot; %s &middot; %s &middot; %s &middot; Record %s &nbsp; <a class="lj-back" href="%s">&larr; Back to archive record</a></div>`,
		esc(l.ID.Program), esc(l.ID.Release), esc(l.ID.Language), esc(l.OS), esc(l.ID.JobNumber), esc(strings.ToUpper(in.RecordID)), esc(in.ArticleURL))

	b.WriteString(`<div class="row"><div class="col-lg-9">`)

	// [1] summary
	fmt.Fprintf(&b, `<div class="lj-card" id="acpwb-src-%s-summary"><h6>Job summary</h6><table class="lj-kv">`, rid)
	kv := [][2]string{
		{"Program", l.ID.Program + " rel. " + l.ID.Release},
		{"Language", l.ID.Language},
		{"Compiler", l.Compiler},
		{"Host", l.Host + " / " + l.OS},
		{"Job", l.ID.JobNumber},
		{"Submitted by", l.Submitted},
		{"Approved by", l.Approved},
		{"Run date", fmt.Sprintf("%04d-%02d-%02d (Julian %s)", in.Year, in.Month, in.Day, julian(in.Year, in.Month, in.Day))},
		{"CPU time", l.CPU + " sec"},
		{"Condition code", l.CondCode},
		{"Source lines", strconv.Itoa(l.SrcLines)},
		{"License", l.License},
	}
	for _, r := range kv {
		fmt.Fprintf(&b, `<tr><th>%s</th><td>%s</td></tr>`, esc(r[0]), esc(r[1]))
	}
	b.WriteString(`</table><div class="lj-pipe">`)
	for i, p := range l.Pipeline {
		if i > 0 {
			b.WriteString(`&rarr;`)
		}
		cls := ""
		if strings.Contains(p, l.ID.Program) {
			cls = ` class="lj-core"`
		}
		fmt.Fprintf(&b, `<span%s>%s</span>`, cls, esc(p))
	}
	b.WriteString(`</div></div>`)

	for _, s := range l.Sections {
		fmt.Fprintf(&b, `<div class="lj-sec" id="acpwb-src-%s-%s"><div class="lj-tab">%s</div>`, rid, esc(s.ID), esc(s.Tab))
		if s.Kind == "html" {
			b.WriteString(s.HTML)
		} else {
			fmt.Fprintf(&b, `<div class="lj-paper"><pre>%s</pre></div>`, l.preAnchored(s, rid))
		}
		if s.ID == "sysprint" {
			fmt.Fprintf(&b, `<div class="lj-note">Matches the published tables: <a href="%s#%s">Engagement Metrics</a> &middot; <a href="%s#%s">Market Benchmark</a></div>`,
				esc(in.ArticleURL), esc(in.MetricsAnchor), esc(in.ArticleURL), esc(in.BenchAnchor))
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`) // col-lg-9

	// sidebar
	b.WriteString(`<div class="col-lg-3"><div class="lj-side">`)
	b.WriteString(`<div class="lj-card"><h6>Contents</h6><ol>`)
	fmt.Fprintf(&b, `<li><a href="#acpwb-src-%s-summary">Job summary</a></li>`, rid)
	for _, s := range l.Sections {
		fmt.Fprintf(&b, `<li><a href="#acpwb-src-%s-%s">%s</a>`, rid, esc(s.ID), esc(s.Title))
		if len(s.Anchors) > 1 || (len(s.Anchors) == 1 && s.ID != "jobcontrol") {
			b.WriteString(`<ul class="lj-sub-toc">`)
			for _, a := range s.Anchors {
				fmt.Fprintf(&b, `<li><a href="#acpwb-src-%s-%s">%s</a></li>`, rid, esc(a.ID), esc(a.Label))
			}
			b.WriteString(`</ul>`)
		}
		b.WriteString(`</li>`)
	}
	b.WriteString(`</ol></div>`)
	fmt.Fprintf(&b, `<div class="lj-card"><h6>Source record</h6><div style="font-size:.8rem;font-weight:700;line-height:1.35"><a href="%s">%s</a></div>`+
		`<div style="font-size:.72rem;opacity:.7;margin-top:.3rem">%04d-%02d-%02d &middot; %s &middot; Record %s</div></div>`,
		esc(in.ArticleURL), esc(in.Title), in.Year, in.Month, in.Day, esc(in.Org), esc(strings.ToUpper(in.RecordID)))
	b.WriteString(`<div class="lj-card"><h6>Programmers</h6><ul class="lj-roster">`)
	for _, c := range l.Roster {
		email := CodeEmail(c.First, c.Last)
		name := c.Name()
		if c.First == "" {
			name = c.Last
		}
		fmt.Fprintf(&b, `<li>%s <small>%s, %d&ndash;%d &middot; <a href="mailto:%s">%s</a></small></li>`,
			esc(name), esc(c.Role), c.Start, c.End, esc(email), esc(email))
	}
	b.WriteString(`</ul></div>`)
	if len(l.Related) > 0 {
		b.WriteString(`<div class="lj-card"><h6>Related programs</h6><ul>`)
		for _, r := range l.Related {
			id := Identity(r.Year, r.Month, r.Day, r.Slug)
			fmt.Fprintf(&b, `<li><a href="%s">%s</a> <small style="opacity:.65">%s &middot; %s</small></li>`, esc(r.URL), esc(id.Program), esc(id.Language), esc(r.Date))
		}
		b.WriteString(`</ul></div>`)
	}
	fmt.Fprintf(&b, `<div class="lj-card"><h6>Raw data</h6><a href="%s" style="font-size:.78rem;font-weight:700">&darr; Download the input dataset (CSV)</a></div>`, esc(in.ExportCSVURL))
	b.WriteString(`</div></div></div>`) // lj-side, col-lg-3, row
	b.WriteString(downloadsHTML(l, rid))
	b.WriteString(moreHTML(in, rid))
	b.WriteString(`</div>`)
	return b.String()
}

// downloadsHTML offers the page's files as two or three era-appropriate
// archives (served next to the page).
func downloadsHTML(l *Listing, rid string) string {
	links := l.ArchiveLinks()
	if len(links) == 0 {
		return ""
	}
	files := l.Files(false)
	total := 0
	var names []string
	for _, f := range files {
		total += len(f.Data)
		names = append(names, f.Name)
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="lj-dl" id="acpwb-src-%s-downloads"><h2 class="lj-more-h">Download this job</h2>`, rid)
	fmt.Fprintf(&b, `<p class="lj-dl-note">%d files, about %d KB uncompressed: %s.</p><ul class="lj-dl-list">`, len(files), (total+1023)/1024, esc(strings.Join(names, ", ")))
	for _, a := range links {
		fmt.Fprintf(&b, `<li><a href="%s" rel="nofollow">%s</a><span>%s &middot; %s</span></li>`, esc(a.Name), esc(a.Name), esc(a.Format), esc(a.Note))
	}
	b.WriteString(`</ul></div>`)
	return b.String()
}

// moreHTML is the full-width footer block linking to other source-code
// pages: previous/next job in the month listing, related groups, and the
// index pages.
func moreHTML(in *Input, rid string) string {
	if in.IndexURL == "" && len(in.More) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="lj-more" id="acpwb-src-%s-more">`, rid)
	if in.Prev != nil || in.Next != nil {
		b.WriteString(`<div class="lj-prevnext">`)
		for i, l := range []*MoreLink{in.Prev, in.Next} {
			label := "&larr; Previous job"
			cls := ""
			if i == 1 {
				label, cls = "Next job &rarr;", " lj-next"
			}
			if l == nil {
				b.WriteString(`<span></span>`)
				continue
			}
			fmt.Fprintf(&b, `<a class="lj-pn%s" href="%s"><small>%s &middot; %s</small><span class="lj-pn-prog">%s</span><span class="lj-pn-title">%s</span></a>`,
				cls, esc(l.URL), label, esc(l.Date), esc(l.Program+" ("+l.Language+")"), esc(l.Title))
		}
		b.WriteString(`</div>`)
	}
	if len(in.More) > 0 {
		b.WriteString(`<h2 class="lj-more-h">Related software projects</h2><div class="lj-more-grid">`)
		for _, g := range in.More {
			fmt.Fprintf(&b, `<div class="lj-card lj-more-group"><h6>%s</h6><ul>`, esc(g.Title))
			for _, l := range g.Links {
				fmt.Fprintf(&b, `<li><a href="%s"><span class="lj-more-prog">%s</span><span class="lj-more-title">%s</span></a><small>%s &middot; %s &middot; %s</small></li>`,
					esc(l.URL), esc(l.Program), esc(l.Title), esc(l.Language), esc(l.Date), esc(l.Job))
			}
			b.WriteString(`</ul>`)
			if g.URL != "" {
				fmt.Fprintf(&b, `<a class="lj-more-all" href="%s">%s &rarr;</a>`, esc(g.URL), esc(g.More))
			}
			b.WriteString(`</div>`)
		}
		b.WriteString(`</div>`)
	}
	if in.IndexURL != "" {
		fmt.Fprintf(&b, `<div class="lj-browse">Browse: <a href="%s">All source code</a> &middot; <a href="%s">%d programs</a> &middot; <a href="%s">%s %d</a>`,
			esc(in.IndexURL), esc(in.YearURL), in.Year, esc(in.MonthURL), monthFull[in.Month-1], in.Year)
		for _, dy := range []int{-1, 1} {
			y := in.Year + dy
			if y >= 1985 && y <= 2025 {
				fmt.Fprintf(&b, ` &middot; <a href="%s%d/">%d</a>`, esc(in.IndexURL), y, y)
			}
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

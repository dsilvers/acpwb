package legacyjob

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	fileSepRe  = regexp.MustCompile(`^(?:#|//) ---- (.+?) ----$`)
	jclStepRe  = regexp.MustCompile(`^//(STEP\d+)\s+EXEC\s+(?:PGM=)?(\S+?)(?:,|$)`)
	batLabelRe = regexp.MustCompile(`^:([A-Z]+)$`)
	dclLabelRe = regexp.MustCompile(`^\$ ([A-Z]+):$`)
	slugRe     = regexp.MustCompile(`[^a-z0-9]+`)
)

func anchorID(s string) string {
	return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// annotate finds the jump targets inside each pre section: the files that
// make up the job control, every routine in the source listing, the card
// groups of the input deck and the report blocks of the output.
func annotate(l *Listing, rep *report) {
	for si := range l.Sections {
		s := &l.Sections[si]
		if s.Kind != "pre" {
			continue
		}
		lines := strings.Split(s.Text, "\n")
		seen := map[string]bool{}
		add := func(i int, label string) {
			id := s.ID + "-" + anchorID(label)
			if seen[id] {
				id = fmt.Sprintf("%s-%d", id, i)
			}
			seen[id] = true
			s.Anchors = append(s.Anchors, Anchor{Line: i, ID: id, Label: label})
		}
		switch s.ID {
		case "jobcontrol":
			first := strings.SplitN(s.Tab, " + ", 2)[0]
			add(0, first)
			for i, ln := range lines {
				t := stripMarkers(ln)
				if m := fileSepRe.FindStringSubmatch(t); m != nil {
					add(i, m[1])
				} else if m := jclStepRe.FindStringSubmatch(t); m != nil {
					add(i, m[1]+" — "+m[2])
				} else if m := batLabelRe.FindStringSubmatch(t); m != nil {
					add(i, ":"+m[1])
				} else if m := dclLabelRe.FindStringSubmatch(t); m != nil {
					add(i, m[1]+":")
				}
			}
		case "source":
			for i, ln := range lines {
				if m := routineRe.FindStringSubmatch(ln); m != nil {
					add(i, strings.ToUpper(m[1]))
				}
			}
		case "sysin":
			add(0, "Client and record card")
			if len(lines) > 3 {
				add(3, "Engagement metric cards")
			}
			if n := 3 + len(rep.MetNames); n < len(lines) {
				add(n, "Benchmark median cards")
			}
		case "sysprint":
			add(1, "Engagement metrics")
			if n := 3 + len(rep.MetNames); n < len(lines) {
				add(n, "Market benchmark")
			}
			add(len(lines)-1, "Run totals and seal")
		}
	}
}

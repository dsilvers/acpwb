package legacyjob

import (
	"fmt"
	"strings"
)

// jobPaths is the per-page filesystem / dataset layout, owned by one person
// from the page's roster and shaped by era.
type jobPaths struct {
	owner     *Coder
	user      string // unix/VMS username
	tso       string // MVS userid (7 chars)
	hlq       string // MVS high-level qualifier
	dept      string // short department token
	host      string // batch host name
	unixHome  string
	unixWork  string
	unixShare string
	unixLog   string
	dosDir    string
	dosShare  string
	dosVolume string
	vmsDir    string
	jenkinsWS string
	container string
}

func alnum(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func buildPaths(j *job) *jobPaths {
	x := j.x.fork("paths")
	in := j.in
	ros := j.g.roster
	owner := ros[len(ros)-1]
	for _, c := range sample(x, ros, len(ros)) {
		if c.First != "" && c.Start <= in.Year && c.End >= in.Year-2 {
			owner = c
			break
		}
	}
	first, last := alnum(owner.First), alnum(owner.Last)
	if first == "" {
		first = "y2k"
	}
	if last == "" {
		last = "team"
	}
	p := &jobPaths{owner: owner}

	// username style by era
	switch {
	case in.Year < 1996:
		p.user = trunc(first[:1]+last, 8)
	case in.Year < 2008:
		p.user = pick(x, []string{trunc(first[:1]+last, 8), trunc(last+first[:1], 8), first[:1] + strings.ToLower(owner.Initials[len(owner.Initials)-1:]) + fmt.Sprint(x.between(1, 99))})
	default:
		p.user = pick(x, []string{first + "." + last, first[:1] + last, last + first[:1]})
	}
	p.tso = trunc(strings.ToUpper(first[:1]+last), 7)
	deptName := pick(x, getLibrary().names["depts"])
	var di strings.Builder
	for _, w := range strings.Fields(deptName) {
		if w[0] >= 'A' && w[0] <= 'Z' {
			di.WriteByte(w[0])
		}
	}
	p.dept = strings.ToLower(trunc(di.String(), 4))
	if len(p.dept) < 2 {
		p.dept = strings.ToLower(trunc(alnum(deptName), 4))
	}
	p.hlq = pick(x, []string{p.tso, strings.ToUpper(p.dept) + "PROD", "RSCH" + strings.ToUpper(p.dept), "ACP" + strings.ToUpper(trunc(p.dept, 3)), p.tso})
	prog := strings.ToLower(j.id.Program)
	client := alnum(strings.Fields(in.Org + " client")[0])

	p.host = pick(x, []string{
		fmt.Sprintf("%s%02d", trunc(p.dept, 4), x.between(1, 12)),
		pick(x, []string{"mercury", "jupiter", "saturn", "pluto", "zeus", "athena", "odin", "loki", "bigvax", "badger", "walleye", "cheddar", "brewer"}),
		fmt.Sprintf("mke-%s-%02d", pick(x, []string{"batch", "util", "app", "rpt"}), x.between(1, 9)),
	})

	// unix homes by era
	switch {
	case in.Year < 1996:
		p.unixHome = pick(x, []string{"/usr/users/", "/u/", "/usr/people/", "/home/"}) + p.user
	case in.Year < 2008:
		p.unixHome = pick(x, []string{"/export/home/", "/home/", "/u01/home/", "/users/"}) + p.user
	default:
		p.unixHome = pick(x, []string{"/home/", "/home/", "/srv/home/", "/data/home/"}) + p.user
	}
	p.unixWork = pick(x, []string{
		p.unixHome + "/" + pick(x, []string{"src", "work", "projects", "bench", "jobs"}) + "/" + prog,
		fmt.Sprintf("/%s/%s/%s", pick(x, []string{"u01", "u02", "data", "app", "opt"}), p.dept, prog),
		fmt.Sprintf("/data/%s/%d/%s", p.dept, in.Year, prog),
		fmt.Sprintf("/usr/local/%s/%s", p.dept, prog),
	})
	p.unixShare = pick(x, []string{
		fmt.Sprintf("/net/%s/export/%s", p.host, p.dept),
		fmt.Sprintf("/mnt/%s/%s", pick(x, []string{"share", "reports", "nas01", "filer"}), p.dept),
		fmt.Sprintf("/shared/%s/%s", p.dept, client),
	})
	p.unixLog = pick(x, []string{"/var/log/" + p.dept, "/var/adm/" + p.dept, p.unixHome + "/logs", "/tmp"})

	drive := pick(x, []string{"C", "C", "D", "E"})
	p.dosDir = fmt.Sprintf("%s:\\%s\\%s", drive, pick(x, []string{strings.ToUpper(trunc(last, 8)), strings.ToUpper(p.dept), "WORK\\" + strings.ToUpper(trunc(last, 8)), "PROJECTS"}), j.id.Program)
	p.dosShare = fmt.Sprintf("%s:\\%s\\%s", pick(x, []string{"F", "G", "H", "S"}), pick(x, []string{"SHARED", "PUBLIC", "REPORTS", strings.ToUpper(p.dept)}), strings.ToUpper(trunc(client, 8)))
	p.dosVolume = strings.ToUpper(trunc(pick(x, []string{last, p.dept + "_DISK", first + "_PC", "WORK"}), 11))

	p.vmsDir = fmt.Sprintf("%s:[%s.%s]", pick(x, []string{"DUA0", "DUA1", "DKA100", "$1$DUA4", "USER$DISK", "SYS$USERS"}), strings.ToUpper(trunc(p.user, 12)), pick(x, []string{j.id.Program, "BENCH." + j.id.Program, "SRC", strings.ToUpper(p.dept) + "." + j.id.Program}))

	p.jenkinsWS = fmt.Sprintf("/var/lib/jenkins/workspace/%s", pick(x, []string{prog, p.dept + "-" + prog, prog + "-" + client, "legacy/" + prog}))
	p.container = fmt.Sprintf("/opt/%s/%s", pick(x, []string{p.dept, client, "legacy", "app"}), prog)
	return p
}

func (p *jobPaths) dsn(kind string) string {
	return p.hlq + "." + kind
}

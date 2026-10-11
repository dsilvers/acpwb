package legacyjob

import (
	"fmt"
	"strconv"
	"strings"
)

// coreOpts are the per-page structural choices for the core computation.
// Every combination computes the same published values.
type coreOpts struct {
	split         bool // core steps as separate routines
	reverseDelta  bool
	reverseLadder bool
	unrollLadder  bool
	multStyle     int // 0 DATA array, 1 derived from integer table, 2 inline literals, 3 permuted table
	temps         bool
	identity      bool
	fuseDelta     bool
	checks        bool
	xcheck        bool
	checksum      bool
	debug         bool
	clientBranch  bool
}

type builder struct {
	g     *gen
	x     *rnd
	nm    *naming
	in    *Input
	id    ID
	prog  *Program
	opts  coreOpts
	dens  float64
	pn    map[string]string // logical routine -> emitted name
	big   int
	sealH int
	sealL int
	ckMul int
	ckMod int
}

func newBuilder(g *gen, x *rnd, nm *naming) *builder {
	b := &builder{g: g, x: x, nm: nm, in: g.in, id: g.id, pn: map[string]string{}}
	b.opts = coreOpts{
		split:         x.chance(0.75),
		reverseDelta:  x.chance(0.25),
		reverseLadder: x.chance(0.25),
		unrollLadder:  x.chance(0.3),
		multStyle:     x.intn(4),
		temps:         x.chance(0.5),
		identity:      x.chance(0.4),
		fuseDelta:     x.chance(0.3),
		checks:        x.chance(0.8),
		xcheck:        x.chance(0.5),
		checksum:      true,
		debug:         x.chance(0.85),
		clientBranch:  x.chance(0.5),
	}
	if b.opts.multStyle == 2 {
		b.opts.unrollLadder = true
	}
	b.dens = []float64{0.03, 0.05, 0.08, 0.12, 0.18}[x.intn(5)]
	b.big = pick(x, []int{999999, 9999999, 500000, 2000000})
	tok := strings.ToUpper(g.in.Token)
	if len(tok) == 8 {
		h, _ := strconv.ParseInt(tok[:4], 16, 64)
		l, _ := strconv.ParseInt(tok[4:], 16, 64)
		b.sealH, b.sealL = int(h), int(l)
	}
	return b
}

func (b *builder) v(role string) *Expr { return V(b.nm.get(role)) }

func (b *builder) build() *Program {
	nm := b.nm
	maxMet := pick(b.x, []int{40, 50, 64, 99, 100})
	maxBen := pick(b.x, []int{40, 50, 64, 99, 100})
	p := &Program{Name: b.id.Program}
	b.prog = p

	cblk := func() string {
		return pick(b.x, []string{"RUNCTL", "WORK", "TABLES", "DECK", "LADDER", "CTLBLK", "MAINWK"})
	}
	dataBlk := cblk()
	ctlBlk := cblk()
	for ctlBlk == dataBlk {
		ctlBlk = cblk()
	}
	sealBlk := "W" + strings.ToUpper(b.in.Token)[:5]
	p.Commons = []string{dataBlk, ctlBlk, sealBlk}

	ladder := make([]string, 8)
	copy(ladder, ladderLit[:])
	nums := make([]string, 8)
	for i, n := range ladderNum {
		nums[i] = strconv.Itoa(n)
	}

	add := func(s Sym) { p.Globals = append(p.Globals, s) }
	add(Sym{Name: nm.get("maxmet"), Type: "int", Role: "const", Init: []string{strconv.Itoa(maxMet)}})
	add(Sym{Name: nm.get("maxben"), Type: "int", Role: "const", Init: []string{strconv.Itoa(maxBen)}})
	add(Sym{Name: nm.get("nladr"), Type: "int", Role: "const", Init: []string{"8"}})
	add(Sym{Name: nm.get("nmet"), Type: "int", Common: ctlBlk})
	add(Sym{Name: nm.get("nben"), Type: "int", Common: ctlBlk})
	add(Sym{Name: nm.get("bsln"), Type: "int", Dims: []int{maxMet}, Common: dataBlk, Role: "bsln"})
	add(Sym{Name: nm.get("curr"), Type: "int", Dims: []int{maxMet}, Common: dataBlk, Role: "curr"})
	add(Sym{Name: nm.get("delt"), Type: "int", Dims: []int{maxMet}, Common: dataBlk, Role: "delt"})
	add(Sym{Name: nm.get("medn"), Type: "int", Dims: []int{maxBen}, Common: dataBlk, Role: "medn"})
	add(Sym{Name: nm.get("pctl"), Type: "int", Dims: []int{8, maxBen}, Common: dataBlk, Role: "pctl"})
	switch b.opts.multStyle {
	case 1:
		add(Sym{Name: nm.get("mult"), Type: "dbl", Dims: []int{8}, Common: ctlBlk, Role: "mult"})
		add(Sym{Name: nm.fresh("int"), Type: "int", Dims: []int{8}, Init: nums, Role: "inum"})
	case 3:
		perm := []int{0, 1, 2, 3, 4, 5, 6, 7}
		shuffle(b.x, perm)
		permuted := make([]string, 8)
		idx := make([]string, 8)
		for slot, k := range perm {
			permuted[slot] = ladder[k]
		}
		for k := 0; k < 8; k++ {
			for slot, kk := range perm {
				if kk == k {
					idx[k] = strconv.Itoa(slot + 1)
				}
			}
		}
		add(Sym{Name: nm.get("mult"), Type: "dbl", Dims: []int{8}, Init: permuted, Role: "mult"})
		add(Sym{Name: nm.fresh("int"), Type: "int", Dims: []int{8}, Init: idx, Role: "perm"})
	case 0:
		add(Sym{Name: nm.get("mult"), Type: "dbl", Dims: []int{8}, Init: ladder, Role: "mult"})
	}
	add(Sym{Name: nm.get("mname"), Type: "str", Len: 40, Dims: []int{maxMet}, Common: dataBlk, Role: "mname"})
	add(Sym{Name: nm.get("bname"), Type: "str", Len: 40, Dims: []int{maxBen}, Common: dataBlk, Role: "bname"})
	add(Sym{Name: nm.get("org"), Type: "str", Len: 40, Common: ctlBlk, Role: "org"})
	add(Sym{Name: nm.get("recid"), Type: "str", Len: 8, Common: ctlBlk, Role: "recid"})
	add(Sym{Name: nm.get("peer"), Type: "str", Len: 60, Common: ctlBlk, Role: "peer"})
	add(Sym{Name: nm.get("seal"), Type: "str", Len: 8, Init: []string{strings.ToUpper(b.in.Token)}, Common: sealBlk, Role: "seal"})
	add(Sym{Name: nm.get("idebug"), Type: "int", Common: ctlBlk, Init: []string{"0"}, Role: "idebug"})
	add(Sym{Name: nm.get("ipage"), Type: "int", Common: ctlBlk})
	add(Sym{Name: nm.get("iline"), Type: "int", Common: ctlBlk})
	add(Sym{Name: nm.get("ierr"), Type: "int", Common: ctlBlk})
	add(Sym{Name: nm.get("chksum"), Type: "int", Common: sealBlk})
	add(Sym{Name: nm.get("tmp"), Type: "int", Common: ctlBlk})
	if b.opts.xcheck {
		add(Sym{Name: nm.get("shadow"), Type: "int", Dims: []int{8, maxBen}, Common: dataBlk, Role: "shadow"})
	}

	for _, s := range p.Globals {
		if s.Role != "const" {
			b.g.vars = append(b.g.vars, s.Name)
		}
	}

	// Routine names.
	stems := map[string][]string{
		"init":     {"INIT", "SETUP", "CLRWRK", "HSKPG", "START", "ZEROIZ"},
		"read":     {"RDDECK", "READIN", "GETCRD", "LOADCD", "INCARD", "RDCTL"},
		"validate": {"VALID", "CHKIN", "EDITCK", "VERIFY", "SANITY", "SCREEN"},
		"delta":    {"DELTAS", "CHANGE", "DIFFS", "CMPDLT", "PERCHG", "DLTCLC"},
		"ladder":   {"LADDER", "PCTLAD", "APPLAD", "RUNGS", "QUANTL", "TIERS"},
		"xcheck":   {"XCHECK", "CROSCK", "RECHEK", "DUPCHK", "VERLAD", "PROOF"},
		"chksum":   {"CHKSUM", "DECKCK", "HASHIT", "SEALCK", "CKDECK", "AUDSUM"},
		"print":    {"REPORT", "PRTOUT", "PRINT1", "OUTRPT", "WRTRPT", "LISTIT"},
		"errdsp":   {"ERRDSP", "ABEND1", "ERRMSG", "OOPS", "FAULT", "ERRTRP"},
		"debug":    {"DUMPIT", "DBGOUT", "TRACE", "SNAP", "PEEK", "DIAGPR"},
	}
	for _, k := range []string{"init", "read", "validate", "delta", "ladder", "xcheck", "chksum", "print", "errdsp", "debug"} {
		b.pn[k] = nm.routine(pick(b.x, stems[k]))
	}

	// Dead routines first (so their names are known to the comment filler
	// and core comments can reference them).
	deadProcs := b.deadRoutines()
	for _, d := range deadProcs {
		b.g.routines = append(b.g.routines, d.Name)
	}
	for _, k := range []string{"init", "read", "validate", "delta", "ladder", "chksum", "print", "errdsp"} {
		b.g.routines = append(b.g.routines, b.pn[k])
	}

	core := b.coreProcs()
	mainBody := []*Stmt{}
	for _, step := range core.order {
		if b.opts.split {
			mainBody = append(mainBody, Call(b.pn[step]))
		} else {
			mainBody = append(mainBody, b.banner(sectionTitle(step), nil)...)
			mainBody = append(mainBody, core.bodies[step]...)
		}
	}
	if b.opts.clientBranch {
		mainBody = append(mainBody, b.clientBranch(deadProcs)...)
	}
	mainBody = append(mainBody, &Stmt{K: "stop"})

	mainAuthor := b.g.roster[0]
	mainProc := &Proc{Name: b.id.Program, Body: b.sprinkle(mainBody, mainAuthor, 0), Purpose: "Main line", Author: mainAuthor, Core: true}
	p.Procs = append(p.Procs, mainProc)

	if b.opts.split {
		for _, step := range core.order {
			a := b.g.coder()
			p.Procs = append(p.Procs, &Proc{
				Name: b.pn[step], Body: b.sprinkle(core.bodies[step], a, 0),
				Purpose: corePurpose[step], Author: a, Core: true,
			})
		}
	}
	errAuthor := b.g.coder()
	icode := nm.get("icode")
	p.Procs = append(p.Procs, &Proc{Name: b.pn["errdsp"], Params: []string{icode}, Locals: []Sym{{Name: icode, Type: "int", Role: "param"}},
		Body: b.errDispatcher(icode), Purpose: corePurpose["errdsp"], Author: errAuthor, Core: true})
	if b.opts.debug {
		a := b.g.coder()
		p.Procs = append(p.Procs, &Proc{Name: b.pn["debug"], Body: b.sprinkle(b.debugDump(), a, 0), Purpose: corePurpose["debug"], Author: a, Core: true})
	}

	b.rewrite(p)

	// Interleave dead routines among the core ones.
	for _, d := range deadProcs {
		pos := 1 + b.x.intn(len(p.Procs))
		p.Procs = append(p.Procs[:pos], append([]*Proc{d}, p.Procs[pos:]...)...)
	}
	return p
}

var corePurpose = map[string]string{
	"init":     "Clears the work areas and control counters. Clears some of them twice.",
	"read":     "Reads the control cards, the engagement metric cards and the benchmark median cards.",
	"validate": "Validates every numeric field against the agreed sanity limits.",
	"delta":    "Computes the period-over-period change for each engagement metric.",
	"ladder":   "Applies the published percentile ladder to each benchmark median. Truncates. Never rounds.",
	"xcheck":   "Recomputes the ladder a second time and compares. Has never found a difference.",
	"chksum":   "Computes the deck checksum and the provenance seal for the audit trail.",
	"print":    "Prints the engagement metrics table and the market benchmark table.",
	"errdsp":   "Dispatches error codes to the appropriate message and terminates the job.",
	"debug":    "Dumps the work arrays when the debug flag is set. The debug flag is never set.",
}

func sectionTitle(step string) string {
	return map[string]string{
		"init": "INITIALIZATION", "read": "READ INPUT DECK", "validate": "VALIDATE INPUT",
		"delta": "COMPUTE PERIOD-OVER-PERIOD CHANGE", "ladder": "APPLY PERCENTILE LADDER",
		"xcheck": "CROSS-CHECK LADDER", "chksum": "DECK CHECKSUM AND PROVENANCE SEAL", "print": "PRINT REPORT",
	}[step]
}

type coreSet struct {
	order  []string
	bodies map[string][]*Stmt
}

func (b *builder) coreProcs() coreSet {
	nm := b.nm
	i, j, k := nm.get("i"), nm.get("j"), nm.get("k")
	bsln, curr, delt := nm.get("bsln"), nm.get("curr"), nm.get("delt")
	medn, pctl := nm.get("medn"), nm.get("pctl")
	nmet, nben := b.v("nmet"), b.v("nben")
	cs := coreSet{bodies: map[string][]*Stmt{}}

	// init
	init := []*Stmt{
		Assign(b.v("idebug"), N(0)), Assign(b.v("ipage"), N(0)), Assign(b.v("iline"), N(0)),
		Assign(b.v("ierr"), N(0)), Assign(b.v("chksum"), N(0)),
		Do(i, N(1), b.v("maxmet"), 1,
			Assign(Ix(bsln, V(i)), N(0)), Assign(Ix(curr, V(i)), N(0)), Assign(Ix(delt, V(i)), N(0))),
		Do(j, N(1), b.v("maxben"), 1,
			Assign(Ix(medn, V(j)), N(0)),
			Do(k, N(1), b.v("nladr"), 1, Assign(Ix(pctl, V(k), V(j)), N(0)))),
	}
	if b.x.chance(0.4) { // clears twice, per the comment
		init = append(init, Do(i, N(1), b.v("maxmet"), 1, Assign(Ix(delt, V(i)), N(0))))
	}
	if b.opts.multStyle == 1 {
		inum := b.symByRole("inum")
		init = append(init, Do(k, N(1), b.v("nladr"), 1,
			Assign(Ix(nm.get("mult"), V(k)), Bin("/", Fn("dble", Ix(inum, V(k))), DL("100.0")))))
	}
	cs.bodies["init"] = init

	// read
	metLoop := []*Stmt{IO("readmet", V(i))}
	if b.opts.fuseDelta {
		metLoop = append(metLoop, Assign(Ix(delt, V(i)), Bin("-", Ix(curr, V(i)), Ix(bsln, V(i)))))
	}
	read := []*Stmt{
		IO("readhdr"),
		Do(i, N(1), nmet, 1, metLoop...),
		Do(j, N(1), nben, 1, IO("readben", V(j))),
	}
	cs.bodies["read"] = read

	// validate
	var val []*Stmt
	val = append(val,
		If(C("gt", nmet, b.v("maxmet")), []*Stmt{Call(b.pn["errdsp"], N(3))}, nil),
		If(C("gt", nben, b.v("maxben")), []*Stmt{Call(b.pn["errdsp"], N(3))}, nil),
	)
	if b.opts.checks {
		val = append(val,
			Do(i, N(1), nmet, 1,
				If(Or(C("lt", Ix(bsln, V(i)), N(0)), C("gt", Ix(bsln, V(i)), N(b.big))), []*Stmt{Call(b.pn["errdsp"], N(1))}, nil),
				If(Or(C("lt", Ix(curr, V(i)), N(0)), C("gt", Ix(curr, V(i)), N(b.big))), []*Stmt{Call(b.pn["errdsp"], N(1))}, nil)),
			Do(j, N(1), nben, 1,
				If(C("le", Ix(medn, V(j)), N(0)), []*Stmt{Call(b.pn["errdsp"], N(2))}, nil)),
		)
	}
	cs.bodies["validate"] = val

	// delta
	from, to, step := N(1), nmet, 1
	if b.opts.reverseDelta {
		from, to, step = nmet, N(1), -1
	}
	var dBody []*Stmt
	rhs := Bin("-", Ix(curr, V(i)), Ix(bsln, V(i)))
	if b.opts.identity {
		rhs = Bin("+", rhs, N(0))
	}
	if b.opts.temps {
		dBody = []*Stmt{Assign(b.v("tmp"), Ix(curr, V(i))), Assign(Ix(delt, V(i)), Bin("-", b.v("tmp"), Ix(bsln, V(i))))}
	} else {
		dBody = []*Stmt{Assign(Ix(delt, V(i)), rhs)}
	}
	cs.bodies["delta"] = []*Stmt{Do(i, from, to, step, dBody...)}

	// ladder
	cs.bodies["ladder"] = b.ladderStmts(pctl)

	// cross-check
	if b.opts.xcheck {
		sh := nm.get("shadow")
		xs := b.ladderStmts(sh)
		xs = append(xs, Do(j, N(1), nben, 1, Do(k, N(1), b.v("nladr"), 1,
			If(C("ne", Ix(pctl, V(k), V(j)), Ix(sh, V(k), V(j))), []*Stmt{Call(b.pn["errdsp"], N(4))}, nil))))
		cs.bodies["xcheck"] = xs
	}

	// checksum + watermark seal halves
	ck := b.v("chksum")
	b.ckMod = pick(b.x, []int{65521, 32749, 99991, 9973})
	b.ckMul = pick(b.x, []int{31, 37, 131, 7, 17})
	mod, mul := N(b.ckMod), N(b.ckMul)
	cs.bodies["chksum"] = []*Stmt{
		Assign(ck, N(b.sealH)),
		Do(i, N(1), nmet, 1, Assign(ck, Fn("mod", Bin("+", Bin("*", ck, mul), Bin("+", Ix(bsln, V(i)), Ix(curr, V(i)))), mod))),
		Do(j, N(1), nben, 1, Assign(ck, Fn("mod", Bin("+", Bin("*", ck, mul), Ix(medn, V(j))), mod))),
		Assign(ck, Fn("mod", Bin("+", ck, N(b.sealL)), mod)),
	}

	// print
	pr := []*Stmt{
		Assign(b.v("ipage"), Bin("+", b.v("ipage"), N(1))),
		IO("pagehdr"), IO("title1"), IO("colhdr1"),
		Do(i, N(1), nmet, 1, IO("rowmet", V(i)), Assign(b.v("iline"), Bin("+", b.v("iline"), N(1)))),
		IO("title2"), IO("colhdr2"),
		Do(j, N(1), nben, 1, IO("rowben", V(j)), Assign(b.v("iline"), Bin("+", b.v("iline"), N(1)))),
		IO("trailer"),
	}
	if b.opts.debug {
		pr = append(pr, If(C("eq", b.v("idebug"), N(1)), []*Stmt{Call(b.pn["debug"])}, nil))
	}
	cs.bodies["print"] = pr

	cs.order = []string{"init", "read", "validate", "delta", "ladder"}
	if b.opts.fuseDelta {
		cs.order = []string{"init", "read", "validate", "ladder"}
	}
	if b.opts.xcheck {
		cs.order = append(cs.order, "xcheck")
	}
	cs.order = append(cs.order, "chksum", "print")
	return cs
}

func (b *builder) symByRole(role string) string {
	for _, s := range b.prog.Globals {
		if s.Role == role {
			return s.Name
		}
	}
	return role
}

func (b *builder) multExpr(k *Expr, kconst int) *Expr {
	mult := b.nm.get("mult")
	switch b.opts.multStyle {
	case 2:
		if b.x.chance(0.5) {
			return Bin("/", DL(strconv.Itoa(ladderNum[kconst])+".0"), DL("100.0"))
		}
		return DL(ladderLit[kconst])
	case 3:
		return Ix(mult, Ix(b.symByRole("perm"), k))
	default:
		return Ix(mult, k)
	}
}

func (b *builder) ladderStmts(target string) []*Stmt {
	nm := b.nm
	j, k := nm.get("j"), nm.get("k")
	medn := nm.get("medn")
	from, to, step := N(1), b.v("nben"), 1
	if b.opts.reverseLadder {
		from, to, step = b.v("nben"), N(1), -1
	}
	var inner []*Stmt
	if b.opts.unrollLadder {
		for kk := 0; kk < 8; kk++ {
			inner = append(inner, Assign(Ix(target, N(kk+1), V(j)), Ladder(Ix(medn, V(j)), b.multExpr(N(kk+1), kk))))
		}
	} else {
		val := Ladder(Ix(medn, V(j)), b.multExpr(V(k), 0))
		inner = []*Stmt{Do(k, N(1), b.v("nladr"), 1, Assign(Ix(target, V(k), V(j)), val))}
	}
	return []*Stmt{Do(j, from, to, step, inner...)}
}

func (b *builder) errDispatcher(icode string) []*Stmt {
	var out []*Stmt
	for code := 1; code <= 4; code++ {
		out = append(out, If(C("eq", V(icode), N(code)), []*Stmt{IO("errmsg", N(code))}, nil))
	}
	out = append(out, Assign(b.v("ierr"), V(icode)), &Stmt{K: "stop"})
	return out
}

func (b *builder) debugDump() []*Stmt {
	nm := b.nm
	i, j := nm.get("i"), nm.get("j")
	return []*Stmt{
		IO("debugmsg", N(100)),
		Do(i, N(1), b.v("nmet"), 1, IO("debugmet", V(i))),
		Do(j, N(1), b.v("nben"), 1, IO("debugben", V(j))),
		IO("debugmsg", N(101)),
	}
}

// clientBranch: dead "special handling" for the page's client, guarded by a
// debug value that is never set.
func (b *builder) clientBranch(dead []*Proc) []*Stmt {
	var body []*Stmt
	if len(dead) > 0 {
		d := pick(b.x, dead)
		if len(d.Params) == 0 {
			body = append(body, Call(d.Name))
		}
	}
	body = append(body, Assign(b.v("ierr"), N(b.x.between(10, 99))))
	text := fmt.Sprintf("Special handling for %s (engagement %s). Retained per %s.", b.in.Org, b.in.EngCode, b.g.memo())
	return []*Stmt{Comment(text), If(C("eq", b.v("idebug"), N(b.x.between(2, 9))), body, nil)}
}

// banner returns a boxed section comment.
func (b *builder) banner(title string, author *Coder) []*Stmt {
	if author == nil {
		author = b.g.coder()
	}
	text := ""
	if b.x.chance(0.3) {
		text = b.g.comment(author)
	}
	return []*Stmt{{K: "comment", Banner: true, Title: title, Text: text}}
}

// sprinkle inserts generated comments between statements (and inside loop /
// IF bodies) at the page's comment density.
func (b *builder) sprinkle(stmts []*Stmt, author *Coder, depth int) []*Stmt {
	var out []*Stmt
	for _, s := range stmts {
		if s.K != "comment" && b.x.float() < b.dens {
			a := author
			if b.x.chance(0.35) {
				a = b.g.coder()
			}
			out = append(out, Comment(b.g.comment(a)))
			if b.dens > 0.15 && b.x.chance(0.15) {
				out = append(out, Comment(b.g.comment(a)))
			}
		}
		if depth < 3 {
			s.Body = b.sprinkle(s.Body, author, depth+1)
			s.Then = b.sprinkle(s.Then, author, depth+1)
			s.Else = b.sprinkle(s.Else, author, depth+1)
		}
		out = append(out, s)
	}
	return out
}

// ── Dead code ───────────────────────────────────────────────────────────────

type deadSpec struct {
	stems   []string
	purpose string
	build   func(b *builder, p *Proc)
}

func (b *builder) deadRoutines() []*Proc {
	specs := append(deadLibrary(), extraDeadLibrary()...)
	n := b.x.between(3, 10)
	var out []*Proc
	for _, sp := range sample(b.x, specs, n) {
		a := b.g.coder()
		p := &Proc{Name: b.nm.routine(pick(b.x, sp.stems)), Purpose: sp.purpose, Author: a, Dead: true}
		sp.build(b, p)
		p.Body = b.sprinkle(p.Body, a, 0)
		out = append(out, p)
	}
	// Plus 1-5 entirely synthetic routines.
	for r := b.x.between(1, 5); r > 0; r-- {
		a := b.g.coder()
		p := &Proc{Name: b.nm.routine(pick(b.x, []string{"UTIL", "WRKRTN", "SUBR", "AUXCLC", "MISC", "OLDRTN", "TEMP", "SPARE", "PATCH", "XTRA"}) + strconv.Itoa(b.x.between(1, 9))),
			Purpose: b.g.comment(a, "routine_purpose", "history"), Author: a, Dead: true}
		b.randomBody(p, out)
		p.Body = b.sprinkle(p.Body, a, 0)
		out = append(out, p)
	}
	return out
}

func (b *builder) local(p *Proc, typ string, dims ...int) string {
	n := b.nm.fresh(typ)
	p.Locals = append(p.Locals, Sym{Name: n, Type: typ, Dims: dims})
	return n
}

func (b *builder) param(p *Proc, typ string, dims ...int) string {
	n := b.nm.fresh(typ)
	p.Params = append(p.Params, n)
	p.Locals = append(p.Locals, Sym{Name: n, Type: typ, Dims: dims, Role: "param"})
	return n
}

func deadLibrary() []deadSpec {
	return []deadSpec{
		{[]string{"RANDU", "RNDGEN", "URAND", "RANDOM"}, "Generates pseudo-random numbers using the IBM RANDU method. Disabled per audit.", func(b *builder, p *Proc) {
			s, v := b.param(p, "int"), b.param(p, "int")
			lo, hi := b.x.between(-20, -5), b.x.between(20, 60)
			p.Body = []*Stmt{
				Assign(V(s), Fn("mod", Bin("*", V(s), N(65539)), N(2147483647))),
				If(C("lt", V(s), N(0)), []*Stmt{Assign(V(s), Bin("+", V(s), N(2147483647)))}, nil),
				Assign(V(v), Bin("+", Fn("mod", V(s), N(hi-lo+1)), N(lo))),
			}
		}},
		{[]string{"LEAPYR", "LEAPCK", "ISLEAP", "LYCHK"}, "Determines whether a year is a leap year. Certified for 2000.", func(b *builder, p *Proc) {
			y, l := b.param(p, "int"), b.param(p, "int")
			p.Body = []*Stmt{
				Assign(V(l), N(0)),
				If(C("eq", Fn("mod", V(y), N(4)), N(0)), []*Stmt{Assign(V(l), N(1))}, nil),
				If(C("eq", Fn("mod", V(y), N(100)), N(0)), []*Stmt{Assign(V(l), N(0))}, nil),
				If(C("eq", Fn("mod", V(y), N(400)), N(0)), []*Stmt{Assign(V(l), N(1))}, nil),
			}
		}},
		{[]string{"JULIAN", "JULDAT", "TOJUL", "YYDDD"}, "Converts a calendar date to Julian (YYDDD) format for the page header.", func(b *builder, p *Proc) {
			y, m, d, out := b.param(p, "int"), b.param(p, "int"), b.param(p, "int"), b.param(p, "int")
			cum := b.local(p, "int", 12)
			ii := b.local(p, "int")
			p.Locals[len(p.Locals)-2].Init = []string{"0", "31", "59", "90", "120", "151", "181", "212", "243", "273", "304", "334"}
			p.Body = []*Stmt{
				Assign(V(ii), Ix(cum, V(m))),
				Assign(V(out), Bin("+", Bin("*", Fn("mod", V(y), N(100)), N(1000)), Bin("+", V(ii), V(d)))),
				If(And(C("gt", V(m), N(2)), C("eq", Fn("mod", V(y), N(4)), N(0))), []*Stmt{Assign(V(out), Bin("+", V(out), N(1)))}, nil),
			}
		}},
		{[]string{"ORDSFX", "SUFFIX", "NTHSFX", "ORDNAL"}, "Produces the ordinal suffix code for a percentile label. Handles 1, 2 and 3. Everything else is TH.", func(b *builder, p *Proc) {
			n, c := b.param(p, "int"), b.param(p, "int")
			m := b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(c), N(4)),
				Assign(V(m), Fn("mod", V(n), N(10))),
				If(C("eq", V(m), N(1)), []*Stmt{Assign(V(c), N(1))}, nil),
				If(C("eq", V(m), N(2)), []*Stmt{Assign(V(c), N(2))}, nil),
				If(C("eq", V(m), N(3)), []*Stmt{Assign(V(c), N(3))}, nil),
			}
		}},
		{[]string{"SORTIX", "BUBSRT", "SRTIDX", "ORDERX"}, "Sorts the index table into ascending order. The table is always already in order.", func(b *builder, p *Proc) {
			arr, n := b.param(p, "int", 40), b.param(p, "int")
			ii, jj, t := b.local(p, "int"), b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Do(ii, N(1), Bin("-", V(n), N(1)), 1,
					Do(jj, N(1), Bin("-", V(n), V(ii)), 1,
						If(C("gt", Ix(arr, V(jj)), Ix(arr, Bin("+", V(jj), N(1)))), []*Stmt{
							Assign(V(t), Ix(arr, V(jj))),
							Assign(Ix(arr, V(jj)), Ix(arr, Bin("+", V(jj), N(1)))),
							Assign(Ix(arr, Bin("+", V(jj), N(1))), V(t)),
						}, nil))),
			}
		}},
		{[]string{"STDDEV", "STDEV", "SIGMA", "DISPER"}, "Computes a standard deviation in integer arithmetic. Approximately.", func(b *builder, p *Proc) {
			arr, n, sd := b.param(p, "int", 40), b.param(p, "int"), b.param(p, "int")
			ii, sum, mean, ss, g := b.local(p, "int"), b.local(p, "int"), b.local(p, "int"), b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(sum), N(0)),
				Do(ii, N(1), V(n), 1, Assign(V(sum), Bin("+", V(sum), Ix(arr, V(ii))))),
				Assign(V(mean), Bin("/", V(sum), V(n))),
				Assign(V(ss), N(0)),
				Do(ii, N(1), V(n), 1, Assign(V(ss), Bin("+", V(ss), Bin("*", Bin("-", Ix(arr, V(ii)), V(mean)), Bin("-", Ix(arr, V(ii)), V(mean)))))),
				Assign(V(ss), Bin("/", V(ss), V(n))),
				Assign(V(g), Bin("/", V(ss), N(2))),
				Do(ii, N(1), N(b.x.between(8, 20)), 1, If(C("gt", V(g), N(0)), []*Stmt{Assign(V(g), Bin("/", Bin("+", V(g), Bin("/", V(ss), V(g))), N(2)))}, nil)),
				Assign(V(sd), V(g)),
			}
		}},
		{[]string{"GAUSSN", "GAUSS", "INTERP", "NORMAL"}, "Interpolates between ladder rungs under a Gaussian assumption. Never called.", func(b *builder, p *Proc) {
			xv, yv := b.param(p, "dbl"), b.param(p, "dbl")
			c := func() *Expr { return DL(fmt.Sprintf("%d.%04d", b.x.between(0, 3), b.x.between(0, 9999))) }
			p.Body = []*Stmt{
				Assign(V(yv), Bin("+", Bin("*", Bin("+", Bin("*", Bin("+", Bin("*", c(), V(xv)), c()), V(xv)), c()), V(xv)), c())),
			}
		}},
		{[]string{"CRC16", "CRCCHK", "CKCRC", "POLYCK"}, "Computes a 16-bit cyclic redundancy check over the deck using modular arithmetic.", func(b *builder, p *Proc) {
			arr, n, crc := b.param(p, "int", 40), b.param(p, "int"), b.param(p, "int")
			ii, kk := b.local(p, "int"), b.local(p, "int")
			poly := b.x.between(4000, 40000)
			p.Body = []*Stmt{
				Assign(V(crc), N(65535)),
				Do(ii, N(1), V(n), 1,
					Assign(V(crc), Fn("mod", Bin("+", V(crc), Ix(arr, V(ii))), N(65536))),
					Do(kk, N(1), N(8), 1,
						If(C("eq", Fn("mod", V(crc), N(2)), N(1)), []*Stmt{Assign(V(crc), Fn("mod", Bin("+", Bin("/", V(crc), N(2)), N(poly)), N(65536)))},
							[]*Stmt{Assign(V(crc), Bin("/", V(crc), N(2)))}))),
			}
		}},
		{[]string{"BCDPK", "PACKIT", "BCDCNV", "UNPACK"}, "Packs a binary value into binary-coded decimal for the old tape format.", func(b *builder, p *Proc) {
			v, pk := b.param(p, "int"), b.param(p, "int")
			w, ii := b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(pk), N(0)), Assign(V(w), V(v)),
				Do(ii, N(1), N(b.x.between(5, 9)), 1,
					Assign(V(pk), Bin("+", Bin("*", V(pk), N(16)), Fn("mod", V(w), N(10)))),
					Assign(V(w), Bin("/", V(w), N(10)))),
			}
		}},
		{[]string{"TAPLBL", "VOLCHK", "LBLCHK", "TAPECK"}, "Checks the tape label before reading. There is no tape.", func(b *builder, p *Proc) {
			vol, ok := b.param(p, "int"), b.param(p, "int")
			p.Body = []*Stmt{
				Assign(V(ok), N(0)),
				If(And(C("gt", V(vol), N(0)), C("lt", V(vol), N(b.x.between(9000, 99999)))), []*Stmt{Assign(V(ok), N(1))}, nil),
			}
		}},
		{[]string{"ROUND", "RNDVAL", "ROUNDR", "NEARST"}, "Rounds a value to the nearest whole dollar. Not used, per audit.", func(b *builder, p *Proc) {
			v := b.param(p, "int")
			p.Body = []*Stmt{Assign(V(v), Bin("/", Bin("+", Bin("*", V(v), N(100)), N(50)), N(100)))}
		}},
		{[]string{"EBCDIC", "XLATE", "TRTAB", "ASCEBC"}, "Translates between EBCDIC and ASCII codes. Present for historical reasons.", func(b *builder, p *Proc) {
			c, out := b.param(p, "int"), b.param(p, "int")
			tab := b.local(p, "int", 16)
			var vals []string
			for t := 0; t < 16; t++ {
				vals = append(vals, strconv.Itoa(b.x.between(64, 249)))
			}
			p.Locals[len(p.Locals)-1].Init = vals
			p.Body = []*Stmt{
				Assign(V(out), V(c)),
				If(And(C("ge", V(c), N(1)), C("le", V(c), N(16))), []*Stmt{Assign(V(out), Ix(tab, V(c)))}, nil),
			}
		}},
		{[]string{"MEDIAN", "MIDVAL", "FNDMED", "CENTER"}, "Finds the median of an array by sorting it. The median is supplied on the card, so this is never needed.", func(b *builder, p *Proc) {
			arr, n, med := b.param(p, "int", 40), b.param(p, "int"), b.param(p, "int")
			p.Body = []*Stmt{
				Assign(V(med), Ix(arr, Bin("+", Bin("/", V(n), N(2)), N(1)))),
				If(C("eq", Fn("mod", V(n), N(2)), N(0)), []*Stmt{Assign(V(med), Bin("/", Bin("+", Ix(arr, Bin("/", V(n), N(2))), Ix(arr, Bin("+", Bin("/", V(n), N(2)), N(1)))), N(2)))}, nil),
			}
		}},
		{[]string{"PAGCNT", "LINCHK", "EJECTP", "PGCTL"}, "Counts lines and decides when to eject a page. Superseded by the printer's own judgement.", func(b *builder, p *Proc) {
			ln, pg := b.param(p, "int"), b.param(p, "int")
			lim := b.x.between(55, 66)
			p.Body = []*Stmt{
				Assign(V(ln), Bin("+", V(ln), N(1))),
				If(C("gt", V(ln), N(lim)), []*Stmt{Assign(V(pg), Bin("+", V(pg), N(1))), Assign(V(ln), N(b.x.between(4, 8)))}, nil),
			}
		}},
		{[]string{"RECON", "RECNCL", "TIEOUT", "BALNCE"}, "Reconciles the metric count against the control card. They always agree.", func(b *builder, p *Proc) {
			n1, n2, rc := b.param(p, "int"), b.param(p, "int"), b.param(p, "int")
			p.Body = []*Stmt{
				Assign(V(rc), N(0)),
				If(C("ne", V(n1), V(n2)), []*Stmt{Assign(V(rc), N(b.x.between(4, 16)))}, nil),
			}
		}},
		{[]string{"INFLAT", "CPIADJ", "DEFLAT", "ESCALT"}, "Adjusts a value for inflation. Disabled after the client asked why salaries went down.", func(b *builder, p *Proc) {
			v, yrs := b.param(p, "int"), b.param(p, "int")
			ii := b.local(p, "int")
			rate := b.x.between(102, 108)
			p.Body = []*Stmt{
				Do(ii, N(1), V(yrs), 1, Assign(V(v), Bin("/", Bin("*", V(v), N(rate)), N(100)))),
			}
		}},
	}
}

// randomBody generates a type-correct but purposeless routine.
func (b *builder) randomBody(p *Proc, callable []*Proc) {
	nInt := b.x.between(2, 6)
	var ints []string
	for t := 0; t < nInt; t++ {
		if t < 2 && b.x.chance(0.5) {
			ints = append(ints, b.param(p, "int"))
		} else {
			ints = append(ints, b.local(p, "int"))
		}
	}
	var arr string
	if b.x.chance(0.5) {
		arr = b.local(p, "int", b.x.between(8, 64))
	}
	idx := b.local(p, "int")
	var gen func(depth int) []*Stmt
	expr := func() *Expr {
		a := V(pick(b.x, ints))
		switch b.x.intn(5) {
		case 0:
			return Bin(pick(b.x, []string{"+", "-", "*"}), a, N(b.x.between(1, 999)))
		case 1:
			return Fn("mod", a, N(b.x.between(2, 997)))
		case 2:
			return Bin("/", a, N(b.x.between(2, 100)))
		case 3:
			return Bin("+", a, V(pick(b.x, ints)))
		default:
			return N(b.x.between(0, 9999))
		}
	}
	gen = func(depth int) []*Stmt {
		var out []*Stmt
		debugged := false
		for s := b.x.between(2, 6); s > 0; s-- {
			r := b.x.intn(10)
			// at most one checkpoint per block; nesting-limited rolls become
			// assignments instead of falling through to more checkpoints
			if (r == 9 && debugged) || (r >= 4 && r < 8 && depth >= 3) {
				r = 0
			}
			switch {
			case r < 4:
				out = append(out, Assign(V(pick(b.x, ints)), expr()))
			case r < 6 && depth < 3:
				body := gen(depth + 1)
				if arr != "" && b.x.chance(0.5) {
					body = append(body, Assign(Ix(arr, V(idx)), expr()))
				}
				out = append(out, Do(idx, N(1), N(b.x.between(2, 8)), 1, body...))
			case r < 8 && depth < 3:
				var els []*Stmt
				if b.x.chance(0.4) {
					els = gen(depth + 1)
				}
				out = append(out, If(C(pick(b.x, []string{"lt", "gt", "eq", "ne", "ge"}), V(pick(b.x, ints)), N(b.x.between(0, 500))), gen(depth+1), els))
			case r == 8 && len(callable) > 0:
				c := pick(b.x, callable)
				if len(c.Params) == 0 {
					out = append(out, Call(c.Name))
				} else {
					var args []*Expr
					for range c.Params {
						args = append(args, V(pick(b.x, ints)))
					}
					if !hasArrayParam(c) {
						out = append(out, Call(c.Name, args...))
					}
				}
			case r == 9:
				debugged = true
				out = append(out, IO("debugmsg", N(b.x.between(200, 899))))
			default:
				out = append(out, Assign(V(pick(b.x, ints)), expr()))
			}
		}
		return out
	}
	p.Body = noDebugRuns(gen(0), func() *Stmt { return Assign(V(pick(b.x, ints)), expr()) })
}

// lastIsDebug reports whether a statement's final emitted line is a debug
// checkpoint (a block that ends in one, with no closing line in some languages).
func lastIsDebug(s *Stmt) bool {
	switch s.K {
	case "io":
		return s.Name == "debugmsg"
	case "do":
		return len(s.Body) > 0 && lastIsDebug(s.Body[len(s.Body)-1])
	case "if":
		b := s.Then
		if s.Else != nil {
			b = s.Else
		}
		return len(b) > 0 && lastIsDebug(b[len(b)-1])
	}
	return false
}

// noDebugRuns replaces any checkpoint that would print right after another.
func noDebugRuns(ss []*Stmt, repl func() *Stmt) []*Stmt {
	for i, s := range ss {
		s.Body = noDebugRuns(s.Body, repl)
		s.Then = noDebugRuns(s.Then, repl)
		if s.Else != nil {
			s.Else = noDebugRuns(s.Else, repl)
		}
		if i > 0 && s.K == "io" && s.Name == "debugmsg" && lastIsDebug(ss[i-1]) {
			ss[i] = repl()
		}
	}
	return ss
}

func hasArrayParam(p *Proc) bool {
	for _, l := range p.Locals {
		if l.Role == "param" && len(l.Dims) > 0 {
			return true
		}
	}
	return false
}

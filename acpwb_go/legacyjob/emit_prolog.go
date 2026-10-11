package legacyjob

import (
	"fmt"
	"strings"
)

// prologEmitter: Prolog has no loops or assignment, so the core is written
// the way Prolog programmers actually wrote it (recursion over the card
// list, accumulators, between/3 for the ladder, format/2 for the report)
// and the dead routines the way the rest of the shop wrote it (assert /
// retract used as global variables, failure-driven loops).
type prologEmitter struct {
	j   *job
	out []Line
	n   int
}

func newPrologEmitter(j *job) *prologEmitter { return &prologEmitter{j: j} }

func (e *prologEmitter) code(t string) {
	e.out = append(e.out, Line{Text: strings.TrimRight(t, " "), Code: true})
}
func (e *prologEmitter) codef(f string, a ...interface{}) { e.code(fmt.Sprintf(f, a...)) }
func (e *prologEmitter) blank()                           { e.out = append(e.out, Line{Text: ""}) }
func (e *prologEmitter) cmt(lines []string) {
	for _, l := range lines {
		e.out = append(e.out, Line{Text: strings.TrimRight("% "+l, " ")})
	}
}
func (e *prologEmitter) note(text string) {
	cs := e.j.cs
	cs.width = 70
	e.cmt(cs.para(text))
}

// atom renders an IR name as a Prolog atom (lower case, no sigils).
func patom(n string) string {
	return strings.ToLower(strings.NewReplacer("$", "", "#", "", "-", "_", ".", "_", "*", "").Replace(n))
}

// pvar renders an IR name as a Prolog variable.
func pvar(n string) string {
	a := patom(n)
	if a == "" {
		return "V"
	}
	return strings.ToUpper(a[:1]) + a[1:]
}

func (e *prologEmitter) emit() []Line {
	j := e.j
	r := j.rep
	tok := strings.ToUpper(j.in.Token)
	turbo := j.id.Language == "Turbo Prolog"
	q := func(s string) string {
		return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'", "~", "~~").Replace(s) + "'"
	}
	e.cmt(j.cs.box(j.id.Program+" -- "+j.in.Title, "American Corporation for Public Well Being, Research Division. "+j.g.comment(j.g.roster[0], "history", "philosophy")))
	e.cmt(j.historyLines())
	e.cmt(j.cs.box("Provenance", "Seal "+tok+", record "+strings.ToUpper(j.in.RecordID)+". "+j.g.comment(nil, "security", "regulatory", "memos")))
	e.blank()
	if turbo {
		e.code("domains")
		e.code("    card = string")
		e.code("    cards = card*")
		e.code("    ints = integer*")
		e.code("database")
		e.code("    work(symbol, integer)")
		e.code("predicates")
		e.code("    main")
		e.code("    rung(integer, real)")
		e.code("clauses")
	} else {
		e.code(":- dynamic work/2.")
		e.code(":- initialization(main, main).")
	}
	e.blank()
	e.codef("seal(%s).", q(tok))
	e.blank()
	e.note("The percentile ladder. Ratified 1981-02-17; facts, not estimates. " + j.g.comment(nil, "magic", "lore"))
	for i, v := range ladderLit {
		e.codef("rung(%d, %s).", i+1, v)
	}
	e.blank()
	e.cmt(j.prologue(j.p.Procs[0], nil, nil))
	e.code("main :-")
	e.code("    read_deck('deck.dat', Cards),")
	e.code("    Cards = [H1, H2, H3 | Rest],")
	e.code("    sub_atom(H1, 0, 40, _, Org), sub_atom(H1, 40, 8, _, RecId),")
	e.code("    sub_atom(H2, 0, 60, _, Peer),")
	e.code("    field_int(H3, 0, 3, NMet), field_int(H3, 3, 3, NBen),")
	e.code("    length(MetCards, NMet), append(MetCards, BenCards, Rest),")
	e.code("    length(BenCards, NBen),")
	if j.x.chance(0.5) {
		e.code("    maplist(metric, MetCards, Metrics),")
		e.code("    maplist(benchmark, BenCards, Benches),")
	} else {
		e.code("    findall(M, (member(C, MetCards), metric(C, M)), Metrics),")
		e.code("    findall(B, (member(C, BenCards), benchmark(C, B)), Benches),")
	}
	e.code("    report(Org, RecId, Peer, Metrics, Benches, NMet, NBen).")
	e.blank()
	e.note("Read every card image into a list of atoms. " + j.g.comment(nil, "deck", "operators"))
	e.code("read_deck(File, Cards) :-")
	e.code("    open(File, read, S),")
	e.code("    read_lines(S, Cards),")
	e.code("    close(S).")
	e.blank()
	switch j.x.intn(3) {
	case 0:
		e.code("read_lines(S, Cards) :-")
		e.code("    read_line_to_string(S, L),")
		e.code("    (   L == end_of_file")
		e.code("    ->  Cards = []")
		e.code("    ;   atom_string(A, L), Cards = [A | More], read_lines(S, More)")
		e.code("    ).")
	case 1:
		e.code("read_lines(S, []) :- at_end_of_stream(S), !.")
		e.code("read_lines(S, [A | More]) :-")
		e.code("    read_line_to_codes(S, Codes),")
		e.code("    atom_codes(A, Codes),")
		e.code("    read_lines(S, More).")
	default:
		e.code("read_lines(S, Cards) :- read_lines(S, [], Rev), reverse(Rev, Cards).")
		e.code("read_lines(S, Acc, Acc) :- at_end_of_stream(S), !.")
		e.code("read_lines(S, Acc, Out) :-")
		e.code("    read_line_to_string(S, L), atom_string(A, L),")
		e.code("    read_lines(S, [A | Acc], Out).")
	}
	e.blank()
	e.code("field_int(Card, Start, Len, N) :-")
	e.code("    sub_atom(Card, Start, Len, _, A),")
	e.code("    normalize_space(atom(T), A),")
	e.code("    (   T == '' -> N = 0 ; atom_number(T, N) ).")
	e.blank()
	e.note("Change = current - baseline. " + j.g.comment(nil, "units", "warnings"))
	e.code("metric(Card, m(Name, Base, Curr, Delta)) :-")
	e.code("    sub_atom(Card, 0, 40, _, Name),")
	e.code("    field_int(Card, 40, 6, Base),")
	e.code("    field_int(Card, 46, 6, Curr),")
	e.code("    Delta is Curr - Base.")
	e.blank()
	e.note("Truncate, never round (audit 83-07). " + j.g.comment(nil, "magic", "memos"))
	e.code("benchmark(Card, b(Name, Ladder)) :-")
	e.code("    sub_atom(Card, 0, 40, _, Name),")
	e.code("    field_int(Card, 40, 7, Median),")
	switch j.x.intn(3) {
	case 0:
		e.code("    findall(P, ( between(1, 8, K), rung(K, M), P is truncate(Median * M) ), Ladder).")
	case 1:
		e.code("    numlist(1, 8, Ks),")
		e.code("    maplist(apply_rung(Median), Ks, Ladder).")
		e.blank()
		e.code("apply_rung(Median, K, P) :- rung(K, M), P is truncate(Median * M).")
	default:
		e.code("    ladder(1, Median, Ladder).")
		e.blank()
		e.code("ladder(K, _, []) :- K > 8, !.")
		e.code("ladder(K, Median, [P | Ps]) :-")
		e.code("    rung(K, M), P is truncate(Median * M),")
		e.code("    K1 is K + 1, ladder(K1, Median, Ps).")
	}
	e.blank()
	e.note("The report. " + j.g.comment(nil, "printing", "clients"))
	e.code("report(Org, RecId, Peer, Metrics, Benches, NMet, NBen) :-")
	e.codef("    format(\"~w~w~w~n\", [%s, RecId, %s]),", q(r.HdrLeft), q(r.HdrMid+"    1"))
	e.codef("    format(\"~w~w~n\", [%s, Org]),", q(r.Title1))
	e.codef("    format(\"~w~n\", [%s]),", q(r.ColHdr1))
	e.code("    forall(member(m(N, B, C, D), Metrics),")
	e.code("           ( sign_text(D, DS), format(\" ~w~t~41|~t~d~50|~t~d~59|~t~w~68|~n\", [N, B, C, DS]) )),")
	e.codef("    format(\"~w~w~n\", [%s, Peer]),", q(r.Title2))
	e.codef("    format(\"~w~n\", [%s]),", q(r.ColHdr2))
	e.code("    forall(member(b(N, L), Benches),")
	e.code("           ( format(\" ~w~t~41|\", [N]), forall(member(P, L), format(\"~t~d~9+\", [P])), nl )),")
	e.code("    seal(Seal),")
	e.code("    format(\"0RECORDS PROCESSED~t~d~24| METRICS~t~d~38| BENCHMARKS   SEAL ~w~n\", [NMet, NBen, Seal]).")
	e.blank()
	e.code("sign_text(D, T) :- D > 0, !, format(atom(T), \"+~d\", [D]).")
	e.code("sign_text(D, T) :- format(atom(T), \"~d\", [D]).")
	for _, p := range j.p.Procs[1:] {
		if p.Dead {
			e.deadPred(p)
		}
	}
	if j.x.chance(0.4) {
		e.blank()
		e.note(j.g.comment(nil, "handoff", "todo", "misc"))
	}
	return e.out
}

// ── dead predicates: assert/retract "variables", failure-driven loops ─────

func (e *prologEmitter) ex(x *Expr) string {
	switch x.K {
	case "var":
		return "V_" + patom(x.Name)
	case "int":
		return fmt.Sprint(x.Val)
	case "dlit":
		return x.S
	case "idx":
		return "E_" + patom(x.Name)
	case "bin":
		op := x.Op
		if op == "/" && !isDouble(x.Args[0]) && !isDouble(x.Args[1]) {
			op = "//"
		}
		return "(" + e.ex(x.Args[0]) + " " + op + " " + e.ex(x.Args[1]) + ")"
	case "call":
		if x.Name == "mod" {
			return "(" + e.ex(x.Args[0]) + " mod " + e.ex(x.Args[1]) + ")"
		}
		return e.ex(x.Args[0])
	case "ladder":
		return "truncate(" + e.ex(x.Args[0]) + " * " + e.ex(x.Args[1]) + ")"
	}
	return "0"
}

// loads lists the work/2 lookups an expression needs first.
func (e *prologEmitter) loads(xs ...*Expr) []string {
	var out []string
	seen := map[string]bool{}
	for _, x := range xs {
		walkExpr(x, func(y *Expr) {
			if y.K == "var" || y.K == "idx" {
				v := e.ex(&Expr{K: y.K, Name: y.Name})
				if !seen[v] {
					seen[v] = true
					out = append(out, fmt.Sprintf("work(%s, %s)", patom(y.Name), v))
				}
			}
		})
	}
	return out
}

func (e *prologEmitter) cond(c *Cond) string {
	if c.Op == "and" || c.Op == "or" {
		op := ", "
		if c.Op == "or" {
			op = " ; "
		}
		return "(" + e.cond(c.A) + op + e.cond(c.B) + ")"
	}
	rel := map[string]string{"lt": "<", "le": "=<", "eq": "=:=", "ne": "=\\=", "gt": ">", "ge": ">="}[c.Op]
	return e.ex(c.L) + " " + rel + " " + e.ex(c.R)
}

func condExprs(c *Cond) []*Expr {
	if c.A != nil {
		return append(condExprs(c.A), condExprs(c.B)...)
	}
	return []*Expr{c.L, c.R}
}

func (e *prologEmitter) goals(ss []*Stmt, ind string) []string {
	var out []string
	for _, s := range ss {
		switch s.K {
		case "assign":
			out = append(out, e.loads(s.R)...)
			e.n++
			t := fmt.Sprintf("T%d", e.n)
			out = append(out, t+" is "+e.ex(s.R), fmt.Sprintf("retractall(work(%s, _))", patom(s.L.Name)), fmt.Sprintf("assertz(work(%s, %s))", patom(s.L.Name), t))
		case "io":
			if s.Name == "debugmsg" && len(s.Args) > 0 {
				out = append(out, fmt.Sprintf("format(\" DEBUG CHECKPOINT ~w~n\", ['%04d'])", s.Args[0].Val))
			}
		case "call":
			out = append(out, patom(s.Name))
		case "if":
			pre := e.loads(condExprs(s.Cond)...)
			then := e.goals(s.Then, ind+"    ")
			if len(then) == 0 {
				then = []string{"true"}
			}
			els := []string{"true"}
			if s.Else != nil {
				if g := e.goals(s.Else, ind+"    "); len(g) > 0 {
					els = g
				}
			}
			out = append(out, pre...)
			out = append(out, "(   "+e.cond(s.Cond)+"\n"+ind+"->  "+strings.Join(then, ",\n"+ind+"    ")+"\n"+ind+";   "+strings.Join(els, ",\n"+ind+"    ")+"\n"+ind+")")
		case "do":
			pre := e.loads(s.From, s.To)
			body := e.goals(s.Body, ind+"    ")
			if len(body) == 0 {
				body = []string{"true"}
			}
			v := pvar(s.Var)
			out = append(out, pre...)
			out = append(out, fmt.Sprintf("(   between(%s, %s, %s),\n%s    retractall(work(%s, _)), assertz(work(%s, %s)),\n%s    %s,\n%s    fail\n%s;   true\n%s)",
				e.ex(s.From), e.ex(s.To), v, ind, patom(s.Var), patom(s.Var), v, ind, strings.Join(body, ",\n"+ind+"    "), ind, ind, ind))
		}
	}
	return out
}

func (e *prologEmitter) deadPred(p *Proc) {
	e.blank()
	e.cmt(e.j.prologue(p, nil, nil))
	head := patom(p.Name)
	var args []string
	for _, q := range p.Params {
		args = append(args, pvar(q))
	}
	if len(args) > 0 {
		head += "(" + strings.Join(args, ", ") + ")"
	}
	gs := []string{}
	for _, a := range p.Params {
		gs = append(gs, fmt.Sprintf("retractall(work(%s, _))", patom(a)), fmt.Sprintf("assertz(work(%s, %s))", patom(a), pvar(a)))
	}
	gs = append(gs, e.goals(p.Body, "    ")...)
	if len(gs) == 0 {
		gs = []string{"true"}
	}
	e.code(head + " :-")
	for i, g := range gs {
		sep := ","
		if i == len(gs)-1 {
			sep = "."
		}
		for k, l := range strings.Split("    "+g+sep, "\n") {
			if k > 0 {
				l = "    " + l
			}
			e.code(l)
		}
	}
}

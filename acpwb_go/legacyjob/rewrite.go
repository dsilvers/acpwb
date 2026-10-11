package legacyjob

import (
	"fmt"
	"strconv"
)

// Semantics-preserving (core) and semantics-free (dead) rewrites applied to
// the finished program, so the same computation takes a different shape on
// every page. Core rewrites must keep the published numbers exact.

func (b *builder) rewrite(p *Program) {
	ints := []string{b.nm.get("ierr"), b.nm.get("tmp"), b.nm.get("iline")}
	for _, pr := range p.Procs {
		pr.Body = b.invertIfs(pr.Body)
		if pr.Dead {
			pr.Body = b.deadRewrites(pr, pr.Body)
		} else if b.x.chance(0.45) && len(pr.Body) > 1 {
			// a never-taken branch guarded by a debug value nobody sets
			pos := 1 + b.x.intn(len(pr.Body))
			if pos > len(pr.Body) {
				pos = len(pr.Body)
			}
			if pr.Body[pos-1].K == "stop" {
				pos--
			}
			g := b.guardBranch(ints)
			pr.Body = append(pr.Body[:pos], append(g, pr.Body[pos:]...)...)
		}
	}
}

// invertIfs swaps THEN/ELSE under a negated condition (exactly equivalent).
func (b *builder) invertIfs(ss []*Stmt) []*Stmt {
	for _, s := range ss {
		s.Body = b.invertIfs(s.Body)
		s.Then = b.invertIfs(s.Then)
		s.Else = b.invertIfs(s.Else)
		if s.K == "if" && len(s.Else) > 0 && len(s.Then) > 0 && b.x.chance(0.4) {
			s.Cond = negate(s.Cond)
			s.Then, s.Else = s.Else, s.Then
		}
	}
	return ss
}

func (b *builder) guardBranch(ints []string) []*Stmt {
	flag := b.v("idebug")
	var body []*Stmt
	for n := b.x.between(1, 4); n > 0; n-- {
		t := pick(b.x, ints)
		switch b.x.intn(4) {
		case 0:
			body = append(body, Assign(V(t), N(b.x.between(1, 9999))))
		case 1:
			body = append(body, Assign(V(t), Fn("mod", Bin("+", V(t), N(b.x.between(1, 999))), N(b.x.between(7, 997)))))
		case 2:
			body = append(body, IO("debugmsg", N(b.x.between(900, 999))))
		default:
			body = append(body, Assign(V(t), Bin("*", V(pick(b.x, ints)), N(b.x.between(2, 9)))))
		}
	}
	body = noDebugRuns(body, func() *Stmt { return Assign(V(ints[0]), N(0)) })
	why := pick(b.x, []string{
		"Diagnostic path for audit finding %s. Enable by setting the debug flag to %d.",
		"Left in after the %s investigation. Only runs when the debug flag is %d.",
		"Test harness hook (%s). Debug flag %d.",
		"Per %s, retained but disabled. Flag value %d was never used in production.",
	})
	k := b.x.between(2, 9)
	return []*Stmt{Comment(fmt.Sprintf(why, b.g.memo(), k)), If(C("eq", flag, N(k)), body, nil)}
}

// deadRewrites reshape dead code freely (it never runs).
func (b *builder) deadRewrites(p *Proc, ss []*Stmt) []*Stmt {
	var out []*Stmt
	for _, s := range ss {
		s.Body = b.deadRewrites(p, s.Body)
		s.Then = b.deadRewrites(p, s.Then)
		s.Else = b.deadRewrites(p, s.Else)
		switch {
		case s.K == "assign" && s.R.K == "bin" && s.R.Op == "*" && s.R.Args[1].K == "int" && s.R.Args[1].Val == 2 && b.x.chance(0.7):
			s.R = Bin("+", s.R.Args[0], s.R.Args[0]) // strength reduction
			out = append(out, s)
		case s.K == "assign" && s.R.K == "bin" && s.L.K == "var" && b.x.chance(0.2):
			t := b.local(p, "int")
			out = append(out, Assign(V(t), s.R.Args[0]), Assign(s.L, Bin(s.R.Op, V(t), s.R.Args[1])))
		case s.K == "do" && len(s.Body) >= 4 && b.x.chance(0.3):
			// loop fission
			h := len(s.Body) / 2
			out = append(out, Do(s.Var, s.From, s.To, s.Step, s.Body[:h]...), Do(s.Var, s.From, s.To, s.Step, s.Body[h:]...))
		default:
			out = append(out, s)
		}
	}
	return out
}

// extraDeadLibrary: more named utility routines, each with internals that
// vary page to page.
func extraDeadLibrary() []deadSpec {
	return []deadSpec{
		{[]string{"LUHN", "MOD10", "CHKDIG", "LUHNCK"}, "Validates an account number with the mod-10 check digit. No account numbers are read.", func(b *builder, p *Proc) {
			v, ok := b.param(p, "int"), b.param(p, "int")
			w, s, d, pos := b.local(p, "int"), b.local(p, "int"), b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(w), V(v)), Assign(V(s), N(0)), Assign(V(pos), N(0)),
				Do(b.nm.get("k"), N(1), N(b.x.between(8, 16)), 1,
					Assign(V(d), Fn("mod", V(w), N(10))),
					If(C("eq", Fn("mod", V(pos), N(2)), N(1)), []*Stmt{
						Assign(V(d), Bin("*", V(d), N(2))),
						If(C("gt", V(d), N(9)), []*Stmt{Assign(V(d), Bin("-", V(d), N(9)))}, nil),
					}, nil),
					Assign(V(s), Bin("+", V(s), V(d))),
					Assign(V(w), Bin("/", V(w), N(10))),
					Assign(V(pos), Bin("+", V(pos), N(1)))),
				Assign(V(ok), N(0)),
				If(C("eq", Fn("mod", V(s), N(10)), N(0)), []*Stmt{Assign(V(ok), N(1))}, nil),
			}
		}},
		{[]string{"DOWEEK", "ZELLER", "WKDAY", "DAYNUM"}, "Computes the day of the week (Zeller's congruence) for the run date. The run date is printed but the day never is.", func(b *builder, p *Proc) {
			y, m, d, dow := b.param(p, "int"), b.param(p, "int"), b.param(p, "int"), b.param(p, "int")
			yy, mm, k, j := b.local(p, "int"), b.local(p, "int"), b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(yy), V(y)), Assign(V(mm), V(m)),
				If(C("lt", V(mm), N(3)), []*Stmt{Assign(V(mm), Bin("+", V(mm), N(12))), Assign(V(yy), Bin("-", V(yy), N(1)))}, nil),
				Assign(V(k), Fn("mod", V(yy), N(100))),
				Assign(V(j), Bin("/", V(yy), N(100))),
				Assign(V(dow), Fn("mod", Bin("+", Bin("+", Bin("+", V(d), Bin("/", Bin("*", N(13), Bin("+", V(mm), N(1))), N(5))), Bin("+", V(k), Bin("/", V(k), N(4)))), Bin("+", Bin("/", V(j), N(4)), Bin("*", N(5), V(j)))), N(7))),
			}
		}},
		{[]string{"INSSRT", "INSORT", "SHUFFL", "ORDINS"}, "Insertion sort, written after the bubble sort was declared too slow. Also never called.", func(b *builder, p *Proc) {
			arr, n := b.param(p, "int", 40), b.param(p, "int")
			i, j, key := b.local(p, "int"), b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Do(i, N(2), V(n), 1,
					Assign(V(key), Ix(arr, V(i))),
					Assign(V(j), Bin("-", V(i), N(1))),
					Do(b.nm.fresh("int"), N(1), V(i), 1,
						If(And(C("gt", V(j), N(0)), C("gt", Ix(arr, V(j)), V(key))), []*Stmt{
							Assign(Ix(arr, Bin("+", V(j), N(1))), Ix(arr, V(j))),
							Assign(V(j), Bin("-", V(j), N(1))),
						}, nil)),
					Assign(Ix(arr, Bin("+", V(j), N(1))), V(key))),
			}
			p.Locals = append(p.Locals, Sym{Name: p.Body[0].Body[2].Var, Type: "int"})
		}},
		{[]string{"BSRCH", "BINSRC", "LOOKUP", "FINDIX"}, "Binary search of the sorted index table. The table has never been sorted.", func(b *builder, p *Proc) {
			arr, n, key, pos := b.param(p, "int", 40), b.param(p, "int"), b.param(p, "int"), b.param(p, "int")
			lo, hi, mid, it := b.local(p, "int"), b.local(p, "int"), b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(lo), N(1)), Assign(V(hi), V(n)), Assign(V(pos), N(0)),
				Do(it, N(1), N(b.x.between(6, 12)), 1,
					If(C("le", V(lo), V(hi)), []*Stmt{
						Assign(V(mid), Bin("/", Bin("+", V(lo), V(hi)), N(2))),
						If(C("eq", Ix(arr, V(mid)), V(key)), []*Stmt{Assign(V(pos), V(mid)), Assign(V(lo), Bin("+", V(hi), N(1)))},
							[]*Stmt{If(C("lt", Ix(arr, V(mid)), V(key)), []*Stmt{Assign(V(lo), Bin("+", V(mid), N(1)))}, []*Stmt{Assign(V(hi), Bin("-", V(mid), N(1)))})}),
					}, nil)),
			}
		}},
		{[]string{"GCD", "EUCLID", "HCF", "COMDIV"}, "Greatest common divisor, for reducing fractions in a ratio report that was cancelled in planning.", func(b *builder, p *Proc) {
			a, c, g := b.param(p, "int"), b.param(p, "int"), b.param(p, "int")
			t, it := b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Do(it, N(1), N(b.x.between(20, 64)), 1,
					If(C("ne", V(c), N(0)), []*Stmt{Assign(V(t), Fn("mod", V(a), V(c))), Assign(V(a), V(c)), Assign(V(c), V(t))}, nil)),
				Assign(V(g), V(a)),
			}
		}},
		{[]string{"ISQRT", "SQROOT", "NEWTON", "ROOTN"}, "Integer square root by Newton's method. The FORTRAN library had one; this one is ours.", func(b *builder, p *Proc) {
			v, r := b.param(p, "int"), b.param(p, "int")
			it := b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(r), Bin("/", V(v), N(2))),
				If(C("le", V(r), N(0)), []*Stmt{Assign(V(r), N(1))}, nil),
				Do(it, N(1), N(b.x.between(10, 30)), 1, Assign(V(r), Bin("/", Bin("+", V(r), Bin("/", V(v), V(r))), N(2)))),
			}
		}},
		{[]string{"B36ENC", "BASE36", "ENCODE", "TOKGEN"}, "Encodes the record number in base 36 for the tape label. The tape label program was retired.", func(b *builder, p *Proc) {
			v, out := b.param(p, "int"), b.param(p, "int")
			w, d, i, mul := b.local(p, "int"), b.local(p, "int"), b.local(p, "int"), b.local(p, "int")
			base := pick(b.x, []int{36, 32, 26, 16})
			p.Body = []*Stmt{
				Assign(V(w), V(v)), Assign(V(out), N(0)), Assign(V(mul), N(1)),
				Do(i, N(1), N(b.x.between(4, 7)), 1,
					Assign(V(d), Fn("mod", V(w), N(base))),
					Assign(V(out), Bin("+", V(out), Bin("*", V(d), V(mul)))),
					Assign(V(mul), Bin("*", V(mul), N(100))),
					Assign(V(w), Bin("/", V(w), N(base)))),
			}
		}},
		{[]string{"PARITY", "PARCHK", "ODDEVN", "PARBIT"}, "Computes odd parity for the paper-tape punch. " + strconv.Itoa(8) + "-level tape, specifically.", func(b *builder, p *Proc) {
			v, par := b.param(p, "int"), b.param(p, "int")
			w, i := b.local(p, "int"), b.local(p, "int")
			p.Body = []*Stmt{
				Assign(V(w), V(v)), Assign(V(par), N(1)),
				Do(i, N(1), N(b.x.between(7, 16)), 1,
					If(C("eq", Fn("mod", V(w), N(2)), N(1)), []*Stmt{Assign(V(par), Bin("-", N(1), V(par)))}, nil),
					Assign(V(w), Bin("/", V(w), N(2)))),
			}
		}},
	}
}

package legacyjob

import (
	"fmt"
	"strings"
)

// Dialect tables for emit_dialect.go. Card columns arrive as a 0-based
// start offset plus width; each dialect converts to its own convention.

var stdRel = map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "<>", "gt": ">", "ge": ">="}
var cRelOps = map[string]string{"lt": "<", "le": "<=", "eq": "==", "ne": "!=", "gt": ">", "ge": ">="}

func dialectFor(j *job) *dialect {
	switch j.id.LangKey {
	case "ada":
		return adaDialect(j)
	case "pascal", "pascalbtr":
		return pascalDialect(j)
	case "pli":
		return pliDialect(j)
	case "pal":
		return palDialect(j)
	case "dbase", "clipper", "foxpro":
		return xbaseDialect(j)
	case "natural":
		return naturalDialect(j)
	case "informix":
		return informixDialect(j)
	case "vb", "vba":
		return vbDialect(j)
	case "vbnet":
		return vbnetDialect(j)
	case "csharp":
		return csharpDialect(j)
	case "php":
		return phpDialect(j)
	case "pick":
		return pickDialect(j)
	case "mumps":
		return mumpsDialect(j)
	case "plsql":
		return plsqlDialect(j)
	}
	if f, ok := moreDialects[j.id.LangKey]; ok {
		d := f(j)
		d.mainEnds = true
		return d
	}
	return nil
}

// initList renders an array's DATA values as a comma list.
func initList(s Sym, q func(string) string) string {
	var vals []string
	for _, v := range s.Init {
		if s.Type == "str" {
			vals = append(vals, q(v))
		} else {
			vals = append(vals, v)
		}
	}
	return strings.Join(vals, ", ")
}

// elementInits assigns DATA values one element at a time (languages without
// array initialisers).
func elementInits(e *dialectEmitter, depth int) {
	for _, s := range e.j.p.Globals {
		if len(s.Init) == 0 || len(s.Dims) == 0 {
			continue
		}
		for i, v := range s.Init {
			if s.Type == "str" {
				v = e.d.lit(v)
			}
			e.line(depth, fmt.Sprintf(e.d.assign, fmt.Sprintf(e.d.idx1, e.name(s.Name), fmt.Sprint(i+1)), v), s.Name)
		}
	}
}

// scalarInits assigns initial values of scalar globals.
func scalarInits(e *dialectEmitter, depth int) {
	for _, s := range e.j.p.Globals {
		if s.Role == "const" || len(s.Init) == 0 || len(s.Dims) > 0 {
			continue
		}
		v := s.Init[0]
		if s.Type == "str" {
			v = e.d.lit(v)
		}
		e.line(depth, fmt.Sprintf(e.d.assign, e.name(s.Name), v), s.Name)
	}
}

func constInits(e *dialectEmitter, depth int) {
	for _, s := range e.j.p.Globals {
		if s.Role == "const" {
			e.line(depth, fmt.Sprintf(e.d.assign, e.name(s.Name), s.Init[0]), s.Name)
		}
	}
}

func (e *dialectEmitter) dimList(s Sym, f func(d string) string, sep string) string {
	var ds []string
	for _, d := range s.Dims {
		ds = append(ds, f(e.dimName(d)))
	}
	return strings.Join(ds, sep)
}

// ── Ada ───────────────────────────────────────────────────────────────────

func adaDialect(j *job) *dialect {
	x := j.x
	ada83 := j.in.Year < 1996
	typeName := func(e *dialectEmitter, s Sym) string {
		base := map[string]string{"int": "Integer", "dbl": "Long_Float", "str": fmt.Sprintf("String (1 .. %d)", maxi(s.Len, 1))}[s.Type]
		if base == "" {
			base = "Integer"
		}
		if len(s.Dims) == 0 {
			return base
		}
		elem := map[string]string{"int": "Int", "dbl": "Flt", "str": fmt.Sprintf("Str%d", s.Len)}[s.Type]
		var parts []string
		for _, d := range s.Dims {
			parts = append(parts, fmt.Sprint(d))
		}
		return elem + "_Array_" + strings.Join(parts, "x")
	}
	d := &dialect{ind: pick(x, []string{"   ", "   ", "    "}), cmt: "-- ",
		idx1: "%s (%s)", idx2: "%s (%s, %s)", intDiv: "%s / %s", mod: "%s mod %s", dble: "Long_Float (%s)",
		ladder: "Integer (Long_Float'Truncation (Long_Float (%s) * %s))",
		rel:    map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "/=", "gt": ">", "ge": ">="},
		and:    " and then ", or: " or else ", assign: "%s := %s;", call: "%s (%s);", callNone: "%s;", stop: "return;",
		ifOpen: "if %s then", ifElse: "else", ifClose: "end if;",
		forUp: "for %s in %s .. %s loop", forDown: "for %[1]s in reverse %[3]s .. %[2]s loop", forClose: "end loop;",
		concat: " & ", lit: dq, order: "nested",
		pstr:  func(x string, w int) string { return fmt.Sprintf("Pad (%s, %d)", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("Img (%s, %d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("Signed_Img (%s, %d)", x, w) },
		print: func(p string) string { return "Put_Line (" + p + ");" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("%-10s : constant := %s;", e.name(s.Name), s.Init[0])}
		}
		t := typeName(e, s)
		init := ""
		switch {
		case len(s.Init) > 0 && len(s.Dims) > 0:
			init = " := (" + initList(s, dq) + ")"
		case len(s.Init) > 0 && s.Type == "str":
			init = " := " + dq(padRight(s.Init[0], s.Len))
		case len(s.Init) > 0:
			init = " := " + s.Init[0]
		case len(s.Dims) > 0 && s.Type != "str":
			init = " := (others => 0)"
		case s.Type == "str" && len(s.Dims) == 0:
			init = " := (others => ' ')"
		case s.Type == "int":
			init = " := 0"
		}
		return []string{fmt.Sprintf("%-10s : %s%s;", e.name(s.Name), t, init)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("with Ada.Text_IO; use Ada.Text_IO;")
		if !ada83 {
			e.raw("with Ada.Strings.Fixed; use Ada.Strings.Fixed;")
		}
		e.blank()
		e.raw("procedure " + strings.Title(strings.ToLower(e.j.id.Program)) + " is")
		e.blank()
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				e.defLine(1, d.decl(e, s, "global")[0], s.Name)
			}
		}
		e.blank()
		seen := map[string]bool{}
		all := append([]Sym{}, e.j.p.Globals...)
		for _, p := range e.j.p.Procs {
			all = append(all, p.Locals...)
		}
		for _, s := range all {
			if len(s.Dims) == 0 {
				continue
			}
			t := typeName(e, s)
			if seen[t] {
				continue
			}
			seen[t] = true
			elem := map[string]string{"int": "Integer", "dbl": "Long_Float", "str": fmt.Sprintf("String (1 .. %d)", s.Len)}[s.Type]
			e.line(1, fmt.Sprintf("type %s is array (%s) of %s;", t, e.dimList(s, func(dn string) string { return "1 .. " + dn }, ", "), elem))
		}
		e.blank()
		for _, s := range e.j.p.Globals {
			if s.Role != "const" {
				e.defLine(1, d.decl(e, s, "global")[0], s.Name)
			}
		}
		for _, l := range e.j.p.Procs[0].Locals {
			if l.Role != "param" {
				e.defLine(1, d.decl(e, l, "local")[0], l.Name)
			}
		}
		e.defLine(1, "Card       : String (1 .. 80) := (others => ' ');", "card")
		e.defLine(1, "Last       : Natural := 0;")
		e.blank()
		e.comment(1, "Formatting helpers. "+e.j.g.comment(nil, "formats", "printing", "workarounds"))
		e.line(1, "function Pad (S : String; W : Positive) return String is")
		e.line(1, "begin")
		e.line(2, "if S'Length >= W then return S (S'First .. S'First + W - 1); end if;")
		e.line(2, "return S & (1 .. W - S'Length => ' ');")
		e.line(1, "end Pad;")
		e.line(1, "function Img (N : Integer; W : Positive) return String is")
		e.line(2, "S : constant String := Integer'Image (N);")
		e.line(1, "begin")
		e.line(2, "if S'Length >= W then return S; end if;")
		e.line(2, "return (1 .. W - S'Length => ' ') & S;")
		e.line(1, "end Img;")
		e.line(1, "function Signed_Img (N : Integer; W : Positive) return String is")
		e.line(1, "begin")
		e.line(2, "if N > 0 then return Img (N, W - 1) & \"\"; end if;  -- TODO: the plus sign (see memo)")
		e.line(2, "return Img (N, W);")
		e.line(1, "end Signed_Img;")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("begin")
			return
		}
		var ps []string
		var copies []Sym
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			mode := "in"
			n := q
			if len(s.Dims) > 0 {
				mode = "in out"
			} else if assigns(p.Body, q) {
				n = q + "_In"
				copies = append(copies, s)
			}
			ps = append(ps, fmt.Sprintf("%s : %s %s", n, mode, typeName(e, s)))
		}
		head := "procedure " + p.Name
		if len(ps) > 0 {
			head += " (" + strings.Join(ps, "; ") + ")"
		}
		e.line(1, head+" is")
		for _, c := range copies {
			e.defLine(2, fmt.Sprintf("%-10s : %s := %s_In;", c.Name, typeName(e, c), c.Name), c.Name)
		}
		for _, l := range p.Locals {
			if l.Role != "param" {
				e.defLine(2, d.decl(e, l, "local")[0], l.Name)
			}
		}
		e.line(1, "begin")
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("end " + strings.Title(strings.ToLower(e.j.id.Program)) + ";")
			return
		}
		e.line(1, "end "+p.Name+";")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "Get_Line (Card, Last);") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s := Integer'Value (Card (%d .. %d));", t, start+1, start+width)
		}
		return fmt.Sprintf("%s := Pad (Card (%d .. %d), %d);", t, start+1, start+width, width)
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 1
		}
		return 2
	}
	return d
}

// ── Pascal (Turbo Pascal, optionally with Btrieve) ─────────────────────────

func pascalDialect(j *job) *dialect {
	x := j.x
	btr := j.id.LangKey == "pascalbtr"
	intT := "LongInt" // Turbo Pascal 4.0 (1987) introduced LongInt
	if j.in.Year < 1987 {
		intT = "Integer"
	}
	realT := pick(x, []string{"Real", "Real", "Double"})
	padFn := pick(x, []string{"PadR", "PadRight", "LJust", "RPad", "FillTo"})
	sgnFn := pick(x, []string{"SignStr", "Signed", "FmtSign", "PlusStr", "SgnFmt"})
	fieldStyle := x.intn(3)
	ioCheck := x.chance(0.5)
	// Btrieve history record: names chosen per page
	recT := pick(x, []string{"THistRec", "HistRecord", "TBenchHist", "RunHistRec", "THRec"})
	recV := pick(x, []string{"Hist", "HRec", "HistBuf", "RunRec"})
	fRec, fMet := pick(x, []string{"RecId", "RecordNo", "DocId"}), pick(x, []string{"Metric", "MetName", "Descr"})
	fBase, fCur, fChg := pick(x, []string{"Baseline", "Prior", "BaseVal"}), pick(x, []string{"Current", "Latest", "CurVal"}), pick(x, []string{"Change", "Delta", "Diff"})
	posV := pick(x, []string{"PosBlk", "PosBlock", "BPos", "FCB"})
	openMode := pick(x, []string{"0", "0", "-1", "-2"})
	typeName := func(e *dialectEmitter, s Sym) string {
		base := map[string]string{"int": intT, "dbl": realT, "str": fmt.Sprintf("string[%d]", maxi(s.Len, 1))}[s.Type]
		if base == "" {
			base = intT
		}
		if len(s.Dims) == 0 {
			return base
		}
		var parts []string
		for _, d := range s.Dims {
			parts = append(parts, fmt.Sprint(d))
		}
		return map[string]string{"int": "TInt", "dbl": "TReal", "str": fmt.Sprintf("TStr%d", s.Len)}[s.Type] + strings.Join(parts, "x")
	}
	d := &dialect{ind: pick(x, []string{"  ", "  ", "    ", "   "}), cmt: "{ ", cmtEnd: " }",
		idx1: "%s[%s]", idx2: "%s[%s, %s]", intDiv: "%s div %s", mod: "%s mod %s", dble: "%s", ladder: "Trunc(%s * %s)",
		rel: stdRel, and: " and ", or: " or ", assign: "%s := %s;", call: "%s(%s);", callNone: "%s;", stop: "Halt(0);",
		ifOpen: "if %s then begin", ifElse: "end else begin", ifClose: "end;",
		forUp: "for %s := %s to %s do begin", forDown: "for %s := %s downto %s do begin", forClose: "end;",
		concat: ", ", lit: sq, order: "procs-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("%s(%s, %d)", padFn, x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("%s:%d", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s(%s, %d)", sgnFn, x, w) },
		print: func(p string) string { return "Writeln(Rpt, " + p + ");" },
	}
	if x.chance(0.4) {
		d.cmt, d.cmtEnd = "(* ", " *)"
	}
	if x.chance(0.4) { // begin on its own line
		d.ifOpen, d.ifElse = "if %s then\nbegin", "end\nelse\nbegin"
		d.forUp, d.forDown = "for %s := %s to %s do\nbegin", "for %s := %s downto %s do\nbegin"
	}
	if x.chance(0.25) {
		d.whileFor, d.whileHd, d.whileEnd = true, "while %s do begin", "end;"
	}
	kcase := x.intn(3)
	d.post = func(e *dialectEmitter) {
		if kcase == 0 {
			return
		}
		f := strings.ToUpper
		if kcase == 1 {
			f = func(w string) string { return strings.ToUpper(w[:1]) + w[1:] }
		}
		recaseKeywords(e, []string{"program", "const", "type", "var", "begin", "end", "if", "then", "else", "for", "to", "downto", "do",
			"while", "repeat", "until", "procedure", "function", "array", "of", "record", "string", "div", "mod", "and", "or", "not"}, f, '\'', "{")
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if len(s.Init) > 0 && len(s.Dims) > 0 {
			return []string{fmt.Sprintf("%s: %s = (%s);", e.name(s.Name), typeName(e, s), initList(s, sq))}
		}
		if len(s.Init) > 0 && s.Type == "str" {
			return []string{fmt.Sprintf("%s: %s = %s;", e.name(s.Name), typeName(e, s), sq(s.Init[0]))}
		}
		return []string{fmt.Sprintf("%s: %s;", e.name(s.Name), typeName(e, s))}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw(fmt.Sprintf("program %s;", strings.Title(strings.ToLower(e.j.id.Program))))
		e.raw(pick(x, []string{"{$R+,S+,I+,N+}", "{$R+}{$N+}{$E+}", "{$I+,R-,N+}"}))
		e.blank()
		if btr {
			e.raw(pick(x, []string{"{$I BTRV.INC}", "{$I TURBTRV.INC}", "{$L BTRVTP.OBJ}"}) + "  { Btrieve interface: function BTRV(Op: Integer; var PosBlk, DataBuf; var DataLen: Integer; var KeyBuf; KeyNum: Integer): Integer }")
			e.blank()
		}
		e.raw("const")
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				e.defLine(1, fmt.Sprintf("%s = %s;", e.name(s.Name), s.Init[0]), s.Name)
			}
		}
		if btr {
			e.line(1, "B_OPEN = 0; B_CLOSE = 1; B_INSERT = 2; B_GET_EQUAL = 5; B_STAT = 15;")
		}
		e.blank()
		e.raw("type")
		seen := map[string]bool{}
		all := append([]Sym{}, e.j.p.Globals...)
		for _, p := range e.j.p.Procs {
			all = append(all, p.Locals...)
		}
		for _, s := range all {
			if len(s.Dims) == 0 {
				continue
			}
			t := typeName(e, s)
			if seen[t] {
				continue
			}
			seen[t] = true
			elem := map[string]string{"int": "LongInt", "dbl": "Real", "str": fmt.Sprintf("string[%d]", s.Len)}[s.Type]
			e.line(1, fmt.Sprintf("%s = array[%s] of %s;", t, e.dimList(s, func(dn string) string { return "1.." + dn }, ", "), elem))
		}
		if btr {
			e.line(1, recT+" = record")
			e.line(2, fmt.Sprintf("%s: string[8]; %s: string[40]; %s, %s, %s: %s;", fRec, fMet, fBase, fCur, fChg, intT))
			if x.chance(0.5) {
				e.line(2, pick(x, []string{"RunDate: string[8];", "Filler: array[1..16] of Char;", "Operator: string[3];", "Reserved: array[1..8] of Byte;"}))
			}
			e.line(1, "end;")
		}
		e.blank()
		e.raw("const { typed constants: initialised tables }")
		for _, s := range e.j.p.Globals {
			if len(s.Init) > 0 && s.Role != "const" && (len(s.Dims) > 0 || s.Type == "str") {
				e.defLine(1, d.decl(e, s, "global")[0], s.Name)
			}
		}
		e.blank()
		e.raw("var")
		for _, s := range e.j.p.Globals {
			if s.Role == "const" || (len(s.Init) > 0 && (len(s.Dims) > 0 || s.Type == "str")) {
				continue
			}
			e.defLine(1, d.decl(e, s, "global")[0], s.Name)
		}
		e.defLine(1, "Card: string[80];", "card")
		e.line(1, "Code: Integer;")
		if fieldStyle == 2 {
			e.line(1, "Buf: string[16];")
		}
		e.line(1, "Deck, Rpt: Text;")
		if btr {
			e.line(1, posV+": array[0..127] of Byte;")
			e.line(1, recV+": "+recT+";")
			e.line(1, "KeyBuf: string[64];")
			e.line(1, "BLen, BStat: Integer;")
		}
		e.blank()
		e.line(0, "function "+padFn+"(S: string; W: Integer): string;")
		switch x.intn(3) {
		case 0:
			e.line(0, "begin")
			e.line(1, "while Length(S) < W do S := S + ' ';")
			e.line(1, padFn+" := Copy(S, 1, W);")
		case 1:
			e.line(0, "var T: string;")
			e.line(0, "begin")
			e.line(1, "FillChar(T, SizeOf(T), ' ');")
			e.line(1, "T := S + Copy(T, 1, W);")
			e.line(1, padFn+" := Copy(T, 1, W);")
		default:
			e.line(0, "var I: Integer;")
			e.line(0, "begin")
			e.line(1, "for I := Length(S) + 1 to W do S[I] := ' ';")
			e.line(1, "S[0] := Chr(W);  { length byte: Turbo strings are ShortStrings }")
			e.line(1, padFn+" := S;")
		}
		e.line(0, "end;")
		e.blank()
		e.line(0, "function "+sgnFn+"(N: "+intT+"; W: Integer): string;")
		e.line(0, "var S: string[20];")
		e.line(0, "begin")
		if x.chance(0.5) {
			e.line(1, "Str(N, S);")
			e.line(1, "if N > 0 then S := '+' + S;")
			e.line(1, "while Length(S) < W do S := ' ' + S;")
		} else {
			e.line(1, "Str(N:W - 1, S);")
			e.line(1, "if N > 0 then S := '+' + S else S := ' ' + S;")
			e.line(1, "if Length(S) > W then Delete(S, 1, Length(S) - W);")
		}
		e.line(1, sgnFn+" := S;")
		e.line(0, "end;")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("begin")
			deck := pick(x, []string{"'DECK.DAT'", "'DECK.DAT'", "ParamStr(1)", "'A:\\DECK.DAT'", "'C:\\RSCH\\DECK.DAT'"})
			if ioCheck {
				e.line(1, "Assign(Deck, "+deck+");")
				e.line(1, "{$I-} Reset(Deck); {$I+}")
				e.line(1, "if IOResult <> 0 then begin Writeln('CANNOT OPEN DECK'); Halt(2) end;")
			} else {
				e.line(1, "Assign(Deck, "+deck+"); Reset(Deck);")
			}
			e.line(1, "Assign(Rpt, "+pick(x, []string{"'REPORT.TXT'", "'PRN'", "'LPT1'", "'REPORT.LST'"})+"); Rewrite(Rpt);")
			return
		}
		var ps []string
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			if len(s.Dims) > 0 {
				ps = append(ps, "var "+q+": "+typeName(e, s))
			} else {
				ps = append(ps, q+": "+typeName(e, s))
			}
		}
		head := "procedure " + p.Name
		if len(ps) > 0 {
			head += "(" + strings.Join(ps, "; ") + ")"
		}
		e.raw(head + ";")
		var locals []Sym
		for _, l := range p.Locals {
			if l.Role != "param" {
				locals = append(locals, l)
			}
		}
		if len(locals) > 0 {
			e.raw("var")
			for _, l := range locals {
				e.defLine(1, fmt.Sprintf("%s: %s;", l.Name, typeName(e, l)), l.Name)
			}
		}
		e.raw("begin")
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			if btr {
				nm := e.j.nm
				e.comment(1, "Append this run to the Btrieve history file. "+e.j.g.comment(nil, "data", "workarounds", "vendors"))
				file := pick(x, []string{"BNCHHIST.BTR", "HISTORY.DAT", "RUNHIST.BTR", "F:\\RSCH\\HIST.BTR"})
				call := func(op string) string {
					return fmt.Sprintf("BTRV(%s, %s, %s, BLen, KeyBuf[1], %s)", op, posV, recV, map[bool]string{true: openMode, false: "0"}[op == "B_OPEN"])
				}
				e.line(1, "BLen := SizeOf("+recV+"); KeyBuf := '"+file+"'#0;")
				e.line(1, "BStat := "+call("B_OPEN")+";")
				if x.chance(0.4) {
					e.line(1, "if BStat = 0 then BStat := "+call("B_STAT")+";  { confirm the file really is a Btrieve file }")
				}
				msg := pick(x, []string{"ASK THE NETWARE ADMINISTRATOR.", "IS BTRIEVE.EXE LOADED?", "SEE RUN BOOK PAGE 4.", "CALL EXT. 4471."})
				e.line(1, "if BStat <> 0 then Writeln('BTRIEVE STATUS ', BStat, ' ON OPEN. "+msg+"')")
				e.line(1, "else begin")
				fill := func(depth int) {
					e.line(depth, fmt.Sprintf("%s.%s := %s; %s.%s := %s[Code];", recV, fRec, nm.get("recid"), recV, fMet, nm.get("mname")), nm.get("recid"), nm.get("mname"))
					e.line(depth, fmt.Sprintf("%s.%s := %s[Code]; %s.%s := %s[Code]; %s.%s := %s[Code];", recV, fBase, nm.get("bsln"), recV, fCur, nm.get("curr"), recV, fChg, nm.get("delt")), nm.get("bsln"), nm.get("curr"), nm.get("delt"))
					e.line(depth, "BStat := "+call("B_INSERT")+";")
					if x.chance(0.5) {
						e.line(depth, "case BStat of")
						e.line(depth+1, "0: ;")
						e.line(depth+1, "5: BStat := 0;  { duplicate key: loaded by an earlier run }")
						e.line(depth+1, "else Writeln('BTRIEVE STATUS ', BStat, ' ON INSERT ', Code);")
						e.line(depth, "end;")
					} else {
						e.line(depth, "if BStat = 5 then { duplicate key: already loaded by an earlier run } BStat := 0;")
					}
				}
				if x.chance(0.5) {
					e.line(2, "for Code := 1 to "+nm.get("nmet")+" do begin", nm.get("nmet"))
					fill(3)
					e.line(2, "end;")
				} else {
					e.line(2, "Code := 1;")
					e.line(2, "while Code <= "+nm.get("nmet")+" do begin", nm.get("nmet"))
					fill(3)
					e.line(3, "Code := Code + 1;")
					e.line(2, "end;")
				}
				e.line(2, "BStat := "+call("B_CLOSE")+";")
				e.line(1, "end;")
			}
			e.line(1, "Close(Rpt); Close(Deck);")
			e.raw("end.")
			return
		}
		e.raw("end;")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "Readln(Deck, Card);") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		switch {
		case num && fieldStyle == 1:
			return fmt.Sprintf("Val(Copy(Card, %d, %d), %s, Code); if Code <> 0 then %s := 0;", start+1, width, t, t)
		case num && fieldStyle == 2:
			return fmt.Sprintf("Buf := Copy(Card, %d, %d); while (Length(Buf) > 0) and (Buf[1] = ' ') do Delete(Buf, 1, 1); Val(Buf, %s, Code);", start+1, width, t)
		case num:
			return fmt.Sprintf("Val(Copy(Card, %d, %d), %s, Code);", start+1, width, t)
		case fieldStyle == 2:
			return fmt.Sprintf("%s := Copy(Card + '%s', %d, %d);", t, strings.Repeat(" ", 8), start+1, width)
		}
		return fmt.Sprintf("%s := Copy(Card, %d, %d);", t, start+1, width)
	}
	return d
}

// ── PL/I ──────────────────────────────────────────────────────────────────

func pliDialect(j *job) *dialect {
	x := j.x
	typ := func(s Sym) string {
		switch s.Type {
		case "dbl":
			return "FLOAT DEC(16)"
		case "str":
			return fmt.Sprintf("CHAR(%d)", maxi(s.Len, 1))
		}
		return "FIXED BIN(31)"
	}
	split := func(p string) (string, string) {
		i := strings.IndexByte(p, '\x00')
		return p[:i], p[i+1:]
	}
	d := &dialect{ind: pick(x, []string{"   ", "  ", "    "}), cmt: "/* ", cmtEnd: " */", upper: true,
		idx1: "%s(%s)", idx2: "%s(%s,%s)", intDiv: "DIVIDE(%s,%s,31)", mod: "MOD(%s,%s)", dble: "FLOAT(%s,16)", ladder: "TRUNC(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "^=", "gt": ">", "ge": ">="}, and: " & ", or: " | ",
		assign: "%s = %s;", call: "CALL %s(%s);", callNone: "CALL %s;", stop: "STOP;",
		ifOpen: "IF %s THEN DO;", ifElse: "END; ELSE DO;", ifClose: "END;",
		forUp: "DO %s = %s TO %s;", forDown: "DO %s = %s TO %s BY -1;", forClose: "END;",
		concat: "\x01", lit: func(s string) string { return sq(cardText(s)) + "\x00A" },
		pstr: func(x string, w int) string { return fmt.Sprintf("%s\x00A(%d)", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("%s\x00F(%d)", x, w) },
		psgn: func(x string, w int) string { return fmt.Sprintf("%s\x00P'%s9'", x, strings.Repeat("S", w-1)) },
	}
	d.print = func(p string) string {
		var xs, fs []string
		for _, part := range strings.Split(p, "\x01") {
			a, b := split(part)
			xs = append(xs, a)
			fs = append(fs, b)
		}
		return "PUT FILE(RPT) SKIP EDIT (" + strings.Join(xs, ", ") + ") (" + strings.Join(fs, ", ") + ");"
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("%%DCL %s FIXED; %%%s = %s;", s.Name, s.Name, s.Init[0])}
		}
		n := s.Name
		if len(s.Dims) > 0 {
			n += "(" + e.dimList(s, func(dn string) string { return dn }, ",") + ")"
		}
		init := ""
		if len(s.Init) > 0 {
			init = " INIT(" + initList(s, sq) + ")"
		} else if s.Type != "str" && x.chance(0.5) {
			init = " INIT(0)"
			if len(s.Dims) > 0 {
				init = " INIT((*)0)"
			}
		}
		st := ""
		if scope == "global" {
			st = " STATIC"
		}
		return []string{fmt.Sprintf("DCL %-10s %s%s%s;", n, typ(s), st, init)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				e.defLine(0, d.decl(e, s, "global")[0], s.Name)
			}
		}
		e.raw(fmt.Sprintf(" %s: PROCEDURE OPTIONS(MAIN)%s;", e.j.id.Program, pick(x, []string{"", " REORDER", ""})))
		for _, s := range e.j.p.Globals {
			if s.Role != "const" {
				e.defLine(1, d.decl(e, s, "global")[0], s.Name)
			}
		}
		e.defLine(1, "DCL CARD       CHAR(80);", "CARD")
		e.line(1, "DCL (DECK, RPT) FILE STREAM;")
		e.line(1, "DCL (MOD, DIVIDE, TRUNC, FLOAT, SUBSTR) BUILTIN;")
		e.line(1, "ON ENDFILE(DECK) BEGIN; PUT FILE(RPT) SKIP LIST('*** UNEXPECTED END OF DECK ***'); STOP; END;")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "OPEN FILE(DECK) INPUT TITLE('SYSIN'), FILE(RPT) OUTPUT PRINT TITLE('SYSPRINT') PAGESIZE(66) LINESIZE(133);")
			return
		}
		head := p.Name + ": PROC"
		if len(p.Params) > 0 {
			head += "(" + strings.Join(p.Params, ", ") + ")"
		}
		e.line(1, head+";")
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			n := q
			if len(s.Dims) > 0 {
				n += "(" + strings.TrimSuffix(strings.Repeat("*,", len(s.Dims)), ",") + ")"
			}
			e.defLine(2, fmt.Sprintf("DCL %-10s %s;", n, typ(s)), q)
		}
		for _, l := range p.Locals {
			if l.Role != "param" {
				e.defLine(2, d.decl(e, l, "local")[0], l.Name)
			}
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "CLOSE FILE(DECK), FILE(RPT);")
			return
		}
		e.line(1, "END "+p.Name+";")
	}
	d.footer = func(e *dialectEmitter) { e.raw(" END " + e.j.id.Program + ";") }
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "GET FILE(DECK) EDIT(CARD) (A(80));") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		return fmt.Sprintf("%s = SUBSTR(CARD,%d,%d);", t, start+1, width)
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 1
		}
		return 2
	}
	return d
}

// ── Paradox PAL ───────────────────────────────────────────────────────────

func palDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"  ", "   ", "    "}), cmt: "; ",
		idx1: "%s[%s]", idx2: "%s[(%s - 1) * {D1} + %s]", intDiv: "INT(%s / %s)", mod: "MOD(%s, %s)", dble: "%s", ladder: "INT(%s * %s)",
		rel: stdRel, and: " AND ", or: " OR ", assign: "%s = %s", call: "%s(%s)", callNone: "%s()", stop: "QUIT",
		ifOpen: "IF %s THEN", ifElse: "ELSE", ifClose: "ENDIF",
		forUp: "FOR %s FROM %s TO %s", forDown: "FOR %s FROM %s TO %s STEP -1", forClose: "ENDFOR",
		concat: " + ", lit: dq, order: "procs-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("FORMAT(\"W%d,AL\", %s)", w, x) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("FORMAT(\"W%d,AR\", %s)", w, x) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("FORMAT(\"W%d,AR,S+\", %s)", w, x) },
		print: func(p string) string { return "PRINT FILE rptFile " + p + ", \"\\n\"" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if len(s.Dims) == 0 {
			return nil
		}
		size := e.dimName(s.Dims[0])
		if len(s.Dims) == 2 {
			size = e.dimName(s.Dims[0]) + " * " + e.dimName(s.Dims[1])
		}
		return []string{fmt.Sprintf("ARRAY %s[%s]", e.name(s.Name), size)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("; " + strings.ToUpper(e.j.id.Program) + ".SC -- run from the Paradox main menu: Scripts/Play")
		constInits(e, 0)
		e.globalsDecl("global")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.comment(0, "Main script. "+e.j.g.comment(nil, "operators", "printing"))
			e.line(0, "CLEARALL")
			e.line(0, "rptFile = \"REPORT.TXT\"")
			scalarInits(e, 0)
			elementInits(e, 0)
			e.line(0, "VIEW \"DECK\"")
			e.line(0, "MOVETO RECORD 1")
			return
		}
		e.raw(fmt.Sprintf("PROC %s(%s)", p.Name, strings.Join(p.Params, ", ")))
		var priv []string
		for _, l := range p.Locals {
			if l.Role != "param" {
				priv = append(priv, l.Name)
			}
		}
		if len(priv) > 0 {
			e.defLine(1, "PRIVATE "+strings.Join(priv, ", "), priv...)
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "CLEARIMAGE")
			e.line(0, "RELEASE VARS ALL")
			return
		}
		e.raw("ENDPROC")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "card = [Card]")
		e.line(depth, "SKIP 1")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = NUMVAL(SUBSTR(card, %d, %d))", t, start+1, width)
		}
		return fmt.Sprintf("%s = SUBSTR(card, %d, %d)", t, start+1, width)
	}
	return d
}

// ── xBase: dBASE IV, Clipper, FoxPro ──────────────────────────────────────

func xbaseDialect(j *job) *dialect {
	x := j.x
	k := j.id.LangKey
	d := &dialect{ind: pick(x, []string{"   ", "  ", "    "}), cmt: "* ", upper: k != "foxpro",
		idx1: "%s[%s]", idx2: "%s[%s, %s]", intDiv: "INT(%s / %s)", mod: "MOD(%s, %s)", dble: "%s", ladder: "INT(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "#", "gt": ">", "ge": ">="}, and: " .AND. ", or: " .OR. ",
		assign: "%s = %s", call: "DO %s WITH %s", callNone: "DO %s", stop: "QUIT",
		ifOpen: "IF %s", ifElse: "ELSE", ifClose: "ENDIF",
		concat: " + ", lit: dq, print: func(p string) string { return "? " + p },
		pstr: func(x string, w int) string { return fmt.Sprintf("PADR(%s, %d)", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("STR(%s, %d)", x, w) },
		psgn: func(x string, w int) string {
			return fmt.Sprintf("PADL(IIF(%[1]s > 0, \"+\", \"\") + LTRIM(STR(%[1]s)), %[2]d)", x, w)
		},
	}
	switch k {
	case "dbase":
		d.whileFor, d.whileHd, d.whileEnd = true, "DO WHILE %s", "ENDDO"
		d.pstr = func(x string, w int) string { return fmt.Sprintf("LEFT(%s + SPACE(%d), %d)", x, w, w) }
		d.psgn = func(x string, w int) string { return fmt.Sprintf("STR(%s, %d)", x, w) }
	case "clipper":
		d.assign, d.call, d.callNone = "%s := %s", "%s(%s)", "%s()"
		d.forUp, d.forDown, d.forClose = "FOR %s := %s TO %s", "FOR %s := %s TO %s STEP -1", "NEXT"
		d.cmt = pick(x, []string{"// ", "* ", "&& "})
	default:
		d.idx1, d.idx2 = "%s(%s)", "%s(%s, %s)"
		d.forUp, d.forDown, d.forClose = "FOR %s = %s TO %s", "FOR %s = %s TO %s STEP -1", "ENDFOR"
		d.cmt = pick(x, []string{"* ", "&& "})
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if len(s.Dims) == 0 {
			if scope == "global" {
				return []string{"PUBLIC " + e.name(s.Name)}
			}
			return nil
		}
		dims := e.dimList(s, func(dn string) string { return dn }, ", ")
		switch k {
		case "dbase":
			return []string{fmt.Sprintf("DECLARE %s[%s]", e.name(s.Name), dims)}
		case "clipper":
			return []string{fmt.Sprintf("PUBLIC %s[%s]", e.name(s.Name), dims)}
		}
		return []string{fmt.Sprintf("DIMENSION %s(%s)", e.name(s.Name), dims)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		if k == "clipper" {
			e.raw("#include \"inkey.ch\"")
		}
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				if k == "dbase" {
					e.defLine(0, fmt.Sprintf("%s = %s", s.Name, s.Init[0]), s.Name)
				} else {
					e.defLine(0, fmt.Sprintf("#define %s %s", s.Name, s.Init[0]), s.Name)
				}
			}
		}
		e.line(0, "SET TALK OFF")
		e.line(0, "SET SAFETY OFF")
		e.line(0, "SET CONSOLE OFF")
		e.line(0, "SET ALTERNATE TO REPORT.TXT")
		e.line(0, "SET ALTERNATE ON")
		e.globalsDecl("global")
		e.line(0, "PUBLIC m_card")
		scalarInits(e, 0)
		elementInits(e, 0)
		e.line(0, "USE DECK")
		e.line(0, "GO TOP")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			return
		}
		var locals []string
		for _, l := range p.Locals {
			if l.Role != "param" {
				locals = append(locals, l.Name)
			}
		}
		switch k {
		case "clipper":
			e.raw(fmt.Sprintf("FUNCTION %s(%s)", p.Name, strings.Join(p.Params, ", ")))
			if len(locals) > 0 {
				e.defLine(1, "LOCAL "+strings.Join(locals, ", "), locals...)
			}
		default:
			e.raw("PROCEDURE " + p.Name)
			if len(p.Params) > 0 {
				kw := "PARAMETERS"
				if k == "foxpro" && e.j.in.Year >= 1995 {
					kw = "LPARAMETERS"
				}
				e.line(1, kw+" "+strings.Join(p.Params, ", "))
			}
			if len(locals) > 0 {
				e.defLine(1, "PRIVATE "+strings.Join(locals, ", "), locals...)
			}
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "SET ALTERNATE OFF")
			e.line(0, "SET ALTERNATE TO")
			e.line(0, "CLOSE DATABASES")
			e.line(0, "RETURN")
			return
		}
		if k == "clipper" {
			e.raw("RETURN NIL")
		} else {
			e.raw("RETURN")
		}
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "m_card = DECK->CARD")
		e.line(depth, "SKIP")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = VAL(SUBSTR(m_card, %d, %d))", t, start+1, width)
		}
		return fmt.Sprintf("%s = SUBSTR(m_card, %d, %d)", t, start+1, width)
	}
	return d
}

// ── Natural (Software AG, Adabas) ─────────────────────────────────────────

func naturalDialect(j *job) *dialect {
	x := j.x
	assign := pick(x, []string{"COMPUTE %s = %s", "ASSIGN %s = %s", "%s := %s"})
	fmtType := func(s Sym) string {
		switch s.Type {
		case "dbl":
			return "F8"
		case "str":
			return fmt.Sprintf("A%d", maxi(s.Len, 1))
		}
		return "I4"
	}
	d := &dialect{ind: "  ", cmt: "* ", upper: true, byGlobal: true,
		idx1: "%s(%s)", idx2: "%s(%s,%s)", intDiv: "%s / %s", mod: "%[1]s - (%[1]s / %[2]s) * %[2]s", dble: "%s", ladder: "%s * %s",
		rel: map[string]string{"lt": "LT", "le": "LE", "eq": "EQ", "ne": "NE", "gt": "GT", "ge": "GE"}, and: " AND ", or: " OR ",
		assign: assign, callNone: "PERFORM %s", call: "PERFORM %s", stop: "STOP",
		ifOpen: "IF %s", ifElse: "ELSE", ifClose: "END-IF",
		forUp: "FOR %s = %s TO %s", forDown: "FOR %s = %s TO %s STEP -1", forClose: "END-FOR",
		concat: " ", lit: func(s string) string { return sq(cardText(s)) },
		pstr:  func(x string, w int) string { return fmt.Sprintf("%s (AL=%d)", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("%s (EM=Z(%d)9)", x, w-1) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s (EM=+Z(%d)9)", x, w-2) },
		print: func(p string) string { return "WRITE NOTITLE " + p },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		t := fmtType(s)
		if len(s.Dims) > 0 {
			t += "/" + e.dimList(s, func(dn string) string { return "1:" + dn }, ",")
		}
		init := ""
		switch {
		case s.Role == "const":
			init = " CONST <" + s.Init[0] + ">"
		case len(s.Init) > 0 && s.Type == "str":
			init = " INIT <" + sq(s.Init[0]) + ">"
		case len(s.Init) > 0:
			init = " INIT <" + strings.Join(s.Init, ",") + ">"
		}
		return []string{fmt.Sprintf("1 %-12s (%s)%s", s.Name, t, init)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("DEFINE DATA LOCAL")
		for _, s := range e.j.p.Globals {
			e.defLine(0, d.decl(e, s, "global")[0], s.Name)
		}
		for _, s := range e.byGlobalSyms() {
			e.defLine(0, d.decl(e, s, "global")[0], s.Name)
		}
		e.defLine(0, "1 #CARD         (A80)", "#CARD")
		e.raw("END-DEFINE")
		e.raw("*")
		e.line(0, "FORMAT LS=133 PS=66")
		e.line(0, "DEFINE WORK FILE 1 '/rsch/deck/"+strings.ToLower(e.j.id.Program)+".dat'")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if !isMain {
			e.raw("DEFINE SUBROUTINE " + p.Name)
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(0, "ESCAPE ROUTINE")
			return
		}
		e.raw("END-SUBROUTINE")
	}
	d.footer = func(e *dialectEmitter) { e.raw("END") }
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "READ WORK FILE 1 ONCE #CARD") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("MOVE VAL(SUBSTR(#CARD,%d,%d)) TO %s", start+1, width, t)
		}
		return fmt.Sprintf("MOVE SUBSTR(#CARD,%d,%d) TO %s", start+1, width, t)
	}
	return d
}

// ── Informix 4GL ──────────────────────────────────────────────────────────

func informixDialect(j *job) *dialect {
	x := j.x
	typ := func(s Sym) string {
		base := map[string]string{"int": "INTEGER", "dbl": "FLOAT", "str": fmt.Sprintf("CHAR(%d)", maxi(s.Len, 1))}[s.Type]
		if base == "" {
			base = "INTEGER"
		}
		if len(s.Dims) == 0 {
			return base
		}
		var ds []string
		for _, d := range s.Dims {
			ds = append(ds, fmt.Sprint(d))
		}
		return "ARRAY[" + strings.Join(ds, ",") + "] OF " + base
	}
	d := &dialect{ind: pick(x, []string{"   ", "  ", "    "}), cmt: "# ",
		idx1: "%s[%s]", idx2: "%s[%s,%s]", intDiv: "%s / %s", mod: "%s MOD %s", dble: "%s", ladder: "fn_trunc(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "!=", "gt": ">", "ge": ">="}, and: " AND ", or: " OR ",
		assign: "LET %s = %s", call: "CALL %s(%s)", callNone: "CALL %s()", stop: "EXIT PROGRAM",
		ifOpen: "IF %s THEN", ifElse: "ELSE", ifClose: "END IF",
		forUp: "FOR %s = %s TO %s", forDown: "FOR %s = %s TO %s STEP -1", forClose: "END FOR",
		concat: ", ", lit: dq, order: "main-first",
		pstr:  func(x string, w int) string { return x },
		pnum:  func(x string, w int) string { return fmt.Sprintf("%s USING \"%s&\"", x, strings.Repeat("#", w-1)) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s USING \"%s&\"", x, strings.Repeat("+", w-1)) },
		print: func(p string) string { return "DISPLAY " + p },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		return []string{fmt.Sprintf("DEFINE %-12s %s", e.name(s.Name), typ(s))}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("DATABASE rsch")
		e.blank()
		e.raw("GLOBALS")
		e.globalsDecl("global")
		e.defLine(1, "DEFINE l_card        CHAR(80)", "l_card")
		e.raw("END GLOBALS")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("MAIN")
			for _, l := range p.Locals {
				if l.Role != "param" {
					e.defLine(1, d.decl(e, l, "local")[0], l.Name)
				}
			}
			e.line(1, "DEFER INTERRUPT")
			e.line(1, "WHENEVER ERROR STOP")
			constInits(e, 1)
			scalarInits(e, 1)
			elementInits(e, 1)
			e.line(1, "LOAD FROM \"deck.dat\" INSERT INTO deck_cards (card)")
			e.line(1, "DECLARE c_deck CURSOR FOR SELECT card FROM deck_cards ORDER BY rowid")
			e.line(1, "OPEN c_deck")
			return
		}
		e.raw(fmt.Sprintf("FUNCTION %s(%s)", p.Name, strings.Join(p.Params, ", ")))
		for _, l := range p.Locals {
			e.defLine(1, d.decl(e, l, "local")[0], l.Name)
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "CLOSE c_deck")
			e.line(1, "DELETE FROM deck_cards")
			e.raw("END MAIN")
			return
		}
		e.raw("END FUNCTION")
	}
	d.footer = func(e *dialectEmitter) {
		e.blank()
		e.comment(0, "Truncate toward zero. "+e.j.g.comment(nil, "memos", "magic"))
		e.raw("FUNCTION fn_trunc(p_val)")
		e.line(1, "DEFINE p_val DECIMAL(16,4), l_int INTEGER")
		e.line(1, "LET l_int = p_val - (p_val MOD 1)")
		e.line(1, "RETURN l_int")
		e.raw("END FUNCTION")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "FETCH c_deck INTO l_card") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		return fmt.Sprintf("LET %s = l_card[%d,%d]", t, start+1, start+width)
	}
	return d
}

// ── Visual Basic / Excel VBA ──────────────────────────────────────────────

func vbDialect(j *job) *dialect {
	x := j.x
	vba := j.id.LangKey == "vba"
	typ := func(s Sym) string {
		switch s.Type {
		case "dbl":
			return "Double"
		case "str":
			return fmt.Sprintf("String * %d", maxi(s.Len, 1))
		}
		return "Long"
	}
	noParens := x.chance(0.4)
	d := &dialect{ind: pick(x, []string{"    ", "    ", "  ", "\t"}), cmt: "' ",
		idx1: "%s(%s)", idx2: "%s(%s, %s)", intDiv: "%s \\ %s", mod: "%s Mod %s", dble: "CDbl(%s)", ladder: "Fix(CDbl(%s) * %s)",
		rel: stdRel, and: " And ", or: " Or ", assign: "%s = %s", call: "Call %s(%s)", callNone: "Call %s", stop: "End",
		ifOpen: "If %s Then", ifElse: "Else", ifClose: "End If",
		forUp: "For %s = %s To %s", forDown: "For %s = %s To %s Step -1", forClose: "Next %s",
		concat: " & ", lit: dq, order: "main-first",
		pstr: func(x string, w int) string { return fmt.Sprintf("Left$(%s & Space$(%d), %d)", x, w, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("Right$(Space$(%d) & CStr(%s), %d)", w, x, w) },
		psgn: func(x string, w int) string {
			return fmt.Sprintf("Right$(Space$(%d) & Format$(%s, \"+0;-0;0\"), %d)", w, x, w)
		},
		print: func(p string) string { return "Print #2, " + p },
	}
	if noParens {
		d.call, d.callNone = "%s %s", "%s"
	}
	if vba {
		d.print = func(p string) string { return "wsOut.Cells(lRow, 1).Value = " + p + ": lRow = lRow + 1" }
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		kw := "Dim"
		if scope == "global" {
			kw = pick(x, []string{"Dim", "Private", "Public"})
		}
		if s.Role == "const" {
			return []string{fmt.Sprintf("Public Const %s As Long = %s", e.name(s.Name), s.Init[0])}
		}
		n := s.Name
		if len(s.Dims) > 0 {
			n += "(" + e.dimList(s, func(dn string) string { return "1 To " + dn }, ", ") + ")"
		}
		return []string{fmt.Sprintf("%s %s As %s", kw, n, typ(s))}
	}
	d.header = func(e *dialectEmitter) {
		if vba {
			e.raw("Attribute VB_Name = \"mod" + strings.Title(strings.ToLower(e.j.id.Program)) + "\"")
		} else {
			e.raw("Attribute VB_Name = \"Module1\"")
		}
		e.headerComments()
		e.raw("Option Explicit")
		e.raw("Option Base 1")
		e.blank()
		for _, s := range e.j.p.Globals {
			if s.Role == "const" {
				e.defLine(0, d.decl(e, s, "global")[0], s.Name)
			}
		}
		for _, s := range e.j.p.Globals {
			if s.Role != "const" {
				e.defLine(0, d.decl(e, s, "global")[0], s.Name)
			}
		}
		e.defLine(0, "Dim sCard As String", "sCard")
		if vba {
			e.line(0, "Dim wsDeck As Worksheet, wsOut As Worksheet")
			e.line(0, "Dim lCard As Long, lRow As Long")
		}
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			if vba {
				e.raw("Public Sub Run" + strings.Title(strings.ToLower(e.j.id.Program)) + "()")
				e.line(1, "' Assigned to the big grey button on the Control sheet. "+asciiText(e.j.g.comment(nil, "operators", "superstition")))
			} else {
				e.raw("Sub Main()")
			}
			for _, l := range p.Locals {
				if l.Role != "param" {
					e.defLine(1, d.decl(e, l, "local")[0], l.Name)
				}
			}
			e.line(1, "On Error GoTo BadDeck")
			scalarInits(e, 1)
			elementInits(e, 1)
			if vba {
				e.line(1, "Application.ScreenUpdating = False")
				e.line(1, "Set wsDeck = ThisWorkbook.Worksheets(\"Deck\")")
				e.line(1, "Set wsOut = ThisWorkbook.Worksheets(\"Report\")")
				e.line(1, "wsOut.Cells.ClearContents")
				e.line(1, "lCard = 1: lRow = 1")
			} else {
				e.line(1, "Open App.Path & \"\\DECK.DAT\" For Input As #1")
				e.line(1, "Open App.Path & \"\\REPORT.TXT\" For Output As #2")
			}
			return
		}
		var ps []string
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			if len(s.Dims) > 0 {
				ps = append(ps, q+"() As "+typ(s))
			} else {
				ps = append(ps, "ByVal "+q+" As "+typ(s))
			}
		}
		e.raw(fmt.Sprintf("Private Sub %s(%s)", p.Name, strings.Join(ps, ", ")))
		e.locals(p, 1)
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			if vba {
				e.line(1, "Application.ScreenUpdating = True")
			} else {
				e.line(1, "Close #2: Close #1")
			}
			e.line(1, "Exit Sub")
			e.raw("BadDeck:")
			e.line(1, "MsgBox \"Error \" & Err.Number & \" (\" & Err.Description & \"). Call \" & \""+asciiText(e.j.g.coder().Name())+"\" & \".\", vbCritical")
			e.raw("End Sub")
			return
		}
		e.raw("End Sub")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		if vba {
			e.line(depth, "sCard = CStr(wsDeck.Cells(lCard, 1).Value): lCard = lCard + 1")
			return
		}
		e.line(depth, "Line Input #1, sCard")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = Val(Mid$(sCard, %d, %d))", t, start+1, width)
		}
		return fmt.Sprintf("%s = Mid$(sCard, %d, %d)", t, start+1, width)
	}
	return d
}

// ── VB.NET ────────────────────────────────────────────────────────────────

func vbnetDialect(j *job) *dialect {
	x := j.x
	typ := func(s Sym) string {
		return map[string]string{"int": "Integer", "dbl": "Double", "str": "String"}[s.Type]
	}
	d := &dialect{ind: "    ", cmt: "' ",
		idx1: "%s(%s)", idx2: "%s(%s, %s)", intDiv: "%s \\ %s", mod: "%s Mod %s", dble: "CDbl(%s)", ladder: "CInt(Math.Truncate(CDbl(%s) * %s))",
		rel: stdRel, and: " AndAlso ", or: " OrElse ", assign: "%s = %s", call: "%s(%s)", callNone: "%s()", stop: "Environment.Exit(0)",
		ifOpen: "If %s Then", ifElse: "Else", ifClose: "End If",
		forUp: "For %s = %s To %s", forDown: "For %s = %s To %s Step -1", forClose: "Next",
		concat: " & ", lit: dq, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("Left(%s & Space(%d), %d)", x, w, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("%s.ToString().PadLeft(%d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s.ToString(\"+0;-0;0\").PadLeft(%d)", x, w) },
		print: func(p string) string { return "Console.WriteLine(" + p + ")" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		kw := "Dim"
		if scope == "global" {
			kw = pick(x, []string{"Private", "Friend", "Public"})
		}
		if s.Role == "const" {
			return []string{fmt.Sprintf("%s Const %s As Integer = %s", kw, e.name(s.Name), s.Init[0])}
		}
		switch {
		case len(s.Init) > 0 && len(s.Dims) > 0:
			return []string{fmt.Sprintf("%s %s() As %s = {%s, %s}", kw, s.Name, typ(s), map[string]string{"str": "\"\""}[s.Type]+map[string]string{"int": "0", "dbl": "0.0"}[s.Type], initList(s, dq))}
		case len(s.Dims) > 0:
			return []string{fmt.Sprintf("%s %s(%s) As %s", kw, s.Name, e.dimList(s, func(dn string) string { return dn }, ", "), typ(s))}
		case len(s.Init) > 0 && s.Type == "str":
			return []string{fmt.Sprintf("%s %s As String = %s", kw, s.Name, dq(s.Init[0]))}
		case len(s.Init) > 0:
			return []string{fmt.Sprintf("%s %s As %s = %s", kw, s.Name, typ(s), s.Init[0])}
		case s.Type == "str":
			return []string{fmt.Sprintf("%s %s As String = \"\"", kw, s.Name)}
		}
		return []string{fmt.Sprintf("%s %s As %s", kw, s.Name, typ(s))}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("Option Strict Off")
		e.raw("Option Explicit On")
		e.raw("Imports System")
		e.raw("Imports System.IO")
		e.blank()
		e.raw("Module " + strings.Title(strings.ToLower(e.j.id.Program)))
		for _, s := range e.j.p.Globals {
			e.defLine(1, d.decl(e, s, "global")[0], s.Name)
		}
		e.defLine(1, "Private card As String = \"\"", "card")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "Sub Main()")
			e.locals(p, 2)
			return
		}
		var ps []string
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			if len(s.Dims) > 0 {
				ps = append(ps, "ByRef "+q+"() As "+typ(s))
			} else {
				ps = append(ps, "ByVal "+q+" As "+typ(s))
			}
		}
		e.line(1, fmt.Sprintf("Private Sub %s(%s)", p.Name, strings.Join(ps, ", ")))
		e.locals(p, 2)
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) { e.line(1, "End Sub") }
	d.footer = func(e *dialectEmitter) { e.raw("End Module") }
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "card = If(Console.ReadLine(), \"\").PadRight(80)")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = CInt(Val(card.Substring(%d, %d)))", t, start, width)
		}
		return fmt.Sprintf("%s = card.Substring(%d, %d)", t, start, width)
	}
	d.bodyDepth = func(bool) int { return 2 }
	return d
}

// ── C# ────────────────────────────────────────────────────────────────────

func csharpDialect(j *job) *dialect {
	x := j.x
	typ := func(s Sym) string { return map[string]string{"int": "int", "dbl": "double", "str": "string"}[s.Type] }
	csq := func(s string) string { return "\"" + strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(s) + "\"" }
	d := &dialect{ind: "    ", cmt: "// ",
		idx1: "%s[%s]", idx2: "%s[%s, %s]", intDiv: "%s / %s", mod: "%s %% %s", dble: "(double)%s", ladder: "(int)Math.Truncate((double)%s * %s)",
		rel: cRelOps, and: " && ", or: " || ", assign: "%s = %s;", call: "%s(%s);", callNone: "%s();", stop: "Environment.Exit(0);",
		ifOpen: "if (%s) {", ifElse: "} else {", ifClose: "}",
		forUp: "for (%[1]s = %[2]s; %[1]s <= %[3]s; %[1]s++) {", forDown: "for (%[1]s = %[2]s; %[1]s >= %[3]s; %[1]s--) {", forClose: "}",
		concat: " + ", lit: csq, order: "main-first",
		pstr: func(x string, w int) string {
			return fmt.Sprintf("(%s ?? \"\").PadRight(%d).Substring(0, %d)", x, w, w)
		},
		pnum:  func(x string, w int) string { return fmt.Sprintf("%s.ToString().PadLeft(%d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("%s.ToString(\"+0;-0;0\").PadLeft(%d)", x, w) },
		print: func(p string) string { return "Console.WriteLine(" + p + ");" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		st := ""
		if scope == "global" {
			st = pick(x, []string{"static ", "private static ", "internal static "})
		}
		if s.Role == "const" {
			return []string{fmt.Sprintf("%sreadonly int %s = %s;", strings.Replace(st, "static ", "static ", 1), e.name(s.Name), s.Init[0])}
		}
		switch {
		case len(s.Init) > 0 && len(s.Dims) > 0:
			lead := map[string]string{"int": "0", "dbl": "0.0", "str": "\"\""}[s.Type]
			return []string{fmt.Sprintf("%s%s[] %s = { %s, %s };", st, typ(s), s.Name, lead, initList(s, csq))}
		case len(s.Dims) == 2:
			return []string{fmt.Sprintf("%s%s[,] %s = new %s[%s + 1, %s + 1];", st, typ(s), s.Name, typ(s), e.dimName(s.Dims[0]), e.dimName(s.Dims[1]))}
		case len(s.Dims) == 1:
			return []string{fmt.Sprintf("%s%s[] %s = new %s[%s + 1];", st, typ(s), s.Name, typ(s), e.dimName(s.Dims[0]))}
		case len(s.Init) > 0 && s.Type == "str":
			return []string{fmt.Sprintf("%s%s %s = %s;", st, typ(s), s.Name, csq(s.Init[0]))}
		case len(s.Init) > 0:
			return []string{fmt.Sprintf("%s%s %s = %s;", st, typ(s), s.Name, s.Init[0])}
		case s.Type == "str":
			return []string{fmt.Sprintf("%s%s %s = \"\";", st, typ(s), s.Name)}
		}
		return []string{fmt.Sprintf("%s%s %s = 0;", st, typ(s), s.Name)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("using System;")
		e.raw("using System.IO;")
		e.blank()
		e.raw("namespace Acpwb.Research." + strings.Title(strings.ToLower(e.j.id.Program)))
		e.raw("{")
		e.line(1, "internal static class Program")
		e.line(1, "{")
		for _, s := range e.j.p.Globals {
			e.defLine(2, d.decl(e, s, "global")[0], s.Name)
		}
		e.defLine(2, "static string card = \"\";", "card")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(2, "static void Main(string[] args)")
		} else {
			var ps []string
			for _, q := range p.Params {
				s := e.paramSym(p, q)
				t := typ(s)
				if len(s.Dims) == 2 {
					t += "[,]"
				} else if len(s.Dims) == 1 {
					t += "[]"
				}
				ps = append(ps, t+" "+q)
			}
			e.line(2, fmt.Sprintf("static void %s(%s)", p.Name, strings.Join(ps, ", ")))
		}
		e.line(2, "{")
		e.locals(p, 3)
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) { e.line(2, "}") }
	d.footer = func(e *dialectEmitter) {
		e.line(1, "}")
		e.raw("}")
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "card = (Console.ReadLine() ?? \"\").PadRight(80);")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = int.Parse(card.Substring(%d, %d).Trim().PadLeft(1, '0'));", t, start, width)
		}
		return fmt.Sprintf("%s = card.Substring(%d, %d);", t, start, width)
	}
	d.bodyDepth = func(bool) int { return 3 }
	return d
}

// ── PHP ───────────────────────────────────────────────────────────────────

func phpDialect(j *job) *dialect {
	x := j.x
	php7 := j.in.Year >= 2016
	d := &dialect{ind: pick(x, []string{"    ", "  ", "\t"}), cmt: pick(x, []string{"// ", "# ", "// "}), sigil: "$",
		idx1: "%s[%s]", idx2: "%s[%s][%s]", intDiv: "(int)(%s / %s)", mod: "%s %% %s", dble: "(float)%s", ladder: "(int)floor(%s * %s)",
		rel: cRelOps, and: " && ", or: " || ", assign: "%s = %s;", call: "%s(%s);", callNone: "%s();", stop: "exit(0);",
		ifOpen: "if (%s) {", ifElse: "} else {", ifClose: "}",
		forUp: "for (%[1]s = %[2]s; %[1]s <= %[3]s; %[1]s++) {", forDown: "for (%[1]s = %[2]s; %[1]s >= %[3]s; %[1]s--) {", forClose: "}",
		concat: " . ", lit: bsq, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("str_pad(substr(%s, 0, %d), %d)", x, w, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("str_pad(%s, %d, ' ', STR_PAD_LEFT)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("sprintf('%%+%dd', %s)", w, x) },
		print: func(p string) string { return "echo " + p + " . \"\\n\";" },
	}
	if php7 {
		d.intDiv = "intdiv(%s, %s)"
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("define('%s', %s);", strings.ToUpper(s.Name), s.Init[0])}
		}
		switch {
		case len(s.Init) > 0 && len(s.Dims) > 0:
			return []string{fmt.Sprintf("$%s = array(1 => %s);", s.Name, initList(s, bsq))}
		case len(s.Dims) > 0:
			return []string{fmt.Sprintf("$%s = array();", s.Name)}
		case len(s.Init) > 0 && s.Type == "str":
			return []string{fmt.Sprintf("$%s = %s;", s.Name, bsq(s.Init[0]))}
		case len(s.Init) > 0:
			return []string{fmt.Sprintf("$%s = %s;", s.Name, s.Init[0])}
		case s.Type == "str":
			return []string{fmt.Sprintf("$%s = '';", s.Name)}
		}
		return []string{fmt.Sprintf("$%s = 0;", s.Name)}
	}
	d.header = func(e *dialectEmitter) {
		e.raw("<?php")
		e.headerComments()
		if php7 {
			e.raw("declare(strict_types=0); // see the comment about strict types and the ladder")
		} else {
			e.raw("error_reporting(E_ALL ^ E_NOTICE);")
		}
		e.globalsDecl("global")
		e.defLine(0, "$card = '';", "card")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		name := p.Name
		var ps []string
		for _, q := range p.Params {
			if len(e.paramSym(p, q).Dims) > 0 {
				ps = append(ps, "&$"+q)
			} else {
				ps = append(ps, "$"+q)
			}
		}
		e.raw(fmt.Sprintf("function %s(%s)", name, strings.Join(ps, ", ")))
		e.raw("{")
		g := e.usedGlobals(p)
		if isMain || len(g) > 0 {
			var gs []string
			for _, n := range g {
				gs = append(gs, "$"+n)
			}
			gs = append(gs, "$card")
			e.line(1, "global "+strings.Join(gs, ", ")+";")
		}
		for _, l := range p.Locals {
			if l.Role != "param" {
				e.defLine(1, d.decl(e, l, "local")[0], l.Name)
			}
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) { e.raw("}") }
	d.footer = func(e *dialectEmitter) {
		e.blank()
		e.raw(e.j.p.Procs[0].Name + "();")
		e.raw("exit(0);")
		if !php7 {
			e.raw("?>")
		}
	}
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "$card = str_pad(rtrim(fgets(STDIN), \"\\r\\n\"), 80);")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = (int)substr($card, %d, %d);", t, start, width)
		}
		return fmt.Sprintf("%s = substr($card, %d, %d);", t, start, width)
	}
	return d
}

// ── Pick BASIC ────────────────────────────────────────────────────────────

func pickDialect(j *job) *dialect {
	x := j.x
	d := &dialect{ind: pick(x, []string{"   ", "  ", "    "}), cmt: pick(x, []string{"* ", "! ", "* "}), upper: true, byGlobal: true,
		idx1: "%s(%s)", idx2: "%s(%s,%s)", intDiv: "INT(%s / %s)", mod: "MOD(%s, %s)", dble: "%s", ladder: "INT(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "#", "gt": ">", "ge": ">="}, and: " AND ", or: " OR ",
		assign: "%s = %s", call: "GOSUB %s", callNone: "GOSUB %s", stop: "STOP",
		ifOpen: "IF %s THEN", ifElse: "END ELSE", ifClose: "END",
		forUp: "FOR %s = %s TO %s", forDown: "FOR %s = %s TO %s STEP -1", forClose: "NEXT %s",
		concat: ":", lit: func(s string) string { return sq(cardText(s)) }, order: "main-first",
		pstr: func(x string, w int) string { return fmt.Sprintf("FMT(%s, '%dL')", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("FMT(%s, '%dR')", x, w) },
		psgn: func(x string, w int) string {
			return fmt.Sprintf("FMT(IF %[1]s > 0 THEN '+':%[1]s ELSE %[1]s, '%[2]dR')", x, w)
		},
		print: func(p string) string { return "PRINT " + p },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("EQU %s TO %s", s.Name, s.Init[0])}
		}
		if len(s.Dims) == 0 {
			return nil
		}
		return []string{fmt.Sprintf("DIM %s(%s)", s.Name, e.dimList(s, func(dn string) string { return dn }, ","))}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.globalsDecl("global")
		for _, s := range e.byGlobalSyms() {
			for _, t := range d.decl(e, s, "global") {
				e.defLine(0, t, s.Name)
			}
		}
		scalarInits(e, 0)
		elementInits(e, 0)
		e.line(0, "OPEN '', 'DECK' TO F.DECK ELSE STOP 201, 'DECK'")
		e.line(0, "SEQ = 0")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if !isMain {
			e.raw(p.Name + ":")
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			return
		}
		e.line(1, "RETURN")
	}
	d.footer = func(e *dialectEmitter) { e.raw("END") }
	d.readCard = func(e *dialectEmitter, depth int) {
		e.line(depth, "SEQ = SEQ + 1")
		e.line(depth, "READ CARD FROM F.DECK, SEQ ELSE CARD = ''")
	}
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s = CARD[%d,%d] + 0", t, start+1, width)
		}
		return fmt.Sprintf("%s = CARD[%d,%d]", t, start+1, width)
	}
	return d
}

// ── MUMPS ─────────────────────────────────────────────────────────────────

func mumpsDialect(j *job) *dialect {
	d := &dialect{ind: " ", cmt: ";", upper: true, dots: true, parenAll: true,
		idx1: "%s(%s)", idx2: "%s(%s,%s)", intDiv: "%s\\%s", mod: "%s#%s", dble: "%s", ladder: "%s*%s\\1",
		rel: map[string]string{"lt": "<", "le": "'>", "eq": "=", "ne": "'=", "gt": ">", "ge": "'<"}, and: "&", or: "!",
		assign: "S %s=%s", call: "D %s(%s)", callNone: "D %s", stop: "Q",
		ifOpen: "I %s D", ifElse: "E  D", ifClose: "",
		forUp: "F %s=%s:1:%s D", forDown: "F %s=%s:-1:%s D", forClose: "",
		concat: ",", lit: dq, order: "main-first",
		pstr:  func(x string, w int) string { return fmt.Sprintf("$$PAD(%s,%d)", x, w) },
		pnum:  func(x string, w int) string { return fmt.Sprintf("$J(%s,%d)", x, w) },
		psgn:  func(x string, w int) string { return fmt.Sprintf("$J($S(%[1]s>0:\"+\"_%[1]s,1:%[1]s),%[2]d)", x, w) },
		print: func(p string) string { return "W !," + p },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string { return nil }
	d.header = func(e *dialectEmitter) {
		j := e.j
		c := j.g.roster[0]
		e.raw(fmt.Sprintf("%s ;RSCH/%s - %s ;%02d-%s-%04d", j.id.Program, c.Initials, cardText(trunc(j.in.Title, 40)), j.in.Day, monthAbbr[j.in.Month-1], j.in.Year))
		e.raw(fmt.Sprintf(" ;;%s;BENCHMARK TABULATION;**%d,%d**;%s %d, %04d;Build %d", j.id.Release, j.x.between(1, 40), j.x.between(41, 99), strings.Title(strings.ToLower(monthAbbr[j.in.Month-1])), j.in.Day, j.in.Year, j.x.between(2, 60)))
		e.headerComments()
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("EN ; main entry point -- called from TaskMan")
			e.line(1, "N DECK,IO,CARD S IO=$I,DECK=\"DECK.DAT\"")
			constInits(e, 1)
			scalarInits(e, 1)
			elementInits(e, 1)
			e.line(1, "O DECK:(READONLY):5 E  W !,\"*** DECK NOT AVAILABLE ***\" Q")
			return
		}
		head := p.Name
		if len(p.Params) > 0 {
			head += "(" + strings.Join(p.Params, ",") + ")"
		}
		e.raw(head + " ;")
		var locals []string
		for _, l := range p.Locals {
			if l.Role != "param" {
				locals = append(locals, l.Name)
			}
		}
		if len(locals) > 0 {
			e.defLine(1, "N "+strings.Join(locals, ","), locals...)
		}
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "C DECK")
		}
		e.line(1, "Q")
	}
	d.footer = func(e *dialectEmitter) {
		e.raw("PAD(X,W) ; left-justify X in W columns")
		e.line(1, "Q $E(X_$J(\"\",W),1,W)")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "U DECK R CARD:5 U IO") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("S %s=+$E(CARD,%d,%d)", t, start+1, start+width)
		}
		return fmt.Sprintf("S %s=$E(CARD,%d,%d)", t, start+1, start+width)
	}
	return d
}

// ── PL/SQL ────────────────────────────────────────────────────────────────

func plsqlDialect(j *job) *dialect {
	x := j.x
	typeName := func(s Sym) string {
		if len(s.Dims) == 2 {
			return map[string]string{"int": "int_mat", "dbl": "num_mat", "str": "str_mat"}[s.Type]
		}
		if len(s.Dims) == 1 {
			return map[string]string{"int": "int_tab", "dbl": "num_tab", "str": "str_tab"}[s.Type]
		}
		return map[string]string{"int": "PLS_INTEGER", "dbl": "NUMBER", "str": fmt.Sprintf("VARCHAR2(%d)", maxi(s.Len, 1))}[s.Type]
	}
	d := &dialect{ind: pick(x, []string{"  ", "   ", "    "}), cmt: "-- ",
		idx1: "%s(%s)", idx2: "%s(%s)(%s)", intDiv: "TRUNC(%s / %s)", mod: "MOD(%s, %s)", dble: "%s", ladder: "TRUNC(%s * %s)",
		rel: map[string]string{"lt": "<", "le": "<=", "eq": "=", "ne": "!=", "gt": ">", "ge": ">="}, and: " AND ", or: " OR ",
		assign: "%s := %s;", call: "%s(%s);", callNone: "%s;", stop: "RETURN;",
		ifOpen: "IF %s THEN", ifElse: "ELSE", ifClose: "END IF;",
		forUp: "FOR %s IN %s .. %s LOOP", forDown: "FOR %[1]s IN REVERSE %[3]s .. %[2]s LOOP", forClose: "END LOOP;",
		concat: " || ", lit: sq, order: "nested",
		pstr: func(x string, w int) string { return fmt.Sprintf("RPAD(NVL(%s, ' '), %d)", x, w) },
		pnum: func(x string, w int) string { return fmt.Sprintf("LPAD(TO_CHAR(%s), %d)", x, w) },
		psgn: func(x string, w int) string {
			return fmt.Sprintf("LPAD(CASE WHEN %[1]s > 0 THEN '+' END || TO_CHAR(%[1]s), %[2]d)", x, w)
		},
		print: func(p string) string { return "DBMS_OUTPUT.PUT_LINE(" + p + ");" },
	}
	d.decl = func(e *dialectEmitter, s Sym, scope string) []string {
		if s.Role == "const" {
			return []string{fmt.Sprintf("%-12s CONSTANT PLS_INTEGER := %s;", s.Name, s.Init[0])}
		}
		init := ""
		if len(s.Init) > 0 && len(s.Dims) == 0 {
			if s.Type == "str" {
				init = " := " + sq(s.Init[0])
			} else {
				init = " := " + s.Init[0]
			}
		} else if s.Type == "int" && len(s.Dims) == 0 {
			init = " := 0"
		}
		return []string{fmt.Sprintf("%-12s %s%s;", s.Name, typeName(s), init)}
	}
	d.header = func(e *dialectEmitter) {
		e.headerComments()
		e.raw("SET SERVEROUTPUT ON SIZE 1000000")
		e.raw("SET FEEDBACK OFF")
		e.raw("WHENEVER SQLERROR EXIT FAILURE ROLLBACK")
		e.blank()
		e.raw("DECLARE")
		e.line(1, "TYPE int_tab IS TABLE OF PLS_INTEGER INDEX BY BINARY_INTEGER;")
		e.line(1, "TYPE num_tab IS TABLE OF NUMBER INDEX BY BINARY_INTEGER;")
		e.line(1, "TYPE str_tab IS TABLE OF VARCHAR2(60) INDEX BY BINARY_INTEGER;")
		e.line(1, "TYPE int_mat IS TABLE OF int_tab INDEX BY BINARY_INTEGER;")
		e.line(1, "TYPE num_mat IS TABLE OF num_tab INDEX BY BINARY_INTEGER;")
		for _, s := range e.j.p.Globals {
			e.defLine(1, d.decl(e, s, "global")[0], s.Name)
		}
		for _, l := range e.j.p.Procs[0].Locals {
			if l.Role != "param" {
				e.defLine(1, d.decl(e, l, "local")[0], l.Name)
			}
		}
		e.defLine(1, "card         VARCHAR2(80);", "card")
		e.line(1, "deck         UTL_FILE.FILE_TYPE;")
	}
	d.procOpen = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.raw("BEGIN")
			elementInits(e, 1)
			e.line(1, "deck := UTL_FILE.FOPEN('BENCH_DIR', '"+strings.ToLower(e.j.id.Program)+".dat', 'R');")
			return
		}
		var ps []string
		var copies []Sym
		for _, q := range p.Params {
			s := e.paramSym(p, q)
			mode, n := "IN", q
			if len(s.Dims) > 0 {
				mode = "IN OUT"
			} else if assigns(p.Body, q) {
				n = "p_" + q
				copies = append(copies, s)
			}
			ps = append(ps, fmt.Sprintf("%s %s %s", n, mode, strings.Split(typeName(s), "(")[0]))
		}
		head := "PROCEDURE " + p.Name
		if len(ps) > 0 {
			head += " (" + strings.Join(ps, ", ") + ")"
		}
		e.line(1, head+" IS")
		for _, c := range copies {
			e.defLine(2, fmt.Sprintf("%-12s %s := p_%s;", c.Name, typeName(c), c.Name), c.Name)
		}
		for _, l := range p.Locals {
			if l.Role != "param" {
				e.defLine(2, d.decl(e, l, "local")[0], l.Name)
			}
		}
		e.line(1, "BEGIN")
	}
	d.procClose = func(e *dialectEmitter, p *Proc, isMain bool) {
		if isMain {
			e.line(1, "UTL_FILE.FCLOSE(deck);")
			e.raw("EXCEPTION")
			e.line(1, "WHEN NO_DATA_FOUND THEN")
			e.line(2, "DBMS_OUTPUT.PUT_LINE('*** UNEXPECTED END OF DECK -- CHECK THE CARD COUNT ***');")
			e.line(2, "IF UTL_FILE.IS_OPEN(deck) THEN UTL_FILE.FCLOSE(deck); END IF;")
			e.line(2, "RAISE;")
			e.raw("END;")
			e.raw("/")
			e.raw("EXIT")
			return
		}
		e.line(1, "END "+p.Name+";")
	}
	d.readCard = func(e *dialectEmitter, depth int) { e.line(depth, "UTL_FILE.GET_LINE(deck, card);") }
	d.field = func(e *dialectEmitter, t string, start, width int, num bool) string {
		if num {
			return fmt.Sprintf("%s := TO_NUMBER(NVL(TRIM(SUBSTR(card, %d, %d)), '0'));", t, start+1, width)
		}
		return fmt.Sprintf("%s := SUBSTR(card, %d, %d);", t, start+1, width)
	}
	d.bodyDepth = func(isMain bool) int {
		if isMain {
			return 1
		}
		return 2
	}
	return d
}

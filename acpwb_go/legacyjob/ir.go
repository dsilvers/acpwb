package legacyjob

// A deliberately small IR. The core computation (read deck → change →
// percentile ladder → print) and the dead-code bloat are both built as IR,
// randomly transformed, then handed to a language emitter. The emitter also
// reports which symbols each emitted line references, which is what the
// compiler-listing cross-reference and diagnostics are built from.

type Expr struct {
	K    string // var idx int dlit bin call ladder str
	Name string
	Op   string
	Val  int
	S    string
	Args []*Expr
}

func V(name string) *Expr                 { return &Expr{K: "var", Name: name} }
func Ix(arr string, idx ...*Expr) *Expr   { return &Expr{K: "idx", Name: arr, Args: idx} }
func N(v int) *Expr                       { return &Expr{K: "int", Val: v} }
func DL(lit string) *Expr                 { return &Expr{K: "dlit", S: lit} }
func Bin(op string, l, r *Expr) *Expr     { return &Expr{K: "bin", Op: op, Args: []*Expr{l, r}} }
func Fn(name string, args ...*Expr) *Expr { return &Expr{K: "call", Name: name, Args: args} }
func Str(s string) *Expr                  { return &Expr{K: "str", S: s} }

// Ladder is INT(DBLE(med) * mult) — truncating multiply of an integer by a
// double ladder rung.
func Ladder(med, mult *Expr) *Expr { return &Expr{K: "ladder", Args: []*Expr{med, mult}} }

type Cond struct {
	Op   string // lt le eq ne gt ge and or
	L, R *Expr
	A, B *Cond // for and/or
}

func C(op string, l, r *Expr) *Cond { return &Cond{Op: op, L: l, R: r} }
func Or(a, b *Cond) *Cond           { return &Cond{Op: "or", A: a, B: b} }
func And(a, b *Cond) *Cond          { return &Cond{Op: "and", A: a, B: b} }

type Stmt struct {
	K string // assign do if call comment io return stop blank

	L, R *Expr // assign

	Var      string // do
	From, To *Expr
	Step     int
	Body     []*Stmt

	Cond *Cond // if
	Then []*Stmt
	Else []*Stmt

	Name string // call / io kind
	Args []*Expr

	Text   string // comment (unwrapped paragraph)
	Banner bool   // comment drawn as a box
	Title  string // banner title
}

func Assign(l, r *Expr) *Stmt { return &Stmt{K: "assign", L: l, R: r} }
func Do(v string, from, to *Expr, step int, body ...*Stmt) *Stmt {
	return &Stmt{K: "do", Var: v, From: from, To: to, Step: step, Body: body}
}
func If(c *Cond, then []*Stmt, els []*Stmt) *Stmt {
	return &Stmt{K: "if", Cond: c, Then: then, Else: els}
}
func Call(name string, args ...*Expr) *Stmt { return &Stmt{K: "call", Name: name, Args: args} }
func IO(kind string, args ...*Expr) *Stmt   { return &Stmt{K: "io", Name: kind, Args: args} }
func Comment(text string) *Stmt             { return &Stmt{K: "comment", Text: text} }
func Blank() *Stmt                          { return &Stmt{K: "blank"} }

type Sym struct {
	Name   string
	Type   string // int dbl str
	Dims   []int
	Len    int      // str length
	Init   []string // DATA initial values (literals)
	Common string   // FORTRAN COMMON block, "" = local
	Role   string
}

type Proc struct {
	Name    string
	Params  []string
	Locals  []Sym
	Body    []*Stmt
	Purpose string
	Author  *Coder
	Dead    bool
	Core    bool
}

type Program struct {
	Name    string
	Globals []Sym // core data (COMMON / globals / WORKING-STORAGE)
	Procs   []*Proc
	Commons []string
}

// walkExpr visits every expression in a statement list (used for symbol
// reference collection).
func walkExpr(e *Expr, f func(*Expr)) {
	if e == nil {
		return
	}
	f(e)
	for _, a := range e.Args {
		walkExpr(a, f)
	}
}

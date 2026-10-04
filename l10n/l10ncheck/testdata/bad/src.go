package bad

func use(p interface {
	T(string, string, ...any) string
	N(string, int, string, string, ...any) string
}, english string) {
	p.T("b.missing", "Not translated.")
	p.T("b.twice", "One text.")
	p.T("b.twice", "Another text.")
	p.T("b.placeholder", "Hello, {name}.", "name", "x")
	p.T("b.shape", "Plain.")
	p.N("b.count", 2, "{n} key.", "{n} keys.")
	p.T("b.variable", english)
}

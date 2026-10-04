package good

func use(p interface {
	T(string, string, ...any) string
	N(string, int, string, string, ...any) string
}) {
	p.T("a.hello", "Hello, "+
		"{name}.", "name", "x")
	p.T("a.hello", "Hello, {name}.", "name", "y")
	p.N("a.keys", 2, "{n} key in {dir}.", "{n} keys in {dir}.", "dir", "/k")
}

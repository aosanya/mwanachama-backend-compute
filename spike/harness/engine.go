package harness

type Engine interface {
	Name() string
	Compile(c Case, library Library) (Program, error)
}

type Program interface {
	Run(inputs map[string]any) (any, error)
}

type Library map[string]Program

type TreeEngine interface {
	Tree(c Case) (any, error)
	Print(c Case) (string, error)
}

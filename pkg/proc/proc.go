package proc

type ProcFunc func(chan<- error)

type Proc struct {
	funcs []ProcFunc
	errC  chan error
}

func NewProc() *Proc {
	proc := &Proc{
		funcs: make([]ProcFunc, 0),
		errC:  make(chan error, 10),
	}
	return proc
}

func (p *Proc) Add(_func ProcFunc) { p.funcs = append(p.funcs, _func) }

func (p *Proc) Start() chan error {
	for _, _func := range p.funcs {
		go _func(p.errC)
	}
	return p.errC
}

func (p *Proc) Error() chan<- error { return p.errC }

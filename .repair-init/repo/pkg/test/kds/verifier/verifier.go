package verifier

type Executable func(tc TestContext) error

type Verifier interface {
	Exec(Executable) Verifier
	Verify(TestContext) error
}

func New() Verifier {
	return &verifier{}
}

type verifier struct {
	fns []Executable
}

func (v *verifier) Exec(f Executable) Verifier {
	v.fns = append(v.fns, f)
	return v
}

func (v *verifier) Verify(tc TestContext) error {
	for _, f := range v.fns {
		if err := f(tc); err != nil {
			return err
		}
	}
	return nil
}

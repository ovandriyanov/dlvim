package upstream

import (
	"golang.org/x/xerrors"
)

func init() {
	AddStartOptionParser("exec", func(argv []string) (StartOption, error) {
		if len(argv) != 1 {
			return nil, xerrors.Errorf("exactly one argument expected")
		}
		return &Exec{ExecutablePath: argv[0]}, nil
	})
}

type Exec struct {
	ExecutablePath string
}

func (d *Exec) isStartOption() {}

func (d *Exec) Argv() []string {
	return []string{
		"exec",
		d.ExecutablePath,
	}
}

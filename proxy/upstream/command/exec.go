package command

import (
	"github.com/ovandriyanov/dlvim/proxy/upstream"
	"golang.org/x/xerrors"
)

func init() {
	upstream.AddCommandParser("exec", func(argv []string) (upstream.Command, error) {
		if len(argv) != 1 {
			return nil, xerrors.Errorf("exactly one argument expected")
		}
		return &Exec{ExecutablePath: argv[0]}, nil
	})
}

type Exec struct {
	ExecutablePath string
}

func (d *Exec) Argv() []string {
	return []string{
		"exec",
		d.ExecutablePath,
	}
}

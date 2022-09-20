package upstream

import (
	"golang.org/x/xerrors"
)

func init() {
	AddStartOptionParser("exec", func(argv []string) (StartOption, error) {
		if len(argv) < 1 {
			return nil, xerrors.Errorf("at least one argument expected")
		}
		return &Exec{ExecutablePath: argv[0], CmdLineArgs: argv[1:]}, nil
	})
}

type Exec struct {
	ExecutablePath string
	CmdLineArgs    []string
}

func (d *Exec) isStartOption() {}

func (d *Exec) Argv() []string {
	return append([]string{"exec", d.ExecutablePath}, d.CmdLineArgs...)
}

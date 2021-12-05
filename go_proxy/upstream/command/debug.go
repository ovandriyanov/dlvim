package command

import (
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	"golang.org/x/xerrors"
)

func init() {
	upstream.AddCommandParser("debug", func(argv []string) (upstream.Command, error) {
		if len(argv) != 1 {
			return nil, xerrors.Errorf("exactly one argument expected")
		}
		return &Debug{PackagePath: argv[0]}, nil
	})
}

type Debug struct {
	PackagePath string
}

func (d *Debug) Argv() []string {
	return []string{
		"debug",
		d.PackagePath,
	}
}

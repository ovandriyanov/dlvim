package upstream

import (
	"golang.org/x/xerrors"
)

func init() {
	AddStartOptionParser("debug", func(argv []string) (StartOption, error) {
		if len(argv) != 1 {
			return nil, xerrors.Errorf("exactly one argument expected")
		}
		return &Debug{PackagePath: argv[0]}, nil
	})
}

type Debug struct {
	PackagePath string
}

func (d *Debug) isStartOption() {}

func (d *Debug) Argv(listenAddress string) []string {
	return []string{
		"debug",
		d.PackagePath,
		"--listen",
		listenAddress,
		"--headless",
		"--accept-multiclient",
	}
}

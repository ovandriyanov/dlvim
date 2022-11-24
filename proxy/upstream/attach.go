package upstream

import (
	"golang.org/x/xerrors"

	"strconv"
)

func init() {
	AddStartOptionParser("attach", func(argv []string) (StartOption, error) {
		if len(argv) != 1 {
			return nil, xerrors.Errorf("exactly one argument expected")
		}
		pid, err := strconv.Atoi(argv[0])
		if err != nil {
			return nil, xerrors.Errorf("invalid pid: %s", argv[0])
		}
		return &Attach{Pid: pid}, nil
	})
}

type Attach struct {
	Pid int
}

func (a *Attach) isStartOption() {}

func (d *Attach) Argv(listenAddress string) []string {
	return []string{
		"attach",
		strconv.Itoa(d.Pid),
		"--listen",
		listenAddress,
		"--headless",
		"--accept-multiclient",
	}
}

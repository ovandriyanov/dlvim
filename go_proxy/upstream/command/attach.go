package command

import (
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	"golang.org/x/xerrors"

	"strconv"
)

func init() {
	upstream.AddCommandParser("attach", func(argv []string) (upstream.Command, error) {
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

func (d *Attach) Argv() []string {
	return []string{"attach", strconv.Itoa(d.Pid)}
}

package upstream

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

func init() {
	AddStartOptionParser("connect", func(argv []string) (StartOption, error) {
		if len(argv) != 1 {
			return nil, xerrors.Errorf("exactly one argument expected")
		}

		components := strings.Split(argv[0], ":")
		if len(components) != 2 {
			return nil, xerrors.New("Expected exactly one colon in the argument")
		}

		port, err := strconv.Atoi(components[1])
		if err != nil {
			return nil, xerrors.Errorf("Invalid port: %s", components[1])
		}

		return &Connect{Host: components[0], Port: port}, nil
	})
}

// Connect to the running dlv instance, do not actually start a new one
type Connect struct {
	Host string
	Port int
}

func (c *Connect) isStartOption() {}

func (c *Connect) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

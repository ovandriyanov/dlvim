package upstream

import (
	"fmt"

	"golang.org/x/xerrors"
)

type (
	StartOption interface {
		isStartOption()
	}

	// Run the dlv as a separate process
	StartDlvProcess interface {
		isStartOption()
		Argv(listenAddress string) []string
	}

	StartOptionParser func(argv []string) (StartOption, error)
	StartOptionName   string
)

var (
	startOptionParsers = map[string]StartOptionParser{}
)

func AddStartOptionParser(name string, parser StartOptionParser) {
	if _, ok := startOptionParsers[name]; ok {
		panic(fmt.Sprintf("Command parser for name '%s' has already been registered", name))
	}
	startOptionParsers[name] = parser
}

func NewStartOption(argv []string) (StartOption, error) {
	if len(argv) < 1 {
		return nil, xerrors.Errorf("at least one argument expected")
	}
	optionName := argv[0]

	parser, ok := startOptionParsers[optionName]
	if !ok {
		return nil, xerrors.Errorf("unknown command '%s'", optionName)
	}

	startOption, err := parser(argv[1:])
	if err != nil {
		return nil, xerrors.Errorf("%s: %w", optionName, err)
	}

	return startOption, nil
}

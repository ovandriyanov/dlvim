package upstream

import (
	"fmt"

	"golang.org/x/xerrors"
)

type (
	Command       interface{ Argv() []string }
	CommandParser func(argv []string) (Command, error)
	CommandName   string
)

var (
	commandParsers = map[string]CommandParser{}
)

func AddCommandParser(name string, parser CommandParser) {
	if _, ok := commandParsers[name]; ok {
		panic(fmt.Sprintf("Command parser for name '%s' has already been registered", name))
	}
	commandParsers[name] = parser
}

func NewCommand(argv []string) (Command, error) {
	if len(argv) < 1 {
		return nil, xerrors.Errorf("at least one argument expected")
	}
	commandName := argv[0]

	parser, ok := commandParsers[commandName]
	if !ok {
		return nil, xerrors.Errorf("unknown command '%s'", commandName)
	}

	command, err := parser(argv[1:])
	if err != nil {
		return nil, xerrors.Errorf("%s: %w", commandName, err)
	}

	return command, nil
}

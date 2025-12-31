package management

import (
	"context"
	"errors"
	"fmt"
	"os"
)

var (
	ErrCommandNotFound = errors.New("command not found")
	ErrNoArgs          = errors.New("no command provided")
)

// Command defines a management command
type Command interface {
	Name() string
	Help() string
	Run(ctx context.Context, args []string) error
}

// CommandWithFlags adds flag parsing
type CommandWithFlags interface {
	Command
	Flags() *FlagSet
}

// Registry manages available commands
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Global registry
var Commands = NewRegistry()

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, error) {
	cmd, ok := r.commands[name]
	if !ok {
		return nil, ErrCommandNotFound
	}
	return cmd, nil
}

// Execute parses args and runs the appropriate command
func (r *Registry) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrNoArgs
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	cmd, err := r.Get(cmdName)
	if err != nil {
		return err
	}

	// Handle flags if supported
	if fcmd, ok := cmd.(CommandWithFlags); ok {
		fs := fcmd.Flags()
		if err := fs.Parse(cmdArgs); err != nil {
			return err
		}
		// Remaining args after flag parsing
		cmdArgs = fs.Args()
	}

	return cmd.Run(ctx, cmdArgs)
}

// Execute global registry
func Execute(args []string) error {
	return Commands.Execute(context.Background(), args)
}

// Stdout represents standard output for commands
var Stdout = &Output{out: os.Stdout}

// Output helper for styled CLI messages
type Output struct {
	out *os.File
}

func (o *Output) Println(a ...any)               { fmt.Fprintln(o.out, a...) }
func (o *Output) Printf(format string, a ...any) { fmt.Fprintf(o.out, format, a...) }

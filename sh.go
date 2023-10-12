package sh

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Shell is an interface that describes a Shell
type Shell interface {
	// Name of the command. Like "bash" for example
	Name() string

	// Prefix returns commands that should be used before the actual command
	//
	// For example -c for bash
	Prefix() []string

	// Suffix returns commands that should be used after the actual command
	//
	// For example ; for bash if needed
	Suffix() []string
}

type Option func(*Environment)

func WithStdout(w io.Writer) Option {
	return func(e *Environment) {
		e.stdout = w
	}
}

func WithStderr(w io.Writer) Option {
	return func(e *Environment) {
		e.stderr = w
	}
}

func WithEnv(env map[string]string) Option {
	return func(e *Environment) {
		e.env = env
	}
}

func WithWorkingDir(dir string) Option {
	return func(e *Environment) {
		e.workingDir = dir
	}
}

// Environment is a struct that describes the Environment
// in which the shell is executed.
type Environment struct {
	// shell is the shell to use.
	shell Shell

	stdout     io.Writer
	stderr     io.Writer
	env        map[string]string
	workingDir string

	argBuffer []string
}

func NewEnvironment(shell Shell, opts ...Option) *Environment {
	env := &Environment{
		shell: shell,
	}

	for _, opt := range opts {
		opt(env)
	}

	return env
}

// Run runs the script in the environment
//
// Run uses shell as a command
// Arguments passed to the exec.Cmd are:
// 1. Shell.Prefix()...
// 2. script
// 3. Shell.Suffix()...
//
// Extra args are passed as environment variables
func (e *Environment) Run(ctx context.Context, script string, args ...any) error {
	defer e.cleanup()

	cmd, err := e.command(ctx, script, args...)
	if err != nil {
		return err
	}

	return cmd.Run()
}

func (e *Environment) Output(ctx context.Context, script string, args ...any) ([]byte, error) {
	defer e.cleanup()

	cmd, err := e.command(ctx, script, args...)
	if err != nil {
		return nil, err
	}

	return cmd.Output()
}

func (e *Environment) cleanup() {
	e.argBuffer = e.argBuffer[:0]
}

func (e *Environment) command(ctx context.Context, script string, args ...any) (*exec.Cmd, error) {
	e.argBuffer = append(e.argBuffer, e.shell.Prefix()...)
	e.argBuffer = append(e.argBuffer, script)
	if suf := e.shell.Suffix(); len(suf) > 0 {
		e.argBuffer = append(e.argBuffer, suf...)
	}

	cmd := exec.CommandContext(ctx, e.shell.Name(), e.argBuffer...)
	cmd.Stdout = e.stdout
	cmd.Stderr = e.stderr

	if e.workingDir != "" {
		cmd.Dir = e.workingDir
	}

	envs := os.Environ()
	if len(e.env) > 0 {
		for k, v := range e.env {
			envs = append(envs, k+"="+v)
		}
	}

	for i := 0; i < len(args); i++ {
		switch v := args[i].(type) {
		case Arg:
			envs = append(envs, v.String())
		default:
			if i == len(args)-1 {
				return nil, fmt.Errorf("invalid number of arguments")
			}
			key := fmt.Sprintf("%v", args[i])
			val := fmt.Sprintf("%v", args[i+1])
			envs = append(envs, key+"="+val)
			i++
		}
	}
	cmd.Env = envs

	return cmd, nil
}

type Arg struct {
	Key   string
	Value string
}

func (kv Arg) String() string {
	return kv.Key + "=" + kv.Value
}

func Bash() Shell {
	return &bash{}
}

func Sh() Shell {
	return &sh{}
}

type bash struct{}

func (b *bash) Name() string {
	return "bash"
}

func (b *bash) Prefix() []string {
	return []string{"-c"}
}

func (b *bash) Suffix() []string {
	return nil
}

var defaultEnvironment = NewEnvironment(&bash{})

func SetDefaultEnvironment(env *Environment) {
	defaultEnvironment = env
}

func Run(ctx context.Context, script string, args ...any) error {
	return defaultEnvironment.Run(ctx, script, args...)
}

func Output(ctx context.Context, script string, args ...any) ([]byte, error) {
	return defaultEnvironment.Output(ctx, script, args...)
}

type sh struct{}

func (s *sh) Name() string {
	return "sh"
}

func (s *sh) Prefix() []string {
	return []string{"-c"}
}

func (s *sh) Suffix() []string {
	return nil
}

package sh_test

import (
	"bytes"
	"context"
	"fmt"

	"github.com/nikola-jokic/sh"
)

func ExampleEnvironment_Run() {
	var stdout bytes.Buffer
	env := sh.NewEnvironment(sh.Bash(), sh.WithStdout(&stdout))
	env.Run(context.Background(), "echo hello $WHO", "WHO", "world")
	fmt.Print(stdout.String())
	// Output:
	// hello world
}

func ExampleEnvironment_Output() {
	env := sh.NewEnvironment(sh.Bash())
	out, _ := env.Output(context.Background(), "echo hello $WHO", "WHO", "world")
	fmt.Print(string(out))
	// Output:
	// hello world
}

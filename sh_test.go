package sh

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

func TestCommonShellsRun(t *testing.T) {
	shells := []Shell{
		Bash(),
		Sh(),
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	env := Environment{
		stdout: &stdout,
		stderr: &stderr,
		env: map[string]string{
			"TEST_ENV": "test_env_val",
		},
	}

	tt := map[string]struct {
		script         string
		args           []any
		expectedStdout string
		expectedStderr string
	}{
		"SimpleScript": {
			script:         "echo $TEST_ENV",
			expectedStdout: "test_env_val\n",
			expectedStderr: "",
		},
		"ScriptWithArgKV": {
			script:         "echo $TEST_ENV $TEST_ARG && echo $TEST_ARG >&2",
			expectedStdout: "test_env_val test_arg_value\n",
			expectedStderr: "test_arg_value\n",
			args:           []any{Arg{"TEST_ARG", "test_arg_value"}},
		},
		"ScriptWithKeyValueArguments": {
			script:         "echo $TEST_ENV $TEST_ARG && echo $TEST_ARG >&2",
			expectedStdout: "test_env_val test_arg_value\n",
			expectedStderr: "test_arg_value\n",
			args:           []any{"TEST_ARG", "test_arg_value"},
		},
	}

	ctx := context.Background()

	for name, tc := range tt {
		for _, shell := range shells {
			testName := fmt.Sprintf("%s_%s", shell.Name(), name)
			t.Run(testName, func(t *testing.T) {
				env := env
				env.shell = shell

				defer stdout.Reset()
				defer stderr.Reset()

				if err := env.Run(ctx, tc.script, tc.args...); err != nil {
					t.Errorf("Run() error = %v", err)
				}

				if stdout.String() != tc.expectedStdout {
					t.Errorf("Run() stdout = %v, want %v", stdout.String(), tc.expectedStdout)
				}

				if stderr.String() != tc.expectedStderr {
					t.Errorf("Run() stderr = %v, want %v", stderr.String(), tc.expectedStderr)
				}
			})
		}
	}
}

func TestCommonShellsOutputs(t *testing.T) {
	shells := []Shell{
		Bash(),
		Sh(),
	}

	var stderr bytes.Buffer
	env := Environment{
		stderr: &stderr,
		env: map[string]string{
			"TEST_ENV": "test_env_val",
		},
	}

	tt := map[string]struct {
		script         string
		args           []any
		expectedStdout string
		expectedStderr string
	}{
		"SimpleScript": {
			script:         "echo $TEST_ENV",
			expectedStdout: "test_env_val\n",
			expectedStderr: "",
		},
		"ScriptWithArgKV": {
			script:         "echo $TEST_ENV $TEST_ARG && echo $TEST_ARG >&2",
			expectedStdout: "test_env_val test_arg_value\n",
			expectedStderr: "test_arg_value\n",
			args:           []any{Arg{"TEST_ARG", "test_arg_value"}},
		},
		"ScriptWithKeyValueArguments": {
			script:         "echo $TEST_ENV $TEST_ARG && echo $TEST_ARG >&2",
			expectedStdout: "test_env_val test_arg_value\n",
			expectedStderr: "test_arg_value\n",
			args:           []any{"TEST_ARG", "test_arg_value"},
		},
	}

	ctx := context.Background()

	for name, tc := range tt {
		for _, shell := range shells {
			testName := fmt.Sprintf("%s_%s", shell.Name(), name)
			t.Run(testName, func(t *testing.T) {
				env := env
				env.shell = shell

				defer stderr.Reset()

				result, err := env.Output(ctx, tc.script, tc.args...)
				if err != nil {
					t.Errorf("Run() error = %v", err)
				}

				resultStr := string(result)
				if resultStr != tc.expectedStdout {
					t.Errorf("Run() stdout = %v, want %v", resultStr, tc.expectedStdout)
				}

				if stderr.String() != tc.expectedStderr {
					t.Errorf("Run() stderr = %v, want %v", stderr.String(), tc.expectedStderr)
				}
			})
		}
	}
}

func TestArgs(t *testing.T) {
	tt := map[string]struct {
		args      []any
		expectErr bool
	}{
		"EmptyArgs": {
			args:      []any{},
			expectErr: false,
		},
		"ArgKV": {
			args:      []any{Arg{"TEST_ARG_1", "test_arg1_value"}, Arg{"TEST_ARG_2", "test_arg2_value"}},
			expectErr: false,
		},
		"KeyValue": {
			args:      []any{"TEST_ARG_1", "test_arg1_value", "TEST_ARG_2", "test_arg2_value"},
			expectErr: false,
		},
		"ArgKVAndKeyValue": {
			args:      []any{Arg{"TEST_ARG_1", "test_arg1_value"}, "TEST_ARG_2", "test_arg2_value"},
			expectErr: false,
		},
		"InvalidNumberOfArgs": {
			args:      []any{"TEST_ARG_1", "test_arg1_value", "TEST_ARG_2"},
			expectErr: true,
		},
	}

	for name, tc := range tt {
		t.Run(name+"_Run", func(t *testing.T) {
			cmd := "echo $TEST_ARG_1 $TEST_ARG_2"
			err := Run(context.Background(), cmd, tc.args...)
			if tc.expectErr && err == nil {
				t.Fatalf("Run() expected error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("Run() expected no error, got %v", err)
			}
		})
		t.Run(name+"_Output", func(t *testing.T) {
			cmd := "echo $TEST_ARG_1 $TEST_ARG_2"
			_, err := Output(context.Background(), cmd, tc.args...)
			if tc.expectErr && err == nil {
				t.Fatalf("Run() expected error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("Run() expected no error, got %v", err)
			}
		})
	}
}

func TestExitCase(t *testing.T) {
	err := Run(context.Background(), "exit 1")
	if err == nil {
		t.Fatalf("Run() expected error, got nil")
	}

	err = Run(context.Background(), "exit 0")
	if err != nil {
		t.Fatalf("Run() expected no error, got %v", err)
	}
}

func TestWorkigDir(t *testing.T) {
	shells := []Shell{
		Bash(),
		Sh(),
	}

	var stdout bytes.Buffer
	env := Environment{
		workingDir: "/tmp",
		stdout:     &stdout,
	}

	for _, shell := range shells {
		t.Run("WorkingDir_"+shell.Name(), func(t *testing.T) {
			stdout.Reset()
			env.shell = shell
			if err := env.Run(context.Background(), "pwd"); err != nil {
				t.Errorf("Run() error = %v", err)
			}

			if stdout.String() != "/tmp\n" {
				t.Errorf("Run() stdout = %v, want %v", stdout.String(), "/tmp")
			}
		})
	}
}

func TestExportedShells(t *testing.T) {
	for _, shell := range []Shell{Bash(), Sh()} {
		if shell == nil {
			t.Errorf("Shell %q is nil", shell.Name())
		}
	}
}

func TestSetDefaultEnvironment(t *testing.T) {
	if defaultEnvironment.shell.Name() != Bash().Name() {
		t.Errorf("defaultEnvironment.shell = %v, want %v", defaultEnvironment.shell.Name(), Bash().Name())
	}

	SetDefaultEnvironment(NewEnvironment(Sh()))
	if defaultEnvironment.shell.Name() != Sh().Name() {
		t.Errorf("defaultEnvironment.shell = %v, want %v", defaultEnvironment.shell.Name(), Sh().Name())
	}
}

// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/common.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -----------------------------------------------------------------------------
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// -----------------------------------------------------------------------------

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var ErrCmdRequired = errors.New("command is required")

const WaitGroupCount = 2

type E2E struct {
	*Config
	BomctlBinPath string
	CliName       string
}

type Config struct {
	Stdout         io.Writer
	Stderr         io.Writer
	CommandPrinter func(format string, a ...any)
	Dir            string
	Env            []string
	Print          bool
}

func NewE2E(t *testing.T) *E2E {
	t.Helper()

	cliName := GetCLIName()
	binPath := filepath.Join("..", "..", "..", "dist", cliName)

	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("bomctl binary %s not found: %v", binPath, err)
	}

	return &E2E{
		Config:        &Config{Print: true},
		BomctlBinPath: binPath,
		CliName:       cliName,
	}
}

// GetCLIName looks at the OS and CPU architecture to determine which bomctl binary needs to be run.
func GetCLIName() string {
	binaryName := "bomctl"
	binaryExt := ""

	switch runtime.GOOS {
	case "linux":
		binaryName += "-linux"
	case "darwin":
		binaryName += "-darwin"
	case "windows":
		binaryName += "windows"
		binaryExt = ".exe"
	}

	switch runtime.GOARCH {
	case "arm64":
		binaryName += "-arm64"
	case "amd64":
		binaryName += "-amd64"
	}

	return binaryName + binaryExt
}

// Bomctl executes a Bomctl command.
func (e2e *E2E) Bomctl(t *testing.T, args ...string) (stdOut, stdErr string, cmderr error) {
	t.Helper()

	if e2e.BomctlBinPath == "" {
		return "", "", ErrCmdRequired
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Set up the command.
	command := e2e.BomctlBinPath
	cmd, cmdStdout, cmdStderr := e2e.setUpCmd(ctx, t, command, args...)

	var (
		stdoutBuf, stderrBuf bytes.Buffer
		errStdout, errStderr error
		waitGroup            sync.WaitGroup
	)

	stdout, stderr := e2e.addConfiguredWriters(&stdoutBuf, &stderrBuf)

	// If we're printing, print the command.
	if e2e.Config.Print && e2e.Config.CommandPrinter != nil {
		e2e.Config.CommandPrinter("%s %s", command, strings.Join(args, " "))
	}

	// Start the command.
	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start command: %w", err)
	}

	// Add to waitgroup for each goroutine.
	waitGroup.Add(WaitGroupCount)

	// Run a goroutine to capture the command's stdout live.
	go func() {
		_, errStdout = io.Copy(stdout, cmdStdout)

		waitGroup.Done()
	}()

	// Run a goroutine to capture the command's stderr live.
	go func() {
		_, errStderr = io.Copy(stderr, cmdStderr)

		waitGroup.Done()
	}()

	// Wait for the goroutines to finish (if any).
	waitGroup.Wait()

	// Abort if there was an error capturing the command's outputs.
	if errStdout != nil {
		return "", "", fmt.Errorf("failed to capture the stdout command output: %w", errStdout)
	}

	if errStderr != nil {
		return "", "", fmt.Errorf("failed to capture the stderr command output: %w", errStderr)
	}

	// Wait for the command to finish and return the buffered outputs, regardless of whether we printed them.
	return stdoutBuf.String(), stderrBuf.String(), cmd.Wait()
}

func (e2e *E2E) addConfiguredWriters(stdoutBuf, stderrBuf *bytes.Buffer) (stdoutWriter, stdErrWriter io.Writer) {
	stdoutWriters := []io.Writer{
		stdoutBuf,
	}

	stdErrWriters := []io.Writer{
		stderrBuf,
	}

	// Add the writers if requested.
	if e2e.Config.Stdout != nil {
		stdoutWriters = append(stdoutWriters, e2e.Config.Stdout)
	}

	if e2e.Config.Stderr != nil {
		stdErrWriters = append(stdErrWriters, e2e.Config.Stderr)
	}

	// Print to stdout if requested.
	if e2e.Config.Print {
		stdoutWriters = append(stdoutWriters, os.Stdout)
		stdErrWriters = append(stdErrWriters, os.Stderr)
	}

	// Bind all the writers.
	stdout := io.MultiWriter(stdoutWriters...)
	stderr := io.MultiWriter(stdErrWriters...)

	return stdout, stderr
}

func (e2e *E2E) setUpCmd(ctx context.Context, t *testing.T, cmdName string,
	args ...string,
) (c *exec.Cmd, cso, cse io.ReadCloser) {
	t.Helper()

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = e2e.Config.Dir
	cmd.Env = append(os.Environ(), e2e.Config.Env...)

	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to retrieve stdout: %v", err)
	}

	cmdStderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to retrieve stderr: %v", err)
	}

	return cmd, cmdStdout, cmdStderr
}

package invocation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

const (
	// One budget for process discovery AND all identity checks, per HTTP attempt.
	DefaultInspectionTimeout = 5 * time.Second
	inspectionHelperFlag     = "--internal-invocation-inspect"
	maxInspectionOutput      = 1 << 20
	inspectionWaitDelay      = 100 * time.Millisecond
)

// Inspect runs the existing detector in a disposable copy of this executable.
// Native identity checks (notably WinVerifyTrust) cannot be cancelled by a Go
// context. Process isolation bounds them without leaking blocked goroutines.
// Any inspection failure is optional metadata loss, never a business error.
func Inspect(parent context.Context, maxDepth int) Result {
	ctx, cancel := context.WithTimeout(parent, DefaultInspectionTimeout)
	defer cancel()
	if ctx.Err() != nil {
		return unavailableInspection("environment inspection cancelled")
	}
	executable, err := os.Executable()
	if err != nil {
		return unavailableInspection("environment inspection helper unavailable")
	}
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}
	if maxDepth > 128 {
		maxDepth = 128
	}
	command := exec.CommandContext(ctx, executable, inspectionHelperFlag, strconv.Itoa(maxDepth))
	return inspectCommand(ctx, command)
}

func inspectCommand(ctx context.Context, command *exec.Cmd) Result {
	var output inspectionOutput
	command.Stdout = &output
	command.Stderr = io.Discard
	command.WaitDelay = inspectionWaitDelay
	configureInspectionCommand(command)
	err := command.Run() // Wait reaps the helper before any business HTTP call continues.
	if ctx.Err() != nil {
		return unavailableInspection("environment inspection timed out or cancelled")
	}
	if err != nil || output.overflow {
		return unavailableInspection("environment inspection helper failed")
	}
	var result Result
	if err := json.Unmarshal(output.buffer.Bytes(), &result); err != nil ||
		result.ChannelType != "cli" || result.ProductCode != ProductCodeEveryline ||
		result.DetectorVersion != DetectorVersion || result.AgentSourceType == "" {
		return unavailableInspection("environment inspection helper returned invalid output")
	}
	return result
}

func unavailableInspection(reason string) Result {
	result := Analyze(nil, nil)
	result.Platform = runtime.GOOS
	result.Reason = reason
	return result
}

// RunInspectionHelper must be dispatched by main BEFORE CLI/auth/update setup.
// It only reads local process identity and writes one JSON report; it never
// opens a profile, reads tokens, starts another helper or sends a business call.
func RunInspectionHelper(args []string, output io.Writer) (bool, error) {
	if len(args) == 0 || args[0] != inspectionHelperFlag {
		return false, nil
	}
	if len(args) != 2 {
		return true, fmt.Errorf("invalid inspection helper arguments")
	}
	maxDepth, err := strconv.Atoi(args[1])
	if err != nil || maxDepth <= 0 || maxDepth > 128 {
		return true, fmt.Errorf("invalid inspection helper depth")
	}
	return true, json.NewEncoder(output).Encode(inspectProcess(int32(os.Getppid()), maxDepth))
}

// Bound the helper's output without blocking its pipes if output is oversized.
type inspectionOutput struct {
	buffer   bytes.Buffer
	overflow bool
}

func (output *inspectionOutput) Write(value []byte) (int, error) {
	size := len(value)
	remaining := maxInspectionOutput - output.buffer.Len()
	if size > remaining {
		output.overflow = true
		value = value[:remaining]
	}
	_, _ = output.buffer.Write(value)
	return size, nil
}

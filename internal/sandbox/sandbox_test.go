package sandbox

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

// TestClassifyRunResult_Timeout reproduces ossf/package-analysis#142: a
// command killed because its context deadline expired must be classified as
// RunStatusTimeout, not RunStatusFailure. Before the fix, cmd.Wait() returns
// a generic *exec.ExitError ("signal: killed") that is indistinguishable
// from an ordinary non-zero exit, so the timeout was silently reported as a
// regular failure.
func TestClassifyRunResult_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sleep", "5")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected the sleep command to be killed by the context deadline")
	}

	status, gotErr := classifyRunResult(ctx, err)
	if status != RunStatusTimeout {
		t.Errorf("classifyRunResult() status = %v, want RunStatusTimeout", status)
	}
	if gotErr != nil {
		t.Errorf("classifyRunResult() err = %v, want nil", gotErr)
	}
}

func TestClassifyRunResult_Success(t *testing.T) {
	status, err := classifyRunResult(context.Background(), nil)
	if status != RunStatusSuccess {
		t.Errorf("classifyRunResult() status = %v, want RunStatusSuccess", status)
	}
	if err != nil {
		t.Errorf("classifyRunResult() err = %v, want nil", err)
	}
}

func TestClassifyRunResult_Failure(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "sh", "-c", "exit 3")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected the command to exit with a non-zero status")
	}

	status, gotErr := classifyRunResult(ctx, err)
	if status != RunStatusFailure {
		t.Errorf("classifyRunResult() status = %v, want RunStatusFailure", status)
	}
	if gotErr != nil {
		t.Errorf("classifyRunResult() err = %v, want nil", gotErr)
	}
}

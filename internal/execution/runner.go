package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunScript executes a .py or .sh script, passing inputs and config as JSON
// positional arguments. It returns (errMsg, nil) where errMsg is empty on
// success or contains the script's stderr on failure. A non-nil error means
// the process could not be started or was killed (e.g. context cancelled).
func RunScript(ctx context.Context, scriptPath string, inputs map[string]any, config map[string]string) (string, error) {
	inputsJSON, err := json.Marshal(orEmpty(inputs))
	if err != nil {
		return "", fmt.Errorf("marshal inputs: %w", err)
	}
	configJSON, err := json.Marshal(orEmptyStr(config))
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}

	if _, err := os.Stat(scriptPath); err != nil {
		return "", fmt.Errorf("script not found: %w", err)
	}

	var cmd *exec.Cmd
	switch strings.ToLower(filepath.Ext(scriptPath)) {
	case ".py":
		cmd = exec.CommandContext(ctx, "python3", scriptPath, string(inputsJSON), string(configJSON))
	case ".sh":
		cmd = exec.CommandContext(ctx, "bash", scriptPath, string(inputsJSON), string(configJSON))
	default:
		return "", fmt.Errorf("unsupported script type: %s", filepath.Ext(scriptPath))
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Context cancellation or kill — propagate as system error.
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		// Non-zero exit — script reported an error via stderr.
		return strings.TrimSpace(stderr.String()), nil
	}

	return "", nil
}

func orEmpty(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

func orEmptyStr(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

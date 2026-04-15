package execution_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chiragguptadtu/trigger/internal/execution"
)

func writeScript(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0755))
	return path
}

func TestRunScript_PythonSuccess(t *testing.T) {
	path := writeScript(t, "ok.py", `#!/usr/bin/env python3
import sys, json
inputs = json.loads(sys.argv[1])
config = json.loads(sys.argv[2])
assert inputs["env"] == "staging"
assert config["SECRET"] == "s3cr3t"
`)
	errMsg, err := execution.RunScript(t.Context(), path, map[string]any{"env": "staging"}, map[string]string{"SECRET": "s3cr3t"})
	require.NoError(t, err)
	assert.Empty(t, errMsg)
}

func TestRunScript_PythonFailure(t *testing.T) {
	path := writeScript(t, "fail.py", `#!/usr/bin/env python3
import sys
print("target not found", file=sys.stderr)
sys.exit(1)
`)
	errMsg, err := execution.RunScript(t.Context(), path, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "target not found", errMsg)
}

func TestRunScript_BashSuccess(t *testing.T) {
	path := writeScript(t, "ok.sh", `#!/bin/bash
INPUTS=$1
ENV=$(echo $INPUTS | python3 -c "import sys,json; print(json.load(sys.stdin)['env'])")
if [ "$ENV" != "prod" ]; then exit 1; fi
`)
	errMsg, err := execution.RunScript(t.Context(), path, map[string]any{"env": "prod"}, nil)
	require.NoError(t, err)
	assert.Empty(t, errMsg)
}

func TestRunScript_BashFailure(t *testing.T) {
	path := writeScript(t, "fail.sh", `#!/bin/bash
echo "disk full" >&2
exit 1
`)
	errMsg, err := execution.RunScript(t.Context(), path, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "disk full", errMsg)
}

func TestRunScript_InvalidPath(t *testing.T) {
	_, err := execution.RunScript(t.Context(), "/nonexistent/script.py", nil, nil)
	require.Error(t, err)
}

func TestRunScript_ContextCancelled(t *testing.T) {
	path := writeScript(t, "slow.py", `#!/usr/bin/env python3
import time
time.sleep(30)
`)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	_, err := execution.RunScript(ctx, path, nil, nil)
	require.Error(t, err)
}

func TestGetOptions_PythonTwoInputs(t *testing.T) {
	path := writeScript(t, "cmd.py", `#!/usr/bin/env python3
import sys, json

def get_options(input_name, config):
    if input_name == "environment":
        envs = config.get("AVAILABLE_ENVIRONMENTS", "staging,production")
        return [e.strip() for e in envs.split(",") if e.strip()]
    if input_name == "region":
        return ["us-east-1", "eu-west-1"]
    return []

if __name__ == "__main__":
    if sys.argv[1] == "--trigger-get-options":
        input_name = sys.argv[2] if len(sys.argv) > 2 else ""
        config = json.loads(sys.argv[3]) if len(sys.argv) > 3 else {}
        print("\n".join(get_options(input_name, config)))
        sys.exit(0)
`)
	cfg := map[string]string{"AVAILABLE_ENVIRONMENTS": "dev,staging,production"}

	envOpts, err := execution.GetOptions(t.Context(), path, "environment", cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"dev", "staging", "production"}, envOpts)

	regionOpts, err := execution.GetOptions(t.Context(), path, "region", cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"us-east-1", "eu-west-1"}, regionOpts)
}

func TestGetOptions_BashSuccess(t *testing.T) {
	path := writeScript(t, "cmd.sh", `#!/usr/bin/env bash
get_options() {
    local input_name="$1"
    if [[ "$input_name" == "environment" ]]; then
        echo "staging"
        echo "production"
    elif [[ "$input_name" == "region" ]]; then
        echo "us-east-1"
        echo "eu-west-1"
    fi
}
if [[ "${1:-}" == "--trigger-get-options" ]]; then
    get_options "${2:-}" "${3:-}"
    exit 0
fi
`)
	opts, err := execution.GetOptions(t.Context(), path, "environment", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"staging", "production"}, opts)

	opts, err = execution.GetOptions(t.Context(), path, "region", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"us-east-1", "eu-west-1"}, opts)
}

func TestGetOptions_ScriptError(t *testing.T) {
	path := writeScript(t, "err.py", `#!/usr/bin/env python3
import sys
if sys.argv[1] == "--trigger-get-options":
    print("database unavailable", file=sys.stderr)
    sys.exit(1)
`)
	_, err := execution.GetOptions(t.Context(), path, "environment", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database unavailable")
}

func TestGetOptions_EmptyOutput(t *testing.T) {
	path := writeScript(t, "empty.py", `#!/usr/bin/env python3
import sys
if sys.argv[1] == "--trigger-get-options":
    sys.exit(0)
`)
	opts, err := execution.GetOptions(t.Context(), path, "environment", nil)
	require.NoError(t, err)
	assert.Empty(t, opts)
}

func TestGetOptions_InvalidPath(t *testing.T) {
	_, err := execution.GetOptions(t.Context(), "/nonexistent/script.py", "env", nil)
	require.Error(t, err)
}

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

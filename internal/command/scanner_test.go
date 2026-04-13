package command_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/command"
)

func TestScanDir_FindsPythonCommands(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "my_command.py"), []byte(validScript), 0644)
	require.NoError(t, err)

	result, err := command.ScanDir(dir)
	require.NoError(t, err)
	require.Len(t, result.Commands, 1)
	assert.Equal(t, "my-command", result.Commands[0].Slug)
	assert.Equal(t, "Send Weekly Report", result.Commands[0].Command.Name)
	assert.Equal(t, filepath.Join(dir, "my_command.py"), result.Commands[0].ScriptPath)
	assert.Empty(t, result.Errors)
}

func TestScanDir_FindsBashCommands(t *testing.T) {
	dir := t.TempDir()
	bashScript := `#!/bin/bash
# ---trigger---
# name: Deploy Service
# description: Deploys the service
# inputs:
#   - name: env
#     label: Environment
#     type: closed
#     options: [staging, production]
#     required: true
# ---end---
echo "deploying"
`
	err := os.WriteFile(filepath.Join(dir, "deploy_service.sh"), []byte(bashScript), 0644)
	require.NoError(t, err)

	result, err := command.ScanDir(dir)
	require.NoError(t, err)
	require.Len(t, result.Commands, 1)
	assert.Equal(t, "deploy-service", result.Commands[0].Slug)
	assert.Equal(t, "Deploy Service", result.Commands[0].Command.Name)
	assert.Empty(t, result.Errors)
}

func TestScanDir_SkipsFilesWithoutTriggerBlock(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "helper.py"), []byte("# just a helper, no trigger block"), 0644)
	require.NoError(t, err)

	result, err := command.ScanDir(dir)
	require.NoError(t, err)
	assert.Empty(t, result.Commands)
	assert.Empty(t, result.Errors)
}

func TestScanDir_IgnoresNonScriptFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0644))

	result, err := command.ScanDir(dir)
	require.NoError(t, err)
	assert.Empty(t, result.Commands)
	assert.Empty(t, result.Errors)
}

func TestScanDir_ScansSubdirectories(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "reports")
	require.NoError(t, os.MkdirAll(subdir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "weekly_report.py"), []byte(validScript), 0644))

	result, err := command.ScanDir(dir)
	require.NoError(t, err)
	require.Len(t, result.Commands, 1)
	assert.Equal(t, "weekly-report", result.Commands[0].Slug)
	assert.Empty(t, result.Errors)
}

func TestScanDir_DirectoryNotFound(t *testing.T) {
	_, err := command.ScanDir("/nonexistent/path/trigger")
	require.Error(t, err)
}

// TestScanDir_PartialError verifies that a malformed file does not abort the
// scan — valid files are still returned and the error is recorded per-file.
func TestScanDir_PartialError(t *testing.T) {
	dir := t.TempDir()
	// One valid command
	require.NoError(t, os.WriteFile(filepath.Join(dir, "good.py"), []byte(validScript), 0644))
	// One file with a malformed trigger block
	malformed := `# ---trigger---
# name: Bad Command
# inputs:
#   - name: ` + "\t" + `bad_yaml: [unclosed
# ---end---
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bad_command.py"), []byte(malformed), 0644))

	result, err := command.ScanDir(dir)
	require.NoError(t, err) // scan itself must not fail

	assert.Len(t, result.Commands, 1)
	assert.Equal(t, "good", result.Commands[0].Slug)

	require.Len(t, result.Errors, 1)
	assert.Equal(t, "bad_command.py", result.Errors[0].Filename)
	assert.Error(t, result.Errors[0].Err)
}

func TestSlugFromFilename(t *testing.T) {
	cases := []struct {
		filename string
		want     string
	}{
		{"send_weekly_report.py", "send-weekly-report"},
		{"deploy-service.sh", "deploy-service"},
		{"mycommand.py", "mycommand"},
		{"My_Command.py", "my-command"},
	}
	for _, c := range cases {
		t.Run(c.filename, func(t *testing.T) {
			assert.Equal(t, c.want, command.SlugFromFilename(c.filename))
		})
	}
}

package command_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/command"
)

const validScript = `# ---trigger---
# name: Send Weekly Report
# description: Sends the weekly summary to stakeholders
# inputs:
#   - name: recipient_email
#     label: Recipient Email
#     type: open
#     required: true
#   - name: report_type
#     label: Report Type
#     type: closed
#     options: [sales, ops, finance]
#     multi: false
#     required: true
#   - name: regions
#     label: Regions
#     type: closed
#     options: [US, EU, APAC]
#     multi: true
#     required: false
# ---end---

import sys

def run(inputs, config):
    return ""
`

func TestParseContent_Valid(t *testing.T) {
	cmd, err := command.ParseContent(validScript)
	require.NoError(t, err)

	assert.Equal(t, "Send Weekly Report", cmd.Name)
	assert.Equal(t, "Sends the weekly summary to stakeholders", cmd.Description)
	require.Len(t, cmd.Inputs, 3)

	open := cmd.Inputs[0]
	assert.Equal(t, "recipient_email", open.Name)
	assert.Equal(t, "Recipient Email", open.Label)
	assert.Equal(t, command.InputTypeOpen, open.Type)
	assert.True(t, open.Required)
	assert.Nil(t, open.Options)

	closed := cmd.Inputs[1]
	assert.Equal(t, "report_type", closed.Name)
	assert.Equal(t, "Report Type", closed.Label)
	assert.Equal(t, command.InputTypeClosed, closed.Type)
	assert.False(t, closed.Multi)
	assert.True(t, closed.Required)
	assert.Equal(t, []string{"sales", "ops", "finance"}, closed.Options)

	multi := cmd.Inputs[2]
	assert.Equal(t, "regions", multi.Name)
	assert.True(t, multi.Multi)
	assert.False(t, multi.Required)
	assert.Equal(t, []string{"US", "EU", "APAC"}, multi.Options)
}

func TestParseContent_MissingTriggerBlock(t *testing.T) {
	_, err := command.ParseContent("import sys\n\ndef run(inputs, config):\n    return ''\n")
	require.ErrorIs(t, err, command.ErrNoTriggerBlock)
}

func TestParseContent_MissingName(t *testing.T) {
	script := `# ---trigger---
# description: A command with no name
# inputs: []
# ---end---
`
	_, err := command.ParseContent(script)
	require.ErrorIs(t, err, command.ErrMissingName)
}

func TestParseContent_UnknownInputType(t *testing.T) {
	script := `# ---trigger---
# name: Bad Command
# description: Has an invalid input type
# inputs:
#   - name: foo
#     label: Foo
#     type: invalid
#     required: true
# ---end---
`
	_, err := command.ParseContent(script)
	require.ErrorIs(t, err, command.ErrUnknownInputType)
}

func TestParseContent_ClosedInputWithNoOptions(t *testing.T) {
	script := `# ---trigger---
# name: Bad Command
# description: Closed input with no options
# inputs:
#   - name: foo
#     label: Foo
#     type: closed
#     required: true
# ---end---
`
	_, err := command.ParseContent(script)
	require.ErrorIs(t, err, command.ErrClosedInputNoOptions)
}

func TestParseContent_NoInputs(t *testing.T) {
	script := `# ---trigger---
# name: Simple Command
# description: A command with no inputs
# inputs: []
# ---end---
`
	cmd, err := command.ParseContent(script)
	require.NoError(t, err)
	assert.Equal(t, "Simple Command", cmd.Name)
	assert.Empty(t, cmd.Inputs)
}

package command

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type InputType string

const (
	InputTypeOpen   InputType = "open"
	InputTypeClosed InputType = "closed"
)

var (
	ErrNoTriggerBlock       = errors.New("no trigger block found")
	ErrMissingName          = errors.New("command name is required")
	ErrUnknownInputType     = errors.New("unknown input type")
	ErrClosedInputNoOptions = errors.New("closed input must define at least one option")
)

type Input struct {
	Name     string
	Label    string
	Type     InputType
	Options  []string
	Multi    bool
	Required bool
	Dynamic  bool
}

type Command struct {
	Name        string
	Description string
	Inputs      []Input
}

// internal structs for YAML unmarshalling
type commandMeta struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Inputs      []inputMeta `yaml:"inputs"`
}

type inputMeta struct {
	Name     string   `yaml:"name"`
	Label    string   `yaml:"label"`
	Type     string   `yaml:"type"`
	Options  []string `yaml:"options"`
	Multi    bool     `yaml:"multi"`
	Required bool     `yaml:"required"`
	Dynamic  bool     `yaml:"dynamic"`
}

// ParseContent parses the trigger metadata block from a script file's content.
// The block is delimited by # ---trigger--- and # ---end--- comment lines.
func ParseContent(content string) (*Command, error) {
	rawYAML, err := extractTriggerBlock(content)
	if err != nil {
		return nil, err
	}

	var meta commandMeta
	if err := yaml.Unmarshal([]byte(rawYAML), &meta); err != nil {
		return nil, fmt.Errorf("invalid trigger block: %w", err)
	}

	if meta.Name == "" {
		return nil, ErrMissingName
	}

	inputs, err := convertInputs(meta.Inputs)
	if err != nil {
		return nil, err
	}

	return &Command{
		Name:        meta.Name,
		Description: meta.Description,
		Inputs:      inputs,
	}, nil
}

func extractTriggerBlock(content string) (string, error) {
	const startMarker = "# ---trigger---"
	const endMarker = "# ---end---"

	lines := strings.Split(content, "\n")
	inBlock := false
	var blockLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == startMarker {
			inBlock = true
			continue
		}

		if trimmed == endMarker {
			if inBlock {
				return strings.Join(blockLines, "\n"), nil
			}
		}

		if inBlock {
			if strings.HasPrefix(trimmed, "# ") {
				blockLines = append(blockLines, strings.TrimPrefix(trimmed, "# "))
			} else if trimmed == "#" {
				blockLines = append(blockLines, "")
			}
		}
	}

	return "", ErrNoTriggerBlock
}

func convertInputs(metas []inputMeta) ([]Input, error) {
	if len(metas) == 0 {
		return nil, nil
	}

	inputs := make([]Input, 0, len(metas))
	for _, m := range metas {
		t := InputType(m.Type)
		switch t {
		case InputTypeOpen:
			// valid, no extra constraints
		case InputTypeClosed:
			if !m.Dynamic && len(m.Options) == 0 {
				return nil, ErrClosedInputNoOptions
			}
		default:
			return nil, ErrUnknownInputType
		}

		inputs = append(inputs, Input{
			Name:     m.Name,
			Label:    m.Label,
			Type:     t,
			Options:  m.Options,
			Multi:    m.Multi,
			Required: m.Required,
			Dynamic:  m.Dynamic,
		})
	}
	return inputs, nil
}

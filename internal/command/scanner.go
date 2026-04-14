package command

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ScannedCommand is the result of parsing one script file.
type ScannedCommand struct {
	Slug       string
	ScriptPath string
	Command    *Command
}

// ScanError records a per-file parse or read failure.
type ScanError struct {
	Filename string
	Err      error
}

// ScanResult holds the outcome of a directory scan.
// Commands contains successfully parsed commands.
// Errors contains per-file failures that did not abort the scan.
type ScanResult struct {
	Commands []ScannedCommand
	Errors   []ScanError
}

// ScanDir walks dir recursively, parses all .py and .sh files that contain a
// trigger block, and returns a ScanResult. Per-file parse/read errors are
// collected into ScanResult.Errors instead of aborting — other files are still
// processed. Only filesystem-level errors (directory not found, walk failure)
// return a non-nil error.
func ScanDir(dir string) (ScanResult, error) {
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return ScanResult{}, err
	}

	var result ScanResult

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // directory traversal error — abort the whole walk
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".py" && ext != ".sh" {
			return nil
		}

		filename := filepath.Base(path)
		content, err := os.ReadFile(path)
		if err != nil {
			result.Errors = append(result.Errors, ScanError{Filename: filename, Err: err})
			return nil
		}

		cmd, err := ParseContent(string(content))
		if err != nil {
			result.Errors = append(result.Errors, ScanError{Filename: filename, Err: err})
			return nil
		}

		result.Commands = append(result.Commands, ScannedCommand{
			Slug:       SlugFromFilename(filename),
			ScriptPath: path,
			Command:    cmd,
		})
		return nil
	})
	if err != nil {
		return ScanResult{}, err
	}

	return result, nil
}

// SlugFromFilename converts a script filename to a URL-safe slug.
// e.g. "send_weekly_report.py" → "send-weekly-report"
func SlugFromFilename(filename string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	slug := strings.ToLower(base)
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}

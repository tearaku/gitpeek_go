package main

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

var entryTemplate = &promptui.SelectTemplates{
	Label:    "{{ . }}?",
	Active:   "\u279E {{ .Path | bold | underline }} ({{ .GitBranch | green | bold | underline }})",
	Inactive: "  {{ .Path}} ({{ .GitBranch | green }})",
	Selected: "\u279E {{ .Path }} ({{ .GitBranch | green }})",
}

func PromptMenu(items []GitEntry) (*promptui.Select, error) {
	if len(items) <= 0 {
		return nil, fmt.Errorf("no result found")
	}
	return &promptui.Select{
		Label:     "Found directories (press 'Enter' to copy directory path to clipboard)",
		Items:     items,
		Templates: entryTemplate,
		Size:      10,
	}, nil
}

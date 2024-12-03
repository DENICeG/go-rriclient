package cli

import (
	"strings"

	"github.com/DENICeG/go-rriclient/pkg/preset"
	"github.com/sbreitf1/go-console/commandline"
)

type domainOrHandleCompletion struct {
	histDomains *domainHistory
	histHandles *handleHistory
}

// NewCompletion returns a new domainOrHandleCompletion instance.
func NewCompletion() *domainOrHandleCompletion {
	return &domainOrHandleCompletion{
		histDomains: &domainHistory{list: make([]string, 0)},
		histHandles: &handleHistory{list: make([]string, 0)},
	}
}

func (c *domainOrHandleCompletion) PutDomain(domain string) {
	c.histDomains.Put(domain)
}

func (c *domainOrHandleCompletion) PutHandle(handle string) {
	c.histHandles.Put(handle)
}

func (c *domainOrHandleCompletion) GetCompletionOptions(currentCommand []string, entryIndex int) []commandline.CompletionOption {
	if len(currentCommand) < 2 && entryIndex != 2 {
		return nil
	}

	if currentCommand[1] == "domain" || currentCommand[1] == "authinfo1" {
		return c.GetCompletionOptions(currentCommand, entryIndex)
	}

	if currentCommand[1] == "handle" {
		return c.histHandles.GetCompletionOptions(currentCommand, entryIndex)
	}

	return nil
}

// PresetCompletion implements the Completion interface for preset names.
type PresetCompletion struct {
	presets preset.Data
}

// NewPresetCompletion returns a new PresetCompletion instance.
func NewPresetCompletion(presets preset.Data) *PresetCompletion {
	return &PresetCompletion{presets: presets}
}

// GetCompletionOptions returns the completion options for the given command.
func (p *PresetCompletion) GetCompletionOptions(currentCommand []string, entryIndex int) []commandline.CompletionOption {
	currentWord := currentCommand[entryIndex]

	var result []commandline.CompletionOption

	for _, entry := range p.presets.Preset {
		if strings.HasPrefix(strings.ToLower(entry.FileName), strings.ToLower(currentWord)) {
			result = append(result, commandline.NewCompletionOption(entry.FileName, false))
		}
	}

	return result
}

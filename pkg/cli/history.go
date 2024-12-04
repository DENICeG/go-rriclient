package cli

import (
	"slices"

	"github.com/DENICeG/go-console/v2/commandline"
)

type domainHistory struct {
	list []string
}

func (h *domainHistory) Put(domain string) {
	if slices.Contains(h.list, domain) {
		return
	}

	h.list = append([]string{domain}, h.list...)
}

func (h *domainHistory) GetCompletionOptions(currentCommand []string, entryIndex int) []commandline.CompletionOption {
	return commandline.PrepareCompletionOptions(h.list, false)
}

type handleHistory struct {
	list []string
}

func (h *handleHistory) Put(handle string) {
	if slices.Contains(h.list, handle) {
		return
	}

	h.list = append([]string{handle}, h.list...)
}

func (h *handleHistory) GetCompletionOptions(currentCommand []string, entryIndex int) []commandline.CompletionOption {
	return commandline.PrepareCompletionOptions(h.list, false)
}

func (h *handleHistory) GetHistoryEntry(index int) (string, bool) {
	if index >= len(h.list) {
		return "", false
	}

	return h.list[index], true
}

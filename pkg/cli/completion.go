package cli

import "github.com/sbreitf1/go-console/commandline"

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

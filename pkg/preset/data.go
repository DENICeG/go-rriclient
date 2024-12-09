package preset

import "strings"

// Data holds a map of all available presets.
type Data struct {
	Preset        map[int]Entry
	XMLStartIndex int
}

// Get returns the entry for the given name.
// returns nil if the entry does not exist.
func (d *Data) Get(name string) *Entry {
	for _, entry := range d.Preset {
		if strings.EqualFold(entry.FileName, name) {
			return &entry
		}
	}

	return nil
}

// Entry represents a single preset entry.
type Entry struct {
	Type     string
	DirName  string
	FileName string
}

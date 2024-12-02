package preset

// Data holds a map of all available presets.
type Data struct {
	Preset        map[int]Entry
	XMLStartIndex int
}

// Entry represents a single preset entry.
type Entry struct {
	Type     string
	FileName string
}

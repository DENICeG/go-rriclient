package preset

import (
	"embed"
	"io/fs"
)

// Load loads all presets from the embedded binary
func Load(embedFS embed.FS, dirInfo []fs.DirEntry) (*Data, error) {
	result := &Data{
		Preset: make(map[int]Entry),
	}

	counter := 0

	for i, info := range dirInfo {
		directories, err := embedFS.ReadDir("examples/" + info.Name())
		if err != nil {
			return nil, err
		}
		for _, dir := range directories {
			files, err := embedFS.ReadDir("examples/" + info.Name() + "/" + dir.Name())
			if err != nil {
				return nil, err
			}

			for _, file := range files {
				presetType := "xml"
				if i == 0 {
					presetType = "kv"
				}

				entry := Entry{
					Type:     presetType,
					FileName: file.Name(),
					DirName:  dir.Name(),
				}

				result.Preset[counter] = entry
				counter++
			}

			if i == 0 {
				result.XMLStartIndex = counter
			}
		}

	}

	return result, nil
}

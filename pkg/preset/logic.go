package preset

import (
	"embed"
	"io/fs"
)

func Load(embedFS embed.FS, dirInfo []fs.DirEntry) (*Data, error) {
	result := &Data{
		Preset: make(map[int]Entry),
	}

	counter := 0

	for i, info := range dirInfo {
		files, err := embedFS.ReadDir("examples/" + info.Name())
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
			}

			result.Preset[counter] = entry
			counter++
		}

		if i == 0 {
			result.XMLStartIndex = counter
		}
	}

	return result, nil
}

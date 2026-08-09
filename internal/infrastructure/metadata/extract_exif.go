package metadata

import (
	"bytes"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

func extractEXIF(data []byte, result *ExtractedMetadata) {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	tags := map[exif.FieldName]string{
		exif.Software: "",
		exif.Make:     "",
		exif.Model:    "",
		exif.Artist:   "",
	}

	for field := range tags {
		tag, err := x.Get(field)
		if err != nil {
			continue
		}

		value, err := tag.StringVal()
		if err != nil {
			continue
		}

		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		result.RawTags[string(field)] = value

		switch field {
		case exif.Software:
			result.Software = value
		case exif.Make:
			result.Make = value
		case exif.Model:
			result.Model = value
		case exif.Artist:
			result.Artist = value
		}
	}
}

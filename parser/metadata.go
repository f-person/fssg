package parser

import (
	"strconv"
	"strings"
)

func ParseMetadata(md []byte) map[string]string {
	metadata := make(map[string]string)

	if !(len(md) > 4 && string(md[:4]) == "---\n") {
		return metadata
	}

	var lineContent strings.Builder
	md = md[4:]
	for i := 0; i < len(md); i++ {
		if md[i] == '-' && i+4 < len(md) && string(md[i:i+4]) == "---\n" {
			metadata["contentStartsAt"] = strconv.Itoa(i + 8) // two "---\n"-s excluded
			break
		}

		if md[i] == '\n' {
			lineParts := strings.Split(lineContent.String(), ":")
			key := strings.TrimSpace(lineParts[0])
			value := strings.Join(lineParts[1:], ":")
			value = strings.TrimSpace(value)
			metadata[key] = value
			lineContent.Reset()
		} else {
			lineContent.WriteByte(md[i])
		}
	}

	return metadata
}

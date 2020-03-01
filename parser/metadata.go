package parser

import (
	"strconv"
	"strings"
)

func (p *Parser) ParseMetadata() map[string]string {
	metadata := make(map[string]string)

	if !(len(p.MD) > 4 && string(p.MD[:4]) == "---\n") {
		return metadata
	}

	var lineContent strings.Builder
	p.MD = p.MD[4:]
	for i := 0; i < len(p.MD); i++ {
		if p.MD[i] == '-' && i+4 < len(p.MD) && string(p.MD[i:i+4]) == "---\n" {
			metadata["contentStartsAt"] = strconv.Itoa(i + 8) // two "---\n"-s excluded
			break
		}

		if p.MD[i] == '\n' {
			lineParts := strings.Split(lineContent.String(), ":")
			key := strings.TrimSpace(lineParts[0])
			value := strings.Join(lineParts[1:], ":")
			value = strings.TrimSpace(value)
			metadata[key] = value
			lineContent.Reset()
		} else {
			lineContent.WriteByte(p.MD[i])
		}
	}

	return metadata
}

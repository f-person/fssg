package parser

import "strings"

func (p *Parser) ParseMetadata() map[string]interface{} {
	metadata := make(map[string]interface{})

	if !(len(p.MD) > 4 && string(p.MD[:4]) == "---\n") {
		metadata["contentStartsAt"] = 0

		return metadata
	}

	var lineContent strings.Builder
	p.MD = p.MD[4:]
	for i := 0; i < len(p.MD); i++ {
		if p.MD[i] == '-' && i+4 < len(p.MD) && string(p.MD[i:i+4]) == "---\n" {
			metadata["contentStartsAt"] = i + 5 // exclude "---\n\n"
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

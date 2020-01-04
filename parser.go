package main

import (
	"strings"
)

func isEscaped(md []byte, index int) bool {
	// this will cause index out of range error
	if index-1 < 0 {
		return false
	}

	return md[index-1] == '\\'
}

func convertMarkdownToHTML(md []byte) (string, error) {
	var html strings.Builder

	inBulletList := false
	isHeading1 := false
	isHeading2 := false
	isDeleted := false

	for i := 0; i < len(md); i++ {
		switch md[i] {
		case '\n':
			if inBulletList {
				if md[i+1] == '*' {
					html.WriteString("</li>\n")
					html.WriteString("<li>")
				} else {
					inBulletList = false
					html.WriteString("</li>\n</ul>\n")
				}
			}

			if isHeading1 {
				isHeading1 = false
				html.WriteString("</h1>\n")
			} else if isHeading2 {
				isHeading2 = false
				html.WriteString("</h2>\n")
			}
		case '~':
			if isEscaped(md, i) {
				break
			}

			if i+1 <= len(md) && md[i+1] == '~' {
				if !isDeleted {
					isDeleted = true
					html.WriteString("<del>")
				} else {
					isDeleted = false
					html.WriteString("</del>")
				}
			}
		case '*':
			if isEscaped(md, i) {
				break
			}

			index := i - 1
			if !inBulletList && (index < 0 || md[i-1] == '\n') {
				inBulletList = true
				html.WriteString("\n<ul>\n<li>")
			}
		case '#':
			if isEscaped(md, i) {
				break
			}

			index := i - 1
			if index < 0 || md[i-1] == '\n' {
				if md[i+1] == '#' {
					isHeading2 = true
					html.WriteString("<h2>")
				} else {
					isHeading1 = true
					html.WriteString("<h1>")
				}
			}
		case '\\':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
			}
		default:
			html.WriteByte(md[i])
		}
	}

	return html.String(), nil
}

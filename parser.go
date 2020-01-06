package main

import (
	"bytes"
	"strings"
)

func isEscaped(md []byte, index int) bool {
	// This will cause index out of range error
	if index-1 < 0 {
		return false
	}

	return md[index-1] == '\\'
}

func convertMarkdownToHTML(md []byte) (string, error) {
	var html strings.Builder
	var linkDescription bytes.Buffer
	var link bytes.Buffer

	inBulletList := false
	isHeading1 := false
	isHeading2 := false
	isDeleted := false
	isInLinkDescription := false
	isInLink := false

	for i := 0; i < len(md); i++ {
		switch md[i] {
		case '\n':
			if isInLinkDescription {
				html.WriteByte('[')
				html.Write(linkDescription.Bytes())
				linkDescription.Reset()

				isInLinkDescription = false
			} else if isInLink {
				html.WriteByte(']')
				html.WriteByte('(')

			}

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

			html.WriteString("</p>")
		case '~':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
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
				html.WriteByte(md[i])
				break
			}

			index := i - 1
			if !inBulletList && (index < 0 || md[i-1] == '\n') {
				inBulletList = true
				html.WriteString("\n<ul>\n<li>")
			}
		case '#':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
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
		case '[':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
				break
			}

			isInLinkDescription = true
		case ']':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
				break
			}

			if i+1 < len(md) && md[i+1] == '(' {
				isInLinkDescription = false
				isInLink = true
			}
		case '(':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
			}
		case ')':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
				break
			}

			// If before it was in link, it's outside now
			// Write link and it's description as html
			if isInLink {
				html.WriteString("<a href=\"")
				html.Write(link.Bytes())
				html.WriteString("\">")
				html.Write(linkDescription.Bytes())
				html.WriteString("</a>")

				link.Reset()
				linkDescription.Reset()
				isInLink = false
			}
		default:
			if isInLinkDescription {
				linkDescription.WriteByte(md[i])
			} else if isInLink {
				link.WriteByte(md[i])
			} else if i-1 >= len(md) && md[i-1] == '\n' && !inBulletList {
				html.WriteString("<p>")
			} else {
				html.WriteByte(md[i])
			}
		}
	}

	return html.String(), nil
}

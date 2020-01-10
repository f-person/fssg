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

	isParagraphClosed := false
	inBulletList := false
	isHeading1 := false
	isHeading2 := false

	isDeleted := false
	isEmphasized := false
	isStrong := false

	isInLinkDescription := false
	isInLink := false

	html.WriteString("<p>")
	for i := 0; i < len(md); i++ {
		switch md[i] {
		case '\n':
			if isInLinkDescription {
				html.WriteByte('[')
				html.Write(linkDescription.Bytes())

				isInLinkDescription = false
				linkDescription.Reset()
			} else if isInLink {
				html.WriteByte('[')
				html.Write(linkDescription.Bytes())
				html.WriteByte(']')
				html.WriteByte('(')
				html.Write(link.Bytes())

				isInLink = false
				link.Reset()
				linkDescription.Reset()
			}

			if inBulletList {
				if i+1 < len(md) && md[i+1] == '*' {
					html.WriteString("</li>")
					html.WriteString("<li>")
				} else {
					inBulletList = false
					html.WriteString("</li></ul>")
				}
			}

			if isHeading1 {
				isHeading1 = false
				html.WriteString("</h1>")
			} else if isHeading2 {
				isHeading2 = false
				html.WriteString("</h2>")
			}

			if !isParagraphClosed {
				html.WriteString("</p>")
				isParagraphClosed = true
			}
		case '~':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
				break
			}

			if i+1 < len(md) && md[i+1] == '~' {
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

			if inBulletList && i-1 > 0 && md[i-1] == '\n' {
				break
			}

			if (i-1 < 0 || md[i-1] == '\n') && (i+1 < len(md) && md[i+1] == ' ') {
				inBulletList = true
				html.WriteString("<ul><li>")
			} else if i+1 < len(md) && md[i+1] == '*' {
				if !isStrong {
					isStrong = true
					html.WriteString("<strong>")
				} else {
					isStrong = false
					html.WriteString("</strong>")
				}
			} else if i-1 >= 0 && md[i-1] != '*' {
				if !isEmphasized {
					isEmphasized = true
					html.WriteString("<em>")
				} else {
					isEmphasized = false
					html.WriteString("</em>")
				}
			}
		case '#':
			if isEscaped(md, i) {
				html.WriteByte(md[i])
				break
			}

			if i-1 < 0 || md[i-1] == '\n' {
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
			if isEscaped(md, i) || !isInLinkDescription {
				html.WriteByte(md[i])
				break
			}

			if i+1 < len(md) && md[i+1] == '(' && isInLinkDescription {
				isInLinkDescription = false
				isInLink = true
			}
		case '(':
			if isEscaped(md, i) || !isInLink {
				html.WriteByte(md[i])
			}
		case ')':
			if isEscaped(md, i) || !isInLink {
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
		case ' ':
			if isInLinkDescription {
				if md[i-1] == ']' {
					html.WriteByte('[')
					html.Write(linkDescription.Bytes())
					html.WriteByte(']')
					html.WriteByte(md[i])

					isInLinkDescription = false
					linkDescription.Reset()
				} else {
					linkDescription.WriteByte(md[i])
				}
			} else {
				html.WriteByte(md[i])
			}
		default:
			if isInLinkDescription {
				linkDescription.WriteByte(md[i])
			} else if isInLink {
				link.WriteByte(md[i])
			} else if i-1 >= 0 && md[i-1] == '\n' && !inBulletList {
				if isParagraphClosed {
					html.WriteString("<p>")
					isParagraphClosed = false
				}

				html.WriteByte(md[i])
			} else {
				html.WriteByte(md[i])
			}
		}
	}

	return html.String(), nil
}

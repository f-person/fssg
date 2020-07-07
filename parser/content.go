package parser

import (
	"bytes"
	"strings"
)

type Parser struct {
	MD   []byte
	html strings.Builder

	linkDescription bytes.Buffer
	link            bytes.Buffer

	isParagraphClosed bool
	isHeading1        bool
	isHeading2        bool

	inBulletList bool
	inNumberList bool

	isDeleted    bool
	isEmphasized bool
	isStrong     bool

	isInLinkDescription bool
	isInLink            bool
	isProbablyImageLink bool
}

func (p *Parser) isEscaped(index int) bool {
	// This will cause index out of range error
	if index-1 < 0 {
		return false
	}

	return p.MD[index-1] == '\\'
}

func (p *Parser) defaultCase(i int) {
	if p.isInLinkDescription {
		p.linkDescription.WriteByte(p.MD[i])
	} else if p.isInLink {
		p.link.WriteByte(p.MD[i])
	} else if i-1 >= 0 && p.MD[i-1] == '\n' && !p.inBulletList {
		if p.isParagraphClosed {
			p.html.WriteString("<p>")
			p.isParagraphClosed = false
		}

		p.html.WriteByte(p.MD[i])
	} else {
		p.html.WriteByte(p.MD[i])
	}
}

func (p *Parser) ConvertMarkdownToHTML() (string, error) {
	p.html.WriteString("<p>")
	for i := 0; i < len(p.MD); i++ {
		switch p.MD[i] {
		case '\n':
			if p.isInLinkDescription {
				p.html.WriteByte('[')
				p.html.Write(p.linkDescription.Bytes())

				p.isInLinkDescription = false
				p.linkDescription.Reset()
			} else if p.isInLink {
				p.html.WriteByte('[')
				p.html.Write(p.linkDescription.Bytes())
				p.html.WriteByte(']')
				p.html.WriteByte('(')
				p.html.Write(p.link.Bytes())

				p.isInLink = false
				p.link.Reset()
				p.linkDescription.Reset()
			}

			if p.inBulletList || p.inNumberList {
				if i+1 < len(p.MD) && (p.MD[i+1] == '*' || (p.MD[i+1] >= '0' && p.MD[i+1] <= '9')) {
					p.html.WriteString("</li>")
					p.html.WriteString("<li>")

					if p.inNumberList {
						for i < len(p.MD) &&
							(p.MD[i+1] == ' ' || p.MD[i+1] == '.' ||
								(p.MD[i+1] >= '0' && p.MD[i+1] <= '9')) {
							i++
						}
					}
				} else {
					if p.inBulletList {
						p.inBulletList = false
						p.html.WriteString("</li></ul>")
					} else {
						p.inNumberList = false
						p.html.WriteString("</li></ol>")
					}
				}

				break
			}

			if p.isHeading1 {
				p.isHeading1 = false
				p.html.WriteString("</h1>")
			} else if p.isHeading2 {
				p.isHeading2 = false
				p.html.WriteString("</h2>")
			}

			if len(p.MD)-1 > 0 {
				p.html.WriteString("<br/>")
			} else if !p.isParagraphClosed {
				p.html.WriteString("</p>")
				p.isParagraphClosed = true
			}
		case '~':
			if p.isEscaped(i) {
				p.html.WriteByte(p.MD[i])
				break
			}

			if i+1 < len(p.MD) && p.MD[i+1] == '~' {
				if !p.isDeleted {
					p.isDeleted = true
					p.html.WriteString("<del>")
				} else {
					p.isDeleted = false
					p.html.WriteString("</del>")
				}
			}
		case '*':
			if p.isEscaped(i) {
				p.html.WriteByte(p.MD[i])
				break
			}

			if p.inBulletList && i-1 > 0 && p.MD[i-1] == '\n' {
				break
			}

			if (i-1 < 0 || p.MD[i-1] == '\n') && (i+1 < len(p.MD) && p.MD[i+1] == ' ') {
				p.inBulletList = true
				p.html.WriteString("<ul><li>")
			} else if i+1 < len(p.MD) && p.MD[i+1] == '*' {
				if !p.isStrong {
					p.isStrong = true
					p.html.WriteString("<strong>")
				} else {
					p.isStrong = false
					p.html.WriteString("</strong>")
				}
			} else if i-1 >= 0 && p.MD[i-1] != '*' {
				if !p.isEmphasized {
					p.isEmphasized = true
					p.html.WriteString("<em>")
				} else {
					p.isEmphasized = false
					p.html.WriteString("</em>")
				}
			}
		case '#':
			if p.isEscaped(i) {
				p.html.WriteByte(p.MD[i])
				break
			}

			if i-1 < 0 || p.MD[i-1] == '\n' {
				if p.MD[i+1] == '#' {
					p.isHeading2 = true
					p.html.WriteString("<h2>")
				} else {
					p.isHeading1 = true
					p.html.WriteString("<h1>")
				}
			}
		case '\\':
			if p.isEscaped(i) {
				p.html.WriteByte(p.MD[i])
			}
		case '!':
			if p.isEscaped(i) || i+1 > len(p.MD) {
				p.html.WriteByte(p.MD[i])
				break
			}

			p.isProbablyImageLink = true
		case '[':
			if p.isEscaped(i) {
				p.html.WriteByte(p.MD[i])
				break
			}

			p.isInLinkDescription = true
		case ']':
			if p.isEscaped(i) || !p.isInLinkDescription {
				p.html.WriteByte(p.MD[i])
				break
			}

			if i+1 < len(p.MD) && p.MD[i+1] == '(' && p.isInLinkDescription {
				p.isInLinkDescription = false
				p.isInLink = true
			}
		case '(':
			if p.isEscaped(i) || !p.isInLink {
				p.html.WriteByte(p.MD[i])
			}
		case ')':
			if p.isEscaped(i) || !p.isInLink {
				p.html.WriteByte(p.MD[i])
				break
			}

			// If before it was in link, it's outside now
			// Write link and it's description as html
			if p.isInLink {
				if p.isProbablyImageLink {
					p.html.WriteString("<img src=\"")
					p.html.Write(p.link.Bytes())
					p.html.WriteString("\" alt=\"")
					p.html.Write(p.linkDescription.Bytes())
					p.html.WriteByte('"')
					p.html.WriteString(" title=\"")
					p.html.Write(p.linkDescription.Bytes())
					p.html.WriteString("\">")
				} else {
					p.html.WriteString("<a href=\"")
					p.html.Write(p.link.Bytes())
					p.html.WriteString("\">")
					p.html.Write(p.linkDescription.Bytes())
					p.html.WriteString("</a>")
				}
				p.link.Reset()
				p.linkDescription.Reset()
				p.isProbablyImageLink = false
				p.isInLink = false
			}
		case ' ':
			if p.isInLinkDescription {
				if p.MD[i-1] == ']' {
					p.html.WriteByte('[')
					p.html.Write(p.linkDescription.Bytes())
					p.html.WriteByte(']')
					p.html.WriteByte(p.MD[i])

					p.isInLinkDescription = false
					p.linkDescription.Reset()
				} else {
					p.linkDescription.WriteByte(p.MD[i])
				}
			} else {
				p.html.WriteByte(p.MD[i])
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if p.inNumberList && i-1 > 0 && p.MD[i-1] == '\n' {
				break
			}

			if (i-1 < 0 || p.MD[i-1] == '\n') && (i+1 < len(p.MD) && p.MD[i+1] == '.') {
				p.inNumberList = true
				p.html.WriteString("<ol><li>")
				for i < len(p.MD) && (p.MD[i] == ' ' || p.MD[i] == '.' || (p.MD[i] >= '0' && p.MD[i] <= '9')) {
					i++
				}
			} else {
				p.defaultCase(i)
			}
		default:
			p.defaultCase(i)
		}
	}

	return p.html.String(), nil
}

package service

import (
	"bytes"
	"html"
	"net/url"
	"strings"

	"github.com/russross/blackfriday/v2"
	xhtml "golang.org/x/net/html"
)

var allowedElements = map[string]map[string]bool{
	"p":          {},
	"br":         {},
	"hr":         {},
	"blockquote": {},
	"pre":        {"class": true},
	"code":       {"class": true},
	"ul":         {},
	"ol":         {},
	"li":         {},
	"em":         {},
	"strong":     {},
	"a":          {"href": true, "title": true},
	"img":        {"src": true, "alt": true, "title": true},
	"h1":         {"id": true},
	"h2":         {"id": true},
	"h3":         {"id": true},
	"h4":         {"id": true},
	"h5":         {"id": true},
	"h6":         {"id": true},
}

func RenderMarkdown(input string) string {
	// 原始 Markdown 先转 HTML，再经过白名单清洗，避免脚本和危险属性混入。
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.UseXHTML |
			blackfriday.SkipHTML |
			blackfriday.Safelink |
			blackfriday.NofollowLinks |
			blackfriday.NoreferrerLinks |
			blackfriday.NoopenerLinks |
			blackfriday.HrefTargetBlank,
	})

	extensions := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs
	rendered := blackfriday.Run([]byte(input), blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(extensions))
	return sanitizeHTML(string(rendered))
}

func sanitizeHTML(input string) string {
	node, err := xhtml.Parse(strings.NewReader("<div>" + input + "</div>"))
	if err != nil {
		return html.EscapeString(input)
	}

	var root *xhtml.Node
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && n.Data == "div" && root == nil {
			root = n
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	if root == nil {
		return html.EscapeString(input)
	}

	var buf bytes.Buffer
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		renderNode(&buf, child)
	}
	return strings.TrimSpace(buf.String())
}

func renderNode(buf *bytes.Buffer, node *xhtml.Node) {
	switch node.Type {
	case xhtml.TextNode:
		buf.WriteString(html.EscapeString(node.Data))
	case xhtml.ElementNode:
		attrs, ok := allowedElements[node.Data]
		if !ok {
			// 不允许的标签直接剥掉标签壳，只保留安全的文本子节点。
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				renderNode(buf, child)
			}
			return
		}

		buf.WriteByte('<')
		buf.WriteString(node.Data)
		for _, attr := range node.Attr {
			if !attrs[attr.Key] {
				continue
			}
			value, ok := sanitizeAttr(node.Data, attr.Key, attr.Val)
			if !ok {
				continue
			}
			buf.WriteByte(' ')
			buf.WriteString(attr.Key)
			buf.WriteString(`="`)
			buf.WriteString(html.EscapeString(value))
			buf.WriteByte('"')
		}
		if node.Data == "a" {
			buf.WriteString(` rel="nofollow noopener noreferrer"`)
		}
		if node.Data == "br" || node.Data == "hr" || node.Data == "img" {
			buf.WriteString(" />")
			return
		}
		buf.WriteByte('>')

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			renderNode(buf, child)
		}

		buf.WriteString("</")
		buf.WriteString(node.Data)
		buf.WriteByte('>')
	}
}

func sanitizeAttr(tag, key, value string) (string, bool) {
	switch key {
	case "href", "src":
		if safeURL(value) {
			return value, true
		}
		return "", false
	case "class":
		if tag == "code" || tag == "pre" {
			return value, true
		}
		return "", false
	default:
		return value, true
	}
}

func safeURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if u.Scheme == "" {
		return strings.HasPrefix(raw, "/") || !strings.Contains(raw, ":")
	}
	switch u.Scheme {
	case "http", "https", "mailto":
		return true
	default:
		return false
	}
}

package htmlwriter

import (
	"bytes"
)

type HtmlNode struct {
	name      string
	text      string
	attr      map[string]string
	child     []*HtmlNode
	classHash map[string]string
	styleHash map[string]string
	parents   *HtmlNode
}

func CreateHtmlNode(name string) *HtmlNode {
	return &HtmlNode{name, "", make(map[string]string), make([]*HtmlNode, 0, 100), make(map[string]string), make(map[string]string), nil}
}

func (n *HtmlNode) Add(name string) *HtmlNode {
	c := CreateHtmlNode(name)
	n.child = append(n.child, c)
	c.parents = n
	return c
}
func (n *HtmlNode) InsertAfter(c *HtmlNode) *HtmlNode {
	if n.parents != nil {
		n.Detach()
	}

	if len(c.parents.child) > 0 {
		newChild := make([]*HtmlNode, 0, len(c.parents.child)+1)
		for _, v := range c.parents.child {
			if v != c {
				newChild = append(newChild, v)
			} else {
				newChild = append(newChild, c)
				newChild = append(newChild, n)
			}
		}
		c.parents.child = newChild
		n.parents = c.parents
	}
	return n
}

func (n *HtmlNode) Eq(idx int) *HtmlNode {
	if len(n.child) > idx {
		return n.child[idx]
	}
	return nil
}

func (n *HtmlNode) Append(c *HtmlNode) *HtmlNode {
	if c.parents != nil {
		c.Detach()
	}
	n.child = append(n.child, c)
	c.parents = n
	return n
}
func (c *HtmlNode) AppendTo(n *HtmlNode) *HtmlNode {
	n.Append(c)
	return c
}

func (n *HtmlNode) Prepend(c *HtmlNode) *HtmlNode {
	if c.parents != nil {
		c.Detach()
	}
	n.child = append([]*HtmlNode{c}, n.child...)
	c.parents = n
	return n
}

func (n *HtmlNode) Detach() *HtmlNode {
	return n.parents.Remove(n)
}
func (n *HtmlNode) Remove(c *HtmlNode) *HtmlNode {
	if len(n.child) > 0 {
		newChild := make([]*HtmlNode, 0, len(n.child)-1)
		for _, v := range n.child {
			if v != c {
				newChild = append(newChild, v)
			}
		}
		n.child = newChild
		c.parents = nil
	}
	return n
}

func (n *HtmlNode) SetText(text string) *HtmlNode {
	n.text = text
	return n
}

func (n *HtmlNode) Text() string {
	return n.text
}

func (n *HtmlNode) String() string {
	var buffer bytes.Buffer
	n.Write(&buffer)
	return buffer.String()
}

func (n *HtmlNode) Write(buffer *bytes.Buffer) {
	n.WriteWith(buffer, "")
}

func (n *HtmlNode) WriteWith(buffer *bytes.Buffer, with string) {
	n.writeWith(buffer, with, 0)
}

func (n *HtmlNode) writeWith(buffer *bytes.Buffer, with string, depth int) {
	if len(with) > 0 {
		buffer.WriteString("\n")
		for i := 0; i < depth; i++ {
			buffer.WriteString(with)
		}
	}
	buffer.WriteString("<")
	buffer.WriteString(n.name)

	if len(n.classHash) > 0 {
		isFirst := true
		buffer.WriteString(" class=\"")
		for v := range n.classHash {
			if isFirst {
				isFirst = false
			} else {
				buffer.WriteString(" ")
			}
			buffer.WriteString(v)
		}
		buffer.WriteString("\"")
	}

	if len(n.styleHash) > 0 {
		isFirst := true
		buffer.WriteString(" style=\"")
		for k, v := range n.styleHash {
			if isFirst {
				isFirst = false
			} else {
				buffer.WriteString(" ")
			}
			buffer.WriteString(k)
			buffer.WriteString(":")
			buffer.WriteString(v)
			buffer.WriteString(";")
		}
		buffer.WriteString("\"")
	}

	if len(n.attr) > 0 {
		isFirst := true
		for k, v := range n.attr {
			if isFirst {
				isFirst = false
			} else {
				buffer.WriteString(" ")
			}
			buffer.WriteString(" ")
			buffer.WriteString(k)
			buffer.WriteString("=\"")
			buffer.WriteString(v)
			buffer.WriteString("\"")
		}
	}
	buffer.WriteString(">")

	hasContent := false
	if len(n.text) > 0 {
		hasContent = true
		buffer.WriteString("\n")
		if len(with) > 0 {
			for i := 0; i < depth; i++ {
				buffer.WriteString(with)
			}
		}
		buffer.WriteString(n.text)
	}
	if len(n.child) > 0 {
		hasContent = true
		for _, v := range n.child {
			v.writeWith(buffer, with, depth+1)
		}
	}

	if n.name != "br" {
		if hasContent {
			buffer.WriteString("\n")
			if len(with) > 0 {
				for i := 0; i < depth; i++ {
					buffer.WriteString(with)
				}
			}
		}
		buffer.WriteString("</")
		buffer.WriteString(n.name)
		buffer.WriteString(">")
	}
}

func (n *HtmlNode) Class(class string) *HtmlNode {
	ch := n.classHash
	ch[class] = class
	return n
}
func (n *HtmlNode) RemoveClass(class string) *HtmlNode {
	delete(n.classHash, class)
	return n
}

func (n *HtmlNode) Style(style string, value string) *HtmlNode {
	n.styleHash[style] = value
	return n
}
func (n *HtmlNode) RemoveStyle(style string) *HtmlNode {
	delete(n.styleHash, style)
	return n
}

func (n *HtmlNode) Attr(key string, value string) *HtmlNode {
	n.attr[key] = value
	return n
}
func (n *HtmlNode) RemoveAttr(key string) *HtmlNode {
	delete(n.attr, key)
	return n
}

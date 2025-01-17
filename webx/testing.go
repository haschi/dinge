package webx

import (
	"slices"

	"golang.org/x/net/html"
)

// TODO: Move to htmlx
func GetElement(n *html.Node, element string) []*html.Node {

	result := []*html.Node{}

	for i := n.FirstChild; i != nil; i = i.NextSibling {
		result = slices.Concat(result, GetElement(i, element))
	}

	if n.Type == html.ElementNode {
		if n.Data == element {
			result = append(result, n)
		}
	}

	return result
}

// TODO: Move to htmlx
func GetAttributeValue(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}

func GetElementById(node *html.Node, id string) *html.Node {

	if GetAttributeValue(node, "id") == id {
		return node
	}

	next := node.NextSibling
	for next != nil {
		if result := GetElementById(next, id); result != nil {
			return result
		}
		next = next.NextSibling
	}

	for child := range node.Descendants() {
		return GetElementById(child, id)
	}

	return nil
}

func GetSelectedOption(n *html.Node) string {
	for descendant := range n.Descendants() {
		if descendant.Type == html.ElementNode && descendant.Data == "option" {
			if HasAttribute(descendant, "selected") {
				return GetAttributeValue(descendant, "value")
			}
		}
	}
	return ""
}

func HasAttribute(n *html.Node, key string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return true
		}
	}

	return false
}

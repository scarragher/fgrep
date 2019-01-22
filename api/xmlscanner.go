package fgrep

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// XMLScanner scans an XML file
type XMLScanner struct {
}

type Node interface{}
type Element struct {
	Name     string
	Children []Node
}

type node struct {
	name    string
	content string
}

func (n *node) String() string {
	return fmt.Sprintf("<%s>%s</%s>", n.name, n.content, n.name)
}

// Scan scans an XML, using node terminators to split into multiple lines for output
func (s *XMLScanner) Scan(fileContents []byte, searchContent string, printContent bool) ([]string, bool) {
	matches := []string{}
	reader := strings.NewReader(string(fileContents))
	decoder := xml.NewDecoder(reader)
	decoder.Entity = map[string]string{
		"lt":   "<",
		"gt":   ">",
		"amp":  "&",
		"apos": "'",
		"quot": `"`,
	}

	depth := 0

	// xmlnode: {nodename} returns the full node and all children within
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			// error handling
			return matches, false
		}
		switch tok := token.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
		}
	}

	return matches, true
}

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
	raw := scrub(string(fileContents))
	reader := strings.NewReader(raw)
	decoder := xml.NewDecoder(reader)
	node := node{}

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
			node.name = tok.Name.Local
		case xml.EndElement:
			s := node.String()
			if strings.Contains(strings.ToLower(s), strings.ToLower(searchContent)) {
				if printContent {
					fmt.Println(s)
				}

				matches = append(matches, s)
			}
		case xml.CharData:
			node.content = string(tok)
		}

	}

	return matches, true
}

func scrub(xml string) string {
	scrubbed := strings.Replace(xml, `&lt;`, "<", -1)
	scrubbed = strings.Replace(scrubbed, `&gt;`, ">", -1)
	return scrubbed
}

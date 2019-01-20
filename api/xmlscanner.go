package fgrep

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// XMLScanner scans an XML file
type XMLScanner struct {
}

type node struct {
	name    string
	content string
}

// Scan scans an XML, using node terminators to split into multiple lines for output
func (s *XMLScanner) Scan(fileContents []byte, searchContent string, printContent bool) ([]string, bool) {
	matches := []string{}
	reader := bytes.NewReader(fileContents)
	decoder := xml.NewDecoder(reader)

	fmt.Println(string(fileContents))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			// error handling
			return matches, false
		}
		nodes := []node{}
		current := node{}

		switch tok := token.(type) {
		case xml.StartElement:
			current.name = tok.Name.Local
			//matchingNodes = append(matchingNodes, tok.Name.Local)
			matches = append(matches, tok.Name.Local)
		case xml.EndElement:
			nodes = append(nodes, current)
			//matchingNodes = matchingNodes[:len(matchingNodes)-1]
		case xml.CharData:
			current.content = string(tok)
		}

		fmt.Println(nodes)
	}

	return matches, true
	/*
		decode XML document into node tree and go through the tree
		new scanner
	*/
}

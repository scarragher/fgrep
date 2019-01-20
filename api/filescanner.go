package fgrep

import (
	"fmt"
	"net/http"
	"strings"
)

var (
	validContentTypes = []string{"text/"}
)

// FileScanner Defines a scanner that scans file contents for matches
type FileScanner interface {
	Scan(fileContents []byte, searchContent string, printContent bool) (matches []string, ok bool)
}

// TextScanner scans a plain text file
type TextScanner struct {
}

// Scan scans a paint text file for a specific string
func (t *TextScanner) Scan(fileContents []byte, searchContent string, printContent bool) (matches []string, ok bool) {
	plainText := string(fileContents)
	segments := strings.Split(plainText, "\n")

	for _, s := range segments {

		if strings.Contains(strings.ToLower(s), strings.ToLower(searchContent)) {
			if printContent {
				fmt.Println(s)
			}

			matches = append(matches, s)
		}
	}

	return matches, true
}

// Scan scans a specific file's contents using the file extension to select the correct FileScanner to use
func Scan(searchContent string, fileExt string, fileContents []byte, printContent bool) ([]string, bool) {
	matches := []string{}

	if len(fileContents) == 0 {
		return matches, false
	}

	contentType := http.DetectContentType(fileContents)

	if !validContentType(contentType) {
		return matches, false
	}

	var scanner FileScanner
	switch fileExt {
	case "txt":
		scanner = &TextScanner{}
	case "xml":
		scanner = &XMLScanner{}
	default:
		scanner = &TextScanner{}
	}

	matches, ok := scanner.Scan(fileContents, searchContent, printContent)
	return matches, ok
}

func validContentType(contentType string) bool {
	for _, validContentType := range validContentTypes {
		content := contentType[0:len(validContentType)]

		if content == validContentType {
			return true
		}
	}

	return false
}

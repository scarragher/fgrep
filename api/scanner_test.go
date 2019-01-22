package fgrep

import (
	"testing"
)

func TestScanner(t *testing.T) {
	tests := []struct {
		ext     string
		search  string
		matches []string
		ok      bool
		content []byte
	}{
		{"txt", "trumpet", []string{"this is a text trumpet file"}, true, []byte("this is a text trumpet file")},
		{"test", "trumpet", []string{"this is a test trumpet test file"}, true, []byte("this is a test trumpet test file")},
		{"test", "trumpet", []string{"trumpet", "another trumpet"}, true, []byte("this is a test \ntrumpet\n test file\nanother trumpet")},
		{"xml", "dinosaur", []string{"<child2>two dinosaur</child2>"}, true, []byte("<parent><child1>one</child1><child2>two dinosaur</child2></parent>")},
		{"xml", "xmlnode:child2", []string{"<child2>two dinosaur</child2>"}, true, []byte("<parent><child1>one</child1><child2>two dinosaur</child2></parent>")},
		{"xml", "parent", []string{"<parent><child1>one</child1><child2>two dinosaur</child2></parent>"}, true, []byte("<parent><child1>one</child1><child2>two dinosaur</child2></parent>")},
		{"xml", "dinosaur", []string{"<child2>two dinosaur33</child2>"}, true, []byte("&lt;parent&gt;&lt;child1&gt;one&lt;/child1&gt;&lt;child2&gt;two dinosaur33&lt;/child2&gt;&lt;/parent&gt;")},
		{"gif", "", []string{}, false, []byte(`GIF89a...`)},
		{"midi", "", []string{}, false, []byte("MThd\x00\x00\x00\x06\x00\x01")},
		{"txt", "trumpet", []string{}, false, []byte{}},
	}

	for _, test := range tests {
		matches, ok := Scan(test.search, test.ext, test.content, false)
		if ok != test.ok {
			t.Errorf("Expected ok %v, got %v", test.ok, ok)
		}

		got := len(matches)
		want := len(test.matches)
		if len(matches) != len(test.matches) {
			t.Errorf("Expected %d matches, got %d searching for %s in %s\n%v", want, got, test.search, test.content, matches)
		}

		for i, w := range test.matches {
			m := matches[i]

			if w != m {
				t.Errorf("Expected match %s. Got (%d): %v", w, len(matches), matches)
			}
		}
	}
}

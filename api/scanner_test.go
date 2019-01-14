package fgrep

import "testing"

func TestScanner(t *testing.T){
	tests := []struct{		
		ext string 
		search string
		found bool
		ok bool
		content []byte
	}{
		{"txt", "trumpet", true, true, []byte("this is a text trumpet file") },		
		{"test", "trumpet", true, true, []byte("this is a test trumpet test file") },		
		{"xml", "dinosaur", true, true, []byte("<parent><child1>one</child1><child2>two dinosaur</child2></parent>") },
		{"gif", "", false, false, []byte(`GIF89a...`)},
		{"midi", "", false, false, []byte("MThd\x00\x00\x00\x06\x00\x01")},
		{"txt", "trumpet", false, false, []byte{} },
	}

	for _, test := range tests {
		match, ok := Scan( test.search, test.ext, test.content, false)
		if ok != test.ok {
			t.Errorf("Expected ok %v, got %v", test.ok, ok)
		}
		if test.found != match {
			t.Errorf("Expected match %v, got %v", test.found, match)
		}
	}
}

func TestTextScanner(t *testing.T ){
	scanner := TextScanner{}
	tests := []struct { 
		match bool 
		search string 
		content string 
	}{
		{true, "file", "this is a text file\nso it is"},
		{false, "test", "nothing here"},
	}

	for _, test := range tests {
		got, ok := scanner.Scan([]byte(test.content), test.search, false)
		if !ok{
			t.Error("scan not ok")
		}
		if got != test.match {
			t.Errorf("Expected %v, got %v", test.match, got)
		}
	}		
}

func TestXMLScanner( t *testing.T){
	scanner := XMLScanner{}
	tests := []struct { 
		match bool 
		search string 
		content string 
	}{
		{true, "elephant", "<node1><child1>elephant</child1></node1>"},
		{false, "test", "<node5><child5>value</child5></node5>"},
	}

	for _, test := range tests {
		got, ok := scanner.Scan([]byte(test.content), test.search, false)
		if !ok{
			t.Error("scan not ok")
		}
		if got != test.match {
			t.Errorf("Expected %v, got %v", test.match, got)
		}
	}	
}
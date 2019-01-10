package fgrep

import "testing"

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
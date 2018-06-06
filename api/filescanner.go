package fgrep

// FileScanner Defines a scanner that scans file contents for matches
type FileScanner interface {
	Scan(fileContents []byte, searchContent string)
	Ext() string
}

type TextScanner struct {
}

func (t *TextScanner) Scan(fileContents []byte, searchContent string) {

}

func (t *TextScanner) Ext() string {
	return "txt"
}

// Scanner handles scanning a file's contents using the correct FileScanner
type Scanner struct {
	scanners map[string]FileScanner
}

// NewScanner creates and initialises a new Scanner
func NewScanner() *Scanner {
	scanner := Scanner{}

	// register the scanners

	return &scanner
}

// Scan scans a specific file's contents using the file extension to select the correct FileScanner to use
func (scanner *Scanner) Scan(fileExt string, fileContents []byte) {
	fileScanner, ok := scanner.scanners[fileExt]

	if !ok {
		fileScanner = TextScanner{}
	}
}

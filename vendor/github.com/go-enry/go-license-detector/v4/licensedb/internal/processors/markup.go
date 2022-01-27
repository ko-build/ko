package processors

import (
	"bytes"
	"sync"

	rst "github.com/hhatto/gorst"
	"github.com/russross/blackfriday/v2"
)

var (
	parserLock sync.Mutex
)

// Markdown converts Markdown to plain text. It tries to revert all the decorations.
func Markdown(text []byte) []byte {
	html := blackfriday.Run(text)
	// Repeat to times to heal broken HTML
	return HTML(html)
}

// RestructuredText converts ReStructuredText to plain text.
// It tries to revert all the decorations.
func RestructuredText(text []byte) []byte {
	parserLock.Lock()
	defer parserLock.Unlock()
	parser := rst.NewParser(nil)
	input := bytes.NewBuffer(text)
	output := &bytes.Buffer{}
	parser.ReStructuredText(input, rst.ToHTML(output))
	// Repeat to times to heal broken HTML
	return HTML(output.Bytes())
}

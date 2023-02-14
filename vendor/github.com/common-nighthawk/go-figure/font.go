package figure

import (
	"bufio"
	"bytes"
	"io"
	"path"
	"strings"
)

const defaultFont = "standard"

var colors = map[string]string{
	"reset":  "\033[0m",
	"red":    "\033[31m",
	"green":  "\033[32m",
	"yellow": "\033[33m",
	"blue":   "\033[34m",
	"purple": "\033[35m",
	"cyan":   "\033[36m",
	"gray":   "\033[37m",
	"white":  "\033[97m",
}

type font struct {
	name      string
	height    int
	baseline  int
	hardblank byte
	reverse   bool
	letters   [][]string
}

func newFont(name string) (font font) {
	font.setName(name)
	fontBytes, err := Asset(path.Join("fonts", font.name+".flf"))
	if err != nil {
		panic(err)
	}
	fontBytesReader := bytes.NewReader(fontBytes)
	scanner := bufio.NewScanner(fontBytesReader)
	font.setAttributes(scanner)
	font.setLetters(scanner)
	return font
}

func newFontFromReader(reader io.Reader) (font font) {
	scanner := bufio.NewScanner(reader)
	font.setAttributes(scanner)
	font.setLetters(scanner)
	return font
}

func (font *font) setName(name string) {
	font.name = name
	if len(name) < 1 {
		font.name = defaultFont
	}
}

func (font *font) setAttributes(scanner *bufio.Scanner) {
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, signature) {
			font.height = getHeight(text)
			font.baseline = getBaseline(text)
			font.hardblank = getHardblank(text)
			font.reverse = getReverse(text)
			break
		}
	}
}

func (font *font) setLetters(scanner *bufio.Scanner) {
	font.letters = append(font.letters, make([]string, font.height, font.height)) //TODO: set spaces from flf
	for i := range font.letters[0] {
		font.letters[0][i] = "  "
	} //TODO: set spaces from flf
	letterIndex := 0
	for scanner.Scan() {
		text, cutLength, letterIndexInc := scanner.Text(), 1, 0
		if lastCharLine(text, font.height) {
			font.letters = append(font.letters, []string{})
			letterIndexInc = 1
			if font.height > 1 {
				cutLength = 2
			}
		}
		if letterIndex > 0 {
			appendText := ""
			if len(text) > 1 {
				appendText = text[:len(text)-cutLength]
			}
			font.letters[letterIndex] = append(font.letters[letterIndex], appendText)
		}
		letterIndex += letterIndexInc
	}
}

func (font *font) evenLetters() {
	var longest int
	for _, letter := range font.letters {
		if len(letter) > 0 && len(letter[0]) > longest {
			longest = len(letter[0])
		}
	}
	for _, letter := range font.letters {
		for i, row := range letter {
			letter[i] = row + strings.Repeat(" ", longest-len(row))
		}
	}
}

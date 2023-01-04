package figure

import (
	"io"
	"log"
	"reflect"
	"strings"
)

const ascii_offset = 32
const first_ascii = ' '
const last_ascii = '~'

type figure struct {
	phrase string
	font
	strict bool
	color  string
}

func NewFigure(phrase, fontName string, strict bool) figure {
	font := newFont(fontName)
	if font.reverse {
		phrase = reverse(phrase)
	}
	return figure{phrase: phrase, font: font, strict: strict}
}

func NewColorFigure(phrase, fontName string, color string, strict bool) figure {
	color = strings.ToLower(color)
	if _, found := colors[color]; !found {
		log.Fatalf("invalid color. must be one of: %s", reflect.ValueOf(colors).MapKeys())
	}

	fig := NewFigure(phrase, fontName, strict)
	fig.color = color
	return fig
}

func NewFigureWithFont(phrase string, reader io.Reader, strict bool) figure {
	font := newFontFromReader(reader)
	if font.reverse {
		phrase = reverse(phrase)
	}
	return figure{phrase: phrase, font: font, strict: strict}
}

func (figure figure) Slicify() (rows []string) {
	for r := 0; r < figure.font.height; r++ {
		printRow := ""
		for _, char := range figure.phrase {
			if char < first_ascii || char > last_ascii {
				if figure.strict {
					log.Fatal("invalid input.")
				} else {
					char = '?'
				}
			}
			fontIndex := char - ascii_offset
			charRowText := scrub(figure.font.letters[fontIndex][r], figure.font.hardblank)
			printRow += charRowText
		}
		if r < figure.font.baseline || len(strings.TrimSpace(printRow)) > 0 {
			rows = append(rows, strings.TrimRight(printRow, " "))
		}
	}
	return rows
}

func scrub(text string, char byte) string {
	return strings.Replace(text, string(char), " ", -1)
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

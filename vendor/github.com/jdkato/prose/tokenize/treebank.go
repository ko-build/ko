package tokenize

import (
	"regexp"
	"strings"
)

// TreebankWordTokenizer splits a sentence into words.
//
// This implementation is a port of the Sed script written by Robert McIntyre,
// which is available at https://gist.github.com/jdkato/fc8b8c4266dba22d45ac85042ae53b1e.
type TreebankWordTokenizer struct {
}

// NewTreebankWordTokenizer is a TreebankWordTokenizer constructor.
func NewTreebankWordTokenizer() *TreebankWordTokenizer {
	return new(TreebankWordTokenizer)
}

var startingQuotes = map[string]*regexp.Regexp{
	"$1 `` ": regexp.MustCompile(`'([ (\[{<])"`),
	"``":     regexp.MustCompile(`^(")`),
	" ``":    regexp.MustCompile(`( ")`),
}
var startingQuotes2 = map[string]*regexp.Regexp{
	" $1 ": regexp.MustCompile("(``)"),
}
var punctuation = map[string]*regexp.Regexp{
	" $1 $2":   regexp.MustCompile(`([:,])([^\d])`),
	" ... ":    regexp.MustCompile(`\.\.\.`),
	"$1 $2$3 ": regexp.MustCompile(`([^\.])(\.)([\]\)}>"\']*)\s*$`),
	"$1 ' ":    regexp.MustCompile(`([^'])' `),
}
var punctuation2 = []*regexp.Regexp{
	regexp.MustCompile(`([:,])$`),
	regexp.MustCompile(`([;@#$%&?!])`),
}
var brackets = map[string]*regexp.Regexp{
	" $1 ": regexp.MustCompile(`([\]\[\(\)\{\}\<\>])`),
	" -- ": regexp.MustCompile(`--`),
}
var endingQuotes = map[string]*regexp.Regexp{
	" '' ": regexp.MustCompile(`"`),
}
var endingQuotes2 = []*regexp.Regexp{
	regexp.MustCompile(`'(\S)(\'\')'`),
	regexp.MustCompile(`([^' ])('[sS]|'[mM]|'[dD]|') `),
	regexp.MustCompile(`([^' ])('ll|'LL|'re|'RE|'ve|'VE|n't|N'T) `),
}
var contractions = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(can)(not)\b`),
	regexp.MustCompile(`(?i)\b(d)('ye)\b`),
	regexp.MustCompile(`(?i)\b(gim)(me)\b`),
	regexp.MustCompile(`(?i)\b(gon)(na)\b`),
	regexp.MustCompile(`(?i)\b(got)(ta)\b`),
	regexp.MustCompile(`(?i)\b(lem)(me)\b`),
	regexp.MustCompile(`(?i)\b(mor)('n)\b`),
	regexp.MustCompile(`(?i)\b(wan)(na) `),
	regexp.MustCompile(`(?i) ('t)(is)\b`),
	regexp.MustCompile(`(?i) ('t)(was)\b`),
}
var newlines = regexp.MustCompile(`(?:\n|\n\r|\r)`)
var spaces = regexp.MustCompile(`(?: {2,})`)

// Tokenize splits a sentence into a slice of words.
//
// This tokenizer performs the following steps: (1) split on contractions (e.g.,
// "don't" -> [do n't]), (2) split on non-terminating punctuation, (3) split on
// single quotes when followed by whitespace, and (4) split on periods that
// appear at the end of lines.
//
// NOTE: As mentioned above, this function expects a sentence (not raw text) as
// input.
func (t TreebankWordTokenizer) Tokenize(text string) []string {
	for substitution, r := range startingQuotes {
		text = r.ReplaceAllString(text, substitution)
	}

	for substitution, r := range startingQuotes2 {
		text = r.ReplaceAllString(text, substitution)
	}

	for substitution, r := range punctuation {
		text = r.ReplaceAllString(text, substitution)
	}

	for _, r := range punctuation2 {
		text = r.ReplaceAllString(text, " $1 ")
	}

	for substitution, r := range brackets {
		text = r.ReplaceAllString(text, substitution)
	}

	text = " " + text + " "

	for substitution, r := range endingQuotes {
		text = r.ReplaceAllString(text, substitution)
	}

	for _, r := range endingQuotes2 {
		text = r.ReplaceAllString(text, "$1 $2 ")
	}

	for _, r := range contractions {
		text = r.ReplaceAllString(text, " $1 $2 ")
	}

	text = newlines.ReplaceAllString(text, " ")
	text = strings.TrimSpace(spaces.ReplaceAllString(text, " "))
	return strings.Split(text, " ")
}

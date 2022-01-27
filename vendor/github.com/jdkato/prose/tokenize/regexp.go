package tokenize

import "regexp"

// RegexpTokenizer splits a string into substrings using a regular expression.
type RegexpTokenizer struct {
	regex   *regexp.Regexp
	gaps    bool
	discard bool
}

// NewRegexpTokenizer is a RegexpTokenizer constructor that takes three
// arguments: a pattern to base the tokenizer on, a boolean value indicating
// whether or not to look for separators between tokens, and boolean value
// indicating whether or not to discard empty tokens.
func NewRegexpTokenizer(pattern string, gaps, discard bool) *RegexpTokenizer {
	rTok := RegexpTokenizer{
		regex: regexp.MustCompile(pattern), gaps: gaps, discard: discard}
	return &rTok
}

// Tokenize splits text into a slice of tokens according to its regexp pattern.
func (r RegexpTokenizer) Tokenize(text string) []string {
	var tokens []string
	if r.gaps {
		temp := r.regex.Split(text, -1)
		if r.discard {
			for _, s := range temp {
				if s != "" {
					tokens = append(tokens, s)
				}
			}
		} else {
			tokens = temp
		}
	} else {
		tokens = r.regex.FindAllString(text, -1)
	}
	return tokens
}

// NewBlanklineTokenizer is a RegexpTokenizer constructor.
//
// This tokenizer splits on any sequence of blank lines.
func NewBlanklineTokenizer() *RegexpTokenizer {
	return &RegexpTokenizer{
		regex: regexp.MustCompile(`\s*\n\s*\n\s*`), gaps: true, discard: true}
}

// NewWordPunctTokenizer is a RegexpTokenizer constructor.
//
// This tokenizer splits text into a sequence of alphabetic and non-alphabetic
// characters.
func NewWordPunctTokenizer() *RegexpTokenizer {
	return &RegexpTokenizer{
		regex: regexp.MustCompile(`\w+|[^\w\s]+`), gaps: false}
}

// NewWordBoundaryTokenizer is a RegexpTokenizer constructor.
//
// This tokenizer splits text into a sequence of word-like tokens.
func NewWordBoundaryTokenizer() *RegexpTokenizer {
	return &RegexpTokenizer{
		regex: regexp.MustCompile(`(?:[A-Z]\.){2,}|[\p{N}\p{L}']+`),
		gaps:  false}
}

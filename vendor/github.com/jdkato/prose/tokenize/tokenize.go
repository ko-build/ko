/*
Package tokenize implements functions to split strings into slices of substrings.
*/
package tokenize

// ProseTokenizer is the interface implemented by an object that takes a string
// and returns a slice of substrings.
type ProseTokenizer interface {
	Tokenize(text string) []string
}

// TextToWords converts the string text into a slice of words.
//
// It does so by tokenizing text into sentences (using a port of NLTK's punkt
// tokenizer; see https://github.com/neurosnap/sentences) and then tokenizing
// the sentences into words via TreebankWordTokenizer.
func TextToWords(text string) []string {
	sentTokenizer := NewPunktSentenceTokenizer()
	wordTokenizer := NewTreebankWordTokenizer()

	words := []string{}
	for _, s := range sentTokenizer.Tokenize(text) {
		words = append(words, wordTokenizer.Tokenize(s)...)
	}

	return words
}

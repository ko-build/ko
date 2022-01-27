// Package chunk implements functions for finding useful chunks in text previously tagged from parts of speech.
//
package chunk

import (
	"regexp"

	"github.com/jdkato/prose/tag"
)

// quadString creates a string containing all of the tags, each padded to 4
// characters wide.
func quadsString(tagged []tag.Token) string {
	tagQuads := ""
	for _, tok := range tagged {
		padding := ""
		pos := tok.Tag
		switch len(pos) {
		case 0:
			padding = "____" // should not exist
		case 1:
			padding = "___"
		case 2:
			padding = "__"
		case 3:
			padding = "_"
		case 4: // no padding required
		default:
			pos = pos[:4] // longer than 4 ... truncate!
		}
		tagQuads += pos + padding
	}
	return tagQuads
}

// TreebankNamedEntities matches proper names, excluding prior adjectives,
// possibly including numbers and a linkage by preposition or subordinating
// conjunctions (for example "Bank of England").
var TreebankNamedEntities = regexp.MustCompile(
	`((CD__)*(NNP.)+(CD__|NNP.)*)+` +
		`((IN__)*(CD__)*(NNP.)+(CD__|NNP.)*)*`)

// Chunk returns a slice containing the chunks of interest according to the
// regexp.
//
// This is a convenience wrapper around Locate, which should be used if you
// need access the to the in-text locations of each chunk.
func Chunk(tagged []tag.Token, rx *regexp.Regexp) []string {
	chunks := []string{}
	for _, loc := range Locate(tagged, rx) {
		res := ""
		for t, tt := range tagged[loc[0]:loc[1]] {
			if t != 0 {
				res += " "
			}
			res += tt.Text
		}
		chunks = append(chunks, res)
	}
	return chunks
}

// Locate finds the chunks of interest according to the regexp.
func Locate(tagged []tag.Token, rx *regexp.Regexp) [][]int {
	rx.Longest() // make sure we find the longest possible sequences
	rs := rx.FindAllStringIndex(quadsString(tagged), -1)
	for i, ii := range rs {
		for j := range ii {
			// quadsString makes every offset 4x what it should be
			rs[i][j] /= 4
		}
	}
	return rs
}

// Copyright (c) 2015 Eric Bower
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package tokenize

import (
	"regexp"
	"strings"

	"github.com/jdkato/prose/internal/util"
	"gopkg.in/neurosnap/sentences.v1"
	"gopkg.in/neurosnap/sentences.v1/data"
)

// PunktSentenceTokenizer is an extension of the Go implementation of the Punkt
// sentence tokenizer (https://github.com/neurosnap/sentences), with a few
// minor improvements (see https://github.com/neurosnap/sentences/pull/18).
type PunktSentenceTokenizer struct {
	tokenizer *sentences.DefaultSentenceTokenizer
}

// NewPunktSentenceTokenizer creates a new PunktSentenceTokenizer and loads
// its English model.
func NewPunktSentenceTokenizer() *PunktSentenceTokenizer {
	var pt PunktSentenceTokenizer
	var err error

	pt.tokenizer, err = newSentenceTokenizer(nil)
	util.CheckError(err)

	return &pt
}

// Tokenize splits text into sentences.
func (p PunktSentenceTokenizer) Tokenize(text string) []string {
	sents := []string{}
	for _, s := range p.tokenizer.Tokenize(text) {
		sents = append(sents, s.Text)
	}
	return sents
}

type wordTokenizer struct {
	sentences.DefaultWordTokenizer
}

var reAbbr = regexp.MustCompile(`((?:[\w]\.)+[\w]*\.)`)
var reLooksLikeEllipsis = regexp.MustCompile(`(?:\.\s?){2,}\.`)
var reEntities = regexp.MustCompile(`Yahoo!`)

// English customized sentence tokenizer.
func newSentenceTokenizer(s *sentences.Storage) (*sentences.DefaultSentenceTokenizer, error) {
	training := s

	if training == nil {
		b, err := data.Asset("data/english.json")
		if err != nil {
			return nil, err
		}

		training, err = sentences.LoadTraining(b)
		if err != nil {
			return nil, err
		}
	}

	// supervisor abbreviations
	abbrevs := []string{"sgt", "gov", "no", "mt"}
	for _, abbr := range abbrevs {
		training.AbbrevTypes.Add(abbr)
	}

	lang := sentences.NewPunctStrings()
	word := newWordTokenizer(lang)
	annotations := sentences.NewAnnotations(training, lang, word)

	ortho := &sentences.OrthoContext{
		Storage:      training,
		PunctStrings: lang,
		TokenType:    word,
		TokenFirst:   word,
	}

	multiPunct := &multiPunctWordAnnotation{
		Storage:      training,
		TokenParser:  word,
		TokenGrouper: &sentences.DefaultTokenGrouper{},
		Ortho:        ortho,
	}

	annotations = append(annotations, multiPunct)

	tokenizer := &sentences.DefaultSentenceTokenizer{
		Storage:       training,
		PunctStrings:  lang,
		WordTokenizer: word,
		Annotations:   annotations,
	}

	return tokenizer, nil
}

func newWordTokenizer(p sentences.PunctStrings) *wordTokenizer {
	word := &wordTokenizer{}
	word.PunctStrings = p

	return word
}

func (e *wordTokenizer) HasSentEndChars(t *sentences.Token) bool {
	enders := []string{
		`."`, `.)`, `.’`, `.”`,
		`?`, `?"`, `?'`, `?)`, `?’`, `?”`,
		`!`, `!"`, `!'`, `!)`, `!’`, `!”`,
	}

	for _, ender := range enders {
		if strings.HasSuffix(t.Tok, ender) && !reEntities.MatchString(t.Tok) {
			return true
		}
	}

	parens := []string{
		`.[`, `.(`, `."`,
		`?[`, `?(`,
		`![`, `!(`,
	}

	for _, paren := range parens {
		if strings.Contains(t.Tok, paren) {
			return true
		}
	}

	return false
}

// MultiPunctWordAnnotation attempts to tease out custom Abbreviations such as
// "F.B.I."
type multiPunctWordAnnotation struct {
	*sentences.Storage
	sentences.TokenParser
	sentences.TokenGrouper
	sentences.Ortho
}

func (a *multiPunctWordAnnotation) Annotate(tokens []*sentences.Token) []*sentences.Token {
	for _, tokPair := range a.TokenGrouper.Group(tokens) {
		if len(tokPair) < 2 || tokPair[1] == nil {
			tok := tokPair[0].Tok
			if strings.Contains(tok, "\n") && strings.Contains(tok, " ") {
				// We've mislabeled due to an errant newline.
				tokPair[0].SentBreak = false
			}
			continue
		}

		a.tokenAnnotation(tokPair[0], tokPair[1])
	}

	return tokens
}

// looksInternal determines if tok's punctuation could appear
// sentence-internally (i.e., parentheses or quotations).
func looksInternal(tok string) bool {
	internal := []string{")", `’`, `”`, `"`, `'`}
	for _, punc := range internal {
		if strings.HasSuffix(tok, punc) {
			return true
		}
	}
	return false
}

func (a *multiPunctWordAnnotation) tokenAnnotation(tokOne, tokTwo *sentences.Token) {
	// This is an expensive calculation, so we only want to do it once.
	var nextTyp string

	// If both tokOne and tokTwo and periods, we're probably in an ellipsis
	// that wasn't properly tokenized by `WordTokenizer`.
	if strings.HasSuffix(tokOne.Tok, ".") && tokTwo.Tok == "." {
		tokOne.SentBreak = false
		tokTwo.SentBreak = false
		return
	}

	isNonBreak := strings.HasSuffix(tokOne.Tok, ".") && !tokOne.SentBreak
	isEllipsis := reLooksLikeEllipsis.MatchString(tokOne.Tok)
	isInternal := tokOne.SentBreak && looksInternal(tokOne.Tok)

	if isNonBreak || isEllipsis || isInternal {
		nextTyp = a.TokenParser.TypeNoSentPeriod(tokTwo)
		isStarter := a.SentStarters[nextTyp]

		// If the tokOne looks like an ellipsis and tokTwo is either
		// capitalized or a frequent sentence starter, break the sentence.
		if isEllipsis {
			if a.TokenParser.FirstUpper(tokTwo) || isStarter != 0 {
				tokOne.SentBreak = true
				return
			}
		}

		// If the tokOne's sentence-breaking punctuation looks like it could
		// occur sentence-internally, ensure that the following word is either
		// capitalized or a frequent sentence starter.
		if isInternal {
			if a.TokenParser.FirstLower(tokTwo) && isStarter == 0 {
				tokOne.SentBreak = false
				return
			}
		}

		// If the tokOne ends with a period but isn't marked as a sentence
		// break, mark it if tokTwo is capitalized and can occur in _ORTHO_LC.
		if isNonBreak && a.TokenParser.FirstUpper(tokTwo) {
			if a.Storage.OrthoContext[nextTyp]&112 != 0 {
				tokOne.SentBreak = true
			}
		}
	}

	if len(reAbbr.FindAllString(tokOne.Tok, 1)) == 0 {
		return
	}

	if a.IsInitial(tokOne) {
		return
	}

	tokOne.Abbr = true
	tokOne.SentBreak = false

	// [4.1.1. Orthographic Heuristic] Check if there's
	// orthogrpahic evidence about whether the next word
	// starts a sentence or not.
	isSentStarter := a.Ortho.Heuristic(tokTwo)
	if isSentStarter == 1 {
		tokOne.SentBreak = true
		return
	}

	if nextTyp == "" {
		nextTyp = a.TokenParser.TypeNoSentPeriod(tokTwo)
	}

	// [4.1.3. Frequent Sentence Starter Heruistic] If the
	// next word is capitalized, and is a member of the
	// frequent-sentence-starters list, then label tok as a
	// sentence break.
	if a.TokenParser.FirstUpper(tokTwo) && a.SentStarters[nextTyp] != 0 {
		tokOne.SentBreak = true
		return
	}
}

package sentences

import (
	"strings"
)

/*
AnnotateTokens is an interface used for the sentence tokenizer to add properties to
any given token during tokenization.
*/
type AnnotateTokens interface {
	Annotate([]*Token) []*Token
}

/*
TypeBasedAnnotation performs the first pass of annotation, which makes decisions
based purely based on the word type of each word:
	* '?', '!', and '.' are marked as sentence breaks.
	* sequences of two or more periods are marked as ellipsis.
	* any word ending in '.' that's a known abbreviation is marked as an abbreviation.
	* any other word ending in '.' is marked as a sentence break.

Return these annotations as a tuple of three sets:
	* sentbreak_toks: The indices of all sentence breaks.
	* abbrev_toks: The indices of all abbreviations.
	* ellipsis_toks: The indices of all ellipsis marks.
*/
type TypeBasedAnnotation struct {
	*Storage
	PunctStrings
	TokenExistential
}

// NewTypeBasedAnnotation creates an instance of the TypeBasedAnnotation struct
func NewTypeBasedAnnotation(s *Storage, p PunctStrings, e TokenExistential) *TypeBasedAnnotation {
	return &TypeBasedAnnotation{
		Storage:          s,
		PunctStrings:     p,
		TokenExistential: e,
	}
}

// NewAnnotations is the default AnnotateTokens struct  that the tokenizer uses
func NewAnnotations(s *Storage, p PunctStrings, word WordTokenizer) []AnnotateTokens {
	return []AnnotateTokens{
		&TypeBasedAnnotation{s, p, word},
		&TokenBasedAnnotation{s, p, word, &DefaultTokenGrouper{}, &OrthoContext{
			s, p, word, word,
		}},
	}
}

// Annotate iterates over all tokens and applies the type annotation on them
func (a *TypeBasedAnnotation) Annotate(tokens []*Token) []*Token {
	for _, augTok := range tokens {
		a.typeAnnotation(augTok)
	}
	return tokens
}

func (a *TypeBasedAnnotation) typeAnnotation(token *Token) {
	chars := []rune(token.Tok)

	if a.HasSentEndChars(token) {
		token.SentBreak = true
	} else if a.HasPeriodFinal(token) && !strings.HasSuffix(token.Tok, "..") {
		tokNoPeriod := strings.ToLower(token.Tok[:len(chars)-1])
		tokNoPeriodHypen := strings.Split(tokNoPeriod, "-")
		tokLastHyphEl := string(tokNoPeriodHypen[len(tokNoPeriodHypen)-1])

		if a.IsAbbr(tokNoPeriod, tokLastHyphEl) {
			token.Abbr = true
		} else {
			token.SentBreak = true
		}
	}
}

/*
TokenBasedAnnotation performs a token-based classification (section 4) over the given
tokens, making use of the orthographic heuristic (4.1.1), collocation
heuristic (4.1.2) and frequent sentence starter heuristic (4.1.3).
*/
type TokenBasedAnnotation struct {
	*Storage
	PunctStrings
	TokenParser
	TokenGrouper
	Ortho
}

// Annotate iterates groups tokens in pairs of two and then iterates over them to apply token annotation
func (a *TokenBasedAnnotation) Annotate(tokens []*Token) []*Token {
	for _, tokPair := range a.TokenGrouper.Group(tokens) {
		a.tokenAnnotation(tokPair[0], tokPair[1])
	}

	return tokens
}

func (a *TokenBasedAnnotation) tokenAnnotation(tokOne, tokTwo *Token) {
	if tokTwo == nil {
		return
	}

	if !a.TokenParser.HasPeriodFinal(tokOne) {
		return
	}

	typ := a.TokenParser.TypeNoPeriod(tokOne)
	nextTyp := a.TokenParser.TypeNoSentPeriod(tokTwo)
	tokIsInitial := a.TokenParser.IsInitial(tokOne)

	/*
	   [4.1.2. Collocation Heuristic] If there's a
	   collocation between the word before and after the
	   period, then label tok as an abbreviation and NOT
	   a sentence break. Note that collocations with
	   frequent sentence starters as their second word are
	   excluded in training.
	*/
	collocation := strings.Join([]string{typ, nextTyp}, ",")
	if a.Collocations[collocation] != 0 {
		tokOne.SentBreak = false
		tokOne.Abbr = true
		return
	}

	/*
		[4.2. Token-Based Reclassification of Abbreviations] If
		the token is an abbreviation or an ellipsis, then decide
		whether we should *also* classify it as a sentbreak.
	*/
	if (tokOne.Abbr || a.TokenParser.IsEllipsis(tokOne)) && !tokIsInitial {
		/*
			[4.1.1. Orthographic Heuristic] Check if there's
			orthogrpahic evidence about whether the next word
			starts a sentence or not.
		*/
		isSentStarter := a.Ortho.Heuristic(tokTwo)
		if isSentStarter == 1 {
			tokOne.SentBreak = true
			return
		}

		/*
			[4.1.3. Frequent Sentence Starter Heruistic] If the
			next word is capitalized, and is a member of the
			frequent-sentence-starters list, then label tok as a
			sentence break.
		*/
		if a.TokenParser.FirstUpper(tokTwo) && a.SentStarters[nextTyp] != 0 {
			tokOne.SentBreak = true
			return
		}
	}

	/*
		Sometimes there are two consecutive tokens with a lone "."
		which probably means it is part of a spaced ellipsis ". . ."
		so set those tokens and not sentence breaks
	*/
	if tokOne.Tok == "." && tokTwo.Tok == "." {
		tokOne.SentBreak = false
		tokTwo.SentBreak = false
		return
	}

	/*
		[4.3. Token-Based Detection of Initials and Ordinals]
		Check if any initials or ordinals tokens that are marked
		as sentbreaks should be reclassified as abbreviations.
	*/
	if tokIsInitial || typ == "##number##" {
		isSentStarter := a.Ortho.Heuristic(tokTwo)

		if isSentStarter == 0 {
			tokOne.SentBreak = false
			tokOne.Abbr = true
			return
		}

		/*
			Special heuristic for initials: if orthogrpahic
			heuristc is unknown, and next word is always
			capitalized, then mark as abbrev (eg: J. Bach).
		*/
		if isSentStarter == -1 &&
			tokIsInitial &&
			a.TokenParser.FirstUpper(tokTwo) &&
			a.OrthoContext[nextTyp]&orthoLc == 0 {

			tokOne.SentBreak = false
			tokOne.Abbr = true
			return
		}
	}
}

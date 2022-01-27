package sentences

// PunctStrings implements all the functions necessary for punctuation strings.
// They are used to detect punctuation in the sentence
// tokenizer.
type PunctStrings interface {
	NonPunct() string
	Punctuation() string
	HasSentencePunct(string) bool
}

// DefaultPunctStrings are used to detect punctuation in the sentence
// tokenizer.
type DefaultPunctStrings struct{}

// NewPunctStrings creates a default set of properties
func NewPunctStrings() *DefaultPunctStrings {
	return &DefaultPunctStrings{}
}

// NonPunct regex string to detect non-punctuation.
func (p *DefaultPunctStrings) NonPunct() string {
	return `[^\W\d]`
}

// Punctuation characters
func (p *DefaultPunctStrings) Punctuation() string {
	return ";:,.!?"
}

// HasSentencePunct does the supplied text have a known sentence punctuation character?
func (p *DefaultPunctStrings) HasSentencePunct(text string) bool {
	endPunct := `.!?`
	for _, char := range endPunct {
		for _, achar := range text {
			if char == achar {
				return true
			}
		}
	}

	return false
}

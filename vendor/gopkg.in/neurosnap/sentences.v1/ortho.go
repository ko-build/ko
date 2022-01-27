package sentences

/*
The following constants are used to describe the orthographic
contexts in which a word can occur.  BEG=beginning, MID=middle,
UNK=unknown, UC=uppercase, LC=lowercase, NC=no case.
*/
const (
	// Beginning of a sentence with upper case.
	orthoBegUc = 1 << 1
	// Middle of a sentence with upper case.
	orthoMidUc = 1 << 2
	// Unknown position in a sentence with upper case.
	orthoUnkUc = 1 << 3
	// Beginning of a sentence with lower case.
	orthoBegLc = 1 << 4
	// Middle of a sentence with lower case.
	orthoMidLc = 1 << 5
	// Unknown position in a sentence with lower case.
	orthoUnkLc = 1 << 6
	// Occurs with upper case.
	orthoUc = orthoBegUc + orthoMidUc + orthoUnkUc
	// Occurs with lower case.
	orthoLc = orthoBegLc + orthoMidLc + orthoUnkLc
)

/*
A map from context position and first-letter case to the
appropriate orthographic context flag.
*/
var orthoMap = map[[2]string]int{
	[2]string{"initial", "upper"}:  orthoBegUc,
	[2]string{"internal", "upper"}: orthoMidUc,
	[2]string{"unknown", "upper"}:  orthoUnkUc,
	[2]string{"initial", "lower"}:  orthoBegLc,
	[2]string{"internal", "lower"}: orthoMidLc,
	[2]string{"unknown", "lower"}:  orthoUnkLc,
}

// Ortho creates a promise for structs to implement an orthogonal heuristic
// method.
type Ortho interface {
	Heuristic(*Token) int
}

// OrthoContext determines whether a token is capitalized, sentence starter, etc.
type OrthoContext struct {
	*Storage
	PunctStrings
	TokenType
	TokenFirst
}

/*
Heuristic decides whether the given token is the first token in a sentence.
*/
func (o *OrthoContext) Heuristic(token *Token) int {
	if token == nil {
		return 0
	}

	for _, punct := range o.PunctStrings.Punctuation() {
		if token.Tok == string(punct) {
			return 0
		}
	}

	orthoCtx := o.Storage.OrthoContext[o.TokenType.TypeNoSentPeriod(token)]
	/*
	   If the word is capitalized, occurs at least once with a
	   lower case first letter, and never occurs with an upper case
	   first letter sentence-internally, then it's a sentence starter.
	*/
	if o.TokenFirst.FirstUpper(token) && (orthoCtx&orthoLc > 0 && orthoCtx&orthoMidUc == 0) {
		return 1
	}

	/*
		If the word is lower case, and either (a) we've seen it used
		with upper case, or (b) we've never seen it used
		sentence-initially with lower case, then it's not a sentence
		starter.
	*/
	if o.TokenFirst.FirstLower(token) && (orthoCtx&orthoUc > 0 || orthoCtx&orthoBegLc == 0) {
		return 0
	}

	return -1
}

package normalize

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	lineEndingsRe = regexp.MustCompile(`\r\n?`)
	// 3.1.1 All whitespace should be treated as a single blank space.
	whitespaceRe         = regexp.MustCompile(`[ \t\f\r             　​]+`)
	trailingWhitespaceRe = regexp.MustCompile(`(?m)[ \t\f\r             　​]$`)
	licenseHeaderRe      = regexp.MustCompile(`(licen[cs]e)\.?\n\n`)
	leadingWhitespaceRe  = regexp.MustCompile(`(?m)^(( \n?)|\n)`)
	// 5.1.2 Hyphens, Dashes  Any hyphen, dash, en dash, em dash, or other variation should be
	// considered equivalent.
	punctuationRe = regexp.MustCompile(`[-‒–—―⁓⸺⸻~˗‐‑⁃⁻₋−∼⎯⏤─➖𐆑֊﹘﹣－]+`)
	// 5.1.3 Quotes  Any variation of quotations (single, double, curly, etc.) should be considered
	// equivalent.
	quotesRe = regexp.MustCompile(`["'“”‘’„‚«»‹›❛❜❝❞\x60]+`)
	// 7.1.1 Where a line starts with a bullet, number, letter, or some form of a list item
	// (determined where list item is followed by a space, then the text of the sentence), ignore
	// the list item for matching purposes.
	bulletRe = regexp.MustCompile(`(?m)^(([-*✱﹡•●⚫⏺🞄∙⋅])|([(\[{]?\d+[.)\]}] ?)|([(\[{]?[a-z][.)\]}] ?)|([(\[{]?i+[.)\]} ] ?))`)
	// 8.1.1 The words in the following columns are considered equivalent and interchangeable.
	wordReplacer = strings.NewReplacer(
		"acknowledgment", "acknowledgement",
		"analogue", "analog",
		"analyse", "analyze",
		"artefact", "artifact",
		"authorisation", "authorization",
		"authorised", "authorized",
		"calibre", "caliber",
		"cancelled", "canceled",
		"capitalisations", "capitalizations",
		"catalogue", "catalog",
		"categorise", "categorize",
		"centre", "center",
		"emphasised", "emphasized",
		"favour", "favor",
		"favourite", "favorite",
		"fulfil", "fulfill",
		"fulfilment", "fulfillment",
		"initialise", "initialize",
		"judgment", "judgement",
		"labelling", "labeling",
		"labour", "labor",
		"licence", "license",
		"maximise", "maximize",
		"modelled", "modeled",
		"modelling", "modeling",
		"offence", "offense",
		"optimise", "optimize",
		"organisation", "organization",
		"organise", "organize",
		"practise", "practice",
		"programme", "program",
		"realise", "realize",
		"recognise", "recognize",
		"signalling", "signaling",
		"sub-license", "sublicense",
		"sub license", "sub-license",
		"utilisation", "utilization",
		"whilst", "while",
		"wilful", "wilfull",
		"non-commercial", "noncommercial",
		"per cent", "percent",
		"copyright owner", "copyright",
	)

	// 9.1.1 "©", "(c)", or "Copyright" should be considered equivalent and interchangeable.
	copyrightRe = regexp.MustCompile(`copyright|\(c\)`)
	trademarkRe = regexp.MustCompile(`trademark(s?)|\(tm\)`)

	// extra cleanup
	brokenLinkRe    = regexp.MustCompile(`http s ://`)
	urlCleanupRe    = regexp.MustCompile(`[<(](http(s?)://[^\s]+)[)>]`)
	copyrightLineRe = regexp.MustCompile(`(?m)^((©.*)|(all rights reserved(\.)?)|(li[cs]en[cs]e))\n`)
	nonAlphaNumRe   = regexp.MustCompile(`[^- \na-z0-9]`)

	// used in Split()
	splitRe = regexp.MustCompile(`\n\s*[^a-zA-Z0-9_,()]{3,}\s*\n`)
)

// Strictness represents the aggressiveness of the performed normalization. The bigger the number,
// the more aggressive. See `Enforced`, `Moderate` and `Relaxed`.
type Strictness int

const (
	// Enforced is the strictest mode - only the official SPDX guidelines are applied.
	Enforced Strictness = 0
	// Moderate is equivalent to Enforced with some additional normalization: dots are removed, copyright lines too.
	Moderate Strictness = 1
	// Relaxed is the most powerful normalization, Moderate + Unicode normalization and all non-alphanumeric chars removed.
	Relaxed Strictness = 2
)

// LicenseText makes a license text ready for analysis.
// It follows SPDX guidelines at
// https://spdx.org/spdx-license-list/matching-guidelines
func LicenseText(text string, strictness Strictness) string {
	// Line endings
	text = lineEndingsRe.ReplaceAllString(text, "\n")

	// 4. Capitalization
	text = strings.ToLower(text)

	// 3. Whitespace
	text = whitespaceRe.ReplaceAllString(text, " ")
	text = trailingWhitespaceRe.ReplaceAllString(text, "")
	text = licenseHeaderRe.ReplaceAllString(text, "$1\nthisislikelyalicenseheaderplaceholder\n")
	text = leadingWhitespaceRe.ReplaceAllString(text, "")

	// 5. Punctuation
	text = punctuationRe.ReplaceAllString(text, "-")
	text = quotesRe.ReplaceAllString(text, "\"")

	// 7. Bullets and Numbering
	text = bulletRe.ReplaceAllString(text, "")

	// 8. Varietal Word Spelling
	text = wordReplacer.Replace(text)

	// 9. Copyright Symbol
	text = copyrightRe.ReplaceAllString(text, "©")
	text = trademarkRe.ReplaceAllString(text, "™")

	// fix broken URLs in SPDX source texts
	text = brokenLinkRe.ReplaceAllString(text, "https://")

	// fix URLs in <> - erase the decoration
	text = urlCleanupRe.ReplaceAllString(text, "$1")

	// collapse several non-alphanumeric characters
	{
		buffer := &bytes.Buffer{}
		back := '\x00'
		for _, char := range text {
			if !unicode.IsLetter(char) && !unicode.IsDigit(char) && back == char {
				continue
			}
			back = char
			buffer.WriteRune(char)
		}
		text = buffer.String()
	}

	if strictness > Enforced {
		// there are common mismatches because of trailing dots
		text = strings.Replace(text, ".", "", -1)
		// usually copyright lines are custom and occur multiple times
		text = copyrightLineRe.ReplaceAllString(text, "")
	}

	if strictness > Moderate {
		return Relax(text)
	}

	text = leadingWhitespaceRe.ReplaceAllString(text, "")
	text = strings.Replace(text, "thisislikelyalicenseheaderplaceholder", "", -1)

	return text
}

// Relax applies very aggressive normalization rules to text.
func Relax(text string) string {
	buffer := &bytes.Buffer{}
	writer := transform.NewWriter(
		buffer, transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC))
	_, _ = writer.Write([]byte(text))
	_ = writer.Close()
	text = buffer.String()
	text = nonAlphaNumRe.ReplaceAllString(text, "")
	text = leadingWhitespaceRe.ReplaceAllString(text, "")
	text = strings.Replace(text, "  ", " ", -1)
	return text
}

// Split applies heuristics to split the text into several parts
func Split(text string) []string {
	result := []string{text}

	// Always add the full text
	splitted := splitRe.Split(text, -1)
	if len(splitted) > 1 {
		result = append(result, splitted...)
	}
	return result
}

// The MIT License (MIT)
//
// Copyright (c) 2015 Kevin S. Dias
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package tokenize

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jdkato/prose/internal/util"
)

/* Public API */

// PragmaticSegmenter is a multilingual, rule-based sentence boundary detector.
//
// This is a port of the Ruby library by Kevin S. Dias
// (https://github.com/diasks2/pragmatic_segmenter).
type PragmaticSegmenter struct {
	processor languageProcessor
}

// NewPragmaticSegmenter creates a new PragmaticSegmenter according to the
// specified language. If the given language is not supported, an error will be
// returned.
//
// Languages are specified by their two-character ISO 639-1 code. The supported
// languages are "en" (English), "es" (Spanish), "fr" (French) ... (WIP)
func NewPragmaticSegmenter(lang string) (*PragmaticSegmenter, error) {
	if p, ok := langToProcessor[lang]; ok {
		return &PragmaticSegmenter{processor: p}, nil
	}
	return nil, errors.New("unknown language")
}

// Tokenize splits text into sentences.
func (p *PragmaticSegmenter) Tokenize(text string) []string {
	return p.processor.process(text)
}

/* Helper functions, regexps, and types */

// A rule associates a regular expression with a replacement string.
type rule struct {
	pattern     *regexp.Regexp
	replacement string
}

// sub replaces all occurrences of Pattern with Replacement.
func (r *rule) sub(text string) string {
	if !r.pattern.MatchString(text) {
		return text
	}

	orig := len(text)
	diff := 0
	for _, submat := range r.pattern.FindAllStringSubmatchIndex(text, -1) {
		for idx, mat := range submat {
			if mat != -1 && idx > 0 && idx%2 == 0 {
				loc := []int{mat - diff, submat[idx+1] - diff}
				text = text[:loc[0]] + r.replacement + text[loc[1]:]
				diff = orig - len(text)
			}
		}
	}

	return text
}

// numbers

var periodBeforeNumberRule = rule{
	pattern: regexp.MustCompile(`(\.)\d`), replacement: "∯"}
var numberAfterPeriodBeforeLetterRule = rule{
	pattern: regexp.MustCompile(`\d(\.)\S`), replacement: "∯"}
var newLineNumberPeriodSpaceLetterRule = rule{
	pattern: regexp.MustCompile(`[\n\r]\d(\.)(?:[\s\S]|\))`), replacement: "∯"}
var startLineNumberPeriodRule = rule{
	pattern: regexp.MustCompile(`^\d(\.)(?:[\s\S]|\))`), replacement: "∯"}
var startLineTwoDigitNumberPeriodRule = rule{
	pattern: regexp.MustCompile(`^\d\d(\.)(?:[\s\S]|\))`), replacement: "∯"}
var allNumberRules = []rule{
	periodBeforeNumberRule, numberAfterPeriodBeforeLetterRule,
	newLineNumberPeriodSpaceLetterRule, startLineNumberPeriodRule,
	startLineTwoDigitNumberPeriodRule,
}

// common

var cleanRules = []rule{
	{pattern: regexp.MustCompile(`[^\n]\s(\n)\S`), replacement: ""},
	{pattern: regexp.MustCompile(`(\n)[a-z]`), replacement: " "},
}
var exclamationWordsRE = regexp.MustCompile(
	`\s(?:!Xũ|!Kung|ǃʼOǃKung|!Xuun|!Kung-Ekoka|ǃHu|` +
		`ǃKhung|ǃKu|ǃung|ǃXo|ǃXû|ǃXung|ǃXũ|!Xun|Yahoo!|Y!J|Yum!)\s`)
var sentenceBoundaryRE = regexp.MustCompile(
	`\x{ff08}(?:[^\x{ff09}])*\x{ff09}(\s?[A-Z])|` +
		`\x{300c}(?:[^\x{300d}])*\x{300d}(\s[A-Z])|` +
		`\((?:[^\)]){2,}\)(\s[A-Z])|` +
		`'(?:[^'])*[^,]'(\s[A-Z])|` +
		`"(?:[^"])*[^,]"(\s[A-Z])|` +
		`“(?:[^”])*[^,]”(\s[A-Z])|` +
		`\S.*?[。．.！!?？ȸȹ☉☈☇☄]`)
var quotationAtEndOfSentenceRE = regexp.MustCompile(
	`[!?\.-][\"\'\x{201d}\x{201c}]\s{1}[A-Z]`)
var splitSpaceQuotationAtEndOfSentenceRE = regexp.MustCompile(
	`[!?\.-][\"\'\x{201d}\x{201c}](\s{1})[A-Z]`) // lookahead
var continuousPunctuationRE = regexp.MustCompile(`\S(!|\?){3,}(?:\s|\z|$)`)
var possessiveAbbreviationRule = rule{
	pattern: regexp.MustCompile(`(\.)'s\s|(\.)'s$|(\.)'s\z`), replacement: "∯"}
var kommanditgesellschaftRule = rule{
	pattern: regexp.MustCompile(`Co(\.)\sKG`), replacement: "∯"}
var multiPeriodAbbrevRE = regexp.MustCompile(`(?i)\b[a-z](?:\.[a-z])+[.]`)

// var parensBetweenDoubleQuotesRE = regexp.MustCompile(`["”]\s\(.*\)\s["“]`)
// var betweenDoubleQuotesRE2 = regexp.MustCompile(`(?:[^"])*[^,]"|“(?:[^”])*[^,]”`)
// var wordWithLeadingApostropheRE = regexp.MustCompile(`\s'(?:[^']|'[a-zA-Z])*'\S`)

// AM/PM
var upperCasePmRule = rule{
	pattern: regexp.MustCompile(`P∯M(∯)\s[A-Z]`), replacement: "."}
var upperCaseAmRule = rule{
	pattern: regexp.MustCompile(`A∯M(∯)\s[A-Z]`), replacement: "."}
var lowerCasePmRule = rule{
	pattern: regexp.MustCompile(`p∯m(∯)\s[A-Z]`), replacement: "."}
var lowerCaseAmRule = rule{
	pattern: regexp.MustCompile(`a∯m(∯)\s[A-Z]`), replacement: "."}
var allAmPmRules = []rule{
	upperCasePmRule, upperCaseAmRule, lowerCasePmRule, lowerCaseAmRule}

// Searches for periods within an abbreviation and replaces the periods.
var singleUpperCaseLetterAtStartOfLineRule = rule{
	pattern: regexp.MustCompile(`^[A-Z](\.)\s`), replacement: "∯"}
var singleUpperCaseLetterRule = rule{
	pattern: regexp.MustCompile(`\s[A-Z](\.)\s`), replacement: "∯"}
var allSingleUpperCaseLetterRules = []rule{
	singleUpperCaseLetterAtStartOfLineRule, singleUpperCaseLetterRule}

// Searches for ellipses within a string and replaces the periods.
var threeConsecutiveRule = rule{
	pattern: regexp.MustCompile(`[^.](\.\.\.)\s+[A-Z]`), replacement: "☏."}
var fourConsecutiveRule = rule{
	pattern: regexp.MustCompile(`\S(\.{3})\.\s[A-Z]`), replacement: "ƪ"}
var threeSpaceRule = rule{
	pattern: regexp.MustCompile(`((?:\s\.){3}\s)`), replacement: "♟"}
var fourSpaceRule = rule{
	pattern: regexp.MustCompile(`[a-z]((?:\.\s){3}\.(?:\z|$|\n))`), replacement: "♝"}
var otherThreePeriodRule = rule{pattern: regexp.MustCompile(`(\.\.\.)`), replacement: "ƪ"}
var allEllipsesRules = []rule{
	threeConsecutiveRule, fourConsecutiveRule, threeSpaceRule, fourSpaceRule,
	otherThreePeriodRule}

// between_punctuation
var betweenSingleQuotesRE = regexp.MustCompile(`\s'(?:[^']|'[a-zA-Z])*'`)
var betweenDoubleQuotesRE = regexp.MustCompile(`"([^"\\]+|\\{2}|\\.)*"`)
var betweenArrowQuotesRE = regexp.MustCompile(`«([^»\\]+|\\{2}|\\.)*»`)
var betweenSmartQuotesRE = regexp.MustCompile(`“([^”\\]+|\\{2}|\\.)*”`)
var betweenSquareBracketsRE = regexp.MustCompile(`\[([^\]\\]+|\\{2}|\\.)*\]`)
var betweenParensRE = regexp.MustCompile(`\(([^\(\)\\]+|\\{2}|\\.)*\)`)

// subPat replaces all punctuation in the strings that match the regexp pat.
func subPat(text, mtype string, pat *regexp.Regexp) string {
	canidates := []string{}
	for _, s := range pat.FindAllString(text, -1) {
		canidates = append(canidates, strings.TrimSpace(s))
	}
	r := punctuationReplacer{
		matches: canidates, text: text, matchType: mtype}
	return r.replace()
}

// replaceBetweenQuotes replaces punctuation inside quotes.
func replaceBetweenQuotes(text string) string {
	text = subPat(text, "single", betweenSingleQuotesRE)
	text = subPat(text, "double", betweenDoubleQuotesRE)
	text = subPat(text, "double", betweenSquareBracketsRE)
	text = subPat(text, "double", betweenParensRE)
	text = subPat(text, "double", betweenArrowQuotesRE)
	text = subPat(text, "double", betweenSmartQuotesRE)
	return text
}

// applyRules applies each rule in []rules to text.
func applyRules(text string, rules []rule) string {
	for _, rule := range rules {
		text = rule.sub(text)
	}
	return text
}

// substitute replaces the substring sub with the string repl.
func substitute(src, sub, repl string) string {
	idx := strings.Index(src, sub)
	for idx >= 0 {
		src = src[:idx] + repl + src[idx+len(sub):]
		idx = strings.Index(src, sub)
	}
	return src
}

// escape
var escapeRegexReservedCharacters = strings.NewReplacer(
	`(`, `\(`, `)`, `\)`, `[`, `\[`, `]`, `\]`, `-`, `\-`,
)

// unescape
var subEscapeRegexReservedCharacters = strings.NewReplacer(
	`\(`, `(`, `\)`, `)`, `\[`, `[`, `\]`, `]`, `\-`, `-`,
)

/* punctuation_replacer */

type punctuationReplacer struct {
	matches   []string
	text      string
	matchType string
}

func (r *punctuationReplacer) replace() string {
	return r.replacePunctuation(r.matches)
}

func (r *punctuationReplacer) replacePunctuation(matches []string) string {
	r.text = escapeRegexReservedCharacters.Replace(r.text)
	for _, m := range matches {
		m = escapeRegexReservedCharacters.Replace(m)

		s := r.sub(m, ".", "∯")
		sub1 := r.sub(s, "。", "&ᓰ&")
		sub2 := r.sub(sub1, "．", "&ᓱ&")
		sub3 := r.sub(sub2, "！", "&ᓳ&")
		sub4 := r.sub(sub3, "!", "&ᓴ&")
		sub5 := r.sub(sub4, "?", "&ᓷ&")
		sub6 := r.sub(sub5, "? ", "&ᓸ&")
		if r.matchType != "single" {
			r.sub(sub6, "'", "&⎋&")
		}
	}
	return subEscapeRegexReservedCharacters.Replace(r.text)
}

func (r *punctuationReplacer) sub(content, a, b string) string {
	repl := substitute(content, a, b)
	r.text = strings.Replace(r.text, content, repl, -1)
	return repl
}

/* abbreviation_replacer */

type abbreviationReplacer struct {
	definition       languageDefinition
	boundaries       *rule
	prepositiveCache map[string][]rule
	numberCache      map[string][]rule
	periodCache      map[string][]rule
	searchCache      map[string][]*regexp.Regexp
}

func newAbbreviationReplacer(lang string) *abbreviationReplacer {
	var def languageDefinition
	var bounds *rule

	if d, ok := langToDefinition[lang]; ok {
		def = d
	} else {
		def = new(commonDefinition)
	}
	regex := ""
	for _, word := range def.starters() {
		esc := regexp.QuoteMeta(word)
		regex += fmt.Sprintf(`U∯S(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`U\.S(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`U∯K(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`U\.K(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`E∯U(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`E\.U(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`U∯S∯A(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`U\.S\.A(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`I(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`i\.v(∯)\s%s\s|`, esc)
		regex += fmt.Sprintf(`I\.V(∯)\s%s\s|`, esc)
	}

	if regex != "" {
		r := regexp.MustCompile(strings.TrimRight(regex, "|"))
		bounds = &rule{pattern: r, replacement: "."}
	}

	return &abbreviationReplacer{definition: def, boundaries: bounds,
		prepositiveCache: make(map[string][]rule),
		numberCache:      make(map[string][]rule),
		periodCache:      make(map[string][]rule),
		searchCache:      make(map[string][]*regexp.Regexp)}
}

func (r *abbreviationReplacer) replace(text string) string {
	text = possessiveAbbreviationRule.sub(text)
	text = kommanditgesellschaftRule.sub(text)
	text = applyRules(text, allSingleUpperCaseLetterRules)

	text = r.search(text, r.definition.abbreviations()["abbreviations"])
	text = r.replaceMultiPeriods(text)

	for _, rule := range allAmPmRules {
		text = rule.sub(text)
	}

	return r.replaceBoundary(text)
}

func (r *abbreviationReplacer) search(query string, list []string) string {
	var match, next *regexp.Regexp

	downcased := strings.ToLower(query)
	for _, abbr := range list {
		if !strings.Contains(downcased, strings.TrimSpace(abbr)) {
			continue
		}

		text := query
		esc := regexp.QuoteMeta(abbr)
		if data, ok := r.searchCache[esc]; ok {
			match, next = data[0], data[1]
		} else {
			match = regexp.MustCompile(`(?i)(?:^|\s|\r|\n)` + esc)
			next = regexp.MustCompile(fmt.Sprintf(`%s (.{1})`, esc))
			r.searchCache[esc] = []*regexp.Regexp{match, next}
		}

		found := match.FindAllStringSubmatch(text, -1)
		if len(found) > 0 {
			chars := next.FindAllString(query, -1)
			for i, am := range found {
				query = r.scan(query, am[0], i, chars)
			}
		}
	}
	return query
}

func (r *abbreviationReplacer) scan(text, am string, idx int, chars []string) string {
	character := ""
	if len(chars) > idx {
		character = chars[idx]
	}
	prepositive := r.definition.abbreviations()["prepositive"]
	number := r.definition.abbreviations()["number"]
	upper := character != "" && character == strings.ToUpper(character)
	clean := strings.TrimSpace(strings.ToLower(am))
	prep := util.StringInSlice(clean, prepositive)
	if !upper || prep {
		if prep {
			text = r.replacePrepositive(text, am)
		} else if util.StringInSlice(clean, number) {
			text = r.replaceNumber(text, am)
		} else {
			text = r.replacePeriod(text, am)
		}
	}
	return text
}

func (r *abbreviationReplacer) replacePrepositive(text, abbr string) string {
	abbr = strings.ToLower(strings.TrimSpace(abbr))
	if rules, ok := r.prepositiveCache[abbr]; ok {
		return applyRules(text, rules)
	}
	q1 := fmt.Sprintf(`(?i)\s%s(\.)\s|^%s(\.)\s`, abbr, abbr)
	q2 := fmt.Sprintf(`(?i)\s%s(\.):\d+|^%s(\.):\d+`, abbr, abbr)
	r1 := rule{pattern: regexp.MustCompile(q1), replacement: "∯"}
	r2 := rule{pattern: regexp.MustCompile(q2), replacement: "∯"}
	r.prepositiveCache[abbr] = []rule{r1, r2}
	return r2.sub(r1.sub(text))
}

func (r *abbreviationReplacer) replaceNumber(text, abbr string) string {
	abbr = strings.ToLower(strings.TrimSpace(abbr))
	if rules, ok := r.numberCache[abbr]; ok {
		return applyRules(text, rules)
	}
	q1 := fmt.Sprintf(`(?i)\s%s(\.)\s\d|^%s(\.)\s\d`, abbr, abbr)
	q2 := fmt.Sprintf(`(?i)\s%s(\.)\s+\(|^%s(\.)\s+\(`, abbr, abbr)
	r1 := rule{pattern: regexp.MustCompile(q1), replacement: "∯"}
	r2 := rule{pattern: regexp.MustCompile(q2), replacement: "∯"}
	r.numberCache[abbr] = []rule{r1, r2}
	return r2.sub(r1.sub(text))
}

func (r *abbreviationReplacer) replacePeriod(text, abbr string) string {
	abbr = strings.TrimSpace(abbr)
	if rules, ok := r.periodCache[abbr]; ok {
		return applyRules(text, rules)
	}
	q1 := fmt.Sprintf(`\s%s(\.)(?:(?:(?:\.|\:|-|\?)|(?:\s(?:[a-z]|I\s|I'm|I'll|\d))))|^%s(\.)(?:(?:(?:\.|\:|\?)|(?:\s(?:[a-z]|I\s|I'm|I'll|\d))))`, abbr, abbr)
	q2 := fmt.Sprintf(`\s%s(\.),|^%s(\.),`, abbr, abbr)
	r1 := rule{pattern: regexp.MustCompile(q1), replacement: "∯"}
	r2 := rule{pattern: regexp.MustCompile(q2), replacement: "∯"}
	r.periodCache[abbr] = []rule{r1, r2}
	return r2.sub(r1.sub(text))
}

func (r *abbreviationReplacer) replaceBoundary(text string) string {
	if r.boundaries != nil {
		return r.boundaries.sub(text)
	}
	return text
}

func (r *abbreviationReplacer) replaceMultiPeriods(text string) string {
	for _, r := range multiPeriodAbbrevRE.FindAllString(text, -1) {
		text = substitute(text, r, substitute(r, ".", "∯"))
	}
	return text
}

/* language definitions */

var langToDefinition = map[string]languageDefinition{
	"fr": new(frenchDefinition),
	"es": new(spanishDefinition),
}

type languageDefinition interface {
	punctuation() []string
	abbreviations() map[string][]string
	punctRules() map[string]*rule
	doublePunctRules() []rule
	exclamationRules() []rule
	subRules() []rule
	subEllipsis() []rule
	starters() []string
}

type commonDefinition struct{}

func (d *commonDefinition) subEllipsis() []rule {
	return []rule{
		{pattern: regexp.MustCompile(`(ƪ)`), replacement: "..."},
		{pattern: regexp.MustCompile(`(♟)`), replacement: " . . . "},
		{pattern: regexp.MustCompile(`(♝)`), replacement: ". . . ."},
		{pattern: regexp.MustCompile(`(☏)`), replacement: ".."},
		{pattern: regexp.MustCompile(`(∮)`), replacement: "."},
	}
}

func (d *commonDefinition) punctuation() []string {
	return []string{"。", "．", ".", "！", "!", "?", "？"}
}

func (d *commonDefinition) abbreviations() map[string][]string {
	return map[string][]string{
		"abbreviations": {
			"adj", "adm", "adv", "al", "ala", "alta", "apr", "arc", "ariz", "ark",
			"art", "assn", "asst", "attys", "aug", "ave", "bart", "bld", "bldg",
			"blvd", "brig", "bros", "btw", "cal", "calif", "capt", "cl", "cmdr",
			"co", "col", "colo", "comdr", "con", "conn", "corp", "cpl", "cres", "ct",
			"d.phil", "dak", "dec", "del", "dept", "det", "dist", "dr", "dr.phil",
			"dr.philos", "drs", "e.g", "ens", "esp", "esq", "etc", "exp", "expy",
			"ext", "feb", "fed", "fla", "ft", "fwy", "fy", "ga", "gen", "gov", "hon",
			"hosp", "hr", "hway", "hwy", "i.e", "ia", "id", "ida", "ill", "inc",
			"ind", "ing", "insp", "is", "jan", "jr", "jul", "jun", "kan", "kans",
			"ken", "ky", "la", "lt", "ltd", "maj", "man", "mar", "mass", "may", "md",
			"me", "med", "messrs", "mex", "mfg", "mich", "min", "minn", "miss", "mlle",
			"mm", "mme", "mo", "mont", "mr", "mrs", "ms", "msgr", "mssrs", "mt", "mtn",
			"neb", "nebr", "nev", "no", "nos", "nov", "nr", "oct", "ok", "okla", "ont",
			"op", "ord", "ore", "p", "pa", "pd", "pde", "penn", "penna", "pfc", "ph",
			"ph.d", "pl", "plz", "pp", "prof", "pvt", "que", "rd", "ref", "rep",
			"reps", "res", "rev", "rt", "sask", "sec", "sen", "sens", "sep", "sept",
			"sfc", "sgt", "sr", "st", "supt", "surg", "tce", "tenn", "tex", "univ",
			"usafa", "u.s", "ut", "va", "v", "ver", "vs", "vt", "wash", "wis", "wisc",
			"wy", "wyo", "yuk"},
		"prepositive": {
			"adm", "attys", "brig", "capt", "cmdr", "col", "cpl", "det", "dr",
			"gen", "gov", "ing", "lt", "maj", "mr", "mrs", "ms", "mt", "messrs",
			"mssrs", "prof", "ph", "rep", "reps", "rev", "sen", "sens", "sgt",
			"st", "supt", "v", "vs"},
		"number": {"art", "ext", "no", "nos", "p", "pp"},
	}
}

func (d *commonDefinition) punctRules() map[string]*rule {
	return map[string]*rule{
		"withMultiplePeriodsAndEmail": {
			pattern: regexp.MustCompile(`\w(\.)\w`), replacement: "∮"},
		"geoLocation": {pattern: regexp.MustCompile(`[a-zA-z]°(\.)\s*\d+`),
			replacement: "∯"},
		"questionMarkInQuotation": {
			pattern: regexp.MustCompile(`(\?)(?:\'|\")`), replacement: "&ᓷ&"},
		"singleNewLine": {
			pattern: regexp.MustCompile(`(\s{3,})`), replacement: " "},
		"extraWhiteSpace": {
			pattern: regexp.MustCompile(`(\n)`), replacement: "ȹ"},
		"subSingleQuote": {
			pattern: regexp.MustCompile(`(&⎋&)`), replacement: "'"},
	}
}

func (d *commonDefinition) subRules() []rule {
	return []rule{
		{pattern: regexp.MustCompile(`(∯)`), replacement: "."},
		{pattern: regexp.MustCompile(`(♬)`), replacement: "،"},
		{pattern: regexp.MustCompile(`(♭)`), replacement: ":"},
		{pattern: regexp.MustCompile(`(&ᓰ&)`), replacement: "。"},
		{pattern: regexp.MustCompile(`(&ᓱ&)`), replacement: "．"},
		{pattern: regexp.MustCompile(`(&ᓳ&)`), replacement: "！"},
		{pattern: regexp.MustCompile(`(&ᓴ&)`), replacement: "!"},
		{pattern: regexp.MustCompile(`(&ᓷ&)`), replacement: "?"},
		{pattern: regexp.MustCompile(`(&ᓸ&)`), replacement: "？"},
		{pattern: regexp.MustCompile(`(☉)`), replacement: "?!"},
		{pattern: regexp.MustCompile(`(☇)`), replacement: "??"},
		{pattern: regexp.MustCompile(`(☈)`), replacement: "!?"},
		{pattern: regexp.MustCompile(`(☄)`), replacement: "!!"},
		{pattern: regexp.MustCompile(`(&✂&)`), replacement: "("},
		{pattern: regexp.MustCompile(`(&⌬&)`), replacement: ")"},
		{pattern: regexp.MustCompile(`(ȸ)`), replacement: ""},
		{pattern: regexp.MustCompile(`(ȹ)`), replacement: "\n"},
	}
}

func (d *commonDefinition) doublePunctRules() []rule {
	return []rule{
		{pattern: regexp.MustCompile(`(\?!)`), replacement: "☉"},
		{pattern: regexp.MustCompile(`(!\?)`), replacement: "☈"},
		{pattern: regexp.MustCompile(`(\?\?)`), replacement: "☇"},
		{pattern: regexp.MustCompile(`(!!)`), replacement: "☄"},
	}
}

func (d *commonDefinition) exclamationRules() []rule {
	return []rule{
		{pattern: regexp.MustCompile(`(!)(?:\'|\")`), replacement: "&ᓴ&"},
		{pattern: regexp.MustCompile(`(!)(?:\,\s[a-z])`), replacement: "&ᓴ&"},
		{pattern: regexp.MustCompile(`(!)(?:\s[a-z])`), replacement: "&ᓴ&"},
	}
}

func (d *commonDefinition) starters() []string {
	return []string{
		"A", "Being", "Did", "For", "He", "How", "However", "I", "In", "It",
		"Millions", "More", "She", "That", "The", "There", "They", "We", "What",
		"When", "Where", "Who", "Why"}
}

type frenchDefinition struct {
	commonDefinition
}

func (f *frenchDefinition) abbreviations() map[string][]string {
	return map[string][]string{
		"abbreviations": {
			"a.c.n", "a.m", "al", "ann", "apr", "art", "auj", "av", "b.p", "boul",
			"c.-à-d", "c.n", "c.n.s", "c.p.i", "c.q.f.d", "c.s", "ca", "cf",
			"ch.-l", "chap", "co", "co", "contr", "dir", "e.g", "e.v", "env",
			"etc", "ex", "fasc", "fig", "fr", "fém", "hab", "i.e", "ibid", "id",
			"inf", "l.d", "lib", "ll.aa", "ll.aa.ii", "ll.aa.rr", "ll.aa.ss",
			"ll.ee", "ll.mm", "ll.mm.ii.rr", "loc.cit", "ltd", "ltd", "masc",
			"mm", "ms", "n.b", "n.d", "n.d.a", "n.d.l.r", "n.d.t", "n.p.a.i",
			"n.s", "n/réf", "nn.ss", "p.c.c", "p.ex", "p.j", "p.s", "pl", "pp",
			"r.-v", "r.a.s", "r.i.p", "r.p", "s.a", "s.a.i", "s.a.r", "s.a.s",
			"s.e", "s.m", "s.m.i.r", "s.s", "sec", "sect", "sing", "sq", "sqq",
			"ss", "suiv", "sup", "suppl", "t.s.v.p", "tél", "vb", "vol", "vs",
			"x.o", "z.i", "éd"},
		"prepositive": {},
		"number":      {},
	}
}

func (f *frenchDefinition) starters() []string { return []string{} }

type spanishDefinition struct {
	commonDefinition
}

func (s *spanishDefinition) abbreviations() map[string][]string {
	return map[string][]string{
		"abbreviations": {
			"a.c", "a/c", "abr", "adj", "admón", "afmo", "ago", "almte", "ap",
			"apdo", "arq", "art", "atte", "av", "avda", "bco", "bibl", "bs. as",
			"c", "c.f", "c.g", "c/c", "c/u", "cap", "cc.aa", "cdad", "cm", "co",
			"cra", "cta", "cv", "d.e.p", "da", "dcha", "dcho", "dep", "dic",
			"dicc", "dir", "dn", "doc", "dom", "dpto", "dr", "dra", "dto", "ee",
			"ej", "en", "entlo", "esq", "etc", "excmo", "ext", "f.c", "fca",
			"fdo", "febr", "ff. aa", "ff.cc", "fig", "fil", "fra", "g.p", "g/p",
			"gob", "gr", "gral", "grs", "hnos", "hs", "igl", "iltre", "imp",
			"impr", "impto", "incl", "ing", "inst", "izdo", "izq", "izqdo",
			"j.c", "jue", "jul", "jun", "kg", "km", "lcdo", "ldo", "let", "lic",
			"ltd", "lun", "mar", "may", "mg", "min", "mié", "mm", "máx", "mín",
			"mt", "n. del t", "n.b", "no", "nov", "ntra. sra", "núm", "oct", "p",
			"p.a", "p.d", "p.ej", "p.v.p", "párrf", "ppal", "prev", "prof",
			"prov", "ptas", "pts", "pza", "pág", "págs", "párr", "q.e.g.e",
			"q.e.p.d", "q.e.s.m", "reg", "rep", "rr. hh", "rte", "s", "s. a",
			"s.a.r", "s.e", "s.l", "s.r.c", "s.r.l", "s.s.s", "s/n", "sdad",
			"seg", "sept", "sig", "sr", "sra", "sres", "srta", "sta", "sto",
			"sáb", "t.v.e", "tamb", "tel", "tfno", "ud", "uu", "uds", "univ",
			"v.b", "v.e", "vd", "vds", "vid", "vie", "vol", "vs", "vto", "a",
			"aero", "ambi", "an", "anfi", "ante", "anti", "archi", "arci",
			"auto", "bi", "bien", "bis", "co", "com", "con", "contra", "crio",
			"cuadri", "cuasi", "cuatri", "de", "deci", "des", "di", "dis", "dr",
			"ecto", "en", "endo", "entre", "epi", "equi", "ex", "extra", "geo",
			"hemi", "hetero", "hiper", "hipo", "homo", "i", "im", "in", "infra",
			"inter", "intra", "iso", "lic", "macro", "mega", "micro", "mini",
			"mono", "multi", "neo", "omni", "para", "pen", "ph", "ph.d", "pluri",
			"poli", "pos", "post", "pre", "pro", "pseudo", "re", "retro", "semi",
			"seudo", "sobre", "sub", "super", "supra", "trans", "tras", "tri",
			"ulter", "ultra", "un", "uni", "vice", "yuxta"},
		"prepositive": {
			"a", "aero", "ambi", "an", "anfi", "ante", "anti", "archi", "arci",
			"auto", "bi", "bien", "bis", "co", "com", "con", "contra", "crio",
			"cuadri", "cuasi", "cuatri", "de", "deci", "des", "di", "dis", "dr",
			"ecto", "ee", "en", "endo", "entre", "epi", "equi", "ex", "extra",
			"geo", "hemi", "hetero", "hiper", "hipo", "homo", "i", "im", "in",
			"infra", "inter", "intra", "iso", "lic", "macro", "mega", "micro",
			"mini", "mono", "mt", "multi", "neo", "omni", "para", "pen", "ph",
			"ph.d", "pluri", "poli", "pos", "post", "pre", "pro", "prof",
			"pseudo", "re", "retro", "semi", "seudo", "sobre", "sub", "super",
			"supra", "sra", "srta", "trans", "tras", "tri", "ulter", "ultra",
			"un", "uni", "vice", "yuxta"},
		"number": {"cra", "ext", "no", "nos", "p", "pp", "tel"},
	}
}

func (s *spanishDefinition) starters() []string { return []string{} }

/* language processors */

var langToProcessor = map[string]languageProcessor{
	"en": newProcessor("en"),
	"fr": newProcessor("fr"),
	"es": newProcessor("es"),
}

type languageProcessor interface {
	process(text string) []string
}

type processor struct {
	abbrReplacer *abbreviationReplacer
}

func newProcessor(lang string) *processor {
	r := newAbbreviationReplacer(lang)
	return &processor{abbrReplacer: r}
}

func (p *processor) cleanQuotations(text string) string {
	return substitute(text, "`", "'")
}

func (p *processor) process(text string) []string {
	text = p.abbrReplacer.replace(applyRules(text, cleanRules))
	text = applyRules(text, allNumberRules)

	text = continuousPunctuationRE.ReplaceAllStringFunc(text, func(s string) string {
		return substitute(substitute(s, "!", "&ᓴ&"), "?", "&ᓷ&")
	})

	pRules := p.abbrReplacer.definition.punctRules()
	text = pRules["withMultiplePeriodsAndEmail"].sub(text)
	text = pRules["geoLocation"].sub(text)

	return p.split(text)
}

func (p *processor) split(text string) []string {
	segments := []string{}
	nLineRule := p.abbrReplacer.definition.punctRules()["singleNewLine"]
	for _, segment := range strings.Split(text, "\n") {
		segment = nLineRule.sub(segment)
		segment = applyRules(segment, allEllipsesRules)
		segments = append(segments, p.checkPunct(segment)...)
	}
	return segments
}

func (p *processor) checkPunct(text string) []string {
	segments := []string{}

	chars := p.abbrReplacer.definition.punctuation()
	if util.ContainsAny(text, chars) {
		segments = append(segments, p.processText(text)...)
	} else {
		segments = append(segments, text)
	}

	sentences := []string{}
	singq := p.abbrReplacer.definition.punctRules()["subSingleQuote"]
	for _, segment := range segments {
		segment = applyRules(segment, p.abbrReplacer.definition.subRules())
		segment = singq.sub(segment)
		sentences = append(sentences, p.postProcess(segment)...)
	}
	return sentences
}

func (p *processor) processText(text string) []string {
	pRules := p.abbrReplacer.definition.punctRules()
	if !util.HasAnySuffix(text, p.abbrReplacer.definition.punctuation()) {
		text = text + "ȸ"
	}
	text = subPat(text, "double", exclamationWordsRE)
	text = replaceBetweenQuotes(text)
	text = applyRules(text, p.abbrReplacer.definition.doublePunctRules())
	text = applyRules(text, p.abbrReplacer.definition.exclamationRules())
	text = pRules["questionMarkInQuotation"].sub(text)
	return sentenceBoundaryRE.FindAllString(text, -1)
}

var earlyExit = regexp.MustCompile(`\A[a-zA-Z]*\z`)

func (p *processor) postProcess(text string) []string {
	if len(text) < 2 && earlyExit.MatchString(text) {
		return []string{text}
	}

	text = applyRules(text, p.abbrReplacer.definition.subEllipsis())
	if quotationAtEndOfSentenceRE.MatchString(text) {
		l := splitSpaceQuotationAtEndOfSentenceRE.FindStringSubmatchIndex(text)
		return []string{text[:l[3]-1], text[l[2]+1:]}
	}
	return []string{strings.TrimSpace(text)}
}

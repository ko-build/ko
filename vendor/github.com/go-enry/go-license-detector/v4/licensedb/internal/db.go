package internal

import (
	"archive/tar"
	"bytes"
	"encoding/csv"
	"fmt"
	"index/suffixarray"
	"io"
	"log"
	"os"
	paths "path"
	"regexp"
	"sort"
	"strings"

	minhashlsh "github.com/ekzhu/minhash-lsh"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/go-enry/go-license-detector/v4/licensedb/internal/assets"
	"github.com/go-enry/go-license-detector/v4/licensedb/internal/fastlog"
	"github.com/go-enry/go-license-detector/v4/licensedb/internal/normalize"
	"github.com/go-enry/go-license-detector/v4/licensedb/internal/wmh"
)

var (
	licenseReadmeMentionRe = regexp.MustCompile(
		fmt.Sprintf("(?i)[^\\s]+/[^/\\s]*(%s)[^\\s]*",
			strings.Join(licenseFileNames, "|")))
)

// database holds the license texts, their hashes and the hashtables to query for nearest
// neighbors.
type database struct {
	debug bool

	// license name -> text
	licenseTexts map[string]string
	// minimum license text length
	minLicenseLength int
	// official license URLs
	urls map[string]string
	// all URLs joined
	urlRe *regexp.Regexp
	// first line of each license OR-ed - used to split
	firstLineRe *regexp.Regexp
	// unique unigrams -> index
	tokens map[string]int
	// document frequencies of the unigrams, indexes match with `tokens`
	docfreqs []int
	// Weighted MinHash hashtables
	lsh *minhashlsh.MinhashLSH
	// turns a license text into a hash
	hasher *wmh.WeightedMinHasher
	// part of license short name (e,g, BSL-1.0) -> list of containing license names
	nameShortSubstrings map[string][]substring
	// number of substrings per short license name
	nameShortSubstringSizes map[string]int
	// part of license name (e,g, Boost Software License 1.0) -> list of containing license names
	nameSubstrings map[string][]substring
	// number of substrings per license name
	nameSubstringSizes map[string]int
}

type substring struct {
	value string
	count int
}

const (
	numHashes           = 154
	similarityThreshold = 0.75
)

// Length returns the number of registered licenses.
func (db database) Length() int {
	return len(db.licenseTexts)
}

// VocabularySize returns the number of unique unigrams.
func (db database) VocabularySize() int {
	return len(db.tokens)
}

func loadUrls(db *database) {
	urlCSVBytes, err := assets.Asset("urls.csv")
	if err != nil {
		log.Fatalf("failed to load urls.csv from the assets: %v", err)
	}
	urlReader := csv.NewReader(bytes.NewReader(urlCSVBytes))
	records, err := urlReader.ReadAll()
	if err != nil || len(records) == 0 {
		log.Fatalf("failed to parse urls.csv from the assets: %v", err)
	}
	db.urls = map[string]string{}
	urlReWriter := &bytes.Buffer{}
	for i, record := range records {
		db.urls[record[1]] = record[0]
		urlReWriter.Write([]byte(regexp.QuoteMeta(record[1])))
		if i < len(records)-1 {
			urlReWriter.WriteRune('|')
		}
	}
	db.urlRe = regexp.MustCompile(urlReWriter.String())
}

func loadNames(db *database) {
	namesBytes, err := assets.Asset("names.csv")
	if err != nil {
		log.Fatalf("failed to load banes.csv from the assets: %v", err)
	}
	namesReader := csv.NewReader(bytes.NewReader(namesBytes))
	records, err := namesReader.ReadAll()
	if err != nil || len(records) == 0 {
		log.Fatalf("failed to parse names.csv from the assets: %v", err)
	}
	db.nameSubstringSizes = map[string]int{}
	db.nameSubstrings = map[string][]substring{}
	for _, record := range records {
		registerNameSubstrings(record[1], record[0], db.nameSubstringSizes, db.nameSubstrings)
	}
}

func registerNameSubstrings(
	name string, key string, sizes map[string]int, substrs map[string][]substring) {
	parts := splitLicenseName(name)
	sizes[key] = 0
	for _, part := range parts {
		if licenseReadmeRe.MatchString(part.value) {
			continue
		}
		sizes[key]++
		list := substrs[part.value]
		if list == nil {
			list = []substring{}
		}
		list = append(list, substring{value: key, count: part.count})
		substrs[part.value] = list
	}
}

// Load takes the licenses from the embedded storage, normalizes, hashes them and builds the
// LSH hashtables.
func loadLicenses() *database {
	db := &database{}
	if os.Getenv("LICENSE_DEBUG") != "" {
		db.debug = true
	}
	loadUrls(db)
	loadNames(db)
	tarBytes, err := assets.Asset("licenses.tar")
	if err != nil {
		log.Fatalf("failed to load licenses.tar from the assets: %v", err)
	}
	tarStream := bytes.NewBuffer(tarBytes)
	archive := tar.NewReader(tarStream)
	db.licenseTexts = map[string]string{}
	tokenFreqs := map[string]map[string]int{}
	firstLineWriter := &bytes.Buffer{}
	firstLineWriter.WriteString("(^|\\n)((.*licen[cs]e\\n\\n)|(")
	for header, err := archive.Next(); err != io.EOF; header, err = archive.Next() {
		if len(header.Name) <= 6 {
			continue
		}
		key := header.Name[2 : len(header.Name)-4]
		text := make([]byte, header.Size)
		readSize, readErr := archive.Read(text)
		if readErr != nil && readErr != io.EOF {
			log.Fatalf("failed to load licenses.tar from the assets: %s: %v", header.Name, readErr)
		}
		if int64(readSize) != header.Size {
			log.Fatalf("failed to load licenses.tar from the assets: %s: incomplete read", header.Name)
		}
		normedText := normalize.LicenseText(string(text), normalize.Moderate)
		if db.minLicenseLength == 0 || db.minLicenseLength > len(normedText) {
			db.minLicenseLength = len(normedText)
		}
		db.licenseTexts[key] = normedText
		newLinePos := strings.Index(normedText, "\n")
		if newLinePos >= 0 {
			firstLineWriter.WriteString(regexp.QuoteMeta(normedText[:newLinePos]))
			firstLineWriter.WriteRune('|')
		}
		normedText = normalize.Relax(normedText)
		lines := strings.Split(normedText, "\n")
		myUniqueTokens := map[string]int{}
		tokenFreqs[key] = myUniqueTokens
		for _, line := range lines {
			tokens := strings.Split(line, " ")
			for _, token := range tokens {
				myUniqueTokens[token]++
			}
		}
	}
	if db.debug {
		log.Println("Minimum license length:", db.minLicenseLength)
		log.Println("Number of supported licenses:", len(db.licenseTexts))
	}
	firstLineWriter.Truncate(firstLineWriter.Len() - 1)
	firstLineWriter.WriteString("))")
	db.firstLineRe = regexp.MustCompile(firstLineWriter.String())
	docfreqs := map[string]int{}
	for _, tokens := range tokenFreqs {
		for token := range tokens {
			docfreqs[token]++
		}
	}
	uniqueTokens := make([]string, len(docfreqs))
	{
		i := 0
		for token := range docfreqs {
			uniqueTokens[i] = token
			i++
		}
	}
	sort.Strings(uniqueTokens)
	db.tokens = map[string]int{}
	db.docfreqs = make([]int, len(uniqueTokens))
	for i, token := range uniqueTokens {
		db.tokens[token] = i
		db.docfreqs[i] = docfreqs[token]
	}
	db.lsh = minhashlsh.NewMinhashLSH64(numHashes, similarityThreshold)
	if db.debug {
		k, l := db.lsh.Params()
		log.Println("LSH:", k, l)
	}
	db.hasher = wmh.NewWeightedMinHasher(len(uniqueTokens), numHashes, 7)
	db.nameShortSubstrings = map[string][]substring{}
	db.nameShortSubstringSizes = map[string]int{}
	for key, tokens := range tokenFreqs {
		indices := make([]int, len(tokens))
		values := make([]float32, len(tokens))
		{
			i := 0
			for t, freq := range tokens {
				indices[i] = db.tokens[t]
				values[i] = tfidf(freq, db.docfreqs[indices[i]], len(db.licenseTexts))
				i++
			}
		}
		db.lsh.Add(key, db.hasher.Hash(values, indices))
		registerNameSubstrings(key, key, db.nameShortSubstringSizes, db.nameShortSubstrings)
	}
	db.lsh.Index()
	return db
}

// QueryLicenseText returns the most similar registered licenses.
func (db *database) QueryLicenseText(text string) map[string]float32 {
	parts := normalize.Split(text)
	licenses := map[string]float32{}
	for _, part := range parts {
		for key, val := range db.queryLicenseAbstract(part) {
			if licenses[key] < val {
				licenses[key] = val
			}
		}
	}
	return licenses
}

func (db *database) queryLicenseAbstract(text string) map[string]float32 {
	normalizedModerate := normalize.LicenseText(text, normalize.Moderate)
	titlePositions := db.firstLineRe.FindAllStringIndex(normalizedModerate, -1)
	candidates := db.queryLicenseAbstractNormalized(normalizedModerate)
	var prevPos int
	var prevMatch string
	for i, titlePos := range titlePositions {
		begPos := titlePos[0]
		match := normalizedModerate[titlePos[0]:titlePos[1]]
		if match[0] == '\n' {
			match = match[1:]
		}
		if match == prevMatch {
			begPos = prevPos
		}
		if normalizedModerate[begPos] == '\n' {
			begPos++
		}
		var endPos int
		if i < len(titlePositions)-1 {
			endPos = titlePositions[i+1][0]
		} else {
			endPos = len(normalizedModerate)
		}
		part := normalizedModerate[begPos:endPos]
		prevMatch = match
		prevPos = begPos
		if float64(len(part)) < float64(db.minLicenseLength)*similarityThreshold {
			continue
		}
		newCandidates := db.queryLicenseAbstractNormalized(part)
		if len(newCandidates) == 0 {
			continue
		}
		for key, val := range newCandidates {
			if candidates[key] < val {
				candidates[key] = val
			}
		}
	}
	db.addURLMatches(candidates, text)
	return candidates
}

func (db *database) addURLMatches(candidates map[string]float32, text string) {
	for key := range db.scanForURLs(text) {
		if db.debug {
			println("URL:", key)
		}
		if conf := candidates[key]; conf < similarityThreshold {
			if conf == 0 {
				candidates[key] = 1
			} else {
				candidates[key] = similarityThreshold
			}
		}
	}
}

func (db *database) queryLicenseAbstractNormalized(normalizedModerate string) map[string]float32 {
	normalizedRelaxed := normalize.Relax(normalizedModerate)
	if db.debug {
		println("\nqueryAbstractNormed --------\n")
		println(normalizedModerate)
		println("\n========\n")
		println(normalizedRelaxed)
	}
	tokens := map[int]int{}
	for _, line := range strings.Split(normalizedRelaxed, "\n") {
		for _, token := range strings.Split(line, " ") {
			if index, exists := db.tokens[token]; exists {
				tokens[index]++
			}
		}
	}
	indices := make([]int, len(tokens))
	values := make([]float32, len(tokens))
	{
		i := 0
		for key, val := range tokens {
			indices[i] = key
			values[i] = tfidf(val, db.docfreqs[key], len(db.licenseTexts))
			i++
		}
	}
	found := db.lsh.Query(db.hasher.Hash(values, indices))
	candidates := map[string]float32{}
	if len(found) == 0 {
		return candidates
	}
	for _, keyint := range found {
		key := keyint.(string)
		licenseText := db.licenseTexts[key]
		yourRunes := make([]rune, 0, len(licenseText)/6)
		vocabulary := map[string]int{}
		for _, line := range strings.Split(licenseText, "\n") {
			for _, token := range strings.Split(line, " ") {
				index, exists := vocabulary[token]
				if !exists {
					index = len(vocabulary)
					vocabulary[token] = index
				}
				yourRunes = append(yourRunes, rune(index))
			}
		}

		oovRune := rune(len(vocabulary))
		myRunes := make([]rune, 0, len(normalizedModerate)/6)
		for _, line := range strings.Split(normalizedModerate, "\n") {
			for _, token := range strings.Split(line, " ") {
				if index, exists := vocabulary[token]; exists {
					myRunes = append(myRunes, rune(index))
				} else if len(myRunes) == 0 || myRunes[len(myRunes)-1] != oovRune {
					myRunes = append(myRunes, oovRune)
				}
			}
		}

		dmp := diffmatchpatch.New()
		diff := dmp.DiffMainRunes(myRunes, yourRunes, false)

		if db.debug {
			tokarr := make([]string, len(db.tokens)+1)
			for key, val := range vocabulary {
				tokarr[val] = key
			}
			tokarr[len(db.tokens)] = "!"
			println(dmp.DiffPrettyText(dmp.DiffCharsToLines(diff, tokarr)))
		}
		distance := dmp.DiffLevenshtein(diff)
		candidates[key] = float32(1) - float32(distance)/float32(len(myRunes))
	}
	weak := make([]string, 0, len(candidates))
	for key, val := range candidates {
		if val < similarityThreshold {
			weak = append(weak, key)
		}
	}
	if len(weak) < len(candidates) {
		for _, key := range weak {
			delete(candidates, key)
		}
	}
	return candidates
}

func (db *database) scanForURLs(text string) map[string]bool {
	byteText := []byte(text)
	index := suffixarray.New(byteText)
	urlMatches := index.FindAllIndex(db.urlRe, -1)
	licenses := map[string]bool{}
	for _, match := range urlMatches {
		url := byteText[match[0]:match[1]]
		licenses[db.urls[string(url)]] = true
	}
	return licenses
}

// QueryReadmeText tries to detect licenses mentioned in the README.
func (db *database) QueryReadmeText(text string, fs filer.Filer) map[string]float32 {
	candidates := map[string]float32{}
	append := func(others map[string]float32) {
		for key, val := range others {
			if candidates[key] < val {
				candidates[key] = val
			}
		}
	}
	for _, match := range licenseReadmeMentionRe.FindAllString(text, -1) {
		match = strings.TrimRight(match, ".,:;-")
		content, err := fs.ReadFile(match)
		if err == nil {
			if preprocessor, exists := filePreprocessors[paths.Ext(match)]; exists {
				content = preprocessor(content)
			}
			append(db.QueryLicenseText(string(content)))
		}
	}
	if len(candidates) == 0 {
		append(investigateReadmeFile(text, db.nameSubstrings, db.nameSubstringSizes))
		append(investigateReadmeFile(text, db.nameShortSubstrings, db.nameShortSubstringSizes))
	}
	if db.debug {
		for key, val := range candidates {
			println("NLP", key, val)
		}
	}
	db.addURLMatches(candidates, text)
	return candidates
}

func tfidf(freq int, docfreq int, ndocs int) float32 {
	weight := fastlog.Log(1+float32(freq)) * fastlog.Log(float32(ndocs)/float32(docfreq))
	if weight < 0 {
		// logarithm is approximate
		return 0
	}
	return weight
}

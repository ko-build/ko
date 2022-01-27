package processors

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var (
	skipHTMLRe   = regexp.MustCompile(`^(head|script|style|object)$`)
	htmlHeaderRe = regexp.MustCompile(`^h[2-6]$`)
	htmlEntityRe = regexp.MustCompile(`&((#\\d+)|([a-zA-Z]+));`)
	marksRe      = regexp.MustCompile(`[#$%*\/\\|><~\x60=!?.,:;\"'\])}-]`)
)

func parseHTMLEntity(entName []byte) []byte {
	entNameStr := strings.ToLower(string(entName[1 : len(entName)-1]))

	if entNameStr[0] == '#' {
		val, err := strconv.Atoi(entNameStr[1:])
		if err != nil {
			return entName
		}
		return []byte(string(rune(val)))
	}
	// the list is not full
	switch entNameStr {
	case "nbsp":
		return []byte(" ")
	case "gt":
		return []byte(">")
	case "lt":
		return []byte("<")
	case "amp":
		return []byte("&")
	case "quot":
		return []byte("\"")
	case "apos":
		return []byte("'")
	case "cent":
		return []byte("¢")
	case "pound":
		return []byte("£")
	case "yen":
		return []byte("¥")
	case "euro":
		return []byte("€")
	case "copy":
		return []byte("©")
	case "reg":
		return []byte("®")
	case "ldquo":
		return []byte("“")
	case "rdquo":
		return []byte("”")
	case "lsquo":
		return []byte("‘")
	case "rsquo":
		return []byte("’")
	case "sbquo":
		return []byte("‚")
	case "rbquo":
		return []byte("\"")
	case "bdquo":
		return []byte("„")
	case "ndash":
		return []byte("–")
	case "mdash":
		return []byte("—")
	case "bull":
		return []byte("•")
	case "hellip":
		return []byte("…")
	case "prime":
		return []byte("′")
	case "lsaquo":
		return []byte("‹")
	case "rsaquo":
		return []byte("›")
	case "trade":
		return []byte("™")
	case "minus":
		return []byte("−")
	case "raquo":
		return []byte("»")
	case "laquo":
		return []byte("«")
	case "deg":
		return []byte("°")
	case "sect":
		return []byte("§")
	case "iexcl":
		return []byte("¡")
	default:
		return entName
	}
}

// HTML converts HTML to plain text. E.g. it rips all the tags.
func HTML(htmlSource []byte) []byte {
	result := &bytes.Buffer{}
	doc := html.NewTokenizer(bytes.NewReader(htmlSource))
	skip := false
	var href []byte
	for token := doc.Next(); token != html.ErrorToken; token = doc.Next() {
		tagName, _ := doc.TagName()
		if skipHTMLRe.Match(tagName) {
			if doc.Token().Type != html.SelfClosingTagToken {
				skip = !skip
			}
			continue
		}
		if skip {
			continue
		}
		text := doc.Text()
		if href != nil && doc.Token().Type == html.TextToken {
			myhref := href
			href = nil
			if bytes.Equal(myhref, text) {
				continue
			} else {
				result.WriteRune(' ')
			}
		}
		text = htmlEntityRe.ReplaceAllFunc(text, parseHTMLEntity)
		text = bytes.Replace(text, []byte("\u00a0"), []byte(" "), -1)
		result.Write(text)
		strTagName := string(tagName)
		if strTagName == "br" {
			result.WriteRune('\n')
		} else if strTagName == "hr" {
			result.WriteString("---")
		} else if strTagName == "a" {
			for key, val, _ := doc.TagAttr(); key != nil; key, val, _ = doc.TagAttr() {
				if string(key) == "href" {
					result.Write(val)
					href = val
					break
				}
			}
		} else if htmlHeaderRe.Match(tagName) && doc.Token().Type == html.EndTagToken {
			if result.Len() > 0 && !marksRe.MatchString(string(result.Bytes()[result.Len()-1])) {
				result.WriteRune('.')
			}
		}
	}
	return result.Bytes()
}

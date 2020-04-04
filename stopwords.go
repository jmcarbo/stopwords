// Copyright 2015 Benjamin BALET. All rights reserved.
// Use of this source code is governed by the BSD license
// license that can be found in the LICENSE file.

// stopwords package removes most frequent words from a text content.
// It can be used to improve the accuracy of SimHash algo for example.
// It uses a list of most frequent words used in various languages :
//
// arabic, bulgarian, czech, danish, english, finnish, french, german,
// hungarian, italian, japanese, latvian, norwegian, persian, polish,
// portuguese, romanian, russian, slovak, spanish, swedish, turkish

// Package stopwords contains various algorithms of text comparison (Simhash, Levenshtein)
package stopwords

import (
	"bytes"
	"html"
	"regexp"

	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

var (
	remTags      = regexp.MustCompile(`<[^>]*>`)
	oneSpace     = regexp.MustCompile(`\s{2,}`)
	wordSegmenter = regexp.MustCompile(`[\pL\p{Mc}\p{Mn}-_']+`)
    stop = map[string]*(map[string]string) {
      "ar": &arabic,
      "bg": &bulgarian,
      "ca": &catalan,
      "cs": &czech,
      "da": &danish,
      "de": &german,
      "el": &greek,
      "en": &english,
      "es": &spanish,
      "fa": &persian,
      "fr": &french,
      "fi": &finnish,
      "hu": &hungarian,
      "id": &indonesian,
      "it": &italian,
      "ja": &japanese,
      "km": &khmer,
      "lv": &latvian,
      "nl": &dutch,
      "no": &norwegian,
      "pl": &polish,
      "pt": &portuguese,
      "ro": &romanian,
      "ru": &russian,
      "sk": &slovak,
      "sv": &swedish,
      "th": &thai,
      "tr": &turkish,
    }
)

// DontStripDigits changes the behaviour of the default word segmenter
// by including 'Number, Decimal Digit' Unicode Category as words
func DontStripDigits() {
	wordSegmenter = regexp.MustCompile(`[\pL\p{Mc}\p{Mn}\p{Nd}-_']+`)
}

// OverwriteWordSegmenter allows you to overwrite the default word segmenter
// with your own regular expression
func OverwriteWordSegmenter(expression string) {
	wordSegmenter = regexp.MustCompile(expression)
}

func GetLanguage(content []byte, langCodes []string) ([]byte, []string, int, int) {
  maxCount := 0
  counts := []int{}
  guessedLanguages := []string{}
  //Remove HTML tags
    content = remTags.ReplaceAll(content, []byte(" "))
    content = []byte(html.UnescapeString(string(content)))

  for _, l := range langCodes {
    //Parse language

    il, ok := stop[l]
    if ok {
      _, count, _ := removeStopWordsCount(content, *il)
      //Remove stop words by using a list of most frequent words
      if count > maxCount {
        maxCount = count
      }
      counts = append(counts, count)
    }
  }
  total:=0
  for i, c := range counts {
    if c == maxCount {
      guessedLanguages = append(guessedLanguages, langCodes[i])
    }
  }
  if maxCount > 0 && len(guessedLanguages) > 0 {
    content, _, total = removeStopWordsCount(content, *stop[guessedLanguages[0]])
    //Remove duplicated space characters
    content = oneSpace.ReplaceAll(content, []byte(" "))
  }
  return content, guessedLanguages, maxCount, total
}

// CleanString removes useless spaces and stop words from string content.
// BCP 47 or ISO 639-1 language code (if unknown, we'll apply english filters).
// If cleanHTML is TRUE, remove HTML tags from content and unescape HTML entities.
func CleanString(content string, langCode string, cleanHTML bool) string {
	return string(Clean([]byte(content), langCode, cleanHTML))
}

// Clean removes useless spaces and stop words from a byte slice.
// BCP 47 or ISO 639-1 language code (if unknown, we'll apply english filters).
// If cleanHTML is TRUE, remove HTML tags from content and unescape HTML entities.
func Clean(content []byte, langCode string, cleanHTML bool) []byte {
	//Remove HTML tags
	if cleanHTML {
		content = remTags.ReplaceAll(content, []byte(" "))
		content = []byte(html.UnescapeString(string(content)))
	}

	//Parse language
	tag := language.Make(langCode)
	base, _ := tag.Base()
	langCode = base.String()

	//Remove stop words by using a list of most frequent words
	switch langCode {
	case "ar":
		content = removeStopWords(content, arabic)
	case "bg":
		content = removeStopWords(content, bulgarian)
	case "ca":
		content = removeStopWords(content, catalan)
	case "cs":
		content = removeStopWords(content, czech)
	case "da":
		content = removeStopWords(content, danish)
	case "de":
		content = removeStopWords(content, german)
	case "el":
		content = removeStopWords(content, greek)
	case "en":
		content = removeStopWords(content, english)
	case "es":
		content = removeStopWords(content, spanish)
	case "fa":
		content = removeStopWords(content, persian)
	case "fr":
		content = removeStopWords(content, french)
	case "fi":
		content = removeStopWords(content, finnish)
	case "hu":
		content = removeStopWords(content, hungarian)
	case "id":
		content = removeStopWords(content, indonesian)
	case "it":
		content = removeStopWords(content, italian)
	case "ja":
		content = removeStopWords(content, japanese)
	case "km":
		content = removeStopWords(content, khmer)
	case "lv":
		content = removeStopWords(content, latvian)
	case "nl":
		content = removeStopWords(content, dutch)
	case "no":
		content = removeStopWords(content, norwegian)
	case "pl":
		content = removeStopWords(content, polish)
	case "pt":
		content = removeStopWords(content, portuguese)
	case "ro":
		content = removeStopWords(content, romanian)
	case "ru":
		content = removeStopWords(content, russian)
	case "sk":
		content = removeStopWords(content, slovak)
	case "sv":
		content = removeStopWords(content, swedish)
	case "th":
		content = removeStopWords(content, thai)
	case "tr":
		content = removeStopWords(content, turkish)
	}

	//Remove duplicated space characters
	content = oneSpace.ReplaceAll(content, []byte(" "))

	return content
}

func removeStopWords(content []byte, dict map[string]string) []byte {
  b, _, _ := removeStopWordsCount(content, dict)
  return b
}

// removeStopWords iterates through a list of words and removes stop words counting matches and total.
func removeStopWordsCount(content []byte, dict map[string]string) ([]byte, int, int) {
	var result []byte
        count := 0
        total := 0
	content = norm.NFC.Bytes(content)
	content = bytes.ToLower(content)
	words := wordSegmenter.FindAll(content, -1)
	for _, w := range words {
		//log.Println(w)
		if _, ok := dict[string(w)]; ok {
			result = append(result, ' ')
                        count++
		} else {
			result = append(result, []byte(w)...)
			result = append(result, ' ')
		}
                total++
	}
	return result, count, total
}

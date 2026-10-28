package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

// topWordsCount is the number of the most frequent words to return.
const topWordsCount = 10

// edgePunctuation matches punctuation characters at the word edges.
// Everything that is not a letter, a digit or any special chars.
var edgePunctuation = regexp.MustCompile(`^[^\p{L}\p{N}-]+|[^\p{L}\p{N}-]+$`)

// Top10 returns the 10 most frequent words of the text.
func Top10(text string) []string {
	return topWords(text, nil)
}

// Top10Normalized returns the 10 most frequent words of the text.
func Top10Normalized(text string) []string {
	return topWords(text, normalizeWord)
}

// wordNormalizer converts a word into its normalized form and to report
// use this word to add this one to all.
type wordNormalizer func(word string) (string, bool)

func topWords(text string, normalize wordNormalizer) []string {
	words := strings.Fields(text)
	frequency := make(map[string]int)
	for _, word := range words {
		if normalize != nil {
			var keep bool
			word, keep = normalize(word)
			if !keep {
				continue
			}
		}
		frequency[word]++
	}

	result := make([]string, 0, len(frequency))
	for word := range frequency {
		result = append(result, word)
	}

	sort.Slice(result, func(i, j int) bool {
		if frequency[result[i]] != frequency[result[j]] {
			return frequency[result[i]] > frequency[result[j]]
		}
		return result[i] < result[j]
	})

	if len(result) > topWordsCount {
		result = result[:topWordsCount]
	}
	return result
}

// normalizeWord lowercases the word and removes punctuation at its edges.
func normalizeWord(word string) (string, bool) {
	word = strings.ToLower(word)
	word = edgePunctuation.ReplaceAllString(word, "")
	if word == "" || word == "-" {
		return "", false
	}
	return word, true
}

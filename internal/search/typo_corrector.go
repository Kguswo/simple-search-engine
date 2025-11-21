// 오타 수정 검색

package search

import (
	"strings"
)

// TypoCorrector handles typo correction for Korean and English
type TypoCorrector struct {
	// 자주 틀리는 단어 사전
	commonTypos map[string]string
}

func NewTypoCorrector() *TypoCorrector {
	return &TypoCorrector{
		commonTypos: map[string]string{
			// 영어 오타
			"serach":        "search",
			"searhc":        "search",
			"elasticsearh":  "elasticsearch",
			"elastcisearch": "elasticsearch",
			"dokcer":        "docker",
			"dcoker":        "docker",
			"kubernets":     "kubernetes",
			"kuberentes":    "kubernetes",

			// 한글 오타
			"엘라스틱서치": "elasticsearch",
			"도커":     "docker",
			"쿠버네티스":  "kubernetes",
			"검색엔진":   "검색 엔진",
			"데이타베이스": "데이터베이스",
			"프로그래밍":  "프로그래밍",
		},
	}
}

// Correct attempts to correct typos in the query
func (tc *TypoCorrector) Correct(query string) string {
	query = strings.TrimSpace(query)
	lowerQuery := strings.ToLower(query)

	// 1. 사전에 있는 일반적인 오타 교정
	if corrected, exists := tc.commonTypos[lowerQuery]; exists {
		return corrected
	}

	// 2. 단어별로 분리해서 각각 교정
	words := strings.Fields(query)
	correctedWords := make([]string, 0, len(words))

	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if corrected, exists := tc.commonTypos[lowerWord]; exists {
			correctedWords = append(correctedWords, corrected)
		} else {
			correctedWords = append(correctedWords, word)
		}
	}

	return strings.Join(correctedWords, " ")
}

// GetSuggestion checks if the query might have typos and returns suggestion
func (tc *TypoCorrector) GetSuggestion(query string) (string, bool) {
	corrected := tc.Correct(query)
	if corrected != query {
		return corrected, true
	}
	return "", false
}

// LevenshteinDistance calculates edit distance between two strings
func LevenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

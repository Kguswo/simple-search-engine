// 초성검색

package search

import (
	"strings"
)

// ChosungSearcher handles Korean initial consonant search
type ChosungSearcher struct {
	// 초성 리스트 (ㄱ~ㅎ)
	chosungList []rune
}

func NewChosungSearcher() *ChosungSearcher {
	return &ChosungSearcher{
		chosungList: []rune{
			'ㄱ', 'ㄲ', 'ㄴ', 'ㄷ', 'ㄸ', 'ㄹ', 'ㅁ', 'ㅂ', 'ㅃ',
			'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅉ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ',
		},
	}
}

// ExtractChosung extracts initial consonants from Korean text
// Example: "삼성전자" -> "ㅅㅅㅈㅈ"
func (cs *ChosungSearcher) ExtractChosung(text string) string {
	var result strings.Builder

	for _, char := range text {
		if cs.isKoreanSyllable(char) {
			chosung := cs.getChosung(char)
			result.WriteRune(chosung)
		} else if cs.isChosung(char) {
			result.WriteRune(char)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// IsChosungOnly checks if the input contains only initial consonants
func (cs *ChosungSearcher) IsChosungOnly(text string) bool {
	if len(text) == 0 {
		return false
	}

	for _, char := range text {
		if !cs.isChosung(char) && char != ' ' {
			return false
		}
	}
	return true
}

// MatchChosung checks if a text matches the chosung pattern
// Example: "삼성전자" matches "ㅅㅅㅈㅈ"
func (cs *ChosungSearcher) MatchChosung(text string, chosungPattern string) bool {
	extracted := cs.ExtractChosung(text)
	return strings.Contains(extracted, chosungPattern)
}

// isKoreanSyllable checks if character is a complete Korean syllable (가-힣)
func (cs *ChosungSearcher) isKoreanSyllable(char rune) bool {
	return char >= 0xAC00 && char <= 0xD7A3
}

// isChosung checks if character is an initial consonant
func (cs *ChosungSearcher) isChosung(char rune) bool {
	for _, chosung := range cs.chosungList {
		if char == chosung {
			return true
		}
	}
	return false
}

// getChosung extracts the initial consonant from a Korean syllable
func (cs *ChosungSearcher) getChosung(char rune) rune {
	if !cs.isKoreanSyllable(char) {
		return char
	}

	// 한글 유니코드 계산식
	// 초성 = (코드 - 0xAC00) / 588
	chosungIndex := (char - 0xAC00) / 588
	return cs.chosungList[chosungIndex]
}

// GenerateChosungRegex generates regex pattern for chosung search
func (cs *ChosungSearcher) GenerateChosungRegex(chosung string) string {
	var result strings.Builder

	for _, char := range chosung {
		if cs.isChosung(char) {
			// 해당 초성을 가진 모든 한글 음절 범위
			result.WriteString("[")
			result.WriteString(cs.getChosungRange(char))
			result.WriteString("]")
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// getChosungRange returns the Unicode range for a given chosung
func (cs *ChosungSearcher) getChosungRange(chosung rune) string {
	var chosungIndex int
	for i, ch := range cs.chosungList {
		if ch == chosung {
			chosungIndex = i
			break
		}
	}

	start := 0xAC00 + (chosungIndex * 588)
	end := start + 587

	return string(rune(start)) + "-" + string(rune(end))
}

// SearchByChosung performs chosung-based search
func (cs *ChosungSearcher) SearchByChosung(query string, candidates []string) []string {
	var results []string

	for _, candidate := range candidates {
		if cs.MatchChosung(candidate, query) {
			results = append(results, candidate)
		}
	}

	return results
}

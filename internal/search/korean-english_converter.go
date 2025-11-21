// 한영 검색

package search

import (
	"strings"
)

// KeyboardConverter converts between Korean and English keyboard layouts
type KeyboardConverter struct {
	engToKor map[rune]rune
	korToEng map[rune]rune
}

func NewKeyboardConverter() *KeyboardConverter {
	// 영어 자판 → 한글 자판 매핑
	engToKor := map[rune]rune{
		// 자음
		'q': 'ㅂ', 'w': 'ㅈ', 'e': 'ㄷ', 'r': 'ㄱ', 't': 'ㅅ',
		'y': 'ㅛ', 'u': 'ㅕ', 'i': 'ㅑ', 'o': 'ㅐ', 'p': 'ㅔ',
		'a': 'ㅁ', 's': 'ㄴ', 'd': 'ㅇ', 'f': 'ㄹ', 'g': 'ㅎ',
		'h': 'ㅗ', 'j': 'ㅓ', 'k': 'ㅏ', 'l': 'ㅣ',
		'z': 'ㅋ', 'x': 'ㅌ', 'c': 'ㅊ', 'v': 'ㅍ', 'b': 'ㅠ',
		'n': 'ㅜ', 'm': 'ㅡ',

		// 쌍자음 (Shift)
		'Q': 'ㅃ', 'W': 'ㅉ', 'E': 'ㄸ', 'R': 'ㄲ', 'T': 'ㅆ',
		'O': 'ㅒ', 'P': 'ㅖ',
	}

	// 한글 자판 → 영어 자판 매핑 (역방향)
	korToEng := make(map[rune]rune)
	for k, v := range engToKor {
		korToEng[v] = k
	}

	return &KeyboardConverter{
		engToKor: engToKor,
		korToEng: korToEng,
	}
}

// ConvertEngToKor converts English keyboard input to Korean
// Example: "duddj" -> "ㅇㅕㄷㄷㅈ" -> (needs assembly) -> "영어"
func (kc *KeyboardConverter) ConvertEngToKor(input string) string {
	var result strings.Builder

	for _, char := range input {
		if korChar, exists := kc.engToKor[char]; exists {
			result.WriteRune(korChar)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// ConvertKorToEng converts Korean keyboard input to English
func (kc *KeyboardConverter) ConvertKorToEng(input string) string {
	var result strings.Builder

	for _, char := range input {
		if engChar, exists := kc.korToEng[char]; exists {
			result.WriteRune(engChar)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// SmartConvert tries both conversions and returns both possibilities
func (kc *KeyboardConverter) SmartConvert(query string) []string {
	results := []string{query} // 원본 포함

	// 영어를 한글로
	korResult := kc.ConvertEngToKor(query)
	if korResult != query {
		results = append(results, korResult)
	}

	// 한글을 영어로
	engResult := kc.ConvertKorToEng(query)
	if engResult != query {
		results = append(results, engResult)
	}

	return results
}

// IsEnglishOnly checks if string contains only English characters
func IsEnglishOnly(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != ' ' {
			return false
		}
	}
	return true
}

// IsKoreanOnly checks if string contains Korean characters
func IsKoreanOnly(s string) bool {
	hasKorean := false
	for _, r := range s {
		if r >= 0xAC00 && r <= 0xD7A3 { // 완성형 한글
			hasKorean = true
		} else if r >= 0x1100 && r <= 0x11FF { // 자음/모음
			hasKorean = true
		} else if r >= 0x3130 && r <= 0x318F { // 호환 자모
			hasKorean = true
		} else if r != ' ' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return hasKorean
}

package search

import (
	"strings"
	"unicode"
)

// 한글 자모 상수
const (
	hangulBase    = 0xAC00
	chosungCount  = 19
	jungsungCount = 21
	jongsungCount = 28
)

var (
	chosungList = []string{
		"ㄱ", "ㄲ", "ㄴ", "ㄷ", "ㄸ", "ㄹ", "ㅁ", "ㅂ", "ㅃ",
		"ㅅ", "ㅆ", "ㅇ", "ㅈ", "ㅉ", "ㅊ", "ㅋ", "ㅌ", "ㅍ", "ㅎ",
	}

	jungsungList = []string{
		"ㅏ", "ㅐ", "ㅑ", "ㅒ", "ㅓ", "ㅔ", "ㅕ", "ㅖ", "ㅗ", "ㅘ",
		"ㅙ", "ㅚ", "ㅛ", "ㅜ", "ㅝ", "ㅞ", "ㅟ", "ㅠ", "ㅡ", "ㅢ", "ㅣ",
	}

	jongsungList = []string{
		"", "ㄱ", "ㄲ", "ㄳ", "ㄴ", "ㄵ", "ㄶ", "ㄷ", "ㄹ", "ㄺ",
		"ㄻ", "ㄼ", "ㄽ", "ㄾ", "ㄿ", "ㅀ", "ㅁ", "ㅂ", "ㅄ", "ㅅ",
		"ㅆ", "ㅇ", "ㅈ", "ㅊ", "ㅋ", "ㅌ", "ㅍ", "ㅎ",
	}
)

// KeyboardConverter 변환기
type KeyboardConverter struct {
	engToKor map[rune]rune
	korToEng map[rune]rune
}

func NewKeyboardConverter() *KeyboardConverter {
	engToKor := map[rune]rune{
		'q': 'ㅂ', 'w': 'ㅈ', 'e': 'ㄷ', 'r': 'ㄱ', 't': 'ㅅ',
		'y': 'ㅛ', 'u': 'ㅕ', 'i': 'ㅑ', 'o': 'ㅐ', 'p': 'ㅔ',
		'a': 'ㅁ', 's': 'ㄴ', 'd': 'ㅇ', 'f': 'ㄹ', 'g': 'ㅎ',
		'h': 'ㅗ', 'j': 'ㅓ', 'k': 'ㅏ', 'l': 'ㅣ',
		'z': 'ㅋ', 'x': 'ㅌ', 'c': 'ㅊ', 'v': 'ㅍ', 'b': 'ㅠ', 'n': 'ㅜ', 'm': 'ㅡ',

		// 대문자 처리 (Shift 조합)
		'Q': 'ㅃ', 'W': 'ㅉ', 'E': 'ㄸ', 'R': 'ㄲ', 'T': 'ㅆ',
		'O': 'ㅒ', 'P': 'ㅖ',
	}

	korToEng := make(map[rune]rune)
	for eng, kor := range engToKor {
		korToEng[kor] = eng
	}

	return &KeyboardConverter{
		engToKor: engToKor,
		korToEng: korToEng,
	}
}

// ConvertEngToKor - 영어를 한글로 변환 (간단한 방식)
func (kc *KeyboardConverter) ConvertEngToKor(eng string) string {
	var result strings.Builder

	for _, char := range eng {
		if kor, exists := kc.engToKor[char]; exists {
			result.WriteRune(kor)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// ConvertKorToEng - 한글을 영어로 변환
func (kc *KeyboardConverter) ConvertKorToEng(kor string) string {
	var result strings.Builder

	for _, char := range kor {
		if eng, exists := kc.korToEng[char]; exists {
			result.WriteRune(eng)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// SmartConvert - 스마트 변환 (주요 수정 부분)
func (kc *KeyboardConverter) SmartConvert(query string) []string {
	results := []string{query}

	// 영어처럼 보이는 입력인 경우 (spdlqj 같은 경우)
	if kc.looksLikeEnglishTyping(query) {
		// 영어 -> 한글 변환 시도
		korFromEng := kc.ConvertEngToKor(query)
		if korFromEng != query {
			results = append(results, korFromEng)

			// 추가로 조합된 한글도 시도
			assembled := kc.assembleHangulSimple(korFromEng)
			if assembled != korFromEng {
				results = append(results, assembled)
			}
		}
	}

	// 한글처럼 보이는 입력인 경우
	if kc.containsHangul(query) {
		// 한글 -> 영어 변환 시도
		engFromKor := kc.ConvertKorToEng(query)
		if engFromKor != query {
			results = append(results, engFromKor)
		}
	}

	return results
}

// looksLikeEnglishTyping - 영어 타자처럼 보이는지 확인
func (kc *KeyboardConverter) looksLikeEnglishTyping(s string) bool {
	if len(s) == 0 {
		return false
	}

	englishCount := 0
	totalCount := 0

	for _, r := range s {
		if r == ' ' {
			continue
		}
		totalCount++

		// 영어 키보드 레이아웃에 있는 문자들
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			englishCount++
		}
	}

	// 70% 이상이 영어 문자면 영어 타자로 판단
	return totalCount > 0 && float64(englishCount)/float64(totalCount) >= 0.7
}

// containsHangul - 한글이 포함되어 있는지 확인
func (kc *KeyboardConverter) containsHangul(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

// assembleHangulSimple - 간단한 한글 조합 (초성+중성, 초성+중성+종성 기본 조합)
func (kc *KeyboardConverter) assembleHangulSimple(jamos string) string {
	var result strings.Builder
	runes := []rune(jamos)
	i := 0

	for i < len(runes) {
		current := runes[i]

		// 한글 자모인지 확인
		if kc.isChosung(current) && i+1 < len(runes) && kc.isJungsung(runes[i+1]) {
			chosung := current
			jungsung := runes[i+1]

			// 기본 한글 음절 생성
			chosungIdx := -1
			for idx, ch := range chosungList {
				if ch == string(chosung) {
					chosungIdx = idx
					break
				}
			}

			jungsungIdx := -1
			for idx, jung := range jungsungList {
				if jung == string(jungsung) {
					jungsungIdx = idx
					break
				}
			}

			if chosungIdx != -1 && jungsungIdx != -1 {
				// 종성이 있는지 확인
				if i+2 < len(runes) && kc.isJongsung(runes[i+2]) {
					jongsung := runes[i+2]
					jongsungIdx := -1
					for idx, jong := range jongsungList {
						if jong == string(jongsung) {
							jongsungIdx = idx
							break
						}
					}

					if jongsungIdx != -1 {
						// 초성+중성+종성
						hangul := rune(hangulBase +
							chosungIdx*jungsungCount*jongsungCount +
							jungsungIdx*jongsungCount +
							jongsungIdx)
						result.WriteRune(hangul)
						i += 3
						continue
					}
				}

				// 초성+중성만
				hangul := rune(hangulBase +
					chosungIdx*jungsungCount*jongsungCount +
					jungsungIdx*jongsungCount)
				result.WriteRune(hangul)
				i += 2
				continue
			}
		}

		// 조합 불가능하면 그대로 출력
		result.WriteRune(current)
		i++
	}

	return result.String()
}

func (kc *KeyboardConverter) isChosung(r rune) bool {
	for _, ch := range chosungList {
		if ch == string(r) {
			return true
		}
	}
	return false
}

func (kc *KeyboardConverter) isJungsung(r rune) bool {
	for _, jung := range jungsungList {
		if jung == string(r) {
			return true
		}
	}
	return false
}

func (kc *KeyboardConverter) isJongsung(r rune) bool {
	for _, jong := range jongsungList {
		if jong == string(r) && jong != "" {
			return true
		}
	}
	return false
}

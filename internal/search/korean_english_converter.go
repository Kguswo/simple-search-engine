package search

import (
	"strings"
	"unicode"

	hangul "github.com/suapapa/go_hangul"
)

type KeyboardConverter struct {
	engToKor  map[rune]rune
	commonMap map[string]string
}

func NewKeyboardConverter() *KeyboardConverter {
	engToKor := map[rune]rune{
		'q': 'ㅂ', 'w': 'ㅈ', 'e': 'ㄷ', 'r': 'ㄱ', 't': 'ㅅ',
		'y': 'ㅛ', 'u': 'ㅕ', 'i': 'ㅑ', 'o': 'ㅐ', 'p': 'ㅔ',
		'a': 'ㅁ', 's': 'ㄴ', 'd': 'ㅇ', 'f': 'ㄹ', 'g': 'ㅎ',
		'h': 'ㅗ', 'j': 'ㅓ', 'k': 'ㅏ', 'l': 'ㅣ',
		'z': 'ㅋ', 'x': 'ㅌ', 'c': 'ㅊ', 'v': 'ㅍ', 'b': 'ㅠ', 'n': 'ㅜ', 'm': 'ㅡ',
	}

	commonMap := map[string]string{
		"rnrmf": "구글", "spdlqj": "네이버", "rkdcjf": "다음",
		"rhrrkal": "카카오", "wkqk": "자바", "dnxpzh": "우테코",
		"google": "구글", "naver": "네이버", "java": "자바",
		"kakao": "카카오", "daum": "다음",
	}

	return &KeyboardConverter{
		engToKor:  engToKor,
		commonMap: commonMap,
	}
}

// ConvertEngToKor - 영어를 한글로 변환
func (kc *KeyboardConverter) ConvertEngToKor(eng string) string {
	// 1. 일반적인 단어는 바로 변환
	if kor, exists := kc.commonMap[strings.ToLower(eng)]; exists {
		return kor
	}

	// 2. go-hangul 라이브러리를 사용한 전문 변환
	jamos := kc.engToJamo(eng)
	return kc.assembleWithGoHangul(jamos)
}

// engToJamo - 영어를 한글 자모로 변환
func (kc *KeyboardConverter) engToJamo(eng string) string {
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

// assembleWithGoHangul - go-hangul 라이브러리를 사용한 한글 조합
func (kc *KeyboardConverter) assembleWithGoHangul(jamos string) string {
	runes := []rune(jamos)
	var result strings.Builder

	i := 0
	for i < len(runes) {
		// 현재 문자가 초성인지 확인
		if i < len(runes)-1 && kc.isChosung(runes[i]) && kc.isJungsung(runes[i+1]) {
			// 초성 + 중성 조합 시도
			cho := runes[i]
			jung := runes[i+1]

			// go-hangul의 CompatJamo 함수로 호환 자모 변환
			choCompat := hangul.CompatJamo(cho)
			jungCompat := hangul.CompatJamo(jung)

			// 한글 문자 생성 (유니코드 계산)
			choIdx := kc.indexOfChosung(choCompat)
			jungIdx := kc.indexOfJungsung(jungCompat)

			if choIdx != -1 && jungIdx != -1 {
				hangulChar := rune(0xAC00 + (choIdx * 21 * 28) + (jungIdx * 28))
				result.WriteRune(hangulChar)
				i += 2
				continue
			}
		}

		// 조합 불가능하면 그대로 출력
		result.WriteRune(runes[i])
		i++
	}

	return result.String()
}

// isChosung - 초성인지 확인
func (kc *KeyboardConverter) isChosung(r rune) bool {
	chosungList := []rune{'ㄱ', 'ㄲ', 'ㄴ', 'ㄷ', 'ㄸ', 'ㄹ', 'ㅁ', 'ㅂ', 'ㅃ', 'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅉ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ'}
	for _, ch := range chosungList {
		if ch == r {
			return true
		}
	}
	return false
}

// isJungsung - 중성인지 확인
func (kc *KeyboardConverter) isJungsung(r rune) bool {
	jungsungList := []rune{'ㅏ', 'ㅐ', 'ㅑ', 'ㅒ', 'ㅓ', 'ㅔ', 'ㅕ', 'ㅖ', 'ㅗ', 'ㅘ', 'ㅙ', 'ㅚ', 'ㅛ', 'ㅜ', 'ㅝ', 'ㅞ', 'ㅟ', 'ㅠ', 'ㅡ', 'ㅢ', 'ㅣ'}
	for _, jung := range jungsungList {
		if jung == r {
			return true
		}
	}
	return false
}

// indexOfChosung - 초성 인덱스 찾기
func (kc *KeyboardConverter) indexOfChosung(chosung rune) int {
	chosungList := []rune{'ㄱ', 'ㄲ', 'ㄴ', 'ㄷ', 'ㄸ', 'ㄹ', 'ㅁ', 'ㅂ', 'ㅃ', 'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅉ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ'}
	for i, ch := range chosungList {
		if ch == chosung {
			return i
		}
	}
	return -1
}

// indexOfJungsung - 중성 인덱스 찾기
func (kc *KeyboardConverter) indexOfJungsung(jungsung rune) int {
	jungsungList := []rune{'ㅏ', 'ㅐ', 'ㅑ', 'ㅒ', 'ㅓ', 'ㅔ', 'ㅕ', 'ㅖ', 'ㅗ', 'ㅘ', 'ㅙ', 'ㅚ', 'ㅛ', 'ㅜ', 'ㅝ', 'ㅞ', 'ㅟ', 'ㅠ', 'ㅡ', 'ㅢ', 'ㅣ'}
	for i, jung := range jungsungList {
		if jung == jungsung {
			return i
		}
	}
	return -1
}

// ConvertKorToEng - 한글을 영어로 변환
func (kc *KeyboardConverter) ConvertKorToEng(kor string) string {
	var result strings.Builder

	for _, char := range kor {
		if hangul.IsHangul(char) {
			// go-hangul 라이브러리로 한글 분해
			cho, jung, jong := hangul.Split(char)

			// 초성, 중성 변환
			if eng := kc.findEngFromKor(cho); eng != 0 {
				result.WriteRune(eng)
			}
			if eng := kc.findEngFromKor(jung); eng != 0 {
				result.WriteRune(eng)
			}
			// 종성 변환 (있는 경우)
			if jong != 0 {
				if eng := kc.findEngFromKor(jong); eng != 0 {
					result.WriteRune(eng)
				}
			}
		} else if eng, exists := kc.korToEng()[char]; exists {
			result.WriteRune(eng)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// korToEng - 한글 자모를 영어로 매핑
func (kc *KeyboardConverter) korToEng() map[rune]rune {
	korToEng := make(map[rune]rune)
	for eng, kor := range kc.engToKor {
		korToEng[kor] = eng
	}
	return korToEng
}

// findEngFromKor - 한글 자모에 해당하는 영어 찾기
func (kc *KeyboardConverter) findEngFromKor(kor rune) rune {
	for eng, k := range kc.engToKor {
		if k == kor {
			return eng
		}
	}
	return 0
}

// isJamo - 한글 자모인지 확인
func (kc *KeyboardConverter) isJamo(r rune) bool {
	// 한글 자모 범위: ㄱ-ㅎ(0x3131-0x314E), ㅏ-ㅣ(0x314F-0x3163)
	return (r >= 0x3131 && r <= 0x314E) || (r >= 0x314F && r <= 0x3163)
}

// SmartConvert - 스마트 변환
func (kc *KeyboardConverter) SmartConvert(query string) []string {
	results := []string{query}

	if kc.looksLikeEnglishTyping(query) {
		korResult := kc.ConvertEngToKor(query)
		if korResult != query {
			results = append(results, korResult)
		}
	}

	if kc.containsHangul(query) {
		engResult := kc.ConvertKorToEng(query)
		if engResult != query {
			results = append(results, engResult)
		}
	}

	return kc.removeDuplicates(results)
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

		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			englishCount++
		}
	}

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

// removeDuplicates - 중복 제거
func (kc *KeyboardConverter) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

package search

import (
	"strings"
)

// ChosungSearcher handles Korean initial consonant search
type ChosungSearcher struct {
	chosungList   []rune
	chosungSet    map[rune]bool   // O(1) 초성 체크용
	chosungIndex  map[rune]int    // O(1) 인덱스 조회용
	chosungRanges map[rune]string // 미리 계산된 유니코드 범위
}

func NewChosungSearcher() *ChosungSearcher {
	chosungList := []rune{
		'ㄱ', 'ㄲ', 'ㄴ', 'ㄷ', 'ㄸ', 'ㄹ', 'ㅁ', 'ㅂ', 'ㅃ',
		'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅉ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ',
	}

	// 초성 셋 생성 (O(1) 조회)
	chosungSet := make(map[rune]bool)
	chosungIndex := make(map[rune]int)
	for i, ch := range chosungList {
		chosungSet[ch] = true
		chosungIndex[ch] = i
	}

	// 초성별 유니코드 범위 미리 계산 (캐싱)
	chosungRanges := make(map[rune]string)
	for i, ch := range chosungList {
		start := 0xAC00 + (i * 588)
		end := start + 587
		chosungRanges[ch] = string(rune(start)) + "-" + string(rune(end))
	}

	return &ChosungSearcher{
		chosungList:   chosungList,
		chosungSet:    chosungSet,
		chosungIndex:  chosungIndex,
		chosungRanges: chosungRanges,
	}
}

// ExtractChosung extracts initial consonants from Korean text
// Example: "삼성전자" -> "ㅅㅅㅈㅈ"
func (cs *ChosungSearcher) ExtractChosung(text string) string {
	var result strings.Builder
	result.Grow(len(text)) // 미리 메모리 할당

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

	hasChosung := false
	for _, char := range text {
		if cs.isChosung(char) {
			hasChosung = true
		} else if char != ' ' {
			return false
		}
	}
	return hasChosung
}

// MatchChosung checks if a text matches the chosung pattern
// Example: "삼성전자" matches "ㅅㅅㅈㅈ"
func (cs *ChosungSearcher) MatchChosung(text string, chosungPattern string) bool {
	extracted := cs.ExtractChosung(text)
	return strings.Contains(extracted, chosungPattern)
}

// MatchChosungPrefix - 접두사 매칭 (더 정확)
func (cs *ChosungSearcher) MatchChosungPrefix(text string, chosungPattern string) bool {
	extracted := cs.ExtractChosung(text)
	return strings.HasPrefix(extracted, chosungPattern)
}

// isKoreanSyllable checks if character is a complete Korean syllable (가-힣)
func (cs *ChosungSearcher) isKoreanSyllable(char rune) bool {
	return char >= 0xAC00 && char <= 0xD7A3
}

// isChosung checks if character is an initial consonant (O(1))
func (cs *ChosungSearcher) isChosung(char rune) bool {
	return cs.chosungSet[char]
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
	result.WriteString(".*") // 앞에 임의의 문자 허용

	for _, char := range chosung {
		if rangeStr, exists := cs.chosungRanges[char]; exists {
			// 캐시된 범위 사용 (O(1))
			result.WriteString("[")
			result.WriteString(rangeStr)
			result.WriteString("]")
		} else if char == ' ' {
			result.WriteString(".*")
		} else {
			// 특수문자 이스케이프
			result.WriteString(escapeRegex(string(char)))
		}
	}

	result.WriteString(".*") // 뒤에 임의의 문자 허용
	return result.String()
}

// GenerateChosungRegexStrict - 정확한 매칭용 (앞뒤 .* 없음)
func (cs *ChosungSearcher) GenerateChosungRegexStrict(chosung string) string {
	var result strings.Builder
	result.WriteString("^") // 문자열 시작

	for _, char := range chosung {
		if rangeStr, exists := cs.chosungRanges[char]; exists {
			result.WriteString("[")
			result.WriteString(rangeStr)
			result.WriteString("]")
		} else if char == ' ' {
			result.WriteString(" ")
		} else {
			result.WriteString(escapeRegex(string(char)))
		}
	}

	result.WriteString("$") // 문자열 끝
	return result.String()
}

// SearchByChosung performs chosung-based search
func (cs *ChosungSearcher) SearchByChosung(query string, candidates []string) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	results := make([]string, 0, len(candidates)/10) // 대략 10% 예상

	for _, candidate := range candidates {
		if cs.MatchChosung(candidate, query) {
			results = append(results, candidate)
		}
	}

	return results
}

// ChosungResult - 초성 검색 결과
type ChosungResult struct {
	Text  string
	Score int
}

// SearchByChosungWithScore - 점수 포함 검색
func (cs *ChosungSearcher) SearchByChosungWithScore(query string, candidates []string) []ChosungResult {
	results := make([]ChosungResult, 0)
	queryChosung := strings.ToLower(query)

	for _, candidate := range candidates {
		candidateChosung := cs.ExtractChosung(candidate)

		if strings.Contains(candidateChosung, queryChosung) {
			score := calculateChosungScore(candidateChosung, queryChosung)
			results = append(results, ChosungResult{
				Text:  candidate,
				Score: score,
			})
		}
	}

	// 점수순 정렬 (간단한 버블 정렬)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// ChosungSimilarity - 초성 유사도 계산 (0~100)
func (cs *ChosungSearcher) ChosungSimilarity(text1, text2 string) int {
	chosung1 := cs.ExtractChosung(text1)
	chosung2 := cs.ExtractChosung(text2)

	if len(chosung1) == 0 || len(chosung2) == 0 {
		return 0
	}

	// LCS (Longest Common Subsequence)
	lcs := longestCommonSubsequence(chosung1, chosung2)
	maxLen := maxInt(len(chosung1), len(chosung2))

	return (lcs * 100) / maxLen
}

// ConvertChosungToWords - 초성 패턴에 맞는 단어 추천
func (cs *ChosungSearcher) ConvertChosungToWords(chosung string, dictionary []string, limit int) []string {
	results := make([]string, 0, limit)
	seen := make(map[string]bool)

	for _, word := range dictionary {
		if len(results) >= limit {
			break
		}

		if cs.MatchChosung(word, chosung) && !seen[word] {
			results = append(results, word)
			seen[word] = true
		}
	}

	return results
}

// Helper functions

// calculateChosungScore - 초성 매칭 점수 계산
func calculateChosungScore(candidate, query string) int {
	score := 0

	// 완전 일치
	if candidate == query {
		return 100
	}

	// 접두사 매칭
	if strings.HasPrefix(candidate, query) {
		score += 50
	}

	// 포함 여부
	if strings.Contains(candidate, query) {
		score += 30
	}

	// 길이 차이 패널티
	lengthDiff := abs(len(candidate) - len(query))
	score -= lengthDiff * 2

	// LCS 기반 추가 점수
	lcs := longestCommonSubsequence(candidate, query)
	score += (lcs * 20) / maxInt(len(candidate), len(query))

	if score < 0 {
		score = 0
	}

	return score
}

// longestCommonSubsequence - 최장 공통 부분 수열
func longestCommonSubsequence(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)

	m, n := len(r1), len(r2)
	if m == 0 || n == 0 {
		return 0
	}

	// DP 테이블
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if r1[i-1] == r2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = maxInt(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

// escapeRegex - 정규식 특수문자 이스케이프
func escapeRegex(s string) string {
	specialChars := []string{
		".", "*", "+", "?", "[", "]", "(", ")",
		"{", "}", "^", "$", "|", "\\",
	}

	result := s
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}

// abs - 절댓값
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// maxInt - 최댓값
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetChosungStatistics - 초성 분포 통계
func (cs *ChosungSearcher) GetChosungStatistics(text string) map[rune]int {
	stats := make(map[rune]int)

	for _, char := range text {
		if cs.isKoreanSyllable(char) {
			chosung := cs.getChosung(char)
			stats[chosung]++
		}
	}

	return stats
}

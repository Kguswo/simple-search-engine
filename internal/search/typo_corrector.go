package search

import (
	"strings"
	"unicode"
)

// TypoCorrector handles typo correction dynamically
type TypoCorrector struct {
	commonPatterns map[string]string // 키보드 인접 오타 등 매우 일반적인 패턴만
}

func NewTypoCorrector() *TypoCorrector {
	return &TypoCorrector{
		commonPatterns: map[string]string{
			// 키보드 인접 오타 (영어)
			"serach": "search",
			"teh":    "the",
		},
	}
}

// CorrectWithCandidates - 실제 검색 결과(후보)를 기반으로 교정
func (tc *TypoCorrector) CorrectWithCandidates(query string, candidates []string) string {
	if len(candidates) == 0 {
		return query
	}

	query = strings.TrimSpace(query)
	lowerQuery := strings.ToLower(query)

	// 후보 중에서 가장 유사한 것 찾기
	bestMatch := ""
	minDistance := len(query) / 3 // 쿼리 길이의 33%까지만 허용

	for _, candidate := range candidates {
		distance := LevenshteinDistance(lowerQuery, strings.ToLower(candidate))

		// 너무 다르면 무시 (오탈자가 아닐 수 있음)
		if distance > 0 && distance < minDistance {
			minDistance = distance
			bestMatch = candidate
		}
	}

	if bestMatch != "" {
		return bestMatch
	}

	return query
}

// Correct attempts basic correction (정적 패턴만)
func (tc *TypoCorrector) Correct(query string) string {
	query = strings.TrimSpace(query)
	lowerQuery := strings.ToLower(query)

	// 극소수의 일반적인 패턴만 체크
	if corrected, exists := tc.commonPatterns[lowerQuery]; exists {
		return corrected
	}

	// 실제로는 여기서 끝
	// Elasticsearch의 Fuzzy Query나 Suggestion API가 알아서 처리
	return query
}

// GetSuggestion - 간단한 교정 제안
func (tc *TypoCorrector) GetSuggestion(query string) (string, bool) {
	corrected := tc.Correct(query)
	if corrected != query {
		return corrected, true
	}
	return "", false
}

// GetMultipleSuggestions - 후보 기반 다중 제안
func (tc *TypoCorrector) GetMultipleSuggestions(query string, candidates []string, limit int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	lowerQuery := strings.ToLower(query)
	suggestions := make([]struct {
		word     string
		distance int
	}, 0)

	// 모든 후보의 편집 거리 계산
	for _, candidate := range candidates {
		distance := LevenshteinDistance(lowerQuery, strings.ToLower(candidate))

		// 적절한 범위 내의 후보만
		if distance > 0 && distance <= len(query)/2 {
			suggestions = append(suggestions, struct {
				word     string
				distance int
			}{candidate, distance})
		}
	}

	// 거리순 정렬 (간단한 버블 정렬)
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].distance < suggestions[i].distance {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	// 상위 N개만 반환
	result := make([]string, 0, limit)
	for i := 0; i < len(suggestions) && i < limit; i++ {
		result = append(result, suggestions[i].word)
	}

	return result
}

// LevenshteinDistance - 편집 거리 계산 (핵심 알고리즘)
func LevenshteinDistance(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)

	if len(r1) == 0 {
		return len(r2)
	}
	if len(r2) == 0 {
		return len(r1)
	}

	// DP 테이블 생성
	matrix := make([][]int, len(r1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(r2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// DP 계산
	for i := 1; i <= len(r1); i++ {
		for j := 1; j <= len(r2); j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost

			matrix[i][j] = min(deletion, insertion, substitution)

			// Damerau-Levenshtein: 인접 문자 교환
			if i > 1 && j > 1 && r1[i-1] == r2[j-2] && r1[i-2] == r2[j-1] {
				transposition := matrix[i-2][j-2] + cost
				matrix[i][j] = min(matrix[i][j], transposition)
			}
		}
	}

	return matrix[len(r1)][len(r2)]
}

// JaroWinklerSimilarity - 유사도 점수 (0~1)
// 문자열이 얼마나 비슷한지 퍼센트로 계산
func JaroWinklerSimilarity(s1, s2 string) float64 {
	r1 := []rune(s1)
	r2 := []rune(s2)

	if len(r1) == 0 && len(r2) == 0 {
		return 1.0
	}
	if len(r1) == 0 || len(r2) == 0 {
		return 0.0
	}

	// 매칭 거리 계산
	matchDistance := max(len(r1), len(r2))/2 - 1
	if matchDistance < 1 {
		matchDistance = 1
	}

	s1Matches := make([]bool, len(r1))
	s2Matches := make([]bool, len(r2))
	matches := 0.0
	transpositions := 0.0

	// 매칭 찾기
	for i := 0; i < len(r1); i++ {
		start := max(0, i-matchDistance)
		end := min(i+matchDistance+1, len(r2))

		for j := start; j < end; j++ {
			if s2Matches[j] || r1[i] != r2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Transposition 계산
	k := 0
	for i := 0; i < len(r1); i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if r1[i] != r2[k] {
			transpositions++
		}
		k++
	}

	// Jaro 유사도
	jaro := (matches/float64(len(r1)) +
		matches/float64(len(r2)) +
		(matches-transpositions/2)/matches) / 3.0

	// Winkler 보너스 (공통 접두사)
	prefix := 0
	for i := 0; i < min(len(r1), len(r2), 4); i++ {
		if r1[i] == r2[i] {
			prefix++
		} else {
			break
		}
	}

	return jaro + float64(prefix)*0.1*(1.0-jaro)
}

// FindBestMatchBySimilarity - 유사도 기반 최적 매칭
func FindBestMatchBySimilarity(query string, candidates []string, threshold float64) string {
	bestMatch := ""
	bestSimilarity := threshold // 임계값 이상만 고려

	for _, candidate := range candidates {
		similarity := JaroWinklerSimilarity(
			strings.ToLower(query),
			strings.ToLower(candidate),
		)

		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = candidate
		}
	}

	return bestMatch
}

// NormalizeQuery - 쿼리 정규화
func (tc *TypoCorrector) NormalizeQuery(query string) string {
	query = strings.TrimSpace(query)
	query = strings.Join(strings.Fields(query), " ")

	var result strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// IsTypo - 오타 가능성 체크
func (tc *TypoCorrector) IsTypo(word string, candidates []string) bool {
	lowerWord := strings.ToLower(word)

	// 후보 중에 정확히 일치하는 게 있으면 오타 아님
	for _, candidate := range candidates {
		if lowerWord == strings.ToLower(candidate) {
			return false
		}
	}

	// 유사한 후보가 있으면 오타일 가능성
	for _, candidate := range candidates {
		distance := LevenshteinDistance(lowerWord, strings.ToLower(candidate))
		if distance <= 2 && distance <= len(word)/3 {
			return true
		}
	}

	return false
}

// Helper functions
func min(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	result := nums[0]
	for _, n := range nums[1:] {
		if n < result {
			result = n
		}
	}
	return result
}

func max(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	result := nums[0]
	for _, n := range nums[1:] {
		if n > result {
			result = n
		}
	}
	return result
}

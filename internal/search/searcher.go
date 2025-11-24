package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"simple-search-engine/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
)

type Searcher struct {
	client            *elasticsearch.Client
	chosungSearcher   *ChosungSearcher
	keyboardConverter *KeyboardConverter
	typoCorrector     *TypoCorrector
}

func NewSearcher(client *elasticsearch.Client) *Searcher {
	return &Searcher{
		client:            client,
		chosungSearcher:   NewChosungSearcher(),
		keyboardConverter: NewKeyboardConverter(),
		typoCorrector:     NewTypoCorrector(),
	}
}

// SmartSearch - 통합 검색 (5단계 폴백)
func (s *Searcher) SmartSearch(query string, indexName string) ([]models.Document, string, string, error) {
	// 1단계: 기본 검색
	results, err := s.Search(query, indexName)
	if err == nil && len(results) > 0 {
		return results, query, "기본 검색", nil
	}

	// 2단계: 한영 변환
	results, correctedQuery, err := s.SearchWithKeyboardConversion(query, indexName)
	if err == nil && len(results) > 0 {
		return results, correctedQuery, "한영 변환", nil
	}

	// 3단계: 초성 검색
	if s.chosungSearcher.IsChosungOnly(query) {
		results, err := s.SearchWithChosung(query, indexName)
		if err == nil && len(results) > 0 {
			return results, query, "초성 검색", nil
		}
	}

	// 4단계: 오타 교정
	results, correctedQuery, err = s.SearchWithTypoCorrection(query, indexName)
	if err == nil && len(results) > 0 {
		return results, correctedQuery, "오타 교정", nil
	}

	// 5단계: Fuzzy 검색
	results, err = s.FuzzySearch(query, indexName)
	if err == nil && len(results) > 0 {
		return results, query, "유사 검색", nil
	}

	return []models.Document{}, query, "결과 없음", nil
}

// Search - 기본 검색
func (s *Searcher) Search(query string, indexName string) ([]models.Document, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title^3", "content", "tags^2"},
				"type":   "best_fields",
			},
		},
		"size": 20,
	}

	return s.executeSearch(searchQuery, indexName)
}

// FuzzySearch - 퍼지 검색
func (s *Searcher) FuzzySearch(query string, indexName string) ([]models.Document, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":         query,
				"fields":        []string{"title^3", "content"},
				"fuzziness":     "AUTO",
				"prefix_length": 2,
				"type":          "best_fields",
			},
		},
		"size": 20,
	}

	return s.executeSearch(searchQuery, indexName)
}

// SearchWithTypoCorrection - 오타 교정 검색
func (s *Searcher) SearchWithTypoCorrection(query string, indexName string) ([]models.Document, string, error) {
	// 먼저 검색 결과 가져오기
	initialResults, _ := s.Search(query, indexName)

	// 결과가 있으면 후보 단어 추출
	var candidates []string
	for _, doc := range initialResults {
		candidates = append(candidates, doc.Title)
	}

	// 후보 기반 교정
	correctedQuery := s.typoCorrector.CorrectWithCandidates(query, candidates)

	// 교정된 쿼리로 검색
	if correctedQuery != query {
		results, err := s.Search(correctedQuery, indexName)
		if err == nil && len(results) > 0 {
			return results, correctedQuery, nil
		}
	}

	// 교정 실패 시 원본 결과 반환
	return initialResults, query, nil
}

// SearchWithKeyboardConversion - 한영 변환 검색
func (s *Searcher) SearchWithKeyboardConversion(query string, indexName string) ([]models.Document, string, error) {
	queries := s.keyboardConverter.SmartConvert(query)

	var bestResults []models.Document
	var usedQuery string

	for _, q := range queries {
		if q == query {
			continue // 원본은 이미 시도했음
		}

		results, err := s.Search(q, indexName)
		if err != nil {
			continue
		}

		if len(results) > len(bestResults) {
			bestResults = results
			usedQuery = q
		}
	}

	if usedQuery == "" {
		usedQuery = query
	}

	return bestResults, usedQuery, nil
}

// SearchWithChosung - 초성 검색
func (s *Searcher) SearchWithChosung(query string, indexName string) ([]models.Document, error) {
	if !s.chosungSearcher.IsChosungOnly(query) {
		return s.Search(query, indexName)
	}

	log.Printf("🔍 초성 검색: %s", query)

	// Elasticsearch regexp 쿼리 사용
	regexPattern := s.chosungSearcher.GenerateChosungRegex(query)

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"regexp": map[string]interface{}{
							"title": map[string]interface{}{
								"value": regexPattern,
							},
						},
					},
					{
						"regexp": map[string]interface{}{
							"content": map[string]interface{}{
								"value": regexPattern,
							},
						},
					},
				},
				"minimum_should_match": 1,
			},
		},
		"size": 20,
	}

	return s.executeSearch(searchQuery, indexName)
}

// Autocomplete - 자동완성
func (s *Searcher) Autocomplete(prefix string, indexName string) ([]string, error) {
	if len([]rune(prefix)) < 2 {
		return []string{}, nil
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase_prefix": map[string]interface{}{
				"title": map[string]interface{}{
					"query": prefix,
				},
			},
		},
		"size":    10,
		"_source": []string{"title"},
	}

	results, err := s.executeSearch(searchQuery, indexName)
	if err != nil {
		return []string{}, err
	}

	suggestions := make([]string, 0, len(results))
	seen := make(map[string]bool)

	for _, doc := range results {
		if !seen[doc.Title] {
			suggestions = append(suggestions, doc.Title)
			seen[doc.Title] = true
		}
	}

	return suggestions, nil
}

// executeSearch - 공통 검색 실행 로직
func (s *Searcher) executeSearch(searchQuery map[string]interface{}, indexName string) ([]models.Document, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("쿼리 JSON 인코딩 실패: %w", err)
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(indexName),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("검색 요청 실패: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("검색 응답 에러: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	documents := make([]models.Document, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]

		docBytes, _ := json.Marshal(source)
		var doc models.Document
		if err := json.Unmarshal(docBytes, &doc); err != nil {
			log.Printf("문서 변환 실패: %v", err)
			continue
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

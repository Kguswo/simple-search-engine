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
	client *elasticsearch.Client
}

func NewSearcher(client *elasticsearch.Client) *Searcher {
	return &Searcher{client: client}
}

// Search performs a search query on the index
func (s *Searcher) Search(query string, indexName string) ([]models.Document, error) {
	// 검색 쿼리 구성 (multi_match: title, content, tags 검색)
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title^3", "content", "tags^2"},
				"type":   "best_fields",
			},
		},
		"size": 10, // 상위 10개 결과
	}

	// JSON 변환
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("쿼리 JSON 인코딩 실패: %w", err)
	}

	// 검색 요청
	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(indexName),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
		s.client.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("검색 요청 실패: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("검색 응답 에러: %s", res.String())
	}

	// 응답 파싱
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	// 결과 추출
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	documents := make([]models.Document, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		score := hit.(map[string]interface{})["_score"]

		// Document로 변환
		docBytes, _ := json.Marshal(source)
		var doc models.Document
		if err := json.Unmarshal(docBytes, &doc); err != nil {
			log.Printf("문서 변환 실패: %v", err)
			continue
		}

		log.Printf("📄 [Score: %.2f] %s", score, doc.Title)
		documents = append(documents, doc)
	}

	return documents, nil
}

// FuzzySearch performs a fuzzy search (오타 허용 검색)
func (s *Searcher) FuzzySearch(query string, indexName string) ([]models.Document, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     query,
				"fields":    []string{"title^3", "content"},
				"fuzziness": "AUTO",
				"type":      "best_fields",
			},
		},
		"size": 10,
	}

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
		score := hit.(map[string]interface{})["_score"]

		docBytes, _ := json.Marshal(source)
		var doc models.Document
		if err := json.Unmarshal(docBytes, &doc); err != nil {
			continue
		}

		log.Printf("📄 [Fuzzy Score: %.2f] %s", score, doc.Title)
		documents = append(documents, doc)
	}

	return documents, nil
}

// SearchWithTypoCorrection performs search with automatic typo correction
func (s *Searcher) SearchWithTypoCorrection(query string, indexName string) ([]models.Document, string, error) {
	corrector := NewTypoCorrector()

	// 오타 교정 시도
	correctedQuery := corrector.Correct(query)

	// 교정된 쿼리로 검색
	results, err := s.Search(correctedQuery, indexName)
	if err != nil {
		return nil, "", err
	}

	// 검색 결과가 없으면 원본 쿼리로 재시도
	if len(results) == 0 && correctedQuery != query {
		results, err = s.Search(query, indexName)
		correctedQuery = query
	}

	return results, correctedQuery, err
}

// SearchWithKeyboardConversion performs search with keyboard layout conversion
func (s *Searcher) SearchWithKeyboardConversion(query string, indexName string) ([]models.Document, string, error) {
	converter := NewKeyboardConverter()

	// 모든 변환 가능성 생성
	queries := converter.SmartConvert(query)

	var allResults []models.Document
	var usedQuery string

	// 각 변환 결과로 검색 시도
	for _, q := range queries {
		results, err := s.Search(q, indexName)
		if err != nil {
			continue
		}

		if len(results) > len(allResults) {
			allResults = results
			usedQuery = q
		}
	}

	if usedQuery == "" {
		usedQuery = query
	}

	return allResults, usedQuery, nil
}

// SearchWithChosung performs search with initial consonant matching
func (s *Searcher) SearchWithChosung(query string, indexName string) ([]models.Document, error) {
	chosungSearcher := NewChosungSearcher()

	// 초성 검색인지 확인
	if !chosungSearcher.IsChosungOnly(query) {
		// 초성이 아니면 일반 검색
		return s.Search(query, indexName)
	}

	log.Printf("🔍 초성 검색 모드: %s", query)

	// 모든 문서 가져오기 (필터링 위해)
	allDocs, err := s.getAllDocuments(indexName)
	if err != nil {
		return nil, err
	}

	// 초성 매칭되는 문서만 필터링
	var results []models.Document
	for _, doc := range allDocs {
		if chosungSearcher.MatchChosung(doc.Title, query) ||
			chosungSearcher.MatchChosung(doc.Content, query) {
			results = append(results, doc)
			log.Printf("✅ 초성 매칭: %s", doc.Title)
		}
	}

	return results, nil
}

// getAllDocuments retrieves all documents from the index
func (s *Searcher) getAllDocuments(indexName string) ([]models.Document, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 1000, // 최대 1000개
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(indexName),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	documents := make([]models.Document, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		docBytes, _ := json.Marshal(source)
		var doc models.Document
		if err := json.Unmarshal(docBytes, &doc); err != nil {
			continue
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

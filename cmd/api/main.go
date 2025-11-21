package main

import (
	"log"
	"net/http"

	"simple-search-engine/internal/elasticsearch"
	"simple-search-engine/internal/indexer"
	"simple-search-engine/internal/models"
	"simple-search-engine/internal/search"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 SmartSearch API 서버 시작...")

	// Elasticsearch 클라이언트 초기화
	client, err := elasticsearch.NewClient()
	if err != nil {
		log.Fatalf("Elasticsearch 연결 실패: %v", err)
	}

	// Searcher 초기화
	searcher := search.NewSearcher(client)

	// Gin 라우터 설정
	router := gin.Default()

	// CORS 설정 (프론트엔드에서 호출 가능하도록)
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 헬스체크
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "SmartSearch API is running",
		})
	})

	// 검색 API
	router.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "검색어를 입력해주세요",
			})
			return
		}

		log.Printf("🔍 검색 요청: %s", query)

		// 검색 수행
		results, err := searcher.Search(query, indexer.IndexName)
		if err != nil {
			log.Printf("❌ 검색 실패: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "검색 중 오류가 발생했습니다",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   query,
			"total":   len(results),
			"results": results,
		})
	})

	// Fuzzy 검색 API
	router.GET("/search/fuzzy", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "검색어를 입력해주세요",
			})
			return
		}

		log.Printf("🔍 Fuzzy 검색 요청: %s", query)

		results, err := searcher.FuzzySearch(query, indexer.IndexName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "검색 중 오류가 발생했습니다",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   query,
			"total":   len(results),
			"results": results,
		})
	})

	// 스마트 통합 검색 API (모든 기능 자동 적용)
	router.GET("/search/smart", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "검색어를 입력해주세요",
			})
			return
		}

		log.Printf("🔍 스마트 검색 요청: %s", query)

		var results []models.Document
		var correctedQuery string
		var searchMethod string

		// 1단계: 기본 검색 시도
		results, err := searcher.Search(query, indexer.IndexName)
		if err == nil && len(results) > 0 {
			correctedQuery = query
			searchMethod = "기본 검색"
			log.Printf("✅ 기본 검색 성공: %d개 결과", len(results))
		} else {
			// 2단계: 한영 변환 검색
			results, correctedQuery, err = searcher.SearchWithKeyboardConversion(query, indexer.IndexName)
			if err == nil && len(results) > 0 {
				searchMethod = "한영 변환"
				log.Printf("✅ 한영 변환 검색 성공: %s → %s", query, correctedQuery)
			} else {
				// 3단계: 초성 검색
				results, err = searcher.SearchWithChosung(query, indexer.IndexName)
				if err == nil && len(results) > 0 {
					correctedQuery = query
					searchMethod = "초성 검색"
					log.Printf("✅ 초성 검색 성공: %d개 결과", len(results))
				} else {
					// 4단계: 오타 교정 검색
					results, correctedQuery, err = searcher.SearchWithTypoCorrection(query, indexer.IndexName)
					if err == nil && len(results) > 0 {
						searchMethod = "오타 교정"
						log.Printf("✅ 오타 교정 검색 성공: %s → %s", query, correctedQuery)
					} else {
						// 5단계: Fuzzy 검색 (최후의 수단)
						results, err = searcher.FuzzySearch(query, indexer.IndexName)
						correctedQuery = query
						searchMethod = "유사 검색"
						log.Printf("✅ Fuzzy 검색 시도")
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"query":           query,
			"corrected_query": correctedQuery,
			"search_method":   searchMethod,
			"total":           len(results),
			"results":         results,
		})
	})

	// 자동완성 API
	router.GET("/autocomplete", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" || len(query) < 2 {
			c.JSON(http.StatusOK, gin.H{
				"suggestions": []string{},
			})
			return
		}

		log.Printf("🔍 자동완성 요청: %s", query)

		// 간단한 검색으로 결과 가져오기
		results, err := searcher.Search(query, indexer.IndexName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"suggestions": []string{},
			})
			return
		}

		// 제목에서 추천어 추출 (최대 5개)
		suggestions := []string{}
		seen := make(map[string]bool)

		for _, doc := range results {
			if len(suggestions) >= 5 {
				break
			}
			if !seen[doc.Title] {
				suggestions = append(suggestions, doc.Title)
				seen[doc.Title] = true
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"query":       query,
			"suggestions": suggestions,
		})
	})

	// 서버 시작
	log.Println("✅ API 서버 실행 중: http://localhost:8080")
	log.Println("📍 API 엔드포인트:")
	log.Println("   - GET /health")
	log.Println("   - GET /search?q=검색어")
	log.Println("   - GET /search/fuzzy?q=검색어")
	log.Println("   - GET /search/smart?q=검색어 ⭐ 추천!")
	log.Println("   - GET /autocomplete?q=검색어 🎯 자동완성")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}

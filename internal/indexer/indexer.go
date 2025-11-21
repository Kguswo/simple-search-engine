package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"simple-search-engine/internal/models"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Indexer struct {
	client *elasticsearch.Client
}

func NewIndexer(client *elasticsearch.Client) *Indexer {
	return &Indexer{client: client}
}

// sanitizeID는 Elasticsearch Document ID로 사용 가능하도록 정리
func sanitizeID(title string) string {
	// 소문자 변환
	id := strings.ToLower(title)

	// 공백을 언더스코어로
	id = strings.ReplaceAll(id, " ", "_")

	// 개행문자, 탭 제거
	id = strings.ReplaceAll(id, "\n", "")
	id = strings.ReplaceAll(id, "\r", "")
	id = strings.ReplaceAll(id, "\t", "")

	// 특수문자 제거
	id = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-' {
			return r
		}
		return -1
	}, id)

	// 연속된 언더스코어 제거
	for strings.Contains(id, "__") {
		id = strings.ReplaceAll(id, "__", "_")
	}

	// 앞뒤 언더스코어 제거
	id = strings.Trim(id, "_")

	// 최대 길이 제한 (255자)
	if len(id) > 255 {
		id = id[:255]
	}

	return id
}

func (i *Indexer) CreateIndex() error {
	indexName := "tech_blogs"

	// 인덱스 존재 확인
	res, err := i.client.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("인덱스 존재 확인 실패: %w", err)
	}
	defer res.Body.Close()

	// 이미 존재하면 삭제
	if res.StatusCode == 200 {
		log.Println("⚠️  인덱스 'tech_blogs'가 이미 존재합니다. 삭제 후 재생성합니다.")
		delRes, err := i.client.Indices.Delete([]string{indexName})
		if err != nil {
			return fmt.Errorf("인덱스 삭제 실패: %w", err)
		}
		defer delRes.Body.Close()
	}

	// 한국어 분석기 설정
	mapping := `{
		"settings": {
			"analysis": {
				"analyzer": {
					"korean": {
						"type": "custom",
						"tokenizer": "nori_tokenizer",
						"filter": ["lowercase"]
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"title": {
					"type": "text",
					"analyzer": "korean"
				},
				"content": {
					"type": "text",
					"analyzer": "korean"
				},
				"author": {
					"type": "keyword"
				},
				"published_date": {
					"type": "date"
				},
				"url": {
					"type": "keyword"
				},
				"tags": {
					"type": "keyword"
				}
			}
		}
	}`

	createRes, err := i.client.Indices.Create(
		indexName,
		i.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return fmt.Errorf("인덱스 생성 실패: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("인덱스 생성 에러: %s", createRes.String())
	}

	log.Println("✅ 인덱스 'tech_blogs' 생성 완료 (한국어 분석기 Nori 적용)")
	return nil
}

func (i *Indexer) IndexDocument(doc models.Document) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("JSON 변환 실패: %w", err)
	}

	// 제목을 Document ID로 사용 (정리 후)
	docID := sanitizeID(doc.Title)

	req := esapi.IndexRequest{
		Index:      "tech_blogs",
		DocumentID: docID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), i.client)
	if err != nil {
		return fmt.Errorf("문서 인덱싱 실패: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("문서 인덱싱 에러: %s", res.String())
	}

	return nil
}

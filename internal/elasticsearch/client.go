package elasticsearch

import (
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

// NewClient creates a new Elasticsearch client
func NewClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch 클라이언트 생성 실패: %w", err)
	}

	// 연결 테스트
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch 연결 실패: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch 응답 에러: %s", res.String())
	}

	log.Println("✅ Elasticsearch 연결 성공!")
	return client, nil
}

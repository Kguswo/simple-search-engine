package main

import (
	"log"

	// "github.com/Kguswo/simple-search-engine/internal/elasticsearch"
	// "github.com/Kguswo/simple-search-engine/internal/indexer"
	"simple-search-engine/internal/elasticsearch"
	"simple-search-engine/internal/indexer"
)

func main() {
	log.Println("🔍 SmartSearch 데이터 인덱싱 시작...")

	// 1. Elasticsearch 클라이언트 초기화
	client, err := elasticsearch.NewClient()
	if err != nil {
		log.Fatalf("Elasticsearch 클라이언트 생성 실패: %v", err)
	}

	// 2. Indexer 생성
	idx := indexer.NewIndexer(client)

	// 3. 인덱스 생성 (한국어 분석기 Nori 적용)
	log.Println("인덱스 생성 중...")
	if err := idx.CreateIndex(); err != nil {
		log.Fatalf("인덱스 생성 실패: %v", err)
	}

	// 4. 샘플 데이터 인덱싱
	log.Println("샘플 데이터 인덱싱 중...")
	if err := idx.IndexDocuments("data/sample_documents.json"); err != nil {
		log.Fatalf("데이터 인덱싱 실패: %v", err)
	}

	log.Println("\n✅ 모든 작업이 완료되었습니다!")
	log.Println("Kibana에서 확인: http://localhost:5601")
	log.Println("Elasticsearch: http://localhost:9200")
}

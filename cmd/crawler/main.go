package main

import (
	"log"

	"simple-search-engine/internal/crawler"
	"simple-search-engine/internal/elasticsearch"
	"simple-search-engine/internal/indexer"
)

func main() {
	log.Println("🕷️ Velog RSS 크롤링 시작...")

	// Velog RSS 크롤러 초기화
	velogCrawler := crawler.NewVelogRSSCrawler()

	// 공개 트렌딩 피드에서 크롤링
	documents, err := velogCrawler.GetPublicTrending(1000)
	if err != nil {
		log.Fatalf("크롤링 실패: %v", err)
	}

	log.Printf("📄 총 %d개의 포스트를 크롤링했습니다.", len(documents))

	// Elasticsearch 클라이언트 초기화
	client, err := elasticsearch.NewClient()
	if err != nil {
		log.Fatalf("Elasticsearch 연결 실패: %v", err)
	}

	// Indexer 초기화
	idx := indexer.NewIndexer(client)

	// 인덱스 생성 (이미 있으면 스킵)
	if err := idx.CreateIndex(); err != nil {
		log.Printf("⚠️ 인덱스 생성 중 오류: %v", err)
	}

	// 크롤링한 데이터 인덱싱
	log.Println("📊 데이터 인덱싱 중...")
	successCount := 0
	for _, doc := range documents {
		if err := idx.IndexDocument(doc); err != nil {
			log.Printf("❌ 인덱싱 실패: %s - %v", doc.Title, err)
		} else {
			successCount++
		}
	}

	log.Printf("✅ 인덱싱 완료: %d개 성공\n", successCount)
}

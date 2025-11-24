package main

import (
	"log"

	"simple-search-engine/internal/crawler"
	"simple-search-engine/internal/elasticsearch"
	"simple-search-engine/internal/indexer"
)

func main() {
	log.Println("🌐 글로벌 IT 기업 테크블로그 크롤링 시작...")
	log.Println("📚 FAANG + 글로벌 IT 기업 + 한국 IT 기업")
	log.Println("")

	// 테크블로그 크롤러 초기화
	techCrawler := crawler.NewTechBlogCrawler()

	// 모든 테크블로그 크롤링
	documents, err := techCrawler.GetAllTechBlogs()
	if err != nil {
		log.Fatalf("크롤링 실패: %v", err)
	}

	if len(documents) == 0 {
		log.Println("⚠️  크롤링된 문서가 없습니다.")
		return
	}

	log.Printf("\n📄 총 %d개의 기술 블로그 글을 크롤링했습니다.", len(documents))

	// Elasticsearch 클라이언트 초기화
	client, err := elasticsearch.NewClient()
	if err != nil {
		log.Fatalf("Elasticsearch 연결 실패: %v", err)
	}

	// Indexer 초기화
	idx := indexer.NewIndexer(client)

	// 인덱스 생성 (이미 있으면 재생성)
	if err := idx.CreateIndex(); err != nil {
		log.Printf("⚠️ 인덱스 생성 중 오류: %v", err)
	}

	// 크롤링한 데이터 인덱싱
	log.Println("\n📊 데이터 인덱싱 중...")
	successCount := 0
	failCount := 0

	for i, doc := range documents {
		if err := idx.IndexDocument(doc); err != nil {
			log.Printf("❌ [%d/%d] 인덱싱 실패: %s", i+1, len(documents), doc.Title)
			failCount++
		} else {
			successCount++
			if successCount%100 == 0 {
				log.Printf("⏳ 진행중... %d개 완료 (%d개 실패)", successCount, failCount)
			}
		}
	}

	log.Printf("\n✅ 인덱싱 완료!")
	log.Printf("   - 성공: %d개", successCount)
	log.Printf("   - 실패: %d개", failCount)
	log.Printf("   - 총합: %d개\n", len(documents))

	log.Println("📚 크롤링된 블로그:")
	log.Println("   [FAANG]")
	log.Println("   Meta, Apple, Netflix, Google")
	log.Println("")
	log.Println("   [글로벌 IT]")
	log.Println("   Microsoft, Uber, Airbnb, Spotify, LinkedIn,")
	log.Println("   Twitter, Dropbox, GitHub, Slack, Pinterest,")
	log.Println("   Lyft, Stripe, Shopify, Atlassian, Reddit, Cloudflare")
	log.Println("")
	log.Println("   [한국 IT]")
	log.Println("   우아한형제들, 카카오, 네이버, 라인, 토스, 당근마켓,")
	log.Println("   뱅크샐러드, 무신사, 컬리, 쿠팡, 29CM, 야놀자,")
	log.Println("   왓챠, 번개장터, 하이퍼커넥트, NHN Cloud")
}

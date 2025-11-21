package crawler

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"simple-search-engine/internal/models"

	"github.com/mmcdole/gofeed"
)

type VelogRSSCrawler struct {
	parser *gofeed.Parser
}

func NewVelogRSSCrawler() *VelogRSSCrawler {
	return &VelogRSSCrawler{
		parser: gofeed.NewParser(),
	}
}

// stripHTML removes HTML tags and returns plain text
func stripHTML(html string) string {
	// HTML 태그 제거
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	// HTML 엔티티 디코딩
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")

	// 연속된 공백 제거
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// 앞뒤 공백 제거
	return strings.TrimSpace(text)
}

// GetPublicTrending fetches public trending posts from Velog
func (vc *VelogRSSCrawler) GetPublicTrending(limit int) ([]models.Document, error) {
	log.Println("🔍 Velog 공개 트렌딩 피드 크롤링 시작...")

	// Velog 공개 RSS 피드
	rssURL := "https://v2.velog.io/rss"

	feed, err := vc.parser.ParseURL(rssURL)
	if err != nil {
		return nil, fmt.Errorf("❌ RSS 파싱 실패: %v", err)
	}

	// limit 개수만큼 가져오기
	if len(feed.Items) < limit {
		limit = len(feed.Items)
	}

	var allDocuments []models.Document

	for i := 0; i < limit; i++ {
		item := feed.Items[i]

		var publishedDate time.Time
		if item.PublishedParsed != nil {
			publishedDate = *item.PublishedParsed
		} else {
			publishedDate = time.Now()
		}

		// 태그 추출
		tags := []string{}
		for _, category := range item.Categories {
			tags = append(tags, category)
		}

		// HTML 제거하고 순수 텍스트만
		cleanContent := stripHTML(item.Description)

		// 150자로 제한
		if len(cleanContent) > 150 {
			cleanContent = cleanContent[:150] + "..."
		}

		// Author 추출 (링크에서)
		author := extractAuthorFromLink(item.Link)

		doc := models.Document{
			Title:         item.Title,
			Content:       cleanContent,
			Author:        author,
			PublishedDate: publishedDate,
			URL:           item.Link,
			Tags:          tags,
		}

		allDocuments = append(allDocuments, doc)
		log.Printf("✅ 크롤링: %s", item.Title)
	}

	log.Printf("📊 총 %d개의 포스트 크롤링 완료", len(allDocuments))
	return allDocuments, nil
}

// extractAuthorFromLink extracts username from Velog URL
// https://velog.io/@username/post-title -> username
func extractAuthorFromLink(link string) string {
	parts := strings.Split(link, "/@")
	if len(parts) < 2 {
		return "Unknown"
	}

	userPart := strings.Split(parts[1], "/")
	if len(userPart) > 0 {
		return userPart[0]
	}

	return "Unknown"
}

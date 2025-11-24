package crawler

import (
	"log"
	"simple-search-engine/internal/models"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type TechBlogCrawler struct {
	parser *gofeed.Parser
}

func NewTechBlogCrawler() *TechBlogCrawler {
	return &TechBlogCrawler{
		parser: gofeed.NewParser(),
	}
}

// GetAllTechBlogs - 글로벌 + 한국 IT 기업 테크블로그 크롤링
func (tc *TechBlogCrawler) GetAllTechBlogs() ([]models.Document, error) {
	log.Println("🚀 글로벌 IT 기업 테크블로그 크롤링 시작...")

	// FAANG + 글로벌 IT 기업 + 한국 IT 기업
	rssFeeds := map[string]string{
		// FAANG
		"Meta Engineering":       "https://engineering.fb.com/feed/",
		"Apple Machine Learning": "https://machinelearning.apple.com/rss.xml",
		"Netflix Tech Blog":      "https://netflixtechblog.com/feed",
		"Google Developers":      "https://developers.googleblog.com/feeds/posts/default",

		// 글로벌 IT 기업
		"Microsoft DevBlogs":    "https://devblogs.microsoft.com/feed/",
		"Uber Engineering":      "https://www.uber.com/blog/engineering/rss/",
		"Airbnb Engineering":    "https://medium.com/feed/airbnb-engineering",
		"Spotify Engineering":   "https://engineering.atspotify.com/feed/",
		"LinkedIn Engineering":  "https://www.linkedin.com/blog/engineering/feed",
		"Twitter Engineering":   "https://blog.twitter.com/engineering/en_us/blog.rss",
		"Dropbox Tech":          "https://dropbox.tech/feed",
		"GitHub Engineering":    "https://github.blog/engineering.atom",
		"Slack Engineering":     "https://slack.engineering/feed/",
		"Pinterest Engineering": "https://medium.com/feed/pinterest-engineering",
		"Lyft Engineering":      "https://eng.lyft.com/feed",
		"Stripe Engineering":    "https://stripe.com/blog/feed.rss",
		"Shopify Engineering":   "https://shopify.engineering/blog.atom",
		"Atlassian Engineering": "https://www.atlassian.com/blog/rss",
		"Reddit Engineering":    "https://www.redditinc.com/blog?format=rss",
		"Cloudflare Blog":       "https://blog.cloudflare.com/rss/",

		// 한국 IT 기업
		"우아한형제들":    "https://techblog.woowahan.com/feed/",
		"카카오":       "https://tech.kakao.com/feed/",
		"네이버D2":     "https://d2.naver.com/d2.atom",
		"라인":        "https://engineering.linecorp.com/ko/feed/",
		"토스":        "https://toss.tech/rss.xml",
		"당근마켓":      "https://medium.com/feed/daangn",
		"뱅크샐러드":     "https://blog.banksalad.com/rss.xml",
		"무신사":       "https://medium.com/feed/musinsa-tech",
		"컬리":        "https://helloworld.kurly.com/feed.xml",
		"29CM":      "https://medium.com/feed/29cm",
		"야놀자":       "https://medium.com/feed/yanolja",
		"왓챠":        "https://medium.com/feed/watcha",
		"번개장터":      "https://medium.com/feed/bunjang-tech-blog",
		"하이퍼커넥트":    "https://hyperconnect.github.io/feed.xml",
		"NHN Cloud": "https://meetup.nhncloud.com/rss",
		"쿠팡":        "https://medium.com/feed/coupang-engineering",
	}

	allDocuments := make([]models.Document, 0)
	seenURLs := make(map[string]bool)

	successCount := 0
	failCount := 0

	for blogName, rssURL := range rssFeeds {
		log.Printf("📡 [%s] 크롤링 중...", blogName)

		feed, err := tc.parser.ParseURL(rssURL)
		if err != nil {
			log.Printf("⚠️  [%s] RSS 파싱 실패: %v", blogName, err)
			failCount++
			continue
		}

		count := 0
		// 피드의 모든 항목 가져오기
		for _, item := range feed.Items {
			// 중복 체크
			if seenURLs[item.Link] {
				continue
			}
			seenURLs[item.Link] = true

			var publishedDate time.Time
			if item.PublishedParsed != nil {
				publishedDate = *item.PublishedParsed
			} else {
				publishedDate = time.Now()
			}

			// 태그
			tags := []string{blogName}
			for _, category := range item.Categories {
				tags = append(tags, category)
			}

			// 내용 정리
			content := stripHTML(item.Description)
			if content == "" {
				content = stripHTML(item.Content)
			}
			if len(content) > 200 {
				content = content[:200] + "..."
			}

			// 작성자
			author := blogName // 회사명을 작성자로
			if item.Author != nil && item.Author.Name != "" {
				author = blogName + " - " + item.Author.Name
			}

			doc := models.Document{
				Title:         strings.TrimSpace(item.Title),
				Content:       strings.TrimSpace(content),
				Author:        author,
				PublishedDate: publishedDate,
				URL:           item.Link,
				Tags:          tags,
			}

			allDocuments = append(allDocuments, doc)
			count++
		}

		log.Printf("✅ [%s] %d개 크롤링", blogName, count)
		successCount++
		time.Sleep(300 * time.Millisecond) // Rate limiting
	}

	log.Printf("\n🎉 크롤링 완료!")
	log.Printf("   - 성공: %d개 블로그", successCount)
	log.Printf("   - 실패: %d개 블로그", failCount)
	log.Printf("   - 총 글: %d개 (중복 제거)", len(allDocuments))

	return allDocuments, nil
}

// stripHTML - HTML 태그 제거
func stripHTML(html string) string {
	// 간단한 태그 제거
	text := strings.ReplaceAll(html, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<p>", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n")

	// 모든 HTML 태그 제거
	var result strings.Builder
	inTag := false
	for _, char := range text {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	// HTML 엔티티 디코딩
	finalText := result.String()
	finalText = strings.ReplaceAll(finalText, "&nbsp;", " ")
	finalText = strings.ReplaceAll(finalText, "&lt;", "<")
	finalText = strings.ReplaceAll(finalText, "&gt;", ">")
	finalText = strings.ReplaceAll(finalText, "&amp;", "&")
	finalText = strings.ReplaceAll(finalText, "&quot;", "\"")
	finalText = strings.ReplaceAll(finalText, "&#39;", "'")

	// 연속 공백/줄바꿈 정리
	lines := strings.Split(finalText, "\n")
	cleanLines := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, " ")
}

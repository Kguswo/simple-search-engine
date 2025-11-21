package models

import "time"

// Document는 기술 블로그 글을 나타낸다.
type Document struct {
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Author        string    `json:"author"`
	PublishedDate time.Time `json:"published_date"`
	URL           string    `json:"url"`
	Tags          []string  `json:"tags"`
}

package models

import "time"

// Document represents a tech blog article
type Document struct {
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Author        string    `json:"author"`
	PublishedDate time.Time `json:"published_date"`
	URL           string    `json:"url"`
	Tags          []string  `json:"tags"`
}

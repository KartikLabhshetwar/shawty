package domain

import "time"

// URL defines the structure for storing URL information.
type URL struct {
	ID           string    `json:"id" bson:"_id"` // Unique identifier, also the short URL
	OriginalUrl  string    `json:"original_url" bson:"original_url"`
	ShortUrl     string    `json:"short_url" bson:"short_url"` // Redundant if ID is the short URL, but kept for clarity from original
	CreationDate time.Time `json:"creation_date" bson:"creation_date"`
}

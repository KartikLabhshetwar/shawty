package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"shawty/internal/domain"
	"shawty/internal/store"
)

// UrlServiceInterface defines operations for URL management.
type UrlServiceInterface interface {
	CreateShortURL(ctx context.Context, originalURL string) (domain.URL, error)
	GetOriginalURL(ctx context.Context, shortID string) (string, error)
}

// UrlService implements UrlServiceInterface.
type UrlService struct {
	urlStore store.UrlStoreInterface
}

// NewUrlService creates a new UrlService.
func NewUrlService(s store.UrlStoreInterface) *UrlService {
	return &UrlService{urlStore: s}
}

// generateShortID creates a short identifier from the original URL.
// This implementation uses MD5 and takes the first 8 characters.
func generateShortID(originalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalURL))
	data := hasher.Sum(nil)
	hash := hex.EncodeToString(data)
	return hash[:8]
}

// CreateShortURL generates a short URL for the given original URL and saves it.
func (s *UrlService) CreateShortURL(ctx context.Context, originalURL string) (domain.URL, error) {
	if originalURL == "" {
		return domain.URL{}, fmt.Errorf("original URL cannot be empty")
	}

	shortID := generateShortID(originalURL)

	urlEntry := domain.URL{
		ID:           shortID, // Using generated shortID as the MongoDB _id
		OriginalUrl:  originalURL,
		ShortUrl:     shortID, // Storing shortID also in ShortUrl field for clarity/flexibility
		CreationDate: time.Now().UTC(),
	}

	err := s.urlStore.Save(ctx, urlEntry)
	if err != nil {
		// If it's a duplicate key error, we might want to fetch the existing one
		// or return a specific error indicating duplication.
		// For now, just propagate the error.
		return domain.URL{}, fmt.Errorf("failed to save URL: %w", err)
	}
	return urlEntry, nil
}

// GetOriginalURL retrieves the original URL for a given short ID.
func (s *UrlService) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	if shortID == "" {
		return "", fmt.Errorf("short ID cannot be empty")
	}
	url, err := s.urlStore.GetByShortID(ctx, shortID)
	if err != nil {
		return "", err
	}
	return url.OriginalUrl, nil
}

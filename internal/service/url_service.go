package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"shawty/internal/domain"
	"shawty/internal/store"
)

// ErrHashCollision is returned when two different original URLs generate the same short ID.
var ErrHashCollision = errors.New("hash collision detected")

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
// If the original URL has already been shortened, it returns the existing short URL.
// It returns ErrHashCollision if a different original URL generates the same short ID.
func (s *UrlService) CreateShortURL(ctx context.Context, originalURL string) (domain.URL, error) {
	if originalURL == "" {
		return domain.URL{}, fmt.Errorf("original URL cannot be empty")
	}

	shortID := generateShortID(originalURL)

	urlToSave := domain.URL{
		ID:           shortID,
		OriginalUrl:  originalURL,
		ShortUrl:     shortID,
		CreationDate: time.Now().UTC(),
	}

	err := s.urlStore.Save(ctx, urlToSave)
	if err == nil {
		// Successfully saved a new entry
		return urlToSave, nil
	}

	// Handle error from Save
	if errors.Is(err, store.ErrDuplicateShortID) {
		// The shortID already exists, fetch the existing entry
		existingURL, getErr := s.urlStore.GetByShortID(ctx, shortID)
		if getErr == nil {
			// Successfully fetched the existing URL
			if existingURL.OriginalUrl == originalURL {
				// The original URLs match, so this is the same URL being submitted again
				return existingURL, nil
			}
			// Original URLs do not match: this is a hash collision
			return domain.URL{}, fmt.Errorf("%w: short ID '%s' generated for a different original URL (submitted: '%s', existing: '%s')", ErrHashCollision, shortID, originalURL, existingURL.OriginalUrl)
		}
		// Error fetching the existing URL after duplicate detection
		return domain.URL{}, fmt.Errorf("error retrieving existing URL for short ID '%s' after duplicate detection: %w", shortID, getErr)
	}

	// Some other error occurred during save
	return domain.URL{}, fmt.Errorf("failed to save URL: %w", err)
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

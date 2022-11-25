package services

import (
	"context"
)

// TrackServiceConfig
type TrackServiceConfig struct {

}

// TrackService
type TrackService struct {

}

// NewTrackService
func NewTrackService(c *TrackServiceConfig) *TrackService {
	return &TrackService{

	}
}

// WriteOpening
func (s *TrackService) WriteOpening(ctx context.Context, uid string) error {
	return nil
}

// GetImage
func (s *TrackService) GetImage(ctx context.Context) ([]byte, error) {
	
	result := []byte{}
	
	return result, nil
}

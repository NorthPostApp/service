package services

import (
	"context"
	"fmt"

	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
)

type musicRepository interface {
	GetPresignedMusicURL(ctx context.Context, opts repository.GetPresignedMusicURLOptions) (*repository.GetPresignedMusicURLResponse, error)
	RefreshMusicList(ctx context.Context) (*repository.RefreshMusicListResponse, error)
}

type MusicService struct {
	repo musicRepository
}

func NewMusicService(repo musicRepository) *MusicService {
	return &MusicService{repo: repo}
}

type RefreshMusicListOutput struct {
	Data []models.Music
}

func (s *MusicService) RefreshMusicList(
	ctx context.Context) (*RefreshMusicListOutput, error) {
	musicList, err := s.repo.RefreshMusicList(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &RefreshMusicListOutput{Data: musicList.Data}, nil
}

package services

import (
	"context"
	"errors"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMusicRepository struct {
	mock.Mock
}

func (m *mockMusicRepository) RefreshMusicList(ctx context.Context) (*repository.RefreshMusicListResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.RefreshMusicListResponse), nil
}

func (m *mockMusicRepository) GetAllMusicList(ctx context.Context) (*repository.GetAllMusicListResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.GetAllMusicListResponse), nil
}

func (m *mockMusicRepository) GetPresignedMusicURL(ctx context.Context, opts repository.GetPresignedMusicURLOptions) (
	*repository.GetPresignedMusicURLResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.GetPresignedMusicURLResponse), nil
}

func setupMusicService() (*MusicService, *mockMusicRepository) {
	repo := new(mockMusicRepository)
	service := NewMusicService(repo)
	return service, repo
}

func TestMusicService_RefreshMusicList(t *testing.T) {
	t.Parallel()
	services, repo := setupMusicService()
	ctx := context.Background()
	expectedOutput := &repository.RefreshMusicListResponse{
		Data: []models.Music{{
			Filename: "test",
		},
		},
	}
	repo.On("RefreshMusicList", mock.Anything, mock.Anything).Return(expectedOutput, nil).Once()
	output, err := services.RefreshMusicList(ctx)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, output.Data[0].Filename, expectedOutput.Data[0].Filename)
}

func TestMusicService_RefreshMusicList_Error(t *testing.T) {
	t.Parallel()
	services, repo := setupMusicService()
	ctx := context.Background()
	repo.On(
		"RefreshMusicList",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("data not found")).Once()
	output, err := services.RefreshMusicList(ctx)
	repo.AssertExpectations(t)
	assert.Nil(t, output)
	assert.Error(t, err)
}

func TestMusicService_GetAllMusicList(t *testing.T) {
	t.Parallel()
	services, repo := setupMusicService()
	ctx := context.Background()
	expectedOutput := &repository.GetAllMusicListResponse{
		Data: []models.Music{{
			Filename: "test",
		},
		},
	}
	repo.On(
		"GetAllMusicList",
		mock.Anything,
		mock.Anything,
	).Return(expectedOutput, nil).Once()
	output, err := services.GetAllMusicList(ctx)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, output.Data[0].Filename, expectedOutput.Data[0].Filename)
}

func TestMusicService_GetAllMusicList_Error(t *testing.T) {
	t.Parallel()
	services, repo := setupMusicService()
	ctx := context.Background()
	repo.On(
		"GetAllMusicList",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("data not found")).Once()
	output, err := services.GetAllMusicList(ctx)
	repo.AssertExpectations(t)
	assert.Nil(t, output)
	assert.Error(t, err)
}

func TestMusicService_GetPresignedMusicURL(t *testing.T) {
	t.Parallel()
	services, repo := setupMusicService()
	ctx := context.Background()
	input := GetPresignedMusicURLInput{Filename: "foo/bar"}
	expectedOutput := &repository.GetPresignedMusicURLResponse{URL: "foo/bar"}
	repo.On(
		"GetPresignedMusicURL",
		mock.Anything,
		mock.Anything,
	).Return(expectedOutput, nil)
	output, err := services.GetPresignedMusicURL(ctx, input)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, expectedOutput.URL, output.URL)
}

func TestMusicService_GetPresignedMusicURL_ERROR(t *testing.T) {
	t.Parallel()
	services, repo := setupMusicService()
	ctx := context.Background()
	input := GetPresignedMusicURLInput{Filename: "foo/bar"}
	repo.On(
		"GetPresignedMusicURL",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("failed"))
	output, err := services.GetPresignedMusicURL(ctx, input)
	repo.AssertExpectations(t)
	assert.NotNil(t, err)
	assert.Nil(t, output)
}

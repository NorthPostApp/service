package dto

import "north-post/service/internal/domain/v1/models"

type MusicDTO struct {
	Filename     string  `json:"filename"`
	Title        string  `json:"title"`
	Genre        string  `json:"genre"`
	Size         float64 `json:"size"`
	LastModified int64   `json:"lastModified"`
	DurationSec  int64   `json:"durationSec"`
}

type GetMusicListResponse struct {
	Data []MusicDTO `json:"data"`
}

func ToMusicDTO(music models.Music) MusicDTO {
	return MusicDTO{
		Filename:     music.Filename,
		Title:        music.Title,
		Genre:        music.Genre,
		Size:         music.Size,
		LastModified: music.LastModified,
		DurationSec:  music.DurationSec,
	}
}

func ToMusicDTOs(musics []models.Music) []MusicDTO {
	output := make([]MusicDTO, len(musics))
	for i, music := range musics {
		output[i] = ToMusicDTO(music)
	}
	return output
}

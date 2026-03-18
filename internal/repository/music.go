package repository

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"north-post/service/internal/domain/v1/models"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	musicBucketName     = "northpost-music"
	musicCollectionName = "music_list"
)

type MusicRepository struct {
	client          *s3.Client
	presignedClient *s3.PresignClient
	firestoreClient *firestore.Client
	logger          *slog.Logger
}

func NewMusicRepository(
	client *s3.Client,
	presignedClient *s3.PresignClient,
	firestoreClient *firestore.Client,
	logger *slog.Logger,
) *MusicRepository {
	return &MusicRepository{
		client:          client,
		presignedClient: presignedClient,
		firestoreClient: firestoreClient,
		logger:          logger,
	}
}

type GetPresignedMusicURLOptions struct {
	Filename string
}

type GetPresignedMusicURLResponse struct {
	URL string
}

type RefreshMusicListResponse struct {
	Data []models.Music
}

// Get the presigned music url with the expiration of 15 minutes
func (r *MusicRepository) GetPresignedMusicURL(
	ctx context.Context,
	opts GetPresignedMusicURLOptions) (*GetPresignedMusicURLResponse, error) {
	request, err := r.presignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(musicBucketName),
		Key:    aws.String(opts.Filename),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		r.logger.Error("failed to get music url", "error", err)
		return nil, fmt.Errorf("failed to get music url: %w", err)
	}
	return &GetPresignedMusicURLResponse{URL: request.URL}, nil
}

// Refresh he music list and store it in the database
func (r *MusicRepository) RefreshMusicList(
	ctx context.Context) (*RefreshMusicListResponse, error) {
	var musicList []models.Music
	paginator := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(musicBucketName),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			r.logger.Error("failed to list bucket objects", "error", err)
			return nil, fmt.Errorf("failed to list bucket objects: %w", err)
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			fileSize := float64(aws.ToInt64(obj.Size)) / (1024 * 1024)
			lastModified := aws.ToTime(obj.LastModified).UnixMilli()
			genre, title := splitMusicKey(key)
			// continue if either genre or title is missing
			if genre == "" || title == "" {
				continue
			}
			musicList = append(musicList, models.Music{
				Filename:     key,
				Title:        title,
				Genre:        genre,
				Size:         roundFilesize(fileSize, 2),
				LastModified: lastModified,
				DurationSec:  -1, // default unset, will update when user plays it the first time
			})
		}
	}
	// write data to database
	if err := r.updateMusicList(ctx, musicList); err != nil {
		return nil, err
	}
	return &RefreshMusicListResponse{Data: musicList}, nil
}

func (r *MusicRepository) updateMusicList(ctx context.Context, musicList []models.Music) error {
	identifierFields := []string{"size", "lastModified"}
	type identifier struct {
		Size         float64 `firestore:"size"`
		LastModified int64   `firestore:"lastModified"`
	}

	collection := r.firestoreClient.Collection(musicCollectionName)
	// fetch existing doc IDs only (with only size and lastModifies parts as identifiers)
	existingDocs, err := collection.Select(identifierFields...).Documents(ctx).GetAll()
	if err != nil {
		r.logger.Error("failed to get existing music documents", "error", err)
		return fmt.Errorf("failed to get existing music documents: %w", err)
	}
	existingDocsData := make(map[string]identifier, len(existingDocs))
	for _, doc := range existingDocs {
		var tempData identifier
		if err := doc.DataTo(&tempData); err != nil {
			r.logger.Error(
				"failed to decode music document",
				"docID",
				doc.Ref.ID,
				"error",
				err,
			)
			continue
		}
		existingDocsData[doc.Ref.ID] = tempData
	}

	newIDs := make(map[string]struct{}, len(musicList))
	filesAdded, filesUpdated, filesDeleted := 0, 0, 0
	bulkWriter := r.firestoreClient.BulkWriter(ctx)
	// add new docs or update existing docs
	for _, music := range musicList {
		docId := getDocId(music.Genre, music.Title)
		newIDs[docId] = struct{}{}
		if _, exists := existingDocsData[docId]; !exists {
			bulkWriter.Set(collection.Doc(docId), music)
			filesAdded += 1
		} else if music.Size != existingDocsData[docId].Size ||
			music.LastModified != existingDocsData[docId].LastModified {
			bulkWriter.Set(collection.Doc(docId), music)
			filesUpdated += 1
		}
	}
	// delete non-existing removed docs
	for _, doc := range existingDocs {
		if _, exists := newIDs[doc.Ref.ID]; !exists {
			bulkWriter.Delete(doc.Ref)
			filesDeleted += 1
		}
	}
	bulkWriter.Flush()
	r.logger.Info("Music list refresh completed: ", "added", filesAdded, "updated", filesUpdated, "deleted", filesDeleted)
	return nil
}

// =========== Helper functions ==========
func splitMusicKey(musicKey string) (string, string) {
	parts := strings.SplitN(musicKey, "/", 2)
	genre, filename := "", ""
	if len(parts) == 2 {
		genre, filename = parts[0], parts[1]
	} else {
		return genre, filename
	}
	title := strings.TrimSuffix(filename, ".mp3")
	return genre, title
}

func getDocId(genre string, title string) string {
	return fmt.Sprintf("%s_%s", genre, title)
}

func roundFilesize(size float64, precision uint) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(size*factor) / factor
}

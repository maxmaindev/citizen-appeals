package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"citizen-appeals/internal/models"
)

var (
	ErrPhotoNotFound = errors.New("photo not found")
)

type PhotoRepository struct {
	db *pgxpool.Pool
}

func NewPhotoRepository(db *pgxpool.Pool) *PhotoRepository {
	return &PhotoRepository{db: db}
}

// Create creates a new photo record
func (r *PhotoRepository) Create(ctx context.Context, photo *models.Photo) error {
	query := `
		INSERT INTO photos (appeal_id, comment_id, file_path, file_name, file_size, mime_type, is_result_photo)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, uploaded_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		photo.AppealID,
		photo.CommentID,
		photo.FilePath,
		photo.FileName,
		photo.FileSize,
		photo.MimeType,
		photo.IsResultPhoto,
	).Scan(&photo.ID, &photo.UploadedAt)

	if err != nil {
		return fmt.Errorf("failed to create photo: %w", err)
	}

	return nil
}

// GetByID retrieves a photo by ID
func (r *PhotoRepository) GetByID(ctx context.Context, id int64) (*models.Photo, error) {
	query := `
		SELECT id, appeal_id, comment_id, file_path, file_name, file_size, mime_type, is_result_photo, uploaded_at
		FROM photos
		WHERE id = $1
	`

	var photo models.Photo
	err := r.db.QueryRow(ctx, query, id).Scan(
		&photo.ID,
		&photo.AppealID,
		&photo.CommentID,
		&photo.FilePath,
		&photo.FileName,
		&photo.FileSize,
		&photo.MimeType,
		&photo.IsResultPhoto,
		&photo.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPhotoNotFound
		}
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}

	return &photo, nil
}

// GetByAppealID retrieves all photos for an appeal
func (r *PhotoRepository) GetByAppealID(ctx context.Context, appealID int64) ([]*models.Photo, error) {
	query := `
		SELECT id, appeal_id, comment_id, file_path, file_name, file_size, mime_type, is_result_photo, uploaded_at
		FROM photos
		WHERE appeal_id = $1
		ORDER BY uploaded_at ASC
	`

	rows, err := r.db.Query(ctx, query, appealID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}
	defer rows.Close()

	photos := make([]*models.Photo, 0)
	for rows.Next() {
		var photo models.Photo
		err := rows.Scan(
			&photo.ID,
			&photo.AppealID,
			&photo.CommentID,
			&photo.FilePath,
			&photo.FileName,
			&photo.FileSize,
			&photo.MimeType,
			&photo.IsResultPhoto,
			&photo.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		photos = append(photos, &photo)
	}

	return photos, nil
}

// GetByCommentID retrieves all photos for a comment
func (r *PhotoRepository) GetByCommentID(ctx context.Context, commentID int64) ([]*models.Photo, error) {
	query := `
		SELECT id, appeal_id, comment_id, file_path, file_name, file_size, mime_type, is_result_photo, uploaded_at
		FROM photos
		WHERE comment_id = $1
		ORDER BY uploaded_at ASC
	`

	rows, err := r.db.Query(ctx, query, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}
	defer rows.Close()

	photos := make([]*models.Photo, 0)
	for rows.Next() {
		var photo models.Photo
		err := rows.Scan(
			&photo.ID,
			&photo.AppealID,
			&photo.CommentID,
			&photo.FilePath,
			&photo.FileName,
			&photo.FileSize,
			&photo.MimeType,
			&photo.IsResultPhoto,
			&photo.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		photos = append(photos, &photo)
	}

	return photos, nil
}

// CountByAppealID counts photos for an appeal
func (r *PhotoRepository) CountByAppealID(ctx context.Context, appealID int64) (int, error) {
	query := `SELECT COUNT(*) FROM photos WHERE appeal_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, appealID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count photos: %w", err)
	}

	return count, nil
}

// Delete deletes a photo record
func (r *PhotoRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM photos WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrPhotoNotFound
	}

	return nil
}

// DeleteByAppealID deletes all photos for an appeal
func (r *PhotoRepository) DeleteByAppealID(ctx context.Context, appealID int64) error {
	query := `DELETE FROM photos WHERE appeal_id = $1`

	_, err := r.db.Exec(ctx, query, appealID)
	if err != nil {
		return fmt.Errorf("failed to delete photos: %w", err)
	}

	return nil
}


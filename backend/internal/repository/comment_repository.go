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
	ErrCommentNotFound = errors.New("comment not found")
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (appeal_id, user_id, text, is_internal)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		comment.AppealID,
		comment.UserID,
		comment.Text,
		comment.IsInternal,
	).Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetByID retrieves a comment by ID with user information
func (r *CommentRepository) GetByID(ctx context.Context, id int64) (*models.Comment, error) {
	query := `
		SELECT c.id, c.appeal_id, c.user_id, c.text, c.is_internal, c.created_at,
		       u.id, u.email, u.first_name, u.last_name, u.phone, u.role
		FROM comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`

	var comment models.Comment
	var user models.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&comment.ID, &comment.AppealID, &comment.UserID, &comment.Text,
		&comment.IsInternal, &comment.CreatedAt,
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Role,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	comment.User = &user
	return &comment, nil
}

// GetByAppealID retrieves all comments for an appeal
func (r *CommentRepository) GetByAppealID(ctx context.Context, appealID int64, userID int64, userRole models.UserRole) ([]*models.Comment, error) {
	// Citizens can only see non-internal comments
	// Dispatchers, executors, and admins can see all comments
	var query string
	var args []interface{}

	if userRole == models.RoleCitizen {
		query = `
			SELECT c.id, c.appeal_id, c.user_id, c.text, c.is_internal, c.created_at,
			       u.id, u.email, u.first_name, u.last_name, u.phone, u.role
			FROM comments c
			LEFT JOIN users u ON c.user_id = u.id
			WHERE c.appeal_id = $1 AND c.is_internal = false
			ORDER BY c.created_at ASC
		`
		args = []interface{}{appealID}
	} else {
		query = `
			SELECT c.id, c.appeal_id, c.user_id, c.text, c.is_internal, c.created_at,
			       u.id, u.email, u.first_name, u.last_name, u.phone, u.role
			FROM comments c
			LEFT JOIN users u ON c.user_id = u.id
			WHERE c.appeal_id = $1
			ORDER BY c.created_at ASC
		`
		args = []interface{}{appealID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	comments := make([]*models.Comment, 0)
	for rows.Next() {
		var comment models.Comment
		var user models.User

		err := rows.Scan(
			&comment.ID, &comment.AppealID, &comment.UserID, &comment.Text,
			&comment.IsInternal, &comment.CreatedAt,
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		comment.User = &user
		comments = append(comments, &comment)
	}

	return comments, nil
}

// Update updates a comment
func (r *CommentRepository) Update(ctx context.Context, comment *models.Comment) error {
	query := `
		UPDATE comments
		SET text = $1, is_internal = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(
		ctx,
		query,
		comment.Text,
		comment.IsInternal,
		comment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}

// Delete deletes a comment
func (r *CommentRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}


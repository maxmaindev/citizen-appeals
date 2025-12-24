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
	ErrCategoryNotFound = errors.New("category not found")
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (name, description, default_priority, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		category.Name,
		category.Description,
		category.DefaultPriority,
		category.IsActive,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	query := `
		SELECT id, name, description, default_priority, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var category models.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.DefaultPriority,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// List retrieves all active categories
func (r *CategoryRepository) List(ctx context.Context, includeInactive bool) ([]*models.Category, error) {
	query := `
		SELECT id, name, description, default_priority, is_active, created_at, updated_at
		FROM categories
		WHERE is_active = true OR $1
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*models.Category, 0)
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.DefaultPriority,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &category)
	}

	return categories, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, default_priority = $3,
		    is_active = $4, updated_at = NOW()
		WHERE id = $5
	`

	result, err := r.db.Exec(
		ctx,
		query,
		category.Name,
		category.Description,
		category.DefaultPriority,
		category.IsActive,
		category.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// Delete soft deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE categories SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

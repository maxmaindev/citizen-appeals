package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"citizen-appeals/internal/models"
)

var (
	ErrCategoryServiceNotFound = errors.New("category service not found")
)

type CategoryServiceRepository struct {
	db *pgxpool.Pool
}

func NewCategoryServiceRepository(db *pgxpool.Pool) *CategoryServiceRepository {
	return &CategoryServiceRepository{db: db}
}

// GetByCategoryID retrieves all services assigned to a category
func (r *CategoryServiceRepository) GetByCategoryID(ctx context.Context, categoryID int64) ([]*models.CategoryService, error) {
	query := `
		SELECT cs.id, cs.category_id, cs.service_id, cs.created_at,
		       s.id, s.name, s.description, s.contact_person, s.contact_phone, s.contact_email, s.is_active
		FROM category_services cs
		JOIN services s ON cs.service_id = s.id
		WHERE cs.category_id = $1 AND s.is_active = true
		ORDER BY cs.id ASC
	`

	rows, err := r.db.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category services: %w", err)
	}
	defer rows.Close()

	services := make([]*models.CategoryService, 0)
	for rows.Next() {
		var cs models.CategoryService
		var service models.Service

		err := rows.Scan(
			&cs.ID, &cs.CategoryID, &cs.ServiceID, &cs.CreatedAt,
			&service.ID, &service.Name, &service.Description, &service.ContactPerson,
			&service.ContactPhone, &service.ContactEmail, &service.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category service: %w", err)
		}

		cs.Service = &service
		services = append(services, &cs)
	}

	return services, nil
}

// GetByServiceID retrieves all categories assigned to a service
func (r *CategoryServiceRepository) GetByServiceID(ctx context.Context, serviceID int64) ([]*models.CategoryService, error) {
	query := `
		SELECT cs.id, cs.category_id, cs.service_id, cs.created_at,
		       c.id, c.name, c.description, c.default_priority, c.is_active
		FROM category_services cs
		JOIN categories c ON cs.category_id = c.id
		WHERE cs.service_id = $1 AND c.is_active = true
		ORDER BY cs.id ASC
	`

	rows, err := r.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*models.CategoryService, 0)
	for rows.Next() {
		var cs models.CategoryService
		var category models.Category

		err := rows.Scan(
			&cs.ID, &cs.CategoryID, &cs.ServiceID, &cs.CreatedAt,
			&category.ID, &category.Name, &category.Description, &category.DefaultPriority, &category.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category service: %w", err)
		}

		cs.Category = &category
		categories = append(categories, &cs)
	}

	return categories, nil
}

// GetAllWithDetails retrieves all category-service relationships with full details
func (r *CategoryServiceRepository) GetAllWithDetails(ctx context.Context) ([]*models.CategoryWithServices, error) {
	// First get all categories
	categoryRepo := NewCategoryRepository(r.db)
	categories, err := categoryRepo.List(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	result := make([]*models.CategoryWithServices, 0, len(categories))
	for _, category := range categories {
		services, err := r.GetByCategoryID(ctx, category.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get services for category %d: %w", category.ID, err)
		}

		result = append(result, &models.CategoryWithServices{
			Category: category,
			Services: services,
		})
	}

	return result, nil
}

// AssignServices assigns services to a category (replaces existing assignments)
func (r *CategoryServiceRepository) AssignServices(ctx context.Context, categoryID int64, serviceIDs []int64) error {
	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing assignments
	_, err = tx.Exec(ctx, "DELETE FROM category_services WHERE category_id = $1", categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete existing assignments: %w", err)
	}

	// Insert new assignments (order is preserved by insertion order, id will be sequential)
	if len(serviceIDs) > 0 {
		query := `
			INSERT INTO category_services (category_id, service_id)
			VALUES ($1, $2)
		`

		for _, serviceID := range serviceIDs {
			_, err = tx.Exec(ctx, query, categoryID, serviceID)
			if err != nil {
				return fmt.Errorf("failed to assign service %d: %w", serviceID, err)
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Create creates a new category-service relationship
func (r *CategoryServiceRepository) Create(ctx context.Context, cs *models.CategoryService) error {
	query := `
		INSERT INTO category_services (category_id, service_id)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		cs.CategoryID,
		cs.ServiceID,
	).Scan(&cs.ID, &cs.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create category service: %w", err)
	}

	return nil
}

// Delete removes a category-service relationship
func (r *CategoryServiceRepository) Delete(ctx context.Context, categoryID, serviceID int64) error {
	query := `DELETE FROM category_services WHERE category_id = $1 AND service_id = $2`

	result, err := r.db.Exec(ctx, query, categoryID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete category service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCategoryServiceNotFound
	}

	return nil
}

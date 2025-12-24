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
	ErrServiceNotFound = errors.New("service not found")
)

type ServiceRepository struct {
	db *pgxpool.Pool
}

func NewServiceRepository(db *pgxpool.Pool) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// Create creates a new service
func (r *ServiceRepository) Create(ctx context.Context, service *models.Service) error {
	query := `
		INSERT INTO services (name, description, contact_person, contact_phone, contact_email, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		service.Name,
		service.Description,
		service.ContactPerson,
		service.ContactPhone,
		service.ContactEmail,
		service.IsActive,
	).Scan(&service.ID, &service.CreatedAt, &service.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	return nil
}

// GetByID retrieves a service by ID
func (r *ServiceRepository) GetByID(ctx context.Context, id int64) (*models.Service, error) {
	query := `
		SELECT id, name, description, contact_person, contact_phone, contact_email,
		       is_active, created_at, updated_at
		FROM services
		WHERE id = $1
	`

	var service models.Service
	err := r.db.QueryRow(ctx, query, id).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.ContactPerson,
		&service.ContactPhone,
		&service.ContactEmail,
		&service.IsActive,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

// List retrieves all active services
func (r *ServiceRepository) List(ctx context.Context, includeInactive bool) ([]*models.Service, error) {
	query := `
		SELECT id, name, description, contact_person, contact_phone, contact_email,
		       is_active, created_at, updated_at
		FROM services
		WHERE is_active = true OR $1
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	services := make([]*models.Service, 0)
	for rows.Next() {
		var service models.Service
		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.Description,
			&service.ContactPerson,
			&service.ContactPhone,
			&service.ContactEmail,
			&service.IsActive,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, &service)
	}

	return services, nil
}

// Update updates a service
func (r *ServiceRepository) Update(ctx context.Context, service *models.Service) error {
	query := `
		UPDATE services
		SET name = $1, description = $2, contact_person = $3, contact_phone = $4,
		    contact_email = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.Exec(
		ctx,
		query,
		service.Name,
		service.Description,
		service.ContactPerson,
		service.ContactPhone,
		service.ContactEmail,
		service.IsActive,
		service.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrServiceNotFound
	}

	return nil
}

// Delete soft deletes a service
func (r *ServiceRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE services SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrServiceNotFound
	}

	return nil
}

// GetByName retrieves a service by name (case-insensitive)
func (r *ServiceRepository) GetByName(ctx context.Context, name string) (*models.Service, error) {
	query := `
		SELECT id, name, description, contact_person, contact_phone, contact_email,
		       is_active, created_at, updated_at
		FROM services
		WHERE LOWER(name) = LOWER($1) AND is_active = true
	`

	var service models.Service
	err := r.db.QueryRow(ctx, query, name).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.ContactPerson,
		&service.ContactPhone,
		&service.ContactEmail,
		&service.IsActive,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("failed to get service by name: %w", err)
	}

	return &service, nil
}

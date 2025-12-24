package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"citizen-appeals/internal/models"
)

var (
	ErrUserServiceNotFound = errors.New("user service not found")
)

type UserServiceRepository struct {
	db *pgxpool.Pool
}

func NewUserServiceRepository(db *pgxpool.Pool) *UserServiceRepository {
	return &UserServiceRepository{db: db}
}

// GetByServiceID retrieves all users (executors) assigned to a service
func (r *UserServiceRepository) GetByServiceID(ctx context.Context, serviceID int64) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.first_name, u.last_name, u.phone, u.role, u.is_active, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_services us ON u.id = us.user_id
		WHERE us.service_id = $1 AND u.role = 'executor' AND u.is_active = true
		ORDER BY u.first_name, u.last_name
	`

	rows, err := r.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service users: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&user.Phone,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetByUserID retrieves all services assigned to a user
func (r *UserServiceRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Service, error) {
	query := `
		SELECT s.id, s.name, s.description, s.contact_person, s.contact_phone, s.contact_email, s.is_active, s.created_at, s.updated_at
		FROM services s
		INNER JOIN user_services us ON s.id = us.service_id
		WHERE us.user_id = $1 AND s.is_active = true
		ORDER BY s.name
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user services: %w", err)
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

// GetAllWithDetails retrieves all services with their assigned users
func (r *UserServiceRepository) GetAllWithDetails(ctx context.Context) ([]*models.ServiceWithUsers, error) {
	// First get all services
	serviceRepo := NewServiceRepository(r.db)
	services, err := serviceRepo.List(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	result := make([]*models.ServiceWithUsers, 0, len(services))
	for _, service := range services {
		users, err := r.GetByServiceID(ctx, service.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get users for service %d: %w", service.ID, err)
		}

		result = append(result, &models.ServiceWithUsers{
			Service: service,
			Users:   users,
		})
	}

	return result, nil
}

// AssignUsers assigns users (executors) to a service (replaces existing assignments)
func (r *UserServiceRepository) AssignUsers(ctx context.Context, serviceID int64, userIDs []int64) error {
	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing assignments
	_, err = tx.Exec(ctx, "DELETE FROM user_services WHERE service_id = $1", serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete existing assignments: %w", err)
	}

	// Insert new assignments
	if len(userIDs) > 0 {
		query := `
			INSERT INTO user_services (user_id, service_id)
			VALUES ($1, $2)
		`

		for _, userID := range userIDs {
			_, err = tx.Exec(ctx, query, userID, serviceID)
			if err != nil {
				return fmt.Errorf("failed to assign user %d: %w", userID, err)
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete removes a user-service relationship
func (r *UserServiceRepository) Delete(ctx context.Context, userID, serviceID int64) error {
	query := `DELETE FROM user_services WHERE user_id = $1 AND service_id = $2`

	result, err := r.db.Exec(ctx, query, userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete user service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserServiceNotFound
	}

	return nil
}

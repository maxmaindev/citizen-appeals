package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"citizen-appeals/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAppealNotFound = errors.New("appeal not found")
)

type AppealRepository struct {
	db *pgxpool.Pool
}

func NewAppealRepository(db *pgxpool.Pool) *AppealRepository {
	return &AppealRepository{db: db}
}

// getStringValue returns empty string if pointer is nil, otherwise the value
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Create creates a new appeal
func (r *AppealRepository) Create(ctx context.Context, appeal *models.Appeal) error {
	query := `
		INSERT INTO appeals (
			user_id, category_id, title, description, address,
			latitude, longitude, priority, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		appeal.UserID,
		appeal.CategoryID,
		appeal.Title,
		appeal.Description,
		appeal.Address,
		appeal.Latitude,
		appeal.Longitude,
		appeal.Priority,
		appeal.Status,
	).Scan(&appeal.ID, &appeal.CreatedAt, &appeal.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create appeal: %w", err)
	}

	return nil
}

// GetByID retrieves an appeal by ID with all related data
func (r *AppealRepository) GetByID(ctx context.Context, id int64) (*models.Appeal, error) {
	query := `
		SELECT
			a.id, a.user_id, a.category_id, a.service_id,
			a.status, a.title, a.description, a.address, a.latitude, a.longitude,
			a.priority, a.created_at, a.updated_at, a.closed_at,
			u.id, u.email, u.first_name, u.last_name, u.phone, u.role,
			c.id, c.name, c.description,
			s.id, s.name, s.description
		FROM appeals a
		LEFT JOIN users u ON a.user_id = u.id
		LEFT JOIN categories c ON a.category_id = c.id
		LEFT JOIN services s ON a.service_id = s.id
		WHERE a.id = $1
	`

	var appeal models.Appeal
	var user models.User

	var categoryID, serviceID *int64
	var categoryName, categoryDesc, serviceName, serviceDesc *string

	// Use nullable types for JOIN fields that can be NULL (category, service)
	var categoryIDVal, serviceIDVal *int64

	err := r.db.QueryRow(ctx, query, id).Scan(
		&appeal.ID, &appeal.UserID, &categoryID, &serviceID,
		&appeal.Status, &appeal.Title, &appeal.Description, &appeal.Address,
		&appeal.Latitude, &appeal.Longitude, &appeal.Priority,
		&appeal.CreatedAt, &appeal.UpdatedAt, &appeal.ClosedAt,
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Role,
		&categoryIDVal, &categoryName, &categoryDesc,
		&serviceIDVal, &serviceName, &serviceDesc,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAppealNotFound
		}
		return nil, fmt.Errorf("failed to get appeal: %w", err)
	}

	// Set appeal IDs
	appeal.CategoryID = categoryID
	appeal.ServiceID = serviceID

	// Set user (should always exist since user_id is NOT NULL)
	appeal.User = &user

	// Set category if exists (check categoryIDVal from JOIN, not categoryID from appeal)
	if categoryIDVal != nil && categoryName != nil {
		var category models.Category
		category.ID = *categoryIDVal
		category.Name = *categoryName
		if categoryDesc != nil {
			category.Description = *categoryDesc
		}
		appeal.Category = &category
	}

	// Set service if exists
	if serviceIDVal != nil && serviceName != nil {
		var service models.Service
		service.ID = *serviceIDVal
		service.Name = *serviceName
		if serviceDesc != nil {
			service.Description = *serviceDesc
		}
		appeal.Service = &service
	}

	return &appeal, nil
}

// List retrieves appeals with filters and pagination
func (r *AppealRepository) List(ctx context.Context, filters *models.AppealFilters) ([]*models.Appeal, int64, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argCount := 1

	// Build WHERE clause
	if filters.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.status = $%d", argCount))
		args = append(args, *filters.Status)
		argCount++
	}

	if filters.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.category_id = $%d", argCount))
		args = append(args, *filters.CategoryID)
		argCount++
	}

	if filters.ServiceID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.service_id = $%d", argCount))
		args = append(args, *filters.ServiceID)
		argCount++
	}

	if filters.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.user_id = $%d", argCount))
		args = append(args, *filters.UserID)
		argCount++
	}

	if filters.FromDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.created_at >= $%d", argCount))
		args = append(args, *filters.FromDate)
		argCount++
	}

	if filters.ToDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.created_at <= $%d", argCount))
		args = append(args, *filters.ToDate)
		argCount++
	}

	if filters.Search != nil && *filters.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(a.title ILIKE $%d OR a.description ILIKE $%d)", argCount, argCount))
		searchTerm := "%" + *filters.Search + "%"
		args = append(args, searchTerm)
		argCount++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM appeals a WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count appeals: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "a.created_at DESC"
	if filters.SortBy != "" {
		order := "DESC"
		if filters.SortOrder == "asc" {
			order = "ASC"
		}
		switch filters.SortBy {
		case "priority":
			orderBy = fmt.Sprintf("a.priority %s, a.created_at DESC", order)
		case "created_at":
			orderBy = fmt.Sprintf("a.created_at %s", order)
		case "status":
			orderBy = fmt.Sprintf("a.status %s, a.created_at DESC", order)
		}
	}

	// Pagination
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.Limit

	// Get appeals
	query := fmt.Sprintf(`
		SELECT
			a.id, a.user_id, a.category_id, a.service_id,
			a.status, a.title, a.description, a.address, a.latitude, a.longitude,
			a.priority, a.created_at, a.updated_at, a.closed_at,
			u.first_name, u.last_name,
			c.name AS category_name,
			s.name AS service_name
		FROM appeals a
		LEFT JOIN users u ON a.user_id = u.id
		LEFT JOIN categories c ON a.category_id = c.id
		LEFT JOIN services s ON a.service_id = s.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argCount, argCount+1)

	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list appeals: %w", err)
	}
	defer rows.Close()

	appeals := make([]*models.Appeal, 0)
	for rows.Next() {
		var appeal models.Appeal
		var firstName, lastName string
		var categoryName, serviceName *string

		err := rows.Scan(
			&appeal.ID, &appeal.UserID, &appeal.CategoryID, &appeal.ServiceID,
			&appeal.Status, &appeal.Title, &appeal.Description, &appeal.Address,
			&appeal.Latitude, &appeal.Longitude, &appeal.Priority,
			&appeal.CreatedAt, &appeal.UpdatedAt, &appeal.ClosedAt,
			&firstName, &lastName,
			&categoryName, &serviceName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan appeal: %w", err)
		}

		// Set user info
		appeal.User = &models.User{
			ID:        appeal.UserID,
			FirstName: firstName,
			LastName:  lastName,
		}

		// Set category if exists
		if categoryName != nil {
			appeal.Category = &models.Category{
				ID:   *appeal.CategoryID,
				Name: *categoryName,
			}
		}

		// Set service if exists
		if serviceName != nil {
			appeal.Service = &models.Service{
				ID:   *appeal.ServiceID,
				Name: *serviceName,
			}
		}

		appeals = append(appeals, &appeal)
	}

	return appeals, total, nil
}

// Update updates an appeal
func (r *AppealRepository) Update(ctx context.Context, appeal *models.Appeal) error {
	query := `
		UPDATE appeals
		SET title = $1, description = $2, category_id = $3, address = $4,
		    latitude = $5, longitude = $6, service_id = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := r.db.Exec(
		ctx,
		query,
		appeal.Title,
		appeal.Description,
		appeal.CategoryID,
		appeal.Address,
		appeal.Latitude,
		appeal.Longitude,
		appeal.ServiceID,
		appeal.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update appeal: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAppealNotFound
	}

	return nil
}

// UpdateStatus updates appeal status and records history
func (r *AppealRepository) UpdateStatus(ctx context.Context, appealID int64, newStatus models.AppealStatus, userID int64, comment *string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current status
	var oldStatus models.AppealStatus
	err = tx.QueryRow(ctx, "SELECT status FROM appeals WHERE id = $1", appealID).Scan(&oldStatus)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	// Only record history if status actually changed
	if oldStatus == newStatus {
		log.Printf("Status unchanged for appeal %d: %s, skipping history record", appealID, oldStatus)
		// Still update updated_at timestamp
		updateQuery := `UPDATE appeals SET updated_at = NOW() WHERE id = $1`
		_, err = tx.Exec(ctx, updateQuery, appealID)
		if err != nil {
			return fmt.Errorf("failed to update timestamp: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}

	// Update status
	// Set closed_at for both 'closed' and 'completed' statuses
	updateQuery := `
		UPDATE appeals
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	if newStatus == models.StatusClosed || newStatus == models.StatusCompleted {
		updateQuery = `
			UPDATE appeals
			SET status = $1, updated_at = NOW(), closed_at = NOW()
			WHERE id = $2
		`
	}

	result, err := tx.Exec(ctx, updateQuery, newStatus, appealID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAppealNotFound
	}

	// Record history
	historyQuery := `
		INSERT INTO appeal_history (appeal_id, user_id, old_status, new_status, action, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	// Translate status to Ukrainian for action message
	statusLabels := map[models.AppealStatus]string{
		models.StatusNew:        "Нове",
		models.StatusAssigned:   "Призначене",
		models.StatusInProgress: "В роботі",
		models.StatusCompleted:  "Виконане",
		models.StatusClosed:     "Закрите",
		models.StatusRejected:   "Відхилене",
	}
	oldLabel := statusLabels[oldStatus]
	if oldLabel == "" {
		oldLabel = string(oldStatus)
	}
	newLabel := statusLabels[newStatus]
	if newLabel == "" {
		newLabel = string(newStatus)
	}
	action := fmt.Sprintf("Статус змінено з %s на %s", oldLabel, newLabel)
	result, err = tx.Exec(ctx, historyQuery, appealID, userID, oldStatus, newStatus, action, comment)
	if err != nil {
		log.Printf("Error inserting history: %v", err)
		return fmt.Errorf("failed to record history: %w", err)
	}
	if result.RowsAffected() == 0 {
		log.Printf("Warning: History insert affected 0 rows for appeal %d", appealID)
	} else {
		log.Printf("History recorded for appeal %d: %s -> %s", appealID, oldStatus, newStatus)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Assign assigns an appeal to a service
func (r *AppealRepository) Assign(ctx context.Context, appealID, serviceID int64, priority *int, userID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current status before updating
	var oldStatus models.AppealStatus
	err = tx.QueryRow(ctx, "SELECT status FROM appeals WHERE id = $1", appealID).Scan(&oldStatus)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	// Update appeal with service assignment and set status to 'assigned'
	query := `
		UPDATE appeals
		SET service_id = $1, priority = COALESCE($2, priority),
		    status = 'assigned', updated_at = NOW()
		WHERE id = $3
	`

	result, err := tx.Exec(ctx, query, serviceID, priority, appealID)
	if err != nil {
		return fmt.Errorf("failed to assign appeal: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAppealNotFound
	}

	// Record history only if status actually changed
	if oldStatus != models.StatusAssigned {
		// Translate status to Ukrainian for action message
		statusLabels := map[models.AppealStatus]string{
			models.StatusNew:        "Нове",
			models.StatusAssigned:   "Призначене",
			models.StatusInProgress: "В роботі",
			models.StatusCompleted:  "Виконане",
			models.StatusClosed:     "Закрите",
			models.StatusRejected:   "Відхилене",
		}
		oldLabel := statusLabels[oldStatus]
		if oldLabel == "" {
			oldLabel = string(oldStatus)
		}
		action := fmt.Sprintf("Звернення призначено до служби. Статус змінено з %s на Призначене", oldLabel)

	historyQuery := `
		INSERT INTO appeal_history (appeal_id, user_id, old_status, new_status, action)
			VALUES ($1, $2, $3, 'assigned', $4)
	`
		_, err = tx.Exec(ctx, historyQuery, appealID, userID, oldStatus, action)
	if err != nil {
			log.Printf("Error inserting history: %v", err)
		return fmt.Errorf("failed to record history: %w", err)
		}
	} else {
		// Status already assigned, just record assignment without status change
		historyQuery := `
			INSERT INTO appeal_history (appeal_id, user_id, old_status, new_status, action)
			VALUES ($1, $2, $3, 'assigned', 'Звернення призначено до служби')
		`
		_, err = tx.Exec(ctx, historyQuery, appealID, userID, oldStatus)
		if err != nil {
			log.Printf("Error inserting history: %v", err)
			return fmt.Errorf("failed to record history: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdatePriority updates only the priority of an appeal
func (r *AppealRepository) UpdatePriority(ctx context.Context, appealID int64, priority int, userID int64) error {
	query := `
		UPDATE appeals
		SET priority = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, priority, appealID)
	if err != nil {
		return fmt.Errorf("failed to update priority: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAppealNotFound
	}

	return nil
}

// Delete soft deletes an appeal (not implemented, for future use)
func (r *AppealRepository) Delete(ctx context.Context, id int64) error {
	// For now, we don't implement soft delete
	// Appeals should not be deleted, only status changed
	return errors.New("deleting appeals is not allowed")
}

// GetStatistics returns basic statistics
func (r *AppealRepository) GetStatistics(ctx context.Context, fromDate, toDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	args := []interface{}{}
	argCount := 1

	// Build WHERE clause for queries with JOIN (using alias 'a')
	whereClauseWithAlias := "1=1"
	// Build WHERE clause for simple queries without JOIN
	simpleWhereClause := "1=1"

	if fromDate != nil {
		whereClauseWithAlias += fmt.Sprintf(" AND a.created_at >= $%d", argCount)
		simpleWhereClause += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *fromDate)
		argCount++
	}

	if toDate != nil {
		whereClauseWithAlias += fmt.Sprintf(" AND a.created_at <= $%d", argCount)
		simpleWhereClause += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *toDate)
		argCount++
	}

	// Total count
	var total int64
	err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM appeals WHERE %s", simpleWhereClause), args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count total: %w", err)
	}
	stats["total"] = total

	// By status
	statusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*)
		FROM appeals
		WHERE %s
		GROUP BY status
	`, simpleWhereClause)

	rows, err := r.db.Query(ctx, statusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}
	defer rows.Close()

	byStatus := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		byStatus[status] = count
	}
	stats["by_status"] = byStatus

	// Average processing time (for completed appeals) - in days
	var avgTime *float64
	avgQuery := fmt.Sprintf(`
		SELECT AVG(EXTRACT(EPOCH FROM (closed_at - created_at))/86400)
		FROM appeals
		WHERE status IN ('closed', 'completed') AND closed_at IS NOT NULL AND %s
	`, simpleWhereClause)
	err = r.db.QueryRow(ctx, avgQuery, args...).Scan(&avgTime)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate avg time: %w", err)
	}
	if avgTime != nil {
		stats["avg_processing_days"] = *avgTime
		stats["avg_processing_hours"] = *avgTime * 24 // Для сумісності зі старим API
	}

	// By category
	categoryQuery := fmt.Sprintf(`
		SELECT c.name, COUNT(a.id)
		FROM appeals a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE %s
		GROUP BY c.name
		ORDER BY COUNT(a.id) DESC
		LIMIT 10
	`, whereClauseWithAlias)
	rows, err = r.db.Query(ctx, categoryQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()

	byCategory := make(map[string]int64)
	for rows.Next() {
		var categoryName *string
		var count int64
		if err := rows.Scan(&categoryName, &count); err != nil {
			return nil, err
		}
		name := "Без категорії"
		if categoryName != nil {
			name = *categoryName
		}
		byCategory[name] = count
	}
	stats["by_category"] = byCategory

	// By service
	serviceQuery := fmt.Sprintf(`
		SELECT s.name, COUNT(a.id)
		FROM appeals a
		LEFT JOIN services s ON a.service_id = s.id
		WHERE %s
		GROUP BY s.name
		ORDER BY COUNT(a.id) DESC
		LIMIT 10
	`, whereClauseWithAlias)
	rows, err = r.db.Query(ctx, serviceQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get service stats: %w", err)
	}
	defer rows.Close()

	byService := make(map[string]int64)
	for rows.Next() {
		var serviceName *string
		var count int64
		if err := rows.Scan(&serviceName, &count); err != nil {
			return nil, err
		}
		name := "Не призначено"
		if serviceName != nil {
			name = *serviceName
		}
		byService[name] = count
	}
	stats["by_service"] = byService

	// By executor - removed, appeals are assigned to services, not executors

	// Daily trend (last 30 days)
	trendQuery := fmt.Sprintf(`
		SELECT DATE(a.created_at) as date, COUNT(*) as count
		FROM appeals a
		WHERE a.created_at >= NOW() - INTERVAL '30 days' AND %s
		GROUP BY DATE(a.created_at)
		ORDER BY date ASC
	`, whereClauseWithAlias)
	rows, err = r.db.Query(ctx, trendQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get trend stats: %w", err)
	}
	defer rows.Close()

	dailyTrend := []map[string]interface{}{}
	for rows.Next() {
		var date time.Time
		var count int64
		if err := rows.Scan(&date, &count); err != nil {
			return nil, err
		}
		dailyTrend = append(dailyTrend, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"count": count,
		})
	}
	stats["daily_trend"] = dailyTrend

	// Priority distribution
	priorityQuery := fmt.Sprintf(`
		SELECT priority, COUNT(*)
		FROM appeals
		WHERE %s
		GROUP BY priority
		ORDER BY priority DESC
	`, simpleWhereClause)
	rows, err = r.db.Query(ctx, priorityQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority stats: %w", err)
	}
	defer rows.Close()

	byPriority := make(map[int]int64)
	for rows.Next() {
		var priority int
		var count int64
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, err
		}
		byPriority[priority] = count
	}
	stats["by_priority"] = byPriority

	// Performance Metrics
	// Overdue appeals (more than 30 days)
	var overdueCount int64
	overdueQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM appeals
		WHERE status NOT IN ('closed', 'rejected')
			AND created_at < NOW() - INTERVAL '30 days'
			AND %s
	`, simpleWhereClause)
	err = r.db.QueryRow(ctx, overdueQuery, args...).Scan(&overdueCount)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to count overdue: %w", err)
	}
	stats["overdue_count"] = overdueCount

	// Percentage of appeals completed on time (within 30 days)
	var onTimeCount int64
	var totalCompleted int64
	onTimeQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) FILTER (WHERE closed_at IS NOT NULL AND EXTRACT(EPOCH FROM (closed_at - created_at))/86400 <= 30) as on_time,
			COUNT(*) FILTER (WHERE status IN ('closed', 'completed')) as total_completed
		FROM appeals
		WHERE status IN ('closed', 'completed') AND %s
	`, simpleWhereClause)
	err = r.db.QueryRow(ctx, onTimeQuery, args...).Scan(&onTimeCount, &totalCompleted)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate on-time percentage: %w", err)
	}
	var onTimePercentage float64
	if totalCompleted > 0 {
		onTimePercentage = float64(onTimeCount) / float64(totalCompleted) * 100
	}
	stats["on_time_percentage"] = onTimePercentage
	stats["on_time_count"] = onTimeCount
	stats["total_completed"] = totalCompleted

	return stats, nil
}

// GetDispatcherDashboard returns dashboard data for dispatcher role
func (r *AppealRepository) GetDispatcherDashboard(ctx context.Context) (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// Unclosed appeals with deadlines (approaching or overdue)
	overdueQuery := `
		SELECT a.id, a.title, a.status, a.created_at, a.service_id, a.priority, s.name as service_name
		FROM appeals a
		LEFT JOIN services s ON a.service_id = s.id
		WHERE a.status NOT IN ('closed', 'rejected')
			AND a.created_at < NOW() - INTERVAL '30 days'
		ORDER BY a.created_at ASC
		LIMIT 20
	`
	rows, err := r.db.Query(ctx, overdueQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue appeals: %w", err)
	}
	defer rows.Close()

	overdueAppeals := []map[string]interface{}{}
	for rows.Next() {
		var id, serviceID int64
		var title, status string
		var createdAt time.Time
		var priority int
		var serviceIDPtr *int64
		var serviceName *string
		err := rows.Scan(&id, &title, &status, &createdAt, &serviceIDPtr, &priority, &serviceName)
		if err != nil {
			continue
		}
		if serviceIDPtr != nil {
			serviceID = *serviceIDPtr
		}
		daysOverdue := int(time.Since(createdAt).Hours() / 24)
		overdueAppeals = append(overdueAppeals, map[string]interface{}{
			"id":           id,
			"title":        title,
			"status":       status,
			"created_at":   createdAt,
			"service_id":   serviceID,
			"service_name": getStringValue(serviceName),
			"priority":     priority,
			"days_overdue": daysOverdue,
		})
	}
	dashboard["overdue_appeals"] = overdueAppeals

	// Appeals without status change for more than 7 days
	staleQuery := `
		SELECT a.id, a.title, a.status, a.updated_at, a.service_id, a.priority, s.name as service_name
		FROM appeals a
		LEFT JOIN services s ON a.service_id = s.id
		WHERE a.status NOT IN ('closed', 'rejected')
			AND a.updated_at < NOW() - INTERVAL '7 days'
		ORDER BY a.updated_at ASC
		LIMIT 20
	`
	rows, err = r.db.Query(ctx, staleQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale appeals: %w", err)
	}
	defer rows.Close()

	staleAppeals := []map[string]interface{}{}
	for rows.Next() {
		var id, serviceID int64
		var title, status string
		var updatedAt time.Time
		var priority int
		var serviceIDPtr *int64
		var serviceName *string
		err := rows.Scan(&id, &title, &status, &updatedAt, &serviceIDPtr, &priority, &serviceName)
		if err != nil {
			continue
		}
		if serviceIDPtr != nil {
			serviceID = *serviceIDPtr
		}
		daysStale := int(time.Since(updatedAt).Hours() / 24)
		staleAppeals = append(staleAppeals, map[string]interface{}{
			"id":           id,
			"title":        title,
			"status":       status,
			"updated_at":   updatedAt,
			"service_id":   serviceID,
			"service_name": getStringValue(serviceName),
			"priority":     priority,
			"days_stale":   daysStale,
		})
	}
	dashboard["stale_appeals"] = staleAppeals

	// Appeals approaching deadline (less than 5 days remaining, but not yet overdue)
	approachingQuery := `
		SELECT a.id, a.title, a.status, a.created_at, a.service_id, a.priority, s.name as service_name
		FROM appeals a
		LEFT JOIN services s ON a.service_id = s.id
		WHERE a.status NOT IN ('closed', 'rejected')
			AND a.created_at >= NOW() - INTERVAL '30 days'
			AND a.created_at < NOW() - INTERVAL '25 days'
		ORDER BY a.created_at ASC
		LIMIT 20
	`
	rows, err = r.db.Query(ctx, approachingQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get approaching appeals: %w", err)
	}
	defer rows.Close()

	approachingAppeals := []map[string]interface{}{}
	for rows.Next() {
		var id, serviceID int64
		var title, status string
		var createdAt time.Time
		var priority int
		var serviceIDPtr *int64
		var serviceName *string
		err := rows.Scan(&id, &title, &status, &createdAt, &serviceIDPtr, &priority, &serviceName)
		if err != nil {
			continue
		}
		if serviceIDPtr != nil {
			serviceID = *serviceIDPtr
		}
		daysRemaining := 30 - int(time.Since(createdAt).Hours()/24)
		approachingAppeals = append(approachingAppeals, map[string]interface{}{
			"id":             id,
			"title":          title,
			"status":         status,
			"created_at":     createdAt,
			"service_id":     serviceID,
			"service_name":   getStringValue(serviceName),
			"priority":       priority,
			"days_remaining": daysRemaining,
		})
	}
	dashboard["approaching_appeals"] = approachingAppeals

	return dashboard, nil
}

// GetAdminDashboard returns dashboard data for admin role
func (r *AppealRepository) GetAdminDashboard(ctx context.Context) (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// Top services by processing speed (average time from creation to closure)
	// Include both 'closed' and 'completed' statuses
	// Use closed_at if available, otherwise use updated_at for completed appeals
	serviceSpeedQuery := `
		SELECT 
			s.name,
			COUNT(a.id) as total_appeals,
			AVG(EXTRACT(EPOCH FROM (COALESCE(a.closed_at, a.updated_at) - a.created_at))/3600) as avg_hours
		FROM appeals a
		INNER JOIN services s ON a.service_id = s.id
		WHERE a.status IN ('closed', 'completed')
			AND (a.closed_at IS NOT NULL OR a.updated_at IS NOT NULL)
			AND COALESCE(a.closed_at, a.updated_at) >= NOW() - INTERVAL '90 days'
		GROUP BY s.id, s.name
		HAVING COUNT(a.id) >= 1
		ORDER BY avg_hours ASC
		LIMIT 10
	`
	rows, err := r.db.Query(ctx, serviceSpeedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get service speeds: %w", err)
	}
	defer rows.Close()

	topServices := []map[string]interface{}{}
	for rows.Next() {
		var name string
		var totalAppeals int64
		var avgHours *float64
		err := rows.Scan(&name, &totalAppeals, &avgHours)
		if err != nil {
			continue
		}
		avgHoursVal := 0.0
		if avgHours != nil {
			avgHoursVal = *avgHours
		}
		topServices = append(topServices, map[string]interface{}{
			"name":          name,
			"total_appeals": totalAppeals,
			"avg_hours":     avgHoursVal,
			"avg_days":      avgHoursVal / 24,
		})
	}
	dashboard["top_services_by_speed"] = topServices

	// Monthly trend (last 6 months)
	monthlyTrendQuery := `
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as count
		FROM appeals
		WHERE created_at >= NOW() - INTERVAL '6 months'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month ASC
	`
	rows, err = r.db.Query(ctx, monthlyTrendQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly trend: %w", err)
	}
	defer rows.Close()

	monthlyTrend := []map[string]interface{}{}
	for rows.Next() {
		var month time.Time
		var count int64
		err := rows.Scan(&month, &count)
		if err != nil {
			continue
		}
		monthlyTrend = append(monthlyTrend, map[string]interface{}{
			"month": month.Format("2006-01"),
			"count": count,
		})
	}
	dashboard["monthly_trend"] = monthlyTrend

	// Heat map by day of week
	dayOfWeekQuery := `
		SELECT 
			EXTRACT(DOW FROM created_at) as day_of_week,
			COUNT(*) as count
		FROM appeals
		WHERE created_at >= NOW() - INTERVAL '90 days'
		GROUP BY EXTRACT(DOW FROM created_at)
		ORDER BY day_of_week ASC
	`
	rows, err = r.db.Query(ctx, dayOfWeekQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get day of week stats: %w", err)
	}
	defer rows.Close()

	dayOfWeekStats := []map[string]interface{}{}
	dayNames := []string{"Неділя", "Понеділок", "Вівторок", "Середа", "Четвер", "П'ятниця", "Субота"}
	for rows.Next() {
		var dayOfWeek int
		var count int64
		err := rows.Scan(&dayOfWeek, &count)
		if err != nil {
			continue
		}
		dayOfWeekStats = append(dayOfWeekStats, map[string]interface{}{
			"day":   dayOfWeek,
			"name":  dayNames[dayOfWeek],
			"count": count,
		})
	}
	dashboard["day_of_week_stats"] = dayOfWeekStats

	// Detailed statistics for all services
	allServicesStatsQuery := `
		SELECT 
			s.id,
			s.name,
			COUNT(a.id) FILTER (WHERE a.id IS NOT NULL) as total_appeals,
			COUNT(a.id) FILTER (WHERE a.status = 'new') as new_count,
			COUNT(a.id) FILTER (WHERE a.status = 'assigned') as assigned_count,
			COUNT(a.id) FILTER (WHERE a.status = 'in_progress') as in_progress_count,
			COUNT(a.id) FILTER (WHERE a.status IN ('completed', 'closed')) as completed_count,
			COUNT(a.id) FILTER (WHERE a.status NOT IN ('closed', 'rejected') 
				AND a.created_at < NOW() - INTERVAL '30 days') as overdue_count,
			AVG(EXTRACT(EPOCH FROM (COALESCE(a.closed_at, a.updated_at) - a.created_at))/3600) 
				FILTER (WHERE a.status IN ('closed', 'completed') 
					AND (a.closed_at IS NOT NULL OR a.updated_at IS NOT NULL)) as avg_hours,
			COUNT(a.id) FILTER (WHERE a.status IN ('closed', 'completed') 
				AND a.closed_at IS NOT NULL 
				AND EXTRACT(EPOCH FROM (a.closed_at - a.created_at))/86400 <= 30) as on_time_count,
			COUNT(a.id) FILTER (WHERE a.status IN ('closed', 'completed')) as total_completed
		FROM services s
		LEFT JOIN appeals a ON s.id = a.service_id
		WHERE s.is_active = true
		GROUP BY s.id, s.name
		ORDER BY s.name ASC
	`
	rows, err = r.db.Query(ctx, allServicesStatsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get all services stats: %w", err)
	}
	defer rows.Close()

	allServicesStats := []map[string]interface{}{}
	for rows.Next() {
		var serviceID int64
		var serviceName string
		var totalAppeals, newCount, assignedCount, inProgressCount, completedCount, overdueCount int64
		var avgHours *float64
		var onTimeCount, totalCompleted int64

		err := rows.Scan(
			&serviceID, &serviceName,
			&totalAppeals, &newCount, &assignedCount, &inProgressCount, &completedCount, &overdueCount,
			&avgHours, &onTimeCount, &totalCompleted,
		)
		if err != nil {
			continue
		}

		avgHoursVal := 0.0
		if avgHours != nil {
			avgHoursVal = *avgHours
		}

		onTimePercentage := 0.0
		if totalCompleted > 0 {
			onTimePercentage = float64(onTimeCount) / float64(totalCompleted) * 100
		}

		allServicesStats = append(allServicesStats, map[string]interface{}{
			"id":                 serviceID,
			"name":               serviceName,
			"total_appeals":      totalAppeals,
			"new_count":          newCount,
			"assigned_count":     assignedCount,
			"in_progress_count":  inProgressCount,
			"completed_count":    completedCount,
			"overdue_count":      overdueCount,
			"avg_hours":          avgHoursVal,
			"avg_days":           avgHoursVal / 24,
			"on_time_count":      onTimeCount,
			"total_completed":    totalCompleted,
			"on_time_percentage": onTimePercentage,
		})
	}
	dashboard["all_services_stats"] = allServicesStats

	return dashboard, nil
}

// GetExecutorDashboard returns dashboard data for executor role
func (r *AppealRepository) GetExecutorDashboard(ctx context.Context, userID int64) (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// Get services for this executor
	serviceQuery := `
		SELECT s.id, s.name
		FROM services s
		INNER JOIN user_services us ON s.id = us.service_id
		WHERE us.user_id = $1
	`
	rows, err := r.db.Query(ctx, serviceQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get executor services: %w", err)
	}
	defer rows.Close()

	serviceIDs := []int64{}
	serviceNames := make(map[int64]string)
	for rows.Next() {
		var serviceID int64
		var serviceName string
		err := rows.Scan(&serviceID, &serviceName)
		if err != nil {
			continue
		}
		serviceIDs = append(serviceIDs, serviceID)
		serviceNames[serviceID] = serviceName
	}
	rows.Close()

	if len(serviceIDs) == 0 {
		dashboard["active_appeals"] = []map[string]interface{}{}
		dashboard["my_avg_processing_time"] = 0.0
		dashboard["service_avg_processing_time"] = 0.0
		return dashboard, nil
	}

	// Active appeals for this executor's services
	// Build query with IN clause for service IDs

	placeholders := make([]string, len(serviceIDs))
	queryArgs := make([]interface{}, len(serviceIDs))
	for i, id := range serviceIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		queryArgs[i] = id
	}

	activeQuery := fmt.Sprintf(`
		SELECT id, title, status, created_at, service_id, priority
		FROM appeals
		WHERE service_id IN (%s)
			AND status NOT IN ('closed', 'rejected')
		ORDER BY priority DESC, created_at ASC
		LIMIT 50
	`, strings.Join(placeholders, ","))
	rows, err = r.db.Query(ctx, activeQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get active appeals: %w", err)
	}
	defer rows.Close()

	activeAppeals := []map[string]interface{}{}
	for rows.Next() {
		var id, serviceID int64
		var title, status string
		var createdAt time.Time
		var priority int
		var serviceIDPtr *int64
		err := rows.Scan(&id, &title, &status, &createdAt, &serviceIDPtr, &priority)
		if err != nil {
			continue
		}
		if serviceIDPtr != nil {
			serviceID = *serviceIDPtr
		}
		daysSinceCreation := int(time.Since(createdAt).Hours() / 24)
		activeAppeals = append(activeAppeals, map[string]interface{}{
			"id":                  id,
			"title":               title,
			"status":              status,
			"created_at":          createdAt,
			"service_id":          serviceID,
			"service_name":        serviceNames[serviceID],
			"priority":            priority,
			"days_since_creation": daysSinceCreation,
		})
	}
	dashboard["active_appeals"] = activeAppeals

	// My average processing time (appeals where I changed status to completed/closed) - in days
	myAvgQuery := `
		SELECT AVG(EXTRACT(EPOCH FROM (a.closed_at - a.created_at))/86400)
		FROM appeals a
		INNER JOIN appeal_history ah ON a.id = ah.appeal_id
		WHERE ah.user_id = $1
			AND ah.new_status IN ('completed', 'closed')
			AND a.closed_at IS NOT NULL
			AND a.closed_at >= NOW() - INTERVAL '90 days'
	`
	var myAvgDays *float64
	err = r.db.QueryRow(ctx, myAvgQuery, userID).Scan(&myAvgDays)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get my avg time: %w", err)
	}
	myAvgDaysVal := 0.0
	if myAvgDays != nil {
		myAvgDaysVal = *myAvgDays
	}
	dashboard["my_avg_processing_time"] = myAvgDaysVal

	// Service average processing time - in days
	serviceAvgQuery := fmt.Sprintf(`
		SELECT AVG(EXTRACT(EPOCH FROM (closed_at - created_at))/86400)
		FROM appeals
		WHERE service_id IN (%s)
			AND status IN ('closed', 'completed')
			AND closed_at IS NOT NULL
			AND closed_at >= NOW() - INTERVAL '90 days'
	`, strings.Join(placeholders, ","))
	var serviceAvgDays *float64
	err = r.db.QueryRow(ctx, serviceAvgQuery, queryArgs...).Scan(&serviceAvgDays)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get service avg time: %w", err)
	}
	serviceAvgDaysVal := 0.0
	if serviceAvgDays != nil {
		serviceAvgDaysVal = *serviceAvgDays
	}
	dashboard["service_avg_processing_time"] = serviceAvgDaysVal

	return dashboard, nil
}

// GetHistory retrieves appeal history (status changes)
func (r *AppealRepository) GetHistory(ctx context.Context, appealID int64) ([]*models.AppealHistory, error) {
	query := `
		SELECT 
			ah.id, ah.appeal_id, ah.user_id, ah.old_status, ah.new_status, 
			ah.action, ah.comment, ah.created_at,
			u.id, u.first_name, u.last_name, u.email, u.role
		FROM appeal_history ah
		LEFT JOIN users u ON ah.user_id = u.id
		WHERE ah.appeal_id = $1
		ORDER BY ah.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, appealID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appeal history: %w", err)
	}
	defer rows.Close()

	history := make([]*models.AppealHistory, 0)
	for rows.Next() {
		var h models.AppealHistory
		var oldStatus *string
		var newStatus string // NOT NULL in DB
		var comment *string
		var userID *int64
		var firstName, lastName, email *string
		var role *string

		err := rows.Scan(
			&h.ID, &h.AppealID, &h.UserID, &oldStatus, &newStatus,
			&h.Action, &comment, &h.CreatedAt,
			&userID, &firstName, &lastName, &email, &role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}

		if oldStatus != nil {
			status := models.AppealStatus(*oldStatus)
			h.OldStatus = &status
		}
		h.NewStatus = models.AppealStatus(newStatus)
		h.Comment = comment

		// Set user if exists
		if userID != nil && firstName != nil {
			h.User = &models.User{
				ID:        *userID,
				FirstName: *firstName,
			}
			if lastName != nil {
				h.User.LastName = *lastName
			}
			if email != nil {
				h.User.Email = *email
			}
			if role != nil {
				h.User.Role = models.UserRole(*role)
			}
		}

		history = append(history, &h)
	}

	return history, nil
}

// GetServiceStatistics returns detailed statistics for a specific service
func (r *AppealRepository) GetServiceStatistics(ctx context.Context, serviceID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get service info
	var serviceName string
	var serviceDescription, contactPerson, contactPhone, contactEmail *string
	var isActive bool
	err := r.db.QueryRow(ctx, `
		SELECT name, description, contact_person, contact_phone, contact_email, is_active
		FROM services
		WHERE id = $1
	`, serviceID).Scan(&serviceName, &serviceDescription, &contactPerson, &contactPhone, &contactEmail, &isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get service info: %w", err)
	}
	stats["service"] = map[string]interface{}{
		"id":             serviceID,
		"name":           serviceName,
		"description":    serviceDescription,
		"contact_person": contactPerson,
		"contact_phone":  contactPhone,
		"contact_email":  contactEmail,
		"is_active":      isActive,
	}

	// Overall statistics
	overallStatsQuery := `
		SELECT 
			COUNT(a.id) FILTER (WHERE a.id IS NOT NULL) as total_appeals,
			COUNT(a.id) FILTER (WHERE a.status = 'new') as new_count,
			COUNT(a.id) FILTER (WHERE a.status = 'assigned') as assigned_count,
			COUNT(a.id) FILTER (WHERE a.status = 'in_progress') as in_progress_count,
			COUNT(a.id) FILTER (WHERE a.status IN ('completed', 'closed')) as completed_count,
			COUNT(a.id) FILTER (WHERE a.status = 'rejected') as rejected_count,
			COUNT(a.id) FILTER (WHERE a.status NOT IN ('closed', 'rejected') 
				AND a.created_at < NOW() - INTERVAL '30 days') as overdue_count,
			AVG(EXTRACT(EPOCH FROM (COALESCE(a.closed_at, a.updated_at) - a.created_at))/86400) 
				FILTER (WHERE a.status IN ('closed', 'completed') 
					AND (a.closed_at IS NOT NULL OR a.updated_at IS NOT NULL)) as avg_days,
			COUNT(a.id) FILTER (WHERE a.status IN ('closed', 'completed') 
				AND a.closed_at IS NOT NULL 
				AND EXTRACT(EPOCH FROM (a.closed_at - a.created_at))/86400 <= 30) as on_time_count,
			COUNT(a.id) FILTER (WHERE a.status IN ('closed', 'completed')) as total_completed
		FROM appeals a
		WHERE a.service_id = $1
	`
	var totalAppeals, newCount, assignedCount, inProgressCount, completedCount, rejectedCount, overdueCount int64
	var avgDays *float64
	var onTimeCount, totalCompleted int64
	err = r.db.QueryRow(ctx, overallStatsQuery, serviceID).Scan(
		&totalAppeals, &newCount, &assignedCount, &inProgressCount, &completedCount, &rejectedCount,
		&overdueCount, &avgDays, &onTimeCount, &totalCompleted,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	avgDaysVal := 0.0
	if avgDays != nil {
		avgDaysVal = *avgDays
	}

	onTimePercentage := 0.0
	if totalCompleted > 0 {
		onTimePercentage = float64(onTimeCount) / float64(totalCompleted) * 100
	}

	stats["overall"] = map[string]interface{}{
		"total_appeals":      totalAppeals,
		"new_count":          newCount,
		"assigned_count":     assignedCount,
		"in_progress_count":  inProgressCount,
		"completed_count":    completedCount,
		"rejected_count":     rejectedCount,
		"overdue_count":      overdueCount,
		"avg_days":           avgDaysVal,
		"on_time_count":      onTimeCount,
		"total_completed":    totalCompleted,
		"on_time_percentage": onTimePercentage,
	}

	// Monthly trend (last 6 months)
	monthlyTrendQuery := `
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as count
		FROM appeals
		WHERE service_id = $1
			AND created_at >= NOW() - INTERVAL '6 months'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month ASC
	`
	rows, err := r.db.Query(ctx, monthlyTrendQuery, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly trend: %w", err)
	}
	defer rows.Close()

	monthlyTrend := []map[string]interface{}{}
	for rows.Next() {
		var month time.Time
		var count int64
		if err := rows.Scan(&month, &count); err != nil {
			continue
		}
		monthlyTrend = append(monthlyTrend, map[string]interface{}{
			"month": month.Format("2006-01"),
			"count": count,
		})
	}
	stats["monthly_trend"] = monthlyTrend

	// Status distribution
	statusDistQuery := `
		SELECT status, COUNT(*) as count
		FROM appeals
		WHERE service_id = $1
		GROUP BY status
		ORDER BY count DESC
	`
	rows, err = r.db.Query(ctx, statusDistQuery, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}
	defer rows.Close()

	statusDistribution := []map[string]interface{}{}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		statusDistribution = append(statusDistribution, map[string]interface{}{
			"status": status,
			"count":  count,
		})
	}
	stats["status_distribution"] = statusDistribution

	// Category distribution
	categoryDistQuery := `
		SELECT c.name, COUNT(*) as count
		FROM appeals a
		INNER JOIN categories c ON a.category_id = c.id
		WHERE a.service_id = $1
		GROUP BY c.id, c.name
		ORDER BY count DESC
		LIMIT 10
	`
	rows, err = r.db.Query(ctx, categoryDistQuery, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category distribution: %w", err)
	}
	defer rows.Close()

	categoryDistribution := []map[string]interface{}{}
	for rows.Next() {
		var categoryName string
		var count int64
		if err := rows.Scan(&categoryName, &count); err != nil {
			continue
		}
		categoryDistribution = append(categoryDistribution, map[string]interface{}{
			"category": categoryName,
			"count":    count,
		})
	}
	stats["category_distribution"] = categoryDistribution

	// Recent appeals (last 10)
	recentAppealsQuery := `
		SELECT id, title, status, created_at, closed_at, priority
		FROM appeals
		WHERE service_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`
	rows, err = r.db.Query(ctx, recentAppealsQuery, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent appeals: %w", err)
	}
	defer rows.Close()

	recentAppeals := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var title, status string
		var createdAt time.Time
		var closedAt *time.Time
		var priority int
		if err := rows.Scan(&id, &title, &status, &createdAt, &closedAt, &priority); err != nil {
			continue
		}
		recentAppeals = append(recentAppeals, map[string]interface{}{
			"id":         id,
			"title":      title,
			"status":     status,
			"created_at": createdAt,
			"closed_at":  closedAt,
			"priority":   priority,
		})
	}
	stats["recent_appeals"] = recentAppeals

	return stats, nil
}

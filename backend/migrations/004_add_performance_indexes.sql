-- +migrate Up
-- Add indexes for better query performance

-- Indexes for appeals table
CREATE INDEX IF NOT EXISTS idx_appeals_status ON appeals(status);
CREATE INDEX IF NOT EXISTS idx_appeals_category_id ON appeals(category_id);
CREATE INDEX IF NOT EXISTS idx_appeals_service_id ON appeals(service_id);
CREATE INDEX IF NOT EXISTS idx_appeals_user_id ON appeals(user_id);
CREATE INDEX IF NOT EXISTS idx_appeals_created_at ON appeals(created_at);
CREATE INDEX IF NOT EXISTS idx_appeals_updated_at ON appeals(updated_at);
CREATE INDEX IF NOT EXISTS idx_appeals_priority ON appeals(priority);
CREATE INDEX IF NOT EXISTS idx_appeals_status_created_at ON appeals (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_appeals_category_id_status ON appeals (category_id, status);
CREATE INDEX IF NOT EXISTS idx_appeals_service_id_status ON appeals (service_id, status);
CREATE INDEX IF NOT EXISTS idx_appeals_user_id_created_at ON appeals (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_appeals_status_closed_at_created_at ON appeals (status, closed_at, created_at) WHERE status IN ('completed', 'closed');
CREATE INDEX IF NOT EXISTS idx_appeals_created_at_status_not_closed_rejected ON appeals (created_at) WHERE status NOT IN ('closed', 'rejected');

-- Indexes for appeal_history table
CREATE INDEX IF NOT EXISTS idx_appeal_history_appeal_id ON appeal_history (appeal_id);
CREATE INDEX IF NOT EXISTS idx_appeal_history_user_id ON appeal_history (user_id);
CREATE INDEX IF NOT EXISTS idx_appeal_history_created_at ON appeal_history (created_at);

-- Indexes for comments table
CREATE INDEX IF NOT EXISTS idx_comments_appeal_id_created_at ON comments (appeal_id, created_at DESC);

-- Indexes for notifications table
CREATE INDEX IF NOT EXISTS idx_notifications_user_id_read_created_at ON notifications (user_id, is_read, created_at DESC);

-- +migrate Down
-- Drop indexes
DROP INDEX IF EXISTS idx_appeals_status_created_at;
DROP INDEX IF EXISTS idx_appeals_category_id_status;
DROP INDEX IF EXISTS idx_appeals_service_id_status;
DROP INDEX IF EXISTS idx_appeals_user_id_created_at;
DROP INDEX IF EXISTS idx_appeals_status_closed_at_created_at;
DROP INDEX IF EXISTS idx_appeals_created_at_status_not_closed_rejected;

DROP INDEX IF EXISTS idx_appeal_history_appeal_id;
DROP INDEX IF EXISTS idx_appeal_history_user_id;
DROP INDEX IF EXISTS idx_appeal_history_created_at;

DROP INDEX IF EXISTS idx_comments_appeal_id_created_at;

DROP INDEX IF EXISTS idx_notifications_user_id_read_created_at;


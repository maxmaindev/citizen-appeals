-- +migrate Up
-- Initial database schema

-- Types
CREATE TYPE user_role AS ENUM ('citizen', 'dispatcher', 'executor', 'admin');
CREATE TYPE appeal_status AS ENUM ('new', 'assigned', 'in_progress', 'completed', 'closed', 'rejected');
CREATE TYPE notification_type AS ENUM ('appeal_created', 'appeal_assigned', 'status_changed', 'comment_added', 'appeal_completed');

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    role user_role NOT NULL DEFAULT 'citizen',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    default_priority INT NOT NULL DEFAULT 2 CHECK (default_priority >= 1 AND default_priority <= 3),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_is_active ON categories(is_active);

-- Insert default categories
INSERT INTO categories (name, description, default_priority) VALUES
    ('Інфраструктура та Мережі', 'Проблеми з Водоканалом, Теплопостачанням, Газопостачанням, Електромережами, аварії, витоки, прориви труб, ліфтове господарство', 2),
    ('Благоустрій та Довкілля', 'Дороги, тротуари, ями, озеленення, парки, прибирання сміття, вивезення ТПВ, снігу, кладовища', 2),
    ('Житловий Фонд', 'Внутрішньобудинкові питання (ОСББ, дахи, підвали, підїзди), капітальний ремонт, аварійний стан будинків', 2),
    ('Транспорт', 'Громадський транспорт (маршрути, розклад), паркування, евакуація, дорожня інфраструктура, знаки, світлофори', 2),
    ('Безпека та Порядок', 'Правопорушення, Муніципальна варта, поліція, ДСНС, пожежна безпека, безпритульні тварини', 2),
    ('Соціальна Сфера', 'Пільги, субсидії, освіта, охорона здоровя, лікарні, культура, спорт', 2),
    ('Дозвільна та Регуляторна', 'Архітектура, містобудування, самовільне будівництво, МАФи, земельні питання, незаконна торгівля', 1),
    ('Інше', 'Інші питання', 1)
ON CONFLICT (name) DO NOTHING;

-- Services table
CREATE TABLE IF NOT EXISTS services (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL UNIQUE,
    description TEXT,
    contact_person VARCHAR(200) NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_services_is_active ON services(is_active);

-- Services are seeded in 002_seed_services.sql

-- Appeals table
CREATE TABLE IF NOT EXISTS appeals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    service_id BIGINT REFERENCES services(id) ON DELETE SET NULL,
    status appeal_status NOT NULL DEFAULT 'new',
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    address VARCHAR(500) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    priority INT NOT NULL DEFAULT 2 CHECK (priority >= 1 AND priority <= 3),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP
);

CREATE INDEX idx_appeals_user_id ON appeals(user_id);
CREATE INDEX idx_appeals_category_id ON appeals(category_id);
CREATE INDEX idx_appeals_service_id ON appeals(service_id);
CREATE INDEX idx_appeals_status ON appeals(status);
CREATE INDEX idx_appeals_created_at ON appeals(created_at);
CREATE INDEX idx_appeals_priority ON appeals(priority);
CREATE INDEX idx_appeals_location ON appeals(latitude, longitude);

-- Comments table
CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    appeal_id BIGINT NOT NULL REFERENCES appeals(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_appeal_id ON comments(appeal_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);

-- Photos table
CREATE TABLE IF NOT EXISTS photos (
    id BIGSERIAL PRIMARY KEY,
    appeal_id BIGINT REFERENCES appeals(id) ON DELETE CASCADE,
    comment_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    is_result_photo BOOLEAN NOT NULL DEFAULT false,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_photos_appeal_id ON photos(appeal_id);
CREATE INDEX idx_photos_comment_id ON photos(comment_id);
CREATE INDEX idx_photos_is_result_photo ON photos(is_result_photo);

-- Category-Services relationship table (many-to-many)
CREATE TABLE IF NOT EXISTS category_services (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    service_id BIGINT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(category_id, service_id)
);

CREATE INDEX idx_category_services_category_id ON category_services(category_id);
CREATE INDEX idx_category_services_service_id ON category_services(service_id);

-- User-Services relationship table (many-to-many)
-- Links executors to services they belong to
CREATE TABLE IF NOT EXISTS user_services (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id BIGINT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, service_id)
);

CREATE INDEX idx_user_services_user_id ON user_services(user_id);
CREATE INDEX idx_user_services_service_id ON user_services(service_id);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    appeal_id BIGINT REFERENCES appeals(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_appeal_id ON notifications(appeal_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_sent_at ON notifications(sent_at);

-- Appeal history table
CREATE TABLE IF NOT EXISTS appeal_history (
    id BIGSERIAL PRIMARY KEY,
    appeal_id BIGINT NOT NULL REFERENCES appeals(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_status appeal_status,
    new_status appeal_status NOT NULL,
    action VARCHAR(100) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_appeal_history_appeal_id ON appeal_history(appeal_id);
CREATE INDEX idx_appeal_history_created_at ON appeal_history(created_at);

-- +migrate Down
DROP TABLE IF EXISTS appeal_history;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS user_services;
DROP TABLE IF EXISTS category_services;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS appeals;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS appeal_status;
DROP TYPE IF EXISTS user_role;


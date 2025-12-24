package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"citizen-appeals/config"
	"citizen-appeals/pkg/auth"
	"citizen-appeals/pkg/database"
)

func main() {
	email := flag.String("email", "admin@example.com", "Admin email")
	password := flag.String("password", "Admin123!", "Admin password")
	firstName := flag.String("first-name", "Адміністратор", "First name")
	lastName := flag.String("last-name", "Системи", "Last name")
	phone := flag.String("phone", "+380501234567", "Phone number")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Hash password
	passwordHash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	ctx := context.Background()

	// Check if user already exists
	var existingID int64
	err = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", *email).Scan(&existingID)
	if err == nil {
		// User exists, update it
		_, err = db.Pool.Exec(ctx, `
			UPDATE users 
			SET password_hash = $1, first_name = $2, last_name = $3, phone = $4, role = 'admin', is_active = true, updated_at = NOW()
			WHERE id = $5
		`, passwordHash, *firstName, *lastName, *phone, existingID)
		if err != nil {
			log.Fatalf("Failed to update admin user: %v", err)
		}
		fmt.Printf("✅ Admin user updated successfully!\n")
		fmt.Printf("   Email: %s\n", *email)
		fmt.Printf("   Password: %s\n", *password)
		os.Exit(0)
	}

	// Create new admin user
	var userID int64
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, role, is_active)
		VALUES ($1, $2, $3, $4, $5, 'admin', true)
		RETURNING id
	`, *email, passwordHash, *firstName, *lastName, *phone).Scan(&userID)

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("✅ Admin user created successfully!\n")
	fmt.Printf("   ID: %d\n", userID)
	fmt.Printf("   Email: %s\n", *email)
	fmt.Printf("   Password: %s\n", *password)
	fmt.Printf("\nYou can now login with these credentials.\n")
}


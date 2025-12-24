package main

import (
	"context"
	"fmt"
	"log"

	"citizen-appeals/config"
	"citizen-appeals/pkg/auth"
	"citizen-appeals/pkg/database"
)

type UserSeed struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
	Role      string
}

func main() {
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

	ctx := context.Background()

	// Define users to create
	users := []UserSeed{
		{
			Email:     "admin@example.com",
			Password:  "admin@example.com",
			FirstName: "Адміністратор",
			LastName:  "Системи",
			Phone:     "+380501234567",
			Role:      "admin",
		},
		{
			Email:     "disp@disp.example",
			Password:  "disp@disp.example",
			FirstName: "Диспетчер",
			LastName:  "Тестовий",
			Phone:     "+380501234568",
			Role:      "dispatcher",
		},
		{
			Email:     "exec@example.com",
			Password:  "exec@example.com",
			FirstName: "Виконавець",
			LastName:  "Тестовий",
			Phone:     "+380501234569",
			Role:      "executor",
		},
		{
			Email:     "usr@example.com",
			Password:  "usr@example.com",
			FirstName: "Користувач",
			LastName:  "Тестовий",
			Phone:     "+380501234570",
			Role:      "citizen",
		},
	}

	fmt.Println("Creating users...")
	fmt.Println("")

	for _, userSeed := range users {
		// Hash password
		passwordHash, err := auth.HashPassword(userSeed.Password)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", userSeed.Email, err)
			continue
		}

		// Check if user already exists
		var existingID int64
		err = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", userSeed.Email).Scan(&existingID)
		if err == nil {
			// User exists, update it
			_, err = db.Pool.Exec(ctx, `
				UPDATE users 
				SET password_hash = $1, first_name = $2, last_name = $3, phone = $4, role = $5, is_active = true, updated_at = NOW()
				WHERE id = $6
			`, passwordHash, userSeed.FirstName, userSeed.LastName, userSeed.Phone, userSeed.Role, existingID)
			if err != nil {
				log.Printf("Failed to update user %s: %v", userSeed.Email, err)
				continue
			}
			fmt.Printf("✅ Updated user: %s (%s)\n", userSeed.Email, userSeed.Role)
			continue
		}

		// Create new user
		var userID int64
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, first_name, last_name, phone, role, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)
			RETURNING id
		`, userSeed.Email, passwordHash, userSeed.FirstName, userSeed.LastName, userSeed.Phone, userSeed.Role).Scan(&userID)

		if err != nil {
			log.Printf("Failed to create user %s: %v", userSeed.Email, err)
			continue
		}

		fmt.Printf("✅ Created user: %s (%s) - ID: %d\n", userSeed.Email, userSeed.Role, userID)
	}

	fmt.Println("")
	fmt.Println("All users processed!")
	fmt.Println("")
	fmt.Println("Credentials:")
	fmt.Println("  Admin:      admin@example.com / admin@example.com")
	fmt.Println("  Dispatcher: disp@disp.example / disp@disp.example")
	fmt.Println("  Executor:   exec@example.com / exec@example.com")
	fmt.Println("  Citizen:    usr@example.com / usr@example.com")
}


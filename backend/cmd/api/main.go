package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"citizen-appeals/config"
	"citizen-appeals/internal/handler"
	"citizen-appeals/internal/middleware"
	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
	"citizen-appeals/internal/service"
	"citizen-appeals/pkg/auth"
	"citizen-appeals/pkg/classification"
	"citizen-appeals/pkg/database"
	"citizen-appeals/pkg/storage"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

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

	log.Println("Connected to PostgreSQL database")

	// Initialize MongoDB (required for keywords storage)
	if !cfg.MongoDB.Enabled {
		log.Fatalf("MongoDB must be enabled. Set MONGODB_ENABLED=true")
	}

	mongoDB, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v. MongoDB is required for keywords storage.", err)
	}
	defer mongoDB.Close()

	log.Println("Connected to MongoDB database")

	// Initialize embedding repository
	collection := mongoDB.Database.Collection("service_embeddings")
	embeddingRepo := repository.NewServiceEmbeddingRepository(collection)

	// Create indexes
	if err := embeddingRepo.CreateIndexes(context.Background()); err != nil {
		log.Printf("Warning: Failed to create MongoDB indexes: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Pool)
	appealRepo := repository.NewAppealRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)
	serviceRepo := repository.NewServiceRepository(db.Pool)
	photoRepo := repository.NewPhotoRepository(db.Pool)
	categoryServiceRepo := repository.NewCategoryServiceRepository(db.Pool)
	userServiceRepo := repository.NewUserServiceRepository(db.Pool)
	commentRepo := repository.NewCommentRepository(db.Pool)
	notificationRepo := repository.NewNotificationRepository(db.Pool)

	// Initialize services
	tokenService := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.Expiration)
	classifier := classification.NewClassifier(cfg.Classification.ServiceURL, cfg.Classification.Enabled)
	systemSettingsHandler := handler.NewSystemSettingsHandler("config/system_settings.json")

	// Create a loader function for system settings
	systemSettingsLoader := func(ctx context.Context) (*models.SystemSettings, error) {
		return systemSettingsHandler.GetSettings()
	}

	appealService := service.NewAppealService(appealRepo, serviceRepo, classifier, systemSettingsLoader)
	notificationService := service.NewNotificationService(notificationRepo, userRepo, appealRepo, serviceRepo)

	// Initialize storage
	fileStorage, err := storage.NewLocalStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize handlers
	validator := validator.New()
	authHandler := handler.NewAuthHandler(userRepo, tokenService)
	userHandler := handler.NewUserHandler(userRepo, validator)
	appealHandler := handler.NewAppealHandler(appealRepo, appealService, notificationService)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)
	// Формуємо URL бекенду для синхронізації (використовуємо localhost замість 0.0.0.0)
	backendHost := cfg.Server.Host
	if backendHost == "0.0.0.0" {
		backendHost = "localhost"
	}
	backendURL := "http://" + backendHost + ":" + cfg.Server.Port
	serviceHandler := handler.NewServiceHandler(serviceRepo, embeddingRepo, cfg.Classification.ServiceURL, backendURL)
	categoryServiceHandler := handler.NewCategoryServiceHandler(categoryServiceRepo)
	userServiceHandler := handler.NewUserServiceHandler(userServiceRepo)
	photoHandler := handler.NewPhotoHandler(photoRepo, appealRepo, fileStorage)
	commentHandler := handler.NewCommentHandler(commentRepo, appealRepo, notificationService)
	notificationHandler := handler.NewNotificationHandler(notificationRepo)

	// Setup router
	r := chi.NewRouter()

	// Apply global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(middleware.SecureHeaders)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Static file serving for uploads (public access to uploaded files)
	// Note: This should be before the /api routes to avoid conflicts
	uploadPath := cfg.Upload.UploadPath
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadPath))))

	// Public endpoint for classification service (no auth required)
	r.Get("/api/services/for-classification", serviceHandler.GetForClassification)

	// Auth routes (public + protected)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
		// Protected auth routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(tokenService))
			r.Get("/me", authHandler.Me)
			r.Put("/profile", authHandler.UpdateProfile)
			r.Put("/change-password", authHandler.ChangePassword)
		})
	})

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(tokenService))

		// System settings routes:
		// - GET: доступний для всіх автентифікованих користувачів (щоб карта брала центр/зум)
		// - PUT: лише для адміна (зміна налаштувань системи)
		r.Route("/system-settings", func(r chi.Router) {
			// Read for any authenticated user
			r.Get("/", systemSettingsHandler.Get)

			// Update only for admin
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleAdmin))
				r.Put("/", systemSettingsHandler.Update)
			})
		})

		// Categories routes (public read, admin write)
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", categoryHandler.List)
			r.Get("/{id}", categoryHandler.GetByID)

			// Admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleAdmin))
				r.Post("/", categoryHandler.Create)
				r.Put("/{id}", categoryHandler.Update)
				r.Delete("/{id}", categoryHandler.Delete)
			})
		})

		// Services routes (public read, admin write)
		r.Route("/services", func(r chi.Router) {
			r.Get("/", serviceHandler.List)
			r.Get("/{id}", serviceHandler.GetByID)

			// Admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleAdmin))
				r.Post("/", serviceHandler.Create)
				r.Put("/{id}", serviceHandler.Update)
				r.Delete("/{id}", serviceHandler.Delete)
			})
		})

		// Category-Service assignment routes (dispatcher, admin)
		r.Route("/category-services", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleDispatcher, models.RoleAdmin))
				r.Get("/", categoryServiceHandler.GetAll)
				r.Get("/category/{category_id}", categoryServiceHandler.GetByCategoryID)
				r.Post("/assign", categoryServiceHandler.AssignServices)
				r.Delete("/category/{category_id}/service/{service_id}", categoryServiceHandler.Delete)
			})
		})

		// User-Service assignment routes
		r.Route("/user-services", func(r chi.Router) {
			// Get my services (for executors to see their own services)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleExecutor, models.RoleDispatcher, models.RoleAdmin))
				r.Get("/me", userServiceHandler.GetMyServices)
			})

			// Admin/Dispatcher routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleDispatcher, models.RoleAdmin))
				r.Get("/", userServiceHandler.GetAll)
				r.Get("/service/{service_id}", userServiceHandler.GetByServiceID)
				r.Post("/assign", userServiceHandler.AssignUsers)
				r.Delete("/service/{service_id}/user/{user_id}", userServiceHandler.Delete)
			})
		})

		// Appeals routes
		r.Route("/appeals", func(r chi.Router) {
			r.Get("/", appealHandler.List)
			r.Post("/", appealHandler.Create)

			// Statistics (admin, dispatcher) - must be before /{id}
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleDispatcher, models.RoleAdmin))
				r.Get("/statistics", appealHandler.GetStatistics)
			})

			// Dashboards - must be before /{id}
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleDispatcher, models.RoleAdmin))
				r.Get("/dashboard/dispatcher", appealHandler.GetDispatcherDashboard)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleAdmin))
				r.Get("/dashboard/admin", appealHandler.GetAdminDashboard)
			})
			// Service statistics доступні всім авторизованим користувачам
			r.Get("/services/{service_id}/statistics", appealHandler.GetServiceStatistics)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleExecutor))
				r.Get("/dashboard/executor", appealHandler.GetExecutorDashboard)
			})

			// Classification (public for all authenticated users) - must be before /{id}
			r.Post("/classify", appealHandler.Classify)

			// Specific routes that must come before /{id}
			r.Get("/{id}/history", appealHandler.GetHistory)

			// Photos routes - must be before /{id}
			r.Route("/{id}/photos", func(r chi.Router) {
				r.Get("/", photoHandler.List)
				r.Post("/", photoHandler.Upload)
			})

			// Status update (dispatcher, executor, admin)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleDispatcher, models.RoleExecutor, models.RoleAdmin))
				r.Patch("/{id}/status", appealHandler.UpdateStatus)
				r.Patch("/{id}/priority", appealHandler.UpdatePriority)
			})

			// Assign appeal (dispatcher, admin)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleDispatcher, models.RoleAdmin))
				r.Patch("/{id}/assign", appealHandler.Assign)
			})

			// General routes (must be last)
			r.Get("/{id}", appealHandler.GetByID)
			r.Put("/{id}", appealHandler.Update)
		})

		// Photos routes (standalone)
		r.Route("/photos", func(r chi.Router) {
			r.Get("/{id}", photoHandler.Get)
			r.Delete("/{id}", photoHandler.Delete)
		})

		// Users routes
		r.Route("/users", func(r chi.Router) {
			// List users (admin only) or executors (dispatcher/admin)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleDispatcher, models.RoleExecutor, models.RoleAdmin))
				r.Get("/", userHandler.List) // Can filter by ?role=executor
			})
			// Other user operations (admin only)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(models.RoleAdmin))
				r.Get("/{id}", userHandler.GetByID)
				r.Put("/{id}", userHandler.Update)
				r.Delete("/{id}", userHandler.Delete)
			})
		})

		// Comments routes
		r.Route("/appeals/{appeal_id}/comments", func(r chi.Router) {
			r.Get("/", commentHandler.GetByAppealID)
			r.Post("/", commentHandler.Create)
		})

		r.Route("/comments", func(r chi.Router) {
			r.Put("/{id}", commentHandler.Update)
			r.Delete("/{id}", commentHandler.Delete)
		})

		// Notifications routes
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", notificationHandler.List)
			r.Get("/unread-count", notificationHandler.GetUnreadCount)
			r.Put("/{id}/read", notificationHandler.MarkAsRead)
			r.Put("/read-all", notificationHandler.MarkAllAsRead)
			r.Delete("/{id}", notificationHandler.Delete)
		})
	})

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

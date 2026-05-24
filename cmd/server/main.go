package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/pli/absensi-api/internal/attendance"
	"github.com/pli/absensi-api/internal/auth"
	"github.com/pli/absensi-api/internal/employee"
	"github.com/pli/absensi-api/internal/leave"
	"github.com/pli/absensi-api/pkg/cleanup"
	"github.com/pli/absensi-api/pkg/database"
	"github.com/pli/absensi-api/pkg/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using environment variables")
	}

	db, err := database.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer db.Close()

	storageClient := storage.New(
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY"),
		os.Getenv("R2_SECRET_KEY"),
		os.Getenv("R2_BUCKET"),
		os.Getenv("R2_PUBLIC_URL"),
	)

	jwtSecret := os.Getenv("JWT_SECRET")

	// Wire up
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, jwtSecret)
	authHandler := auth.NewHandler(authService)

	attendanceRepo := attendance.NewRepository(db)
	attendanceService := attendance.NewService(attendanceRepo, storageClient)
	attendanceHandler := attendance.NewHandler(attendanceService)

	leaveRepo := leave.NewRepository(db)
	leaveService := leave.NewService(leaveRepo)
	leaveHandler := leave.NewHandler(leaveService)

	employeeService := employee.NewService(authRepo, authService)
	employeeHandler := employee.NewHandler(employeeService)

	cleanup.New(db, storageClient).Start()

	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api/v1")

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Public
	api.POST("/auth/login", authHandler.Login)

	// Protected (semua karyawan)
	protected := api.Group("")
	protected.Use(auth.JWTMiddleware(jwtSecret))

	protected.POST("/attendance/check-in", attendanceHandler.CheckIn)
	protected.POST("/attendance/check-out", attendanceHandler.CheckOut)
	protected.GET("/attendance/today", attendanceHandler.Today)
	protected.GET("/attendance/history", attendanceHandler.History)

	protected.POST("/leave", leaveHandler.Create)
	protected.GET("/leave", leaveHandler.MyLeaves)

	// Admin only
	admin := protected.Group("")
	admin.Use(auth.AdminMiddleware())

	admin.GET("/attendance/all", attendanceHandler.AllAttendances)
	admin.GET("/leave/all", leaveHandler.AllLeaves)
	admin.PUT("/leave/:id/approve", leaveHandler.Approve)
	admin.PUT("/leave/:id/reject", leaveHandler.Reject)
	admin.GET("/employees", employeeHandler.List)
	admin.POST("/employees", employeeHandler.Create)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server jalan di port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

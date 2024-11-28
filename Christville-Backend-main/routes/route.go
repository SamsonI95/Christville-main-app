package routes

import (
	"time"

	"bossblock/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// User routes
	router.POST("/user", controllers.GetOrCreateUser)
	router.GET("/user/:userId", controllers.GetUserByID)
	router.GET("/referred-users/:userId", controllers.GetReferredUsers)
	router.POST("/claim-daily-bonus/:userId", controllers.ClaimDailyBonus)

	// Game routes
	router.GET("/daily-verse", controllers.GetDailyBibleVerse)

	// Leaderboard routes
	router.GET("/leaderboard", controllers.GetLeaderboard)
	
	// Prayer Wall routes
	router.POST("/prayer", controllers.CreatePrayer)
	router.GET("/all-prayers", controllers.GetAllPrayers)
	router.GET("/prayer/user/:userId", controllers.GetUserPrayers)
	
	// Quiz Routes
	router.POST("/quiz", controllers.CreateQuiz)
	router.GET("/quiz", controllers.GetDailyQuiz)	
	router.POST("/submit-quiz", controllers.SubmitQuiz)
	
}

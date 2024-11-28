package main

import (
	"bossblock/controllers"
	"bossblock/db"
	"bossblock/routes"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	
)

func main() {
	// Connect to MongoDB
	client, err := db.ConnectMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())


	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router)


	log.Println("Adding cron job...")
	c := cron.New(cron.WithSeconds()) // Enable seconds precision for cron expressions

	_, err = c.AddFunc("0 0 0 * * *", func() { // Runs daily at midnight
		log.Println("Cron job triggered: Running daily Bible verse fetch...")
	
		// Fetch and store new random Bible verse from external API via combined function.
		_, err := controllers.FetchRandomBibleVerse()
		if err != nil {
			log.Printf("Error fetching and storing Bible verse: %v", err)
			return
		}
	
		log.Println("Successfully fetched and stored new daily Bible verse.")
	})

	if err != nil {
		log.Fatalf("Error adding cron job: %v", err)
	} else {
		log.Println("Cron job added successfully")
	}

	c.Start() // Start the cron scheduler

	go func() {
		log.Println("BossBlock Server started on port 8080...")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatal(err)
		}
	}()

	select {} // Block forever to keep both cron and HTTP server running
}

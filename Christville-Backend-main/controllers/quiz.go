package controllers

import (
	"christville/db"
	"christville/model"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


func CreateQuiz(c *gin.Context) {
	client := db.GetClient()
	database := client.Database("Christville") 

	var newQuiz model.Quiz

	if err := c.ShouldBindJSON(&newQuiz); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	newQuiz.ID = primitive.NewObjectID() 

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.Collection("quizzes").InsertOne(ctx, newQuiz)
	if err != nil {
		log.Printf("Failed to insert quiz: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert quiz"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Quiz created successfully", "quiz_id": newQuiz.ID})
}

func GetDailyQuiz(c *gin.Context) {
	client := db.GetClient()
	database := client.Database("Christville") 

	// Get today's date in "YYYY-MM-DD" format
	today := time.Now().Format("2006-01-02")

	var quiz model.Quiz

	// Find today's quiz based on the date
	err := database.Collection("quizzes").FindOne(context.Background(), bson.M{"date": today}).Decode(&quiz)
	if err != nil {
		log.Printf("Failed to fetch today's quiz: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch today's quiz"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quiz_id":   quiz.ID,
		"title":     quiz.Title,
		"questions": quiz.Questions,
	})
}

// SubmitQuiz handles user's submission of their answers and calculates their score
func SubmitQuiz(c *gin.Context) {
	client := db.GetClient()
	database := client.Database("Christville") 

	var submission struct {
		UserID      string   `json:"user_id"`
		QuizID      string   `json:"quiz_id"`
		UserAnswers []struct {
			QuestionID  string `json:"question_id"`
			SelectedIdx int    `json:"selected_idx"`
		} `json:"user_answers"`
	}

	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(submission.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var score int

	for _, userAnswer := range submission.UserAnswers {
		
        questionObjectID, err := primitive.ObjectIDFromHex(userAnswer.QuestionID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
            return
        }

        var question model.Question

        // Fetch the question from the database to validate the answer
        err = database.Collection("quizzes").FindOne(context.Background(), bson.M{
            "_id": submission.QuizID,
            "questions._id": questionObjectID,
        }).Decode(&question)
        
        if err != nil {
            log.Printf("Failed to fetch question for validation: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate answer"})
            return
        }

        // Check if user's selected answer is correct
        isCorrect := (userAnswer.SelectedIdx == question.CorrectIdx)

        if isCorrect {
            score++
        }

        // Store user's answer in DB for future reference (optional)
        userAnswerRecord := model.UserAnswer{
            UserID:      userObjectID,
            QuestionID:  questionObjectID,
            SelectedIdx: userAnswer.SelectedIdx,
            IsCorrect:   isCorrect,
        }

        _, err = database.Collection("user_answers").InsertOne(context.Background(), userAnswerRecord)
        if err != nil {
            log.Printf("Failed to store user's answer: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store user's answer"})
            return
        }
    }

    // Return the user's score after validation
    c.JSON(http.StatusOK, gin.H{
        "score": score,
    })
}
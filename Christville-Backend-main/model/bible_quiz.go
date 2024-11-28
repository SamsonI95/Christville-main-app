package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Question struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Text       string             `bson:"text"`        // The question text
	Options    []string           `bson:"options"`     // List of 4 options
	CorrectIdx int                `bson:"correct_idx"` // Index of the correct option (0-3)
}

type Quiz struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`       // Title of the quiz
	Questions []Question         `bson:"questions"`   // List of questions in the quiz
	Date      string             `bson:"date"`        // Date of this quiz (optional)
}

type UserAnswer struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`   // Unique ID for the answer
	UserID      primitive.ObjectID `bson:"user_id"`         // ID of the user who answered
	QuestionID  primitive.ObjectID `bson:"question_id"`     // ID of the question being answered
	SelectedIdx int                `bson:"selected_idx"`    // Index of the selected option (0-3)
	IsCorrect   bool               `bson:"is_correct"`      // Whether the selected answer is correct
}
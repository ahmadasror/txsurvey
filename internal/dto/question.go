package dto

import "github.com/ahmadasror/txsurvey/internal/model"

// QuestionInput is the body for creating/updating a question.
type QuestionInput struct {
	Type        model.QuestionType     `json:"type" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Metadata    model.QuestionMetadata `json:"metadata"`
}

// ReorderRequest is the body for PUT /forms/:id/questions/reorder.
type ReorderRequest struct {
	OrderedIDs []string `json:"ordered_ids" binding:"required"`
}

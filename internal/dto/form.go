package dto

import "github.com/ahmadasror/txsurvey/internal/model"

// CreateFormRequest is the body for POST /forms.
type CreateFormRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// UpdateFormRequest is the body for PATCH /forms/:id.
type UpdateFormRequest struct {
	Title       string             `json:"title" binding:"required"`
	Description string             `json:"description"`
	Settings    model.FormSettings `json:"settings"`
}

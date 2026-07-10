package dto

import "github.com/ahmadasror/txsurvey/internal/model"

// CreateFormRequest is the body for POST /forms.
type CreateFormRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// UpdateFormRequest is the body for PATCH /forms/:id. Slug is optional: when
// non-empty and different from the current slug it retargets the public URL —
// only allowed while the form is a draft (see FormService.Update).
type UpdateFormRequest struct {
	Title       string             `json:"title" binding:"required"`
	Description string             `json:"description"`
	Slug        string             `json:"slug"`
	Settings    model.FormSettings `json:"settings"`
}

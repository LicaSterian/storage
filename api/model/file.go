package model

import "github.com/gofrs/uuid"

// File struct
type File struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt int64     `json:"created_at"`
	Size      int64     `json:"size"`
}

// GetAllRequest is the object that binds to the /files/:tenantID request
type GetAllRequest struct {
	ID           int64           `json:"id"`
	Page         int             `json:"page"`
	PerPage      int             `json:"perPage"`
	FilterFields []string        `json:"filterFields"`
	FilterValues [][]interface{} `json:"filterValues"`
	SortBy       string          `json:"sortBy"`
	SortAsc      bool            `json:"sortAsc"`
	Fields       []string        `json:"fields"`
}

// Response struct
type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// MessageResponse struct
type MessageResponse struct {
	Response
	Message string `json:"message"`
}

// GetAllResponse struct
type GetAllResponse struct {
	RequestID int64 `json:"request_id,omitempty"`
	Response
	Data GetAllResponseData `json:"data"`
}

// GetAllResponseData struct
type GetAllResponseData struct {
	Rows  interface{} `json:"rows"`
	Total int64       `json:"total"`
}

// UploadResponse struct
type UploadResponse struct {
	Response
	File File `json:"file"`
}

package models

import (
	"github.com/google/uuid"
)

type ApiAccountResponse struct {
	Message string `json:"message"`
	Data    struct {
		Role    string `json:"role"`
		EventId int64  `json:"event_id"`
	} `json:"data"`
}

type User struct {
	Id           int       `json:"id"`
	LastName     string    `json:"last_name"`
	FirstName    string    `json:"first_name"`
	FatherName   string    `json:"father_name"`
	Birthday     string    `json:"birthday"`
	Email        string    `json:"email"`
	RunetId      int64     `json:"runet_id"`
	Gender       string    `json:"gender"` //ENUM ``
	Visible      bool      `json:"visible"`
	PrimaryPhone string    `json:"phone"`
	Verified     bool      `json:"verified"`
	Uuid         uuid.UUID `json:"uuid"`
	Photo        string    `json:"photo"`
}

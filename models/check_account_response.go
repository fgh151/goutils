package models

type ApiAccountResponse struct {
	Message string `json:"message"`
	Data    struct {
		Role    string `json:"role"`
		EventId string `json:"event_id"`
	}
}

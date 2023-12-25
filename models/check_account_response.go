package models

type ApiAccountResponse struct {
	Message string `json:"message"`
	Data    struct {
		Role    string `json:"role"`
		EventId int64  `json:"event_id"`
	} `json:"data"`
}

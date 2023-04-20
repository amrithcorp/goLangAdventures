package models

import (
	"github.com/google/uuid"
)

type Account struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Username        string    `json:"username" gorm:"unique"`
	Email           string    `json:"email" gorm:"unique"`
	PaymentRequests []PaymentRequest
}

type PaymentRequest struct {
	AccountID uint
	Amount    float64
	Memo      string
	AssetCode string
}

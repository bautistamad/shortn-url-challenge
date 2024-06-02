package entities

import "github.com/google/uuid"

type URL struct {
	LongURL  string    `json:"longURL" gorm:"column:longurl"`
	ShortURL string    `json:"shortUrl" gorm:"column:shorturl"`
	ID       uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()" `
}

func (URL) TableName() string {
	return "url"
}
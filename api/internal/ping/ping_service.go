package ping

import (
	"github.com/chattycathy/api/db/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type PingResponse struct {
	Message string `json:"message"`
}

// Ping creates a ping record and returns pong
func (s *Service) Ping() (*PingResponse, error) {
	ping := &models.Ping{
		Message: "pong",
	}

	if err := s.db.Create(ping).Error; err != nil {
		return nil, err
	}

	return &PingResponse{Message: ping.Message}, nil
}

package local

import (
	"context"
	"log"
)

type Service struct {
}

func NewLocalSMSService() *Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	log.Printf("send code:%s\n", args[0])
	return nil
}

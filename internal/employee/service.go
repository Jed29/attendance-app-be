package employee

import (
	"context"

	"github.com/pli/absensi-api/internal/auth"
)

type Service struct {
	repo    *auth.Repository
	authSvc *auth.Service
}

func NewService(repo *auth.Repository, authSvc *auth.Service) *Service {
	return &Service{repo: repo, authSvc: authSvc}
}

func (s *Service) List(ctx context.Context) ([]auth.Employee, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, req auth.CreateEmployeeRequest) (*auth.Employee, error) {
	hash, err := s.authSvc.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, &req, hash)
}

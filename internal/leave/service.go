package leave

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, employeeID uuid.UUID, req CreateRequest) (*Request, error) {
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("format start_date salah, gunakan YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("format end_date salah, gunakan YYYY-MM-DD")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end_date harus setelah start_date")
	}
	return s.repo.Create(ctx, employeeID, req.LeaveType, start, end, req.Reason)
}

func (s *Service) MyLeaves(ctx context.Context, employeeID uuid.UUID, status string) ([]Request, error) {
	return s.repo.GetByEmployee(ctx, employeeID, status)
}

func (s *Service) AllLeaves(ctx context.Context, status string) ([]Request, error) {
	return s.repo.GetAll(ctx, status)
}

func (s *Service) Approve(ctx context.Context, requestID, approverID uuid.UUID) (*Request, error) {
	return s.repo.UpdateStatus(ctx, requestID, approverID, "approved")
}

func (s *Service) Reject(ctx context.Context, requestID, approverID uuid.UUID) (*Request, error) {
	return s.repo.UpdateStatus(ctx, requestID, approverID, "rejected")
}

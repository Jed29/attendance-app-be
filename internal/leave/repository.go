package leave

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, employeeID uuid.UUID, leaveType string, start, end time.Time, reason string) (*Request, error) {
	var req Request
	err := r.db.QueryRow(ctx,
		`INSERT INTO leave_requests (employee_id, leave_type, start_date, end_date, reason)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, employee_id, leave_type, start_date, end_date, reason, status, approved_by, created_at`,
		employeeID, leaveType, start, end, reason,
	).Scan(&req.ID, &req.EmployeeID, &req.LeaveType, &req.StartDate, &req.EndDate,
		&req.Reason, &req.Status, &req.ApprovedBy, &req.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create leave request: %w", err)
	}
	return &req, nil
}

func (r *Repository) GetByEmployee(ctx context.Context, employeeID uuid.UUID, status string) ([]Request, error) {
	query := `SELECT id, employee_id, leave_type, start_date, end_date, reason, status, approved_by, created_at
	          FROM leave_requests WHERE employee_id = $1`
	args := []interface{}{employeeID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}

func (r *Repository) GetAll(ctx context.Context, status string) ([]Request, error) {
	query := `SELECT id, employee_id, leave_type, start_date, end_date, reason, status, approved_by, created_at
	          FROM leave_requests`
	var args []interface{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}

func (r *Repository) UpdateStatus(ctx context.Context, id, approverID uuid.UUID, status string) (*Request, error) {
	var req Request
	err := r.db.QueryRow(ctx,
		`UPDATE leave_requests SET status = $1, approved_by = $2
		 WHERE id = $3
		 RETURNING id, employee_id, leave_type, start_date, end_date, reason, status, approved_by, created_at`,
		status, approverID, id,
	).Scan(&req.ID, &req.EmployeeID, &req.LeaveType, &req.StartDate, &req.EndDate,
		&req.Reason, &req.Status, &req.ApprovedBy, &req.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update leave status: %w", err)
	}
	return &req, nil
}

func scanRequests(rows interface{ Next() bool; Scan(...interface{}) error; Close() }) ([]Request, error) {
	defer rows.Close()
	var reqs []Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.EmployeeID, &req.LeaveType, &req.StartDate, &req.EndDate,
			&req.Reason, &req.Status, &req.ApprovedBy, &req.CreatedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}
	return reqs, nil
}

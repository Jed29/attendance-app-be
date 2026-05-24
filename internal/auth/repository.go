package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*Employee, string, error) {
	var e Employee
	var hash string
	err := r.db.QueryRow(ctx,
		`SELECT id, office_id, name, email, role, password_hash, created_at
		 FROM employees WHERE email = $1`, email,
	).Scan(&e.ID, &e.OfficeID, &e.Name, &e.Email, &e.Role, &hash, &e.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("employee not found")
	}
	return &e, hash, nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	var e Employee
	err := r.db.QueryRow(ctx,
		`SELECT id, office_id, name, email, role, created_at
		 FROM employees WHERE id = $1`, id,
	).Scan(&e.ID, &e.OfficeID, &e.Name, &e.Email, &e.Role, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("employee not found")
	}
	return &e, nil
}

func (r *Repository) Create(ctx context.Context, req *CreateEmployeeRequest, passwordHash string) (*Employee, error) {
	var e Employee
	officeID, _ := uuid.Parse(req.OfficeID)
	err := r.db.QueryRow(ctx,
		`INSERT INTO employees (office_id, name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, office_id, name, email, role, created_at`,
		officeID, req.Name, req.Email, passwordHash, req.Role,
	).Scan(&e.ID, &e.OfficeID, &e.Name, &e.Email, &e.Role, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create employee: %w", err)
	}
	return &e, nil
}

func (r *Repository) List(ctx context.Context) ([]Employee, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, office_id, name, email, role, created_at FROM employees ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var e Employee
		if err := rows.Scan(&e.ID, &e.OfficeID, &e.Name, &e.Email, &e.Role, &e.CreatedAt); err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, nil
}

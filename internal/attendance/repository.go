package attendance

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

func (r *Repository) GetOfficeByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*Office, error) {
	var o Office
	err := r.db.QueryRow(ctx,
		`SELECT o.id, o.lat, o.lng, o.radius_meters
		 FROM offices o
		 JOIN employees e ON e.office_id = o.id
		 WHERE e.id = $1`, employeeID,
	).Scan(&o.ID, &o.Lat, &o.Lng, &o.RadiusMeters)
	if err != nil {
		return nil, fmt.Errorf("office not found for employee")
	}
	return &o, nil
}

func (r *Repository) GetToday(ctx context.Context, employeeID uuid.UUID) (*Record, error) {
	var rec Record
	err := r.db.QueryRow(ctx,
		`SELECT id, employee_id, check_in_at, check_out_at, check_in_lat, check_in_lng, photo_url, status, created_at
		 FROM attendances
		 WHERE employee_id = $1 AND DATE(check_in_at AT TIME ZONE 'Asia/Jakarta') = CURRENT_DATE AT TIME ZONE 'Asia/Jakarta'
		 LIMIT 1`, employeeID,
	).Scan(&rec.ID, &rec.EmployeeID, &rec.CheckInAt, &rec.CheckOutAt,
		&rec.CheckInLat, &rec.CheckInLng, &rec.PhotoURL, &rec.Status, &rec.CreatedAt)
	if err != nil {
		return nil, nil // tidak ada record = belum absen
	}
	return &rec, nil
}

func (r *Repository) Create(ctx context.Context, employeeID uuid.UUID, lat, lng float64, photoURL, status string) (*Record, error) {
	var rec Record
	err := r.db.QueryRow(ctx,
		`INSERT INTO attendances (employee_id, check_in_lat, check_in_lng, photo_url, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, employee_id, check_in_at, check_out_at, check_in_lat, check_in_lng, photo_url, status, created_at`,
		employeeID, lat, lng, photoURL, status,
	).Scan(&rec.ID, &rec.EmployeeID, &rec.CheckInAt, &rec.CheckOutAt,
		&rec.CheckInLat, &rec.CheckInLng, &rec.PhotoURL, &rec.Status, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create attendance: %w", err)
	}
	return &rec, nil
}

func (r *Repository) SetCheckOut(ctx context.Context, attendanceID uuid.UUID) (*Record, error) {
	var rec Record
	err := r.db.QueryRow(ctx,
		`UPDATE attendances SET check_out_at = NOW()
		 WHERE id = $1
		 RETURNING id, employee_id, check_in_at, check_out_at, check_in_lat, check_in_lng, photo_url, status, created_at`,
		attendanceID,
	).Scan(&rec.ID, &rec.EmployeeID, &rec.CheckInAt, &rec.CheckOutAt,
		&rec.CheckInLat, &rec.CheckInLng, &rec.PhotoURL, &rec.Status, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("checkout failed: %w", err)
	}
	return &rec, nil
}

func (r *Repository) GetHistory(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]Record, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, employee_id, check_in_at, check_out_at, check_in_lat, check_in_lng, photo_url, status, created_at
		 FROM attendances
		 WHERE employee_id = $1 AND check_in_at BETWEEN $2 AND $3
		 ORDER BY check_in_at DESC`,
		employeeID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.EmployeeID, &rec.CheckInAt, &rec.CheckOutAt,
			&rec.CheckInLat, &rec.CheckInLng, &rec.PhotoURL, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *Repository) GetAllToday(ctx context.Context) ([]Record, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, employee_id, check_in_at, check_out_at, check_in_lat, check_in_lng, photo_url, status, created_at
		 FROM attendances
		 WHERE DATE(check_in_at AT TIME ZONE 'Asia/Jakarta') = CURRENT_DATE AT TIME ZONE 'Asia/Jakarta'
		 ORDER BY check_in_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.EmployeeID, &rec.CheckInAt, &rec.CheckOutAt,
			&rec.CheckInLat, &rec.CheckInLng, &rec.PhotoURL, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

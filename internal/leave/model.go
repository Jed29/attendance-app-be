package leave

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	ID         uuid.UUID  `json:"id"`
	EmployeeID uuid.UUID  `json:"employee_id"`
	LeaveType  string     `json:"leave_type"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    time.Time  `json:"end_date"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	ApprovedBy *uuid.UUID `json:"approved_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateRequest struct {
	LeaveType string `json:"leave_type" binding:"required,oneof=sick annual other"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Reason    string `json:"reason"`
}

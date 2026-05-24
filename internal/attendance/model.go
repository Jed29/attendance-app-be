package attendance

import (
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID          uuid.UUID  `json:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	CheckInAt   time.Time  `json:"check_in_at"`
	CheckOutAt  *time.Time `json:"check_out_at"`
	CheckInLat  float64    `json:"check_in_lat"`
	CheckInLng  float64    `json:"check_in_lng"`
	PhotoURL    string     `json:"photo_url"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Office struct {
	ID           uuid.UUID
	Lat          float64
	Lng          float64
	RadiusMeters int
}

type TodayStatus struct {
	Status string  `json:"status"` // absent | checked_in | checked_out
	Record *Record `json:"record,omitempty"`
}

type HistoryQuery struct {
	From string `form:"from"`
	To   string `form:"to"`
}

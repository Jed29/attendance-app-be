package attendance

import (
	"context"
	"fmt"
	"math"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/pli/absensi-api/pkg/storage"
)

type Service struct {
	repo    *Repository
	storage *storage.Client
}

func NewService(repo *Repository, storage *storage.Client) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) CheckIn(ctx context.Context, employeeID uuid.UUID, lat, lng float64, photo multipart.File, photoHeader *multipart.FileHeader) (*Record, error) {
	existing, err := s.repo.GetToday(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("sudah absen masuk hari ini")
	}

	office, err := s.repo.GetOfficeByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	dist := haversine(lat, lng, office.Lat, office.Lng)
	if dist > float64(office.RadiusMeters) {
		return nil, fmt.Errorf("lokasi terlalu jauh dari kantor (%.0f meter, maks %d meter)", dist, office.RadiusMeters)
	}

	photoURL := ""
	if photo != nil {
		key := fmt.Sprintf("checkin/%s/%d%s", employeeID, time.Now().Unix(), fileExt(photoHeader.Filename))
		url, err := s.storage.Upload(ctx, key, photo, photoHeader.Header.Get("Content-Type"))
		if err != nil {
			return nil, fmt.Errorf("upload foto gagal: %w", err)
		}
		photoURL = url
	}

	status := "present"
	// TODO: bandingkan dengan shift untuk menentukan status 'late'

	return s.repo.Create(ctx, employeeID, lat, lng, photoURL, status)
}

func (s *Service) CheckOut(ctx context.Context, employeeID uuid.UUID) (*Record, error) {
	existing, err := s.repo.GetToday(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("belum absen masuk hari ini")
	}
	if existing.CheckOutAt != nil {
		return nil, fmt.Errorf("sudah absen keluar hari ini")
	}
	return s.repo.SetCheckOut(ctx, existing.ID)
}

func (s *Service) TodayStatus(ctx context.Context, employeeID uuid.UUID) (*TodayStatus, error) {
	rec, err := s.repo.GetToday(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return &TodayStatus{Status: "absent"}, nil
	}
	status := "checked_in"
	if rec.CheckOutAt != nil {
		status = "checked_out"
	}
	return &TodayStatus{Status: status, Record: rec}, nil
}

func (s *Service) History(ctx context.Context, employeeID uuid.UUID, fromStr, toStr string) ([]Record, error) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t.Add(24 * time.Hour)
		}
	}
	return s.repo.GetHistory(ctx, employeeID, from, to)
}

func (s *Service) AllToday(ctx context.Context) ([]Record, error) {
	return s.repo.GetAllToday(ctx)
}

// haversine returns distance in meters between two lat/lng coordinates.
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	lat1R := lat1 * math.Pi / 180
	lat2R := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1R)*math.Cos(lat2R)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func fileExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ".jpg"
}

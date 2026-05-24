package cleanup

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pli/absensi-api/pkg/storage"
)

type Job struct {
	db      *pgxpool.Pool
	storage *storage.Client
}

func New(db *pgxpool.Pool, storage *storage.Client) *Job {
	return &Job{db: db, storage: storage}
}

// Start menjalankan cleanup setiap hari jam 02:00. Hapus foto > 90 hari.
func (j *Job) Start() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))
			j.run()
		}
	}()
	log.Println("Cleanup job aktif: hapus foto > 90 hari setiap hari jam 02:00")
}

func (j *Job) run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cutoff := time.Now().AddDate(0, 0, -90)

	rows, err := j.db.Query(ctx,
		`SELECT id, photo_url FROM attendances
		 WHERE check_in_at < $1 AND photo_url != ''`, cutoff)
	if err != nil {
		log.Printf("cleanup: query error: %v", err)
		return
	}
	defer rows.Close()

	type row struct {
		id  string
		url string
	}
	var targets []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.url); err == nil {
			targets = append(targets, r)
		}
	}

	deleted := 0
	for _, t := range targets {
		key := extractKey(t.url)
		if err := j.storage.Delete(ctx, key); err != nil {
			log.Printf("cleanup: hapus R2 gagal (%s): %v", key, err)
			continue
		}
		j.db.Exec(ctx, `UPDATE attendances SET photo_url = '' WHERE id = $1`, t.id)
		deleted++
	}

	log.Printf("cleanup: %d foto dihapus (cutoff: %s)", deleted, cutoff.Format("2006-01-02"))
}

// extractKey ambil path dari full URL: https://pub-xxx.r2.dev/checkin/xxx → checkin/xxx
func extractKey(url string) string {
	parts := strings.SplitN(url, ".dev/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return url
}

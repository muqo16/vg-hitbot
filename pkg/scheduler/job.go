// Package scheduler provides cron-like job scheduling for simulations
package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Job zamanlı iş tanımı
type Job struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Enabled        bool      `json:"enabled"`
	DaysOfWeek     []string  `json:"days_of_week"`     // "monday","tuesday",... veya "daily","weekday","weekend"
	StartHour      int       `json:"start_hour"`        // Başlangıç saati (0-23)
	StartMinute    int       `json:"start_minute"`      // Başlangıç dakikası (0-59)
	Duration       int       `json:"duration"`          // Süre (dakika)
	Domain         string    `json:"domain"`            // Hedef domain (boşsa mevcut config kullanılır)
	HitsPerMinute  int       `json:"hits_per_minute"`   // HPM override (0 = mevcut config)
	MaxConcurrent  int       `json:"max_concurrent"`    // Concurrent override (0 = mevcut config)
	LastRun        time.Time `json:"last_run"`
	NextRun        time.Time `json:"next_run"`
	RunCount       int       `json:"run_count"`
}

// JobStorage iş kalıcılığı
type JobStorage struct {
	mu       sync.Mutex
	filePath string
	jobs     []*Job
}

// NewJobStorage yeni job storage oluşturur
func NewJobStorage(filePath string) *JobStorage {
	s := &JobStorage{
		filePath: filePath,
		jobs:     make([]*Job, 0),
	}
	_ = s.Load()
	return s
}

// Load dosyadan işleri yükler
func (s *JobStorage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var jobs []*Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return err
	}

	s.jobs = jobs
	return nil
}

// Save işleri dosyaya kaydeder
func (s *JobStorage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.jobs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// AddJob yeni iş ekler
func (s *JobStorage) AddJob(job *Job) error {
	s.mu.Lock()
	// ID kontrolü
	for _, j := range s.jobs {
		if j.ID == job.ID {
			s.mu.Unlock()
			return fmt.Errorf("iş zaten var: %s", job.ID)
		}
	}
	s.jobs = append(s.jobs, job)
	s.mu.Unlock()

	return s.Save()
}

// RemoveJob iş siler
func (s *JobStorage) RemoveJob(id string) error {
	s.mu.Lock()
	var filtered []*Job
	for _, j := range s.jobs {
		if j.ID != id {
			filtered = append(filtered, j)
		}
	}
	s.jobs = filtered
	s.mu.Unlock()

	return s.Save()
}

// UpdateJob işi günceller
func (s *JobStorage) UpdateJob(job *Job) error {
	s.mu.Lock()
	for i, j := range s.jobs {
		if j.ID == job.ID {
			s.jobs[i] = job
			break
		}
	}
	s.mu.Unlock()

	return s.Save()
}

// GetJob ID'ye göre iş döner
func (s *JobStorage) GetJob(id string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, j := range s.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}

// ListJobs tüm işleri listeler
func (s *JobStorage) ListJobs() []*Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*Job, len(s.jobs))
	copy(result, s.jobs)
	return result
}

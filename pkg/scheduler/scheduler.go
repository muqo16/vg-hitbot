// Package scheduler provides cron-like job scheduling for simulations
package scheduler

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

// SimulationStartFunc simülasyon başlatma callback fonksiyonu
type SimulationStartFunc func(domain string, durationMin int, hpm int, maxConcurrent int) error

// SimulationStopFunc simülasyon durdurma callback fonksiyonu
type SimulationStopFunc func() error

// Scheduler zamanlı görev yöneticisi
type Scheduler struct {
	mu             sync.Mutex
	storage        *JobStorage
	running        bool
	cancel         context.CancelFunc
	startFn        SimulationStartFunc
	stopFn         SimulationStopFunc
	activeJobID    string
	activeJobTimer *time.Timer
	location       *time.Location
}

// NewScheduler yeni scheduler oluşturur
func NewScheduler(storage *JobStorage, startFn SimulationStartFunc, stopFn SimulationStopFunc) *Scheduler {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		loc = time.UTC
	}

	return &Scheduler{
		storage:  storage,
		startFn:  startFn,
		stopFn:   stopFn,
		location: loc,
	}
}

// SetTimezone saat dilimini ayarlar
func (s *Scheduler) SetTimezone(tz string) error {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.location = loc
	s.mu.Unlock()
	return nil
}

// Start scheduler'ı başlatır
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true
	s.mu.Unlock()

	go s.loop(ctx)
	log.Println("[SCHEDULER] Scheduler başlatıldı")
}

// Stop scheduler'ı durdurur
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}
	if s.activeJobTimer != nil {
		s.activeJobTimer.Stop()
	}
	s.running = false
	log.Println("[SCHEDULER] Scheduler durduruldu")
}

// IsRunning scheduler çalışıyor mu
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// loop ana scheduler döngüsü - her dakika kontrol eder
func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndRunJobs()
		}
	}
}

// checkAndRunJobs zamanı gelen işleri kontrol eder ve çalıştırır
func (s *Scheduler) checkAndRunJobs() {
	s.mu.Lock()
	if s.activeJobID != "" {
		// Zaten çalışan iş var
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	now := time.Now().In(s.location)
	jobs := s.storage.ListJobs()

	for _, job := range jobs {
		if !job.Enabled {
			continue
		}

		// Bugün çalışmalı mı kontrol et
		if !s.shouldRunToday(job, now) {
			continue
		}

		// Saat ve dakika kontrolü
		if now.Hour() != job.StartHour || now.Minute() != job.StartMinute {
			continue
		}

		// Son 2 dakika içinde çalıştıysa atla (duplicate önleme)
		if !job.LastRun.IsZero() && time.Since(job.LastRun) < 2*time.Minute {
			continue
		}

		// İşi çalıştır
		s.runJob(job)
		break // Bir seferde bir iş
	}
}

// shouldRunToday iş bugün çalışmalı mı
func (s *Scheduler) shouldRunToday(job *Job, now time.Time) bool {
	if len(job.DaysOfWeek) == 0 {
		return true // Boşsa her gün
	}

	today := strings.ToLower(now.Weekday().String())
	isWeekday := now.Weekday() >= time.Monday && now.Weekday() <= time.Friday
	isWeekend := now.Weekday() == time.Saturday || now.Weekday() == time.Sunday

	for _, day := range job.DaysOfWeek {
		day = strings.ToLower(strings.TrimSpace(day))
		switch day {
		case "daily":
			return true
		case "weekday", "weekdays":
			if isWeekday {
				return true
			}
		case "weekend", "weekends":
			if isWeekend {
				return true
			}
		default:
			if day == today {
				return true
			}
		}
	}

	return false
}

// runJob işi çalıştırır
func (s *Scheduler) runJob(job *Job) {
	s.mu.Lock()
	s.activeJobID = job.ID
	s.mu.Unlock()

	domain := job.Domain
	hpm := job.HitsPerMinute
	maxConcurrent := job.MaxConcurrent
	duration := job.Duration

	if duration <= 0 {
		duration = 60
	}

	log.Printf("[SCHEDULER] İş başlatılıyor: %s (Domain: %s, Süre: %d dk, HPM: %d)",
		job.Name, domain, duration, hpm)

	// Simülasyonu başlat
	if s.startFn != nil {
		if err := s.startFn(domain, duration, hpm, maxConcurrent); err != nil {
			log.Printf("[SCHEDULER] İş başlatma hatası: %v", err)
			s.mu.Lock()
			s.activeJobID = ""
			s.mu.Unlock()
			return
		}
	}

	// İş bilgilerini güncelle
	job.LastRun = time.Now()
	job.RunCount++
	_ = s.storage.UpdateJob(job)

	// Süre sonunda otomatik durdur
	s.mu.Lock()
	s.activeJobTimer = time.AfterFunc(time.Duration(duration)*time.Minute, func() {
		log.Printf("[SCHEDULER] İş süresi doldu, durduruluyor: %s", job.Name)
		if s.stopFn != nil {
			_ = s.stopFn()
		}
		s.mu.Lock()
		s.activeJobID = ""
		s.activeJobTimer = nil
		s.mu.Unlock()
	})
	s.mu.Unlock()
}

// GetActiveJobID çalışan işin ID'sini döner
func (s *Scheduler) GetActiveJobID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeJobID
}

// GetStorage storage'ı döner
func (s *Scheduler) GetStorage() *JobStorage {
	return s.storage
}

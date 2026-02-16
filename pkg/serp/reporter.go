// Package serp provides SERP simulation reporting
package serp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SERPReporter SERP simülasyon raporlayıcı
type SERPReporter struct {
	mu        sync.Mutex
	enabled   bool
	reportDir string
	data      *ReportData
	startTime time.Time
}

// ReportData rapor verileri
type ReportData struct {
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	Domain         string                 `json:"domain"`
	TotalSearches  int64                  `json:"total_searches"`
	TotalClicks    int64                  `json:"total_clicks"`
	TotalErrors    int64                  `json:"total_errors"`
	EngineStats    map[string]*EngineStat `json:"engine_stats"`
	KeywordStats   map[string]*KeywordStat `json:"keyword_stats"`
	PositionCTR    map[int]*PositionStat  `json:"position_ctr"`
	HourlyStats    map[int]*HourlyStat    `json:"hourly_stats"`
}

// EngineStat arama motoru istatistikleri
type EngineStat struct {
	Engine     string  `json:"engine"`
	Searches   int64   `json:"searches"`
	Clicks     int64   `json:"clicks"`
	Errors     int64   `json:"errors"`
	AvgCTR     float64 `json:"avg_ctr"`
	SuccessRate float64 `json:"success_rate"`
}

// KeywordStat keyword istatistikleri
type KeywordStat struct {
	Keyword     string  `json:"keyword"`
	Searches    int64   `json:"searches"`
	Clicks      int64   `json:"clicks"`
	AvgPosition float64 `json:"avg_position"`
	CTR         float64 `json:"ctr"`
	LastSeen    time.Time `json:"last_seen"`
}

// PositionStat pozisyon bazlı CTR istatistikleri
type PositionStat struct {
	Position    int     `json:"position"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	CTR         float64 `json:"ctr"`
}

// HourlyStat saatlik istatistikler
type HourlyStat struct {
	Hour      int   `json:"hour"`
	Searches  int64 `json:"searches"`
	Clicks    int64 `json:"clicks"`
	Errors    int64 `json:"errors"`
}

// NewSERPReporter yeni SERP reporter oluşturur
func NewSERPReporter(reportDir string, enabled bool) *SERPReporter {
	r := &SERPReporter{
		enabled:   enabled,
		reportDir: reportDir,
		startTime: time.Now(),
		data: &ReportData{
			StartTime:    time.Now(),
			EngineStats:  make(map[string]*EngineStat),
			KeywordStats: make(map[string]*KeywordStat),
			PositionCTR:  make(map[int]*PositionStat),
			HourlyStats:  make(map[int]*HourlyStat),
		},
	}

	// Rapor dizinini oluştur
	if enabled && reportDir != "" {
		_ = os.MkdirAll(reportDir, 0755)
	}

	return r
}

// RecordSearch arama kaydeder
func (r *SERPReporter) RecordSearch(engine, keyword string) {
	if !r.enabled {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.data.TotalSearches++

	// Engine stat
	es, ok := r.data.EngineStats[engine]
	if !ok {
		es = &EngineStat{Engine: engine}
		r.data.EngineStats[engine] = es
	}
	es.Searches++

	// Keyword stat
	ks, ok := r.data.KeywordStats[keyword]
	if !ok {
		ks = &KeywordStat{Keyword: keyword}
		r.data.KeywordStats[keyword] = ks
	}
	ks.Searches++
	ks.LastSeen = time.Now()

	// Hourly stat
	hour := time.Now().Hour()
	hs, ok := r.data.HourlyStats[hour]
	if !ok {
		hs = &HourlyStat{Hour: hour}
		r.data.HourlyStats[hour] = hs
	}
	hs.Searches++
}

// RecordClick tıklama kaydeder
func (r *SERPReporter) RecordClick(engine, keyword string, position int) {
	if !r.enabled {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.data.TotalClicks++

	// Engine stat
	if es, ok := r.data.EngineStats[engine]; ok {
		es.Clicks++
		if es.Searches > 0 {
			es.AvgCTR = float64(es.Clicks) / float64(es.Searches) * 100
		}
	}

	// Keyword stat
	if ks, ok := r.data.KeywordStats[keyword]; ok {
		ks.Clicks++
		// Pozisyon ortalamasını güncelle
		totalClicks := ks.Clicks
		ks.AvgPosition = (ks.AvgPosition*float64(totalClicks-1) + float64(position)) / float64(totalClicks)
		if ks.Searches > 0 {
			ks.CTR = float64(ks.Clicks) / float64(ks.Searches) * 100
		}
	}

	// Position stat
	ps, ok := r.data.PositionCTR[position]
	if !ok {
		ps = &PositionStat{Position: position}
		r.data.PositionCTR[position] = ps
	}
	ps.Clicks++
	ps.Impressions++
	ps.CTR = float64(ps.Clicks) / float64(ps.Impressions) * 100

	// Hourly stat
	hour := time.Now().Hour()
	if hs, ok := r.data.HourlyStats[hour]; ok {
		hs.Clicks++
	}
}

// RecordError hata kaydeder
func (r *SERPReporter) RecordError(engine string) {
	if !r.enabled {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.data.TotalErrors++

	if es, ok := r.data.EngineStats[engine]; ok {
		es.Errors++
		if es.Searches > 0 {
			es.SuccessRate = float64(es.Searches-es.Errors) / float64(es.Searches) * 100
		}
	}

	hour := time.Now().Hour()
	if hs, ok := r.data.HourlyStats[hour]; ok {
		hs.Errors++
	}
}

// GetReport mevcut rapor verisini döner
func (r *SERPReporter) GetReport() *ReportData {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Kopya oluştur
	data := *r.data
	data.EndTime = time.Now()
	return &data
}

// SaveReport raporu dosyaya kaydeder
func (r *SERPReporter) SaveReport(domain string) error {
	if !r.enabled || r.reportDir == "" {
		return nil
	}

	r.mu.Lock()
	r.data.EndTime = time.Now()
	r.data.Domain = domain
	data, err := json.MarshalIndent(r.data, "", "  ")
	r.mu.Unlock()

	if err != nil {
		return err
	}

	filename := fmt.Sprintf("serp_report_%s_%s.json",
		domain,
		time.Now().Format("2006-01-02_15-04-05"),
	)

	return os.WriteFile(filepath.Join(r.reportDir, filename), data, 0644)
}

// SetDomain domain'i ayarlar
func (r *SERPReporter) SetDomain(domain string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.Domain = domain
}

// GetSummary özet rapor döner
func (r *SERPReporter) GetSummary() map[string]interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	var overallCTR float64
	if r.data.TotalSearches > 0 {
		overallCTR = float64(r.data.TotalClicks) / float64(r.data.TotalSearches) * 100
	}

	return map[string]interface{}{
		"total_searches": r.data.TotalSearches,
		"total_clicks":   r.data.TotalClicks,
		"total_errors":   r.data.TotalErrors,
		"overall_ctr":    overallCTR,
		"duration":       time.Since(r.startTime).String(),
		"engines_count":  len(r.data.EngineStats),
		"keywords_count": len(r.data.KeywordStats),
	}
}

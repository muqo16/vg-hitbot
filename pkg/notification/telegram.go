// Package notification provides notification services for the bot
package notification

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TelegramNotifier Telegram bildirim servisi
type TelegramNotifier struct {
	mu             sync.Mutex
	botToken       string
	chatID         string
	enabled        bool
	reportInterval time.Duration
	httpClient     *http.Client
	lastReport     time.Time
	stopCh         chan struct{}
	running        bool
}

// TelegramConfig Telegram yapılandırması
type TelegramConfig struct {
	BotToken       string
	ChatID         string
	Enabled        bool
	ReportInterval int // dakika cinsinden
}

// SimulationStats simülasyon istatistikleri
type SimulationStats struct {
	TotalHits      int64
	SuccessfulHits int64
	FailedHits     int64
	SuccessRate    float64
	Duration       time.Duration
	HitsPerMinute  float64
	ActiveProxies  int
	Domain         string
}

// NewTelegramNotifier yeni Telegram notifier oluşturur
func NewTelegramNotifier(cfg TelegramConfig) *TelegramNotifier {
	interval := time.Duration(cfg.ReportInterval) * time.Minute
	if interval <= 0 {
		interval = 10 * time.Minute
	}

	return &TelegramNotifier{
		botToken:       cfg.BotToken,
		chatID:         cfg.ChatID,
		enabled:        cfg.Enabled,
		reportInterval: interval,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

// IsEnabled bildirim aktif mi
func (t *TelegramNotifier) IsEnabled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enabled && t.botToken != "" && t.chatID != ""
}

// UpdateConfig yapılandırmayı günceller
func (t *TelegramNotifier) UpdateConfig(cfg TelegramConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.botToken = cfg.BotToken
	t.chatID = cfg.ChatID
	t.enabled = cfg.Enabled
	if cfg.ReportInterval > 0 {
		t.reportInterval = time.Duration(cfg.ReportInterval) * time.Minute
	}
}

// TestConnection bot token ve chat ID doğrulama
func (t *TelegramNotifier) TestConnection() error {
	if t.botToken == "" {
		return fmt.Errorf("bot token boş")
	}
	if t.chatID == "" {
		return fmt.Errorf("chat ID boş")
	}

	// Bot bilgilerini al
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", t.botToken)
	resp, err := t.httpClient.Get(apiURL)
	if err != nil {
		return fmt.Errorf("Telegram API'ya bağlanılamadı: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OK          bool `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("yanıt parse hatası: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("bot token geçersiz: %s", result.Description)
	}

	// Test mesajı gönder
	return t.SendMessage("✅ *VGBot* bağlantı testi başarılı\\!")
}

// SendMessage Telegram mesajı gönderir (MarkdownV2 formatında)
func (t *TelegramNotifier) SendMessage(text string) error {
	t.mu.Lock()
	token := t.botToken
	chatID := t.chatID
	enabled := t.enabled
	t.mu.Unlock()

	if !enabled || token == "" || chatID == "" {
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", text)
	params.Set("parse_mode", "MarkdownV2")
	params.Set("disable_web_page_preview", "true")

	resp, err := t.httpClient.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("mesaj gönderilemedi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API hatası (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// sendRawMessage parse mode olmadan mesaj gönderir
func (t *TelegramNotifier) sendRawMessage(text string) error {
	t.mu.Lock()
	token := t.botToken
	chatID := t.chatID
	enabled := t.enabled
	t.mu.Unlock()

	if !enabled || token == "" || chatID == "" {
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", text)
	params.Set("disable_web_page_preview", "true")

	resp, err := t.httpClient.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("mesaj gönderilemedi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API hatası (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// SendSimulationStart simülasyon başlangıç bildirimi
func (t *TelegramNotifier) SendSimulationStart(domain string, durationMin int, hpm int, concurrent int) error {
	msg := fmt.Sprintf(
		"🚀 Simülasyon Başladı\n\n"+
			"🌐 Domain: %s\n"+
			"⏱ Süre: %d dakika\n"+
			"📊 HPM: %d\n"+
			"🔄 Eşzamanlı: %d\n"+
			"🕐 Başlangıç: %s",
		domain,
		durationMin,
		hpm,
		concurrent,
		time.Now().Format("15:04:05"),
	)
	return t.sendRawMessage(msg)
}

// SendSimulationEnd simülasyon bitiş bildirimi
func (t *TelegramNotifier) SendSimulationEnd(stats SimulationStats) error {
	msg := fmt.Sprintf(
		"✅ Simülasyon Tamamlandı\n\n"+
			"🌐 Domain: %s\n"+
			"📊 Toplam Hit: %d\n"+
			"✓ Başarılı: %d\n"+
			"✗ Başarısız: %d\n"+
			"📈 Başarı Oranı: %.1f%%\n"+
			"⏱ Süre: %s\n"+
			"📊 Ortalama HPM: %.1f\n"+
			"🕐 Bitiş: %s",
		stats.Domain,
		stats.TotalHits,
		stats.SuccessfulHits,
		stats.FailedHits,
		stats.SuccessRate,
		formatDuration(stats.Duration),
		stats.HitsPerMinute,
		time.Now().Format("15:04:05"),
	)
	return t.sendRawMessage(msg)
}

// SendError hata bildirimi
func (t *TelegramNotifier) SendError(errMsg string) error {
	msg := fmt.Sprintf(
		"⚠️ Hata Bildirimi\n\n"+
			"🔴 Hata: %s\n"+
			"🕐 Zaman: %s",
		errMsg,
		time.Now().Format("15:04:05"),
	)
	return t.sendRawMessage(msg)
}

// SendPeriodicReport periyodik durum raporu
func (t *TelegramNotifier) SendPeriodicReport(stats SimulationStats) error {
	t.mu.Lock()
	t.lastReport = time.Now()
	t.mu.Unlock()

	msg := fmt.Sprintf(
		"📊 Durum Raporu\n\n"+
			"🌐 Domain: %s\n"+
			"📊 Toplam Hit: %d\n"+
			"✓ Başarılı: %d\n"+
			"✗ Başarısız: %d\n"+
			"📈 Başarı Oranı: %.1f%%\n"+
			"⏱ Geçen Süre: %s\n"+
			"📊 HPM: %.1f\n"+
			"🔗 Aktif Proxy: %d\n"+
			"🕐 Rapor Zamanı: %s",
		stats.Domain,
		stats.TotalHits,
		stats.SuccessfulHits,
		stats.FailedHits,
		stats.SuccessRate,
		formatDuration(stats.Duration),
		stats.HitsPerMinute,
		stats.ActiveProxies,
		time.Now().Format("15:04:05"),
	)
	return t.sendRawMessage(msg)
}

// ShouldSendReport periyodik rapor zamanı geldi mi
func (t *TelegramNotifier) ShouldSendReport() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.enabled || t.botToken == "" || t.chatID == "" {
		return false
	}
	return time.Since(t.lastReport) >= t.reportInterval
}

// StartPeriodicReporting periyodik rapor gönderimini başlatır
func (t *TelegramNotifier) StartPeriodicReporting(statsFn func() SimulationStats) {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return
	}
	t.running = true
	t.stopCh = make(chan struct{})
	interval := t.reportInterval
	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if t.IsEnabled() {
					stats := statsFn()
					_ = t.SendPeriodicReport(stats)
				}
			case <-t.stopCh:
				return
			}
		}
	}()
}

// StopPeriodicReporting periyodik rapor gönderimini durdurur
func (t *TelegramNotifier) StopPeriodicReporting() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running {
		close(t.stopCh)
		t.running = false
	}
}

// formatDuration süreyi okunabilir formata çevirir
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dsa %ddk %dsn", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%ddk %dsn", m, s)
	}
	return fmt.Sprintf("%dsn", s)
}

// Package referrer provides advanced social media referrer generation
package referrer

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

// SocialPlatform sosyal medya platform tanımı
type SocialPlatform struct {
	Name       string
	Weight     int
	DesktopURLs []string
	MobileURLs  []string
	HasUTM     bool
}

// SocialReferrerEngine gelişmiş sosyal medya referrer motoru
type SocialReferrerEngine struct {
	platforms     map[string]*SocialPlatform
	enableUTM     bool
	utmCampaigns  []string
	totalWeight   int
	weightedOrder []string
}

// SocialConfig sosyal medya referrer yapılandırması
type SocialConfig struct {
	EnableUTM    bool
	UTMCampaigns []string
	Weights      map[string]int // platform adı -> ağırlık
	Platforms    []string       // aktif platform listesi
}

// UTMParams UTM parametreleri
type UTMParams struct {
	Source   string
	Medium   string
	Campaign string
	Content  string
	Term     string
}

// defaultPlatforms tüm desteklenen platformlar
var defaultPlatforms = map[string]*SocialPlatform{
	"facebook": {
		Name:   "facebook",
		Weight: 25,
		DesktopURLs: []string{
			"https://www.facebook.com/",
			"https://www.facebook.com/groups/",
			"https://www.facebook.com/pages/",
			"https://www.facebook.com/share/",
			"https://l.facebook.com/l.php?u=",
			"https://lm.facebook.com/l.php?u=",
		},
		MobileURLs: []string{
			"https://m.facebook.com/",
			"https://m.facebook.com/groups/",
			"https://m.facebook.com/story.php",
			"https://touch.facebook.com/",
		},
		HasUTM: true,
	},
	"twitter": {
		Name:   "twitter",
		Weight: 15,
		DesktopURLs: []string{
			"https://twitter.com/",
			"https://x.com/",
			"https://t.co/",
			"https://twitter.com/search?q=",
			"https://twitter.com/i/web/status/",
		},
		MobileURLs: []string{
			"https://mobile.twitter.com/",
			"https://mobile.x.com/",
		},
		HasUTM: true,
	},
	"instagram": {
		Name:   "instagram",
		Weight: 15,
		DesktopURLs: []string{
			"https://www.instagram.com/",
			"https://www.instagram.com/p/",
			"https://www.instagram.com/explore/",
			"https://l.instagram.com/?u=",
		},
		MobileURLs: []string{
			"https://www.instagram.com/",
			"https://instagram.com/",
		},
		HasUTM: false,
	},
	"linkedin": {
		Name:   "linkedin",
		Weight: 10,
		DesktopURLs: []string{
			"https://www.linkedin.com/feed/",
			"https://www.linkedin.com/",
			"https://www.linkedin.com/posts/",
			"https://www.linkedin.com/pulse/",
			"https://www.linkedin.com/sharing/share-offsite/",
		},
		MobileURLs: []string{
			"https://www.linkedin.com/feed/",
			"https://m.linkedin.com/",
		},
		HasUTM: true,
	},
	"reddit": {
		Name:   "reddit",
		Weight: 10,
		DesktopURLs: []string{
			"https://www.reddit.com/",
			"https://www.reddit.com/r/webdev/",
			"https://www.reddit.com/r/SEO/",
			"https://www.reddit.com/r/marketing/",
			"https://www.reddit.com/r/technology/",
			"https://old.reddit.com/",
		},
		MobileURLs: []string{
			"https://www.reddit.com/",
			"https://m.reddit.com/",
			"https://i.reddit.com/",
		},
		HasUTM: false,
	},
	"tiktok": {
		Name:   "tiktok",
		Weight: 10,
		DesktopURLs: []string{
			"https://www.tiktok.com/",
			"https://www.tiktok.com/@",
			"https://www.tiktok.com/foryou",
			"https://www.tiktok.com/discover/",
		},
		MobileURLs: []string{
			"https://www.tiktok.com/",
			"https://m.tiktok.com/",
			"https://vm.tiktok.com/",
		},
		HasUTM: false,
	},
	"youtube": {
		Name:   "youtube",
		Weight: 10,
		DesktopURLs: []string{
			"https://www.youtube.com/",
			"https://www.youtube.com/watch?v=",
			"https://www.youtube.com/redirect?q=",
			"https://www.youtube.com/results?search_query=",
		},
		MobileURLs: []string{
			"https://m.youtube.com/",
			"https://youtu.be/",
		},
		HasUTM: true,
	},
	"pinterest": {
		Name:   "pinterest",
		Weight: 5,
		DesktopURLs: []string{
			"https://www.pinterest.com/",
			"https://www.pinterest.com/pin/",
			"https://tr.pinterest.com/",
			"https://www.pinterest.com/search/pins/?q=",
		},
		MobileURLs: []string{
			"https://www.pinterest.com/",
			"https://pin.it/",
		},
		HasUTM: true,
	},
	"whatsapp": {
		Name:   "whatsapp",
		Weight: 5,
		DesktopURLs: []string{
			"https://web.whatsapp.com/",
			"https://api.whatsapp.com/send?text=",
		},
		MobileURLs: []string{
			"https://wa.me/",
			"https://api.whatsapp.com/",
		},
		HasUTM: false,
	},
	"telegram": {
		Name:   "telegram",
		Weight: 5,
		DesktopURLs: []string{
			"https://web.telegram.org/",
			"https://t.me/",
			"https://telegram.me/",
		},
		MobileURLs: []string{
			"https://t.me/",
			"https://telegram.me/",
		},
		HasUTM: false,
	},
}

// defaultUTMCampaigns varsayılan UTM kampanya isimleri
var defaultUTMCampaigns = []string{
	"social_share",
	"organic_social",
	"social_post",
	"social_link",
	"community_share",
	"viral_share",
	"content_promo",
	"brand_awareness",
	"engagement",
	"referral",
}

// defaultUTMContents varsayılan UTM content etiketleri
var defaultUTMContents = []string{
	"post", "story", "reel", "video", "link", "bio",
	"comment", "share", "feed", "profile", "group",
	"pin", "tweet", "thread", "article", "blog",
}

// NewSocialReferrerEngine yeni sosyal medya referrer motoru oluşturur
func NewSocialReferrerEngine(cfg *SocialConfig) *SocialReferrerEngine {
	engine := &SocialReferrerEngine{
		platforms:    make(map[string]*SocialPlatform),
		enableUTM:   cfg != nil && cfg.EnableUTM,
		utmCampaigns: defaultUTMCampaigns,
	}

	if cfg != nil && len(cfg.UTMCampaigns) > 0 {
		engine.utmCampaigns = cfg.UTMCampaigns
	}

	// Aktif platformları belirle
	activePlatforms := make(map[string]bool)
	if cfg != nil && len(cfg.Platforms) > 0 {
		for _, p := range cfg.Platforms {
			activePlatforms[strings.ToLower(p)] = true
		}
	}

	// Platformları yükle
	for name, platform := range defaultPlatforms {
		if len(activePlatforms) > 0 && !activePlatforms[name] {
			continue
		}

		p := *platform // kopya
		if cfg != nil && cfg.Weights != nil {
			if w, ok := cfg.Weights[name]; ok {
				p.Weight = w
			}
		}

		engine.platforms[name] = &p
		engine.totalWeight += p.Weight
		engine.weightedOrder = append(engine.weightedOrder, name)
	}

	return engine
}

// GenerateReferrer belirtilen platform veya rastgele platform için referrer üretir
func (e *SocialReferrerEngine) GenerateReferrer(platform string) *ReferrerSource {
	if platform == "" || platform == "random" {
		platform = e.selectWeightedPlatform()
	}

	p, ok := e.platforms[platform]
	if !ok {
		// fallback
		p = defaultPlatforms["facebook"]
		platform = "facebook"
	}

	// Mobile vs Desktop (30% mobile)
	isMobile := rand.Float64() < 0.3
	var urls []string
	if isMobile && len(p.MobileURLs) > 0 {
		urls = p.MobileURLs
	} else {
		urls = p.DesktopURLs
	}

	if len(urls) == 0 {
		return &ReferrerSource{
			Type:     "social",
			URL:      "https://www.facebook.com/",
			Platform: platform,
		}
	}

	referrerURL := urls[rand.Intn(len(urls))]

	// Rastgele path ekleme (bazı platformlar için)
	referrerURL = e.appendRandomPath(referrerURL, platform)

	// UTM parametreleri ekle
	if e.enableUTM && p.HasUTM {
		utm := e.GenerateUTMParams(platform)
		referrerURL = e.appendUTMToURL(referrerURL, utm)
	}

	return &ReferrerSource{
		Type:     "social",
		URL:      referrerURL,
		Platform: platform,
	}
}

// GenerateUTMParams rastgele UTM parametreleri üretir
func (e *SocialReferrerEngine) GenerateUTMParams(platform string) *UTMParams {
	campaign := e.utmCampaigns[rand.Intn(len(e.utmCampaigns))]
	content := defaultUTMContents[rand.Intn(len(defaultUTMContents))]

	return &UTMParams{
		Source:   platform,
		Medium:   "social",
		Campaign: campaign,
		Content:  content,
	}
}

// selectWeightedPlatform ağırlıklı rastgele platform seçer
func (e *SocialReferrerEngine) selectWeightedPlatform() string {
	if e.totalWeight <= 0 || len(e.weightedOrder) == 0 {
		return "facebook"
	}

	roll := rand.Intn(e.totalWeight)
	cumulative := 0
	for _, name := range e.weightedOrder {
		p := e.platforms[name]
		cumulative += p.Weight
		if roll < cumulative {
			return name
		}
	}

	return e.weightedOrder[0]
}

// appendRandomPath platforma göre rastgele path ekler
func (e *SocialReferrerEngine) appendRandomPath(baseURL string, platform string) string {
	// Rastgele hash/id oluştur
	randomID := fmt.Sprintf("%d", rand.Int63n(9999999999))
	randomHash := fmt.Sprintf("%x", rand.Int63())[:8]

	switch platform {
	case "facebook":
		if strings.Contains(baseURL, "/groups/") {
			return baseURL + randomID + "/posts/" + randomID
		}
		if strings.Contains(baseURL, "l.php") {
			return baseURL // redirect URL'si olduğu gibi
		}
		if rand.Float64() < 0.3 {
			return baseURL + randomID
		}
	case "twitter":
		if strings.Contains(baseURL, "/status/") {
			return baseURL + randomID
		}
		if strings.Contains(baseURL, "t.co/") {
			return baseURL + randomHash
		}
	case "instagram":
		if strings.Contains(baseURL, "/p/") {
			return baseURL + randomHash + "/"
		}
	case "youtube":
		if strings.Contains(baseURL, "watch?v=") {
			// Rastgele video ID (11 karakter)
			chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
			videoID := make([]byte, 11)
			for i := range videoID {
				videoID[i] = chars[rand.Intn(len(chars))]
			}
			return baseURL + string(videoID)
		}
	case "tiktok":
		if strings.Contains(baseURL, "vm.tiktok.com/") {
			return baseURL + randomHash + "/"
		}
	case "pinterest":
		if strings.Contains(baseURL, "/pin/") {
			return baseURL + randomID + "/"
		}
	case "reddit":
		subreddits := []string{
			"webdev", "SEO", "marketing", "technology", "programming",
			"startups", "digital_marketing", "web_design", "smallbusiness",
		}
		if strings.HasSuffix(baseURL, "/r/") || strings.Contains(baseURL, "/r/webdev/") ||
			strings.Contains(baseURL, "/r/SEO/") || strings.Contains(baseURL, "/r/marketing/") ||
			strings.Contains(baseURL, "/r/technology/") {
			sub := subreddits[rand.Intn(len(subreddits))]
			return "https://www.reddit.com/r/" + sub + "/comments/" + randomHash + "/"
		}
	}

	return baseURL
}

// appendUTMToURL URL'ye UTM parametreleri ekler
func (e *SocialReferrerEngine) appendUTMToURL(rawURL string, utm *UTMParams) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := u.Query()
	q.Set("utm_source", utm.Source)
	q.Set("utm_medium", utm.Medium)
	q.Set("utm_campaign", utm.Campaign)

	if utm.Content != "" && rand.Float64() < 0.6 {
		q.Set("utm_content", utm.Content)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// GetAvailablePlatforms mevcut platform listesini döner
func (e *SocialReferrerEngine) GetAvailablePlatforms() []string {
	platforms := make([]string, 0, len(e.platforms))
	for name := range e.platforms {
		platforms = append(platforms, name)
	}
	return platforms
}

// GetDefaultPlatformNames tüm varsayılan platform isimlerini döner
func GetDefaultPlatformNames() []string {
	return []string{
		"facebook", "twitter", "instagram", "linkedin", "reddit",
		"tiktok", "youtube", "pinterest", "whatsapp", "telegram",
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

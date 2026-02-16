// Package serp provides search engine profiles for multi-engine SERP simulation
package serp

// SearchEngineProfile arama motoru profili
type SearchEngineProfile struct {
	Name             string            // Motor adı
	BaseURL          string            // Temel URL
	QueryParam       string            // Arama sorgu parametresi
	ResultSelector   string            // CSS seçici: organik sonuçlar
	NextPageSelector string            // CSS seçici: sonraki sayfa
	LinkSelector     string            // CSS seçici: sonuç linki
	TitleSelector    string            // CSS seçici: sonuç başlığı
	SnippetSelector  string            // CSS seçici: sonuç açıklaması
	CountryDomains   map[string]string // Ülke kodu -> domain
	Headers          map[string]string // Ek HTTP headers
	AcceptLanguage   string            // Accept-Language header
}

// EngineRegistry arama motoru kayıt defteri
var EngineRegistry = map[string]*SearchEngineProfile{
	"google": {
		Name:             "Google",
		BaseURL:          "https://www.google.com/search",
		QueryParam:       "q",
		ResultSelector:   "div.g",
		NextPageSelector: "a#pnnext",
		LinkSelector:     "div.g a[href]",
		TitleSelector:    "div.g h3",
		SnippetSelector:  "div.g div.VwiC3b",
		CountryDomains: map[string]string{
			"tr": "www.google.com.tr",
			"de": "www.google.de",
			"fr": "www.google.fr",
			"uk": "www.google.co.uk",
			"jp": "www.google.co.jp",
			"br": "www.google.com.br",
			"in": "www.google.co.in",
			"ru": "www.google.ru",
			"it": "www.google.it",
			"es": "www.google.es",
			"nl": "www.google.nl",
			"pl": "www.google.pl",
			"au": "www.google.com.au",
			"ca": "www.google.ca",
			"mx": "www.google.com.mx",
		},
		Headers:        map[string]string{},
		AcceptLanguage: "en-US,en;q=0.9",
	},
	"bing": {
		Name:             "Bing",
		BaseURL:          "https://www.bing.com/search",
		QueryParam:       "q",
		ResultSelector:   "li.b_algo",
		NextPageSelector: "a.sb_pagN",
		LinkSelector:     "li.b_algo h2 a",
		TitleSelector:    "li.b_algo h2",
		SnippetSelector:  "li.b_algo .b_caption p",
		CountryDomains: map[string]string{
			"de": "www.bing.com",
			"fr": "www.bing.com",
			"uk": "www.bing.com",
			"jp": "www.bing.com",
		},
		Headers:        map[string]string{},
		AcceptLanguage: "en-US,en;q=0.9",
	},
	"duckduckgo": {
		Name:             "DuckDuckGo",
		BaseURL:          "https://duckduckgo.com/",
		QueryParam:       "q",
		ResultSelector:   "article[data-testid='result']",
		NextPageSelector: "button#more-results",
		LinkSelector:     "article a[data-testid='result-title-a']",
		TitleSelector:    "article a[data-testid='result-title-a'] span",
		SnippetSelector:  "article div[data-result='snippet']",
		CountryDomains:   map[string]string{},
		Headers:          map[string]string{},
		AcceptLanguage:   "en-US,en;q=0.9",
	},
	"yandex": {
		Name:             "Yandex",
		BaseURL:          "https://yandex.com/search/",
		QueryParam:       "text",
		ResultSelector:   "li.serp-item",
		NextPageSelector: "a.pager__item_kind_next",
		LinkSelector:     "li.serp-item a.organic__url",
		TitleSelector:    "li.serp-item h2.organic__title-wrapper",
		SnippetSelector:  "li.serp-item div.organic__text",
		CountryDomains: map[string]string{
			"tr": "yandex.com.tr",
			"ru": "yandex.ru",
		},
		Headers:        map[string]string{},
		AcceptLanguage: "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
	},
	"baidu": {
		Name:             "Baidu",
		BaseURL:          "https://www.baidu.com/s",
		QueryParam:       "wd",
		ResultSelector:   "div.result",
		NextPageSelector: "a.n",
		LinkSelector:     "div.result h3 a",
		TitleSelector:    "div.result h3",
		SnippetSelector:  "div.result div.c-abstract",
		CountryDomains:   map[string]string{},
		Headers:          map[string]string{},
		AcceptLanguage:   "zh-CN,zh;q=0.9,en;q=0.8",
	},
	"yahoo": {
		Name:             "Yahoo",
		BaseURL:          "https://search.yahoo.com/search",
		QueryParam:       "p",
		ResultSelector:   "div.algo-sr",
		NextPageSelector: "a.next",
		LinkSelector:     "div.algo-sr h3 a",
		TitleSelector:    "div.algo-sr h3",
		SnippetSelector:  "div.algo-sr div.compText",
		CountryDomains: map[string]string{
			"jp": "search.yahoo.co.jp",
		},
		Headers:        map[string]string{},
		AcceptLanguage: "en-US,en;q=0.9",
	},
}

// GetEngine arama motoru profili döner
func GetEngine(name string) *SearchEngineProfile {
	if engine, ok := EngineRegistry[name]; ok {
		return engine
	}
	return EngineRegistry["google"]
}

// GetSearchURL arama URL'si oluşturur (ülke domain desteği ile)
func (e *SearchEngineProfile) GetSearchURL(query string, countryCode string) string {
	baseURL := e.BaseURL

	// Ülke-spesifik domain kullan
	if countryCode != "" {
		if domain, ok := e.CountryDomains[countryCode]; ok {
			switch e.Name {
			case "Google":
				baseURL = "https://" + domain + "/search"
			case "Yandex":
				baseURL = "https://" + domain + "/search/"
			}
		}
	}

	return baseURL + "?" + e.QueryParam + "=" + query
}

// GetSupportedEngines desteklenen arama motorları listesi
func GetSupportedEngines() []string {
	engines := make([]string, 0, len(EngineRegistry))
	for name := range EngineRegistry {
		engines = append(engines, name)
	}
	return engines
}

// GetCountryCodes belirli motor için ülke kodları
func GetCountryCodes(engineName string) []string {
	engine := GetEngine(engineName)
	codes := make([]string, 0, len(engine.CountryDomains))
	for code := range engine.CountryDomains {
		codes = append(codes, code)
	}
	return codes
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gobackend/models"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type CacheEntry struct {
	Data      *models.CollegeStats
	ExpiresAt time.Time
}

type CollegeCache struct {
	mu    sync.RWMutex
	cache map[string]CacheEntry
	ttl   time.Duration
}

func NewCollegeCache(ttl time.Duration) *CollegeCache {
	cache := &CollegeCache{
		cache: make(map[string]CacheEntry),
		ttl:   ttl,
	}

	// Background cleanup goroutine
	go cache.cleanup()
	return cache
}

func (c *CollegeCache) Get(key string) (*models.CollegeStats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Data, true
}

func (c *CollegeCache) Set(key string, data *models.CollegeStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *CollegeCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.ExpiresAt) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

var collegeCache *CollegeCache

func InitializeCache() {
	collegeCache = NewCollegeCache(1 * time.Hour)
}

func GetFromCache(collegeName string) (*models.CollegeStats, bool) {
	if collegeCache == nil {
		return nil, false
	}
	cacheKey := strings.ToLower(strings.TrimSpace(collegeName))
	return collegeCache.Get(cacheKey)
}

func SaveToCache(collegeName string, data *models.CollegeStats) {
	if collegeCache == nil {
		return
	}
	cacheKey := strings.ToLower(strings.TrimSpace(collegeName))
	collegeCache.Set(cacheKey, data)
}

func getPrompt(university string) string {
	return fmt.Sprintf(`
You are a university data researcher. Provide comprehensive structured details for university: %s

Return ONLY valid JSON format (no markdown, no code blocks, no extra text):
{
  "college_name": "%s",
  "country": "Country name",
  "about": "Detailed description including history, establishment year, and location of the college",
  "location": "City, State/Country",
  "summary": "Brief 2-3 sentence summary about the college's reputation and strengths",
  "ug_programs": ["B.Tech Computer Science", "B.Tech Mechanical Engineering", "B.A Economics"],
  "pg_programs": ["M.Tech Computer Science", "MBA", "M.Sc Physics"],
  "phd_programs": ["PhD Computer Science", "PhD Physics", "PhD Economics"],
  "fees": {
    "ug_yearly_min": 50000,
    "ug_yearly_max": 150000,
    "pg_yearly_min": 100000,
    "pg_yearly_max": 300000,
    "phd_yearly_min": 0,
    "phd_yearly_max": 50000
  },
  "scholarships": ["Merit-based scholarship", "Need-based scholarship", "Government scholarship"],
  "student_gender_ratio": {
    "male_percentage": 60,
    "female_percentage": 40
  },
  "faculty_staff": 500,
  "international_students": 100,
  "global_ranking": "Top 100 or specific rank",
  "departments": ["Computer Science", "Mechanical Engineering", "Civil Engineering"],
  "student_statistics": [
    {"category": "Total students (2025)", "value": 10000},
    {"category": "Undergraduate (UG) students (2025)", "value": 7000},
    {"category": "Postgraduate (PG) students (2025)", "value": 2500},
    {"category": "Male students (2025)", "value": 6000},
    {"category": "Female students (2025)", "value": 4000},
    {"category": "International students (2025)", "value": 100},
    {"category": "Total students placed (2025)", "value": 1500},
    {"category": "UG 4-year students placed (2025)", "value": 1000},
    {"category": "UG 5-year students placed (2025)", "value": 200},
    {"category": "PG 2-year students placed (2025)", "value": 300},
    {"category": "Placement rate (UG 4-year, 2025)", "value": 80}
  ],
  "additional_details": [
    {"category": "NIRF Ranking (Engineering)", "value": "50"},
    {"category": "Times Higher Education World University Rankings", "value": "501-600"},
    {"category": "Student–faculty ratio", "value": 15},
    {"category": "Median CTC (2025)", "value": "INR 10 LPA"},
    {"category": "Median CTC (UG 4-year, 2025)", "value": "₹8 LPA"},
    {"category": "Median CTC (UG 5-year, 2025)", "value": "₹9 LPA"},
    {"category": "Median CTC (PG 2-year, 2025)", "value": "₹12 LPA"}
  ],
  "sources": ["https://university-website.edu", "https://official-source.com"]
}

Provide realistic data based on actual records.
`, university, university)
}

// FetchCollegeDataFromGemini fetches data directly from Gemini API
func FetchCollegeDataFromGemini(collegeName string) (*models.CollegeStats, error) {
	startTime := time.Now()
	log.Printf("� Fetching data for: %s", collegeName)

	// Check cache first
	if cachedData, found := GetFromCache(collegeName); found {
		log.Printf("📦 Cache HIT for: %s", collegeName)
		return cachedData, nil
	}

	// Initialize Gemini client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Printf("❌ GEMINI_API_KEY not set in environment")
		return nil, fmt.Errorf("GEMINI_API_KEY not set in .env file")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("❌ Failed to create Gemini client: %v", err)
		if strings.Contains(err.Error(), "403") {
			log.Printf("🔴 API Key Error: Your Gemini API key may be compromised or invalid")
		}
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	prompt := getPrompt(collegeName)

	// Call Gemini API
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("❌ Gemini API error: %v", err)

		// Better error messages for common issues
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "leaked") {
			log.Printf("🔴 CRITICAL: Your API key has been reported as leaked or is invalid!")
			log.Printf("📌 Action required: Get a new API key from https://aistudio.google.com")
			return nil, fmt.Errorf("API key compromised. Get a new one from https://aistudio.google.com")
		}
		if strings.Contains(err.Error(), "429") {
			return nil, fmt.Errorf("API rate limit exceeded. Please try again later")
		}
		if strings.Contains(err.Error(), "401") {
			return nil, fmt.Errorf("API authentication failed. Check your GEMINI_API_KEY")
		}

		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Printf(" Empty response from Gemini")
		return nil, fmt.Errorf("empty response from Gemini")
	}

	text := fmt.Sprint(resp.Candidates[0].Content.Parts[0])

	// Remove markdown code blocks if present
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		parts := strings.Split(text, "```")
		if len(parts) >= 2 {
			text = parts[1]
			text = strings.TrimPrefix(text, "json")
		}
	}
	text = strings.TrimSpace(text)

	// Parse JSON response
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		log.Printf(" JSON Parse Error: %v", err)
		log.Printf("Response text: %s", text)
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Convert to CollegeStats model
	stats := mapGeminiResponseToCollegeStats(data)

	elapsedTime := time.Since(startTime)
	log.Printf("✅ Successfully fetched data for: %s (⏱️ %dms)", stats.CollegeName, elapsedTime.Milliseconds())

	// Save to cache
	SaveToCache(collegeName, stats)

	return stats, nil
}

// mapGeminiResponseToCollegeStats converts Gemini JSON response to CollegeStats model
func mapGeminiResponseToCollegeStats(data map[string]interface{}) *models.CollegeStats {
	stats := &models.CollegeStats{
		CollegeName:           getStringValue(data, "college_name"),
		Country:               getStringValue(data, "country"),
		About:                 getStringValue(data, "about"),
		Location:              getStringValue(data, "location"),
		Summary:               getStringValue(data, "summary"),
		GlobalRanking:         getStringValue(data, "global_ranking"),
		FacultyStaff:          getIntValue(data, "faculty_staff"),
		InternationalStudents: getIntValue(data, "international_students"),
	}

	// Parse array fields
	stats.UGPrograms = parseStringArray(getSliceValue(data, "ug_programs"))
	stats.PGPrograms = parseStringArray(getSliceValue(data, "pg_programs"))
	stats.PhDPrograms = parseStringArray(getSliceValue(data, "phd_programs"))
	stats.Departments = parseStringArray(getSliceValue(data, "departments"))
	stats.Scholarships = parseStringArray(getSliceValue(data, "scholarships"))

	// Parse gender ratio
	stats.StudentGenderRatio = parseGenderRatio(data["student_gender_ratio"])

	// Parse fees
	stats.Fees = parseFees(data["fees"])

	// Parse statistics
	stats.StudentStatistics = parseStatistics(data["student_statistics"])
	stats.AdditionalDetails = parseStatistics(data["additional_details"])

	// Parse sources
	stats.Sources = parseStringArray(getSliceValue(data, "sources"))

	return stats
}

func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntValue(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			var result int
			fmt.Sscanf(v, "%d", &result)
			return result
		}
	}
	return 0
}

func getSliceValue(data map[string]interface{}, key string) []interface{} {
	if val, ok := data[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			return slice
		}
	}
	return []interface{}{}
}

func parseStringArray(slice []interface{}) []string {
	var result []string
	for _, item := range slice {
		if str, ok := item.(string); ok {
			trimmed := strings.TrimSpace(str)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}

func parseGenderRatio(data interface{}) models.GenderRatio {
	var ratio models.GenderRatio
	if mapData, ok := data.(map[string]interface{}); ok {
		ratio.MalePercentage = getIntValue(mapData, "male_percentage")
		ratio.FemalePercentage = getIntValue(mapData, "female_percentage")
	}
	return ratio
}

func parseFees(data interface{}) models.FeesInfo {
	var fees models.FeesInfo
	if mapData, ok := data.(map[string]interface{}); ok {
		fees.UGYearlyMin = getIntValue(mapData, "ug_yearly_min")
		fees.UGYearlyMax = getIntValue(mapData, "ug_yearly_max")
		fees.PGYearlyMin = getIntValue(mapData, "pg_yearly_min")
		fees.PGYearlyMax = getIntValue(mapData, "pg_yearly_max")
		fees.PhDYearlyMin = getIntValue(mapData, "phd_yearly_min")
		fees.PhDYearlyMax = getIntValue(mapData, "phd_yearly_max")
	}
	return fees
}

func parseStatistics(data interface{}) []models.StatisticItem {
	var stats []models.StatisticItem
	if slice, ok := data.([]interface{}); ok {
		for _, item := range slice {
			if mapItem, ok := item.(map[string]interface{}); ok {
				stat := models.StatisticItem{
					Category: getStringValue(mapItem, "category"),
					Value:    mapItem["value"],
				}
				stats = append(stats, stat)
			}
		}
	}
	return stats
}

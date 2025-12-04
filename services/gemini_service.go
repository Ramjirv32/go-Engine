package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gobackend/models"
)

func FetchCollegeDataFromGemini(collegeName string) (*models.CollegeStats, error) {
	log.Printf("� Calling Django API for: %s", collegeName)

	// Django API endpoint
	djangoURL := "http://127.0.0.1:8000/test-fetch-gemini/"

	// Prepare form data
	formData := url.Values{}
	formData.Set("university_name", collegeName)

	// Make POST request
	response, err := http.PostForm(djangoURL, formData)
	if err != nil {
		log.Printf("❌ Django API error: %v", err)
		return nil, fmt.Errorf("failed to call Django API: %w", err)
	}
	defer response.Body.Close()

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("❌ Failed to read Django response: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("❌ Django API returned status %d: %s", response.StatusCode, string(body))
		return nil, fmt.Errorf("django API returned status %d", response.StatusCode)
	}

	// Parse Django response
	var djangoResp struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &djangoResp); err != nil {
		log.Printf("❌ Failed to parse Django response: %v", err)
		return nil, fmt.Errorf("failed to parse Django response: %w", err)
	}

	if djangoResp.Status != "success" {
		log.Printf("❌ Django API error: %s", djangoResp.Message)
		return nil, fmt.Errorf("django API error: %s", djangoResp.Message)
	}

	// Convert Django data to CollegeStats model
	dataJSON, err := json.Marshal(djangoResp.Data)
	if err != nil {
		log.Printf("❌ Failed to marshal Django data: %v", err)
		return nil, fmt.Errorf("failed to marshal Django data: %w", err)
	}

	var stats models.CollegeStats
	if err := json.Unmarshal(dataJSON, &stats); err != nil {
		log.Printf("❌ Failed to unmarshal to CollegeStats: %v", err)
		// Try mapping from Django format
		return mapDjangoToCollegeStats(djangoResp.Data)
	}

	if stats.CollegeName == "" {
		log.Printf("❌ Empty college name in response")
		return nil, fmt.Errorf("empty college name in response")
	}

	log.Printf("✅ Django API returned data for: %s", stats.CollegeName)
	return &stats, nil
}

// mapDjangoToCollegeStats converts Django API response format to CollegeStats
func mapDjangoToCollegeStats(data interface{}) (*models.CollegeStats, error) {
	djangoData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format from Django")
	}

	stats := &models.CollegeStats{
		CollegeName:           getStringValue(djangoData, "UNIVERSITY"),
		Country:               getStringValue(djangoData, "COUNTRY"),
		About:                 getStringValue(djangoData, "ABOUT"),
		Location:              getStringValue(djangoData, "LOCATION"),
		Summary:               getStringValue(djangoData, "SUMMARY"),
		GlobalRanking:         getStringValue(djangoData, "GLOBAL_RANKING"),
		FacultyStaff:          getIntValue(djangoData, "FACULTY_STAFF"),
		InternationalStudents: getIntValue(djangoData, "INTERNATIONAL_STUDENTS"),
	}

	// Parse array fields
	stats.UGPrograms = parseStringArray(getStringValue(djangoData, "UG_PROGRAMS"))
	stats.PGPrograms = parseStringArray(getStringValue(djangoData, "PG_PROGRAMS"))
	stats.PhDPrograms = parseStringArray(getStringValue(djangoData, "PHD_PROGRAMS"))
	stats.Departments = parseStringArray(getStringValue(djangoData, "DEPARTMENTS"))
	stats.Scholarships = parseStringArray(getStringValue(djangoData, "SCHOLARSHIPS"))

	// Parse gender ratio
	stats.StudentGenderRatio = parseGenderRatio(getStringValue(djangoData, "STUDENT_GENDER_RATIO"))

	// Parse fees
	stats.Fees = parseFees(getStringValue(djangoData, "FEES"))

	// Parse statistics
	stats.StudentStatistics = parseStatistics(getStringValue(djangoData, "STUDENT_STATISTICS"))
	stats.AdditionalDetails = parseStatistics(getStringValue(djangoData, "ADDITIONAL_DETAILS"))

	// Parse sources
	stats.Sources = parseStringArray(getStringValue(djangoData, "SOURCES"))

	return stats, nil
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

func parseStringArray(jsonStr string) []string {
	var result []string
	if jsonStr == "" {
		return result
	}

	// Try parsing as JSON array first
	if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
		return result
	}

	// Fallback: split by comma
	parts := strings.Split(jsonStr, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseGenderRatio(jsonStr string) models.GenderRatio {
	var ratio models.GenderRatio
	if err := json.Unmarshal([]byte(jsonStr), &ratio); err != nil {
		log.Printf("⚠️ Failed to parse gender ratio: %v", err)
	}
	return ratio
}

func parseFees(jsonStr string) models.FeesInfo {
	var fees models.FeesInfo
	if err := json.Unmarshal([]byte(jsonStr), &fees); err != nil {
		log.Printf("⚠️ Failed to parse fees: %v", err)
	}
	return fees
}

func parseStatistics(jsonStr string) []models.StatisticItem {
	var stats []models.StatisticItem
	if err := json.Unmarshal([]byte(jsonStr), &stats); err != nil {
		log.Printf("⚠️ Failed to parse statistics: %v", err)
	}
	return stats
}

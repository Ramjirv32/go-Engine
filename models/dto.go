package models

// APIResponse is a generic API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CollegeStatisticsResponse wraps college data for API responses
type CollegeStatisticsResponse struct {
	CollegeName       string          `json:"college_name"`
	Country           string          `json:"country"`
	About             string          `json:"about"`
	Location          string          `json:"location"`
	Summary           string          `json:"summary"`
	StudentStatistics []StatisticItem `json:"student_statistics"`
	AdditionalDetails []StatisticItem `json:"additional_details"`
	Fees              FeesInfo        `json:"fees"`
	GlobalRanking     string          `json:"global_ranking"`
	FacultyStaff      int             `json:"faculty_staff"`
}

// SearchRequest represents a college search request
type SearchRequest struct {
	CollegeName string `json:"college_name" form:"college_name"`
	Country     string `json:"country" form:"country"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Country string      `json:"country,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// CountryData represents country information
type CountryData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

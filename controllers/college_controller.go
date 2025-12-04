package controllers


import (
	"log"
	"net/http"

	"gobackend/services"
	"gobackend/utils"
)

func GetCollegeStatistics(w http.ResponseWriter, r *http.Request) {
	collegeName := r.URL.Query().Get("college_name")

	if collegeName == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "college_name required"})
		return
	}


	log.Printf("📊 Fetching stats for: %s", collegeName)


	cachedResult, err := services.GetCollegeFromCache(collegeName)
	if err == nil {
		go services.CompareAndUpdateCache(collegeName, *cachedResult)
		utils.RespondJSON(w, http.StatusOK, cachedResult)
		return
	}


	log.Println("🔄 Calling Gemini API directly...")
	stats, err := services.FetchCollegeDataFromGemini(collegeName)
	if err != nil {
		log.Printf("❌ Gemini API error: %v", err)
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{
			"error":  "Failed to fetch data from Gemini",
			"detail": err.Error(),
		})
		return
	}


	err = services.SaveCollegeToCache(stats)
	if err == nil {
		collegeData := map[string]interface{}{
			"id":      stats.CollegeName,
			"name":    stats.CollegeName,
			"country": stats.Country,
			"data":    stats.StudentStatistics,
		}
		services.BroadcastNewCollege(stats.Country, collegeData)
	}

	
	utils.RespondJSON(w, http.StatusOK, stats)
}

func SearchUniversity(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("university_name")
	if name == "" {
		name = r.URL.Query().Get("q")
	}

	if name == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "university_name required"})
		return
	}

	result, err := services.SearchUniversityByName(name)
	if err != nil {
		utils.RespondJSON(w, http.StatusNotFound, map[string]string{"error": "University not found"})
		return
	}

	utils.RespondJSON(w, http.StatusOK, result)
}

func GetAllColleges(w http.ResponseWriter, r *http.Request) {
	colleges, err := services.GetAllColleges()
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, colleges)
}

func GetCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := services.GetDistinctCountries()
	if err != nil {
		log.Printf("❌ Error fetching countries: %v", err)
		defaultCountries := []map[string]string{
			{"id": "1", "name": "India"},
			{"id": "2", "name": "United States"},
			{"id": "3", "name": "United Kingdom"},
			{"id": "4", "name": "Canada"},
			{"id": "5", "name": "Australia"},
		}
		utils.RespondJSON(w, http.StatusOK, defaultCountries)
		return
	}

	countryList := make([]map[string]string, 0)
	for i, country := range countries {
		if country != nil && country != "" {
			countryList = append(countryList, map[string]string{
				"id":   string(rune(i + 49)),
				"name": country.(string),
			})
		}
	}

	if len(countryList) == 0 {
		countryList = []map[string]string{
			{"id": "1", "name": "India"},
			{"id": "2", "name": "United States"},
			{"id": "3", "name": "United Kingdom"},
			{"id": "4", "name": "Canada"},
			{"id": "5", "name": "Australia"},
		}
	}

	utils.RespondJSON(w, http.StatusOK, countryList)
}

func GetCollegesByCountry(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if country == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "country parameter required"})
		return
	}

	colleges, err := services.GetCollegesByCountry(country)
	if err != nil {
		log.Printf("❌ Error fetching colleges: %v", err)
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch colleges"})
		return
	}

	collegeList := make([]map[string]interface{}, 0)
	for _, college := range colleges {
		collegeList = append(collegeList, map[string]interface{}{
			"id":      college.CollegeName,
			"name":    college.CollegeName,
			"country": country,
			"data":    college.StudentStatistics,
		})
	}

	utils.RespondJSON(w, http.StatusOK, collegeList)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"backend": "Go",
		"version": "1.0.0",
	})
}

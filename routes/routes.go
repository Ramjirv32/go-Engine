package routes

import (
	"net/http"

	"gobackend/controllers"
	"gobackend/middleware"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/college-statistics", controllers.GetCollegeStatistics).Methods("GET")
	r.HandleFunc("/api/countries", controllers.GetCountries).Methods("GET")
	r.HandleFunc("/api/colleges-by-country", controllers.GetCollegesByCountry).Methods("GET")
	r.HandleFunc("/api/search", controllers.SearchUniversity).Methods("GET")
	r.HandleFunc("/api/all-colleges", controllers.GetAllColleges).Methods("GET")
	r.HandleFunc("/api/health", controllers.HealthCheck).Methods("GET")

	r.HandleFunc("/ws/colleges", controllers.HandleWebSocketColleges)
	r.HandleFunc("/ws/countries", controllers.HandleWebSocketCountries)
	r.HandleFunc("/ws", controllers.HandleWebSocketCountries) // Fallback

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("../App/static"))))

	r.HandleFunc("/", controllers.HomePage).Methods("GET")
	r.HandleFunc("/college-statistics", controllers.CollegeStatsPage).Methods("GET")

	r.Use(middleware.CorsMiddleware)

	return r
}

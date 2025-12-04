package controllers

import (
	"net/http"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../App/templates/university/college_statistics.html")
}

func CollegeStatsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../App/templates/university/college_statistics.html")
}

package models

type CollegeStats struct {
	CollegeName           string          `json:"college_name" bson:"college_name"`
	Country               string          `json:"country" bson:"country"`
	About                 string          `json:"about" bson:"about"`
	Location              string          `json:"location" bson:"location"`
	Summary               string          `json:"summary" bson:"summary"`
	UGPrograms            []string        `json:"ug_programs" bson:"ug_programs"`
	PGPrograms            []string        `json:"pg_programs" bson:"pg_programs"`
	PhDPrograms           []string        `json:"phd_programs" bson:"phd_programs"`
	Fees                  FeesInfo        `json:"fees" bson:"fees"`
	Scholarships          []string        `json:"scholarships" bson:"scholarships"`
	StudentGenderRatio    GenderRatio     `json:"student_gender_ratio" bson:"student_gender_ratio"`
	FacultyStaff          int             `json:"faculty_staff" bson:"faculty_staff"`
	InternationalStudents int             `json:"international_students" bson:"international_students"`
	GlobalRanking         string          `json:"global_ranking" bson:"global_ranking"`
	Departments           []string        `json:"departments" bson:"departments"`
	StudentStatistics     []StatisticItem `json:"student_statistics" bson:"student_statistics"`
	AdditionalDetails     []StatisticItem `json:"additional_details" bson:"additional_details"`
	Sources               []string        `json:"sources" bson:"sources"`
}

type FeesInfo struct {
	UGYearlyMin  int `json:"ug_yearly_min" bson:"ug_yearly_min"`
	UGYearlyMax  int `json:"ug_yearly_max" bson:"ug_yearly_max"`
	PGYearlyMin  int `json:"pg_yearly_min" bson:"pg_yearly_min"`
	PGYearlyMax  int `json:"pg_yearly_max" bson:"pg_yearly_max"`
	PhDYearlyMin int `json:"phd_yearly_min" bson:"phd_yearly_min"`
	PhDYearlyMax int `json:"phd_yearly_max" bson:"phd_yearly_max"`
}

type GenderRatio struct {
	MalePercentage   int `json:"male_percentage" bson:"male_percentage"`
	FemalePercentage int `json:"female_percentage" bson:"female_percentage"`
}

type StatisticItem struct {
	Category string      `json:"category" bson:"category"`
	Value    interface{} `json:"value" bson:"value"`
}

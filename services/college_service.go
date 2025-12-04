package services

import (
	"context"
	"log"

	"gobackend/config"
	"gobackend/models"

	"go.mongodb.org/mongo-driver/bson"
)

func GetCollegeFromCache(collegeName string) (*models.CollegeStats, error) {
	var cachedResult models.CollegeStats
	err := config.CollegeCollection.FindOne(context.TODO(), bson.M{
		"college_name": bson.M{"$regex": "^" + collegeName + "$", "$options": "i"},
	}).Decode(&cachedResult)

	if err != nil {
		return nil, err
	}

	log.Println("✅ Found in cache")
	return &cachedResult, nil
}

func SaveCollegeToCache(stats *models.CollegeStats) error {
	_, err := config.CollegeCollection.InsertOne(context.TODO(), stats)
	if err != nil {
		log.Printf("⚠️ Cache store failed: %v", err)
		return err
	}

	log.Println("💾 Cached in MongoDB")
	return nil
}

func UpdateCollegeCache(collegeName string, stats *models.CollegeStats) error {
	_, err := config.CollegeCollection.UpdateOne(
		context.TODO(),
		bson.M{"college_name": bson.M{"$regex": "^" + collegeName + "$", "$options": "i"}},
		bson.M{"$set": stats},
	)

	if err != nil {
		log.Printf("⚠️ Cache update failed: %v", err)
		return err
	}

	log.Printf("✅ Cache updated for %s", collegeName)
	return nil
}

func CompareAndUpdateCache(collegeName string, cachedData models.CollegeStats) {
	log.Printf("🔄 Background: Fetching fresh data for %s from Gemini", collegeName)

	freshStats, err := FetchCollegeDataFromGemini(collegeName)
	if err != nil {
		log.Printf("❌ Background Gemini fetch error: %v", err)
		return
	}

	hasChanged := false
	if len(cachedData.StudentStatistics) != len(freshStats.StudentStatistics) {
		hasChanged = true
	} else {
		for i, oldStat := range cachedData.StudentStatistics {
			if i < len(freshStats.StudentStatistics) {
				if oldStat.Value != freshStats.StudentStatistics[i].Value {
					hasChanged = true
					break
				}
			}
		}
	}

	if hasChanged {
		log.Printf("🔄 Changes detected for %s, updating cache...", collegeName)
		UpdateCollegeCache(collegeName, freshStats)
	} else {
		log.Printf("✅ No changes detected for %s", collegeName)
	}
}

func SearchUniversityByName(name string) (*models.CollegeStats, error) {
	var result models.CollegeStats
	err := config.CollegeCollection.FindOne(context.TODO(), bson.M{
		"college_name": bson.M{"$regex": name, "$options": "i"},
	}).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetAllColleges() ([]models.CollegeStats, error) {
	cursor, err := config.CollegeCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var colleges []models.CollegeStats
	if err := cursor.All(context.TODO(), &colleges); err != nil {
		return nil, err
	}

	return colleges, nil
}

func GetDistinctCountries() ([]interface{}, error) {
	countries, err := config.CollegeCollection.Distinct(context.TODO(), "country", bson.M{})
	if err != nil {
		return nil, err
	}

	return countries, nil
}

func GetCollegesByCountry(country string) ([]models.CollegeStats, error) {
	cursor, err := config.CollegeCollection.Find(context.TODO(), bson.M{
		"country": bson.M{"$regex": "^" + country + "$", "$options": "i"},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var colleges []models.CollegeStats
	for cursor.Next(context.TODO()) {
		var college models.CollegeStats
		if err := cursor.Decode(&college); err == nil {
			colleges = append(colleges, college)
		}
	}

	return colleges, nil
}

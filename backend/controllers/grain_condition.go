package controllers

import (
	"net/http"
	"time"

	"grain-management/database"
	"grain-management/middleware"
	"grain-management/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GrainConditionController struct{}

func NewGrainConditionController() *GrainConditionController {
	return &GrainConditionController{}
}

type GrainConditionCreateRequest struct {
	GranaryID       string  `json:"granary_id" binding:"required"`
	RecordTime      string  `json:"record_time" binding:"required"`
	AvgTemperature  float64 `json:"avg_temperature"`
	MaxTemperature  float64 `json:"max_temperature"`
	MinTemperature  float64 `json:"min_temperature"`
	AvgHumidity     float64 `json:"avg_humidity"`
	GrainLevel      float64 `json:"grain_level"`
	PestFound       bool    `json:"pest_found"`
	MoldFound       bool    `json:"mold_found"`
	AbnormalAreas   string  `json:"abnormal_areas"`
	WeatherCondition string  `json:"weather_condition"`
	Remark          string  `json:"remark"`
}

func (gcc *GrainConditionController) Create(c *gin.Context) {
	var req GrainConditionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)
	recordTime, _ := time.Parse(time.RFC3339, req.RecordTime)

	record := models.GrainConditionRecord{
		ID:               uuid.New(),
		GranaryID:        uuid.MustParse(req.GranaryID),
		RecorderID:       uuid.MustParse(userID),
		RecordTime:       recordTime,
		AvgTemperature:   req.AvgTemperature,
		MaxTemperature:   req.MaxTemperature,
		MinTemperature:   req.MinTemperature,
		AvgHumidity:      req.AvgHumidity,
		GrainLevel:       req.GrainLevel,
		PestFound:        req.PestFound,
		MoldFound:        req.MoldFound,
		AbnormalAreas:    req.AbnormalAreas,
		WeatherCondition: req.WeatherCondition,
		Remark:           req.Remark,
	}

	tx := database.DB.Begin()
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否需要生成翻仓建议
	err := CheckAndGenerateTurnoverSuggestion(tx, &record)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成翻仓建议失败: " + err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, record)
}

func (gcc *GrainConditionController) List(c *gin.Context) {
	granaryID := c.Query("granary_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	db := database.DB.Preload("Granary").Preload("Recorder")

	if granaryID != "" {
		db = db.Where("granary_id = ?", granaryID)
	}
	if startDate != "" {
		db = db.Where("record_time >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("record_time <= ?", endDate)
	}

	var records []models.GrainConditionRecord
	if err := db.Order("record_time DESC").Limit(100).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (gcc *GrainConditionController) Get(c *gin.Context) {
	id := c.Param("id")
	var record models.GrainConditionRecord
	if err := database.DB.Preload("Granary").Preload("Recorder").
		First(&record, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "粮情记录不存在"})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (gcc *GrainConditionController) GetSensorReadings(c *gin.Context) {
	granaryID := c.Param("id")
	sensorType := c.Query("type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	limit := c.DefaultQuery("limit", "1000")

	db := database.DB.Table("sensor_readings sr").
		Select("sr.*, s.code as sensor_code, s.type as sensor_type, s.location_desc, s.unit").
		Joins("LEFT JOIN sensors s ON sr.sensor_id = s.id").
		Where("sr.granary_id = ?", granaryID)

	if sensorType != "" {
		db = db.Where("s.type = ?", sensorType)
	}
	if startTime != "" {
		db = db.Where("sr.reading_time >= ?", startTime)
	}
	if endTime != "" {
		db = db.Where("sr.reading_time <= ?", endTime)
	}

	type ReadingResult struct {
		ID           int64     `json:"id"`
		SensorID     string    `json:"sensor_id"`
		GranaryID    string    `json:"granary_id"`
		ReadingTime  time.Time `json:"reading_time"`
		Value        float64   `json:"value"`
		IsAbnormal   bool      `json:"is_abnormal"`
		SensorCode   string    `json:"sensor_code"`
		SensorType   string    `json:"sensor_type"`
		LocationDesc string    `json:"location_desc"`
		Unit         string    `json:"unit"`
	}

	var results []ReadingResult
	if err := db.Order("sr.reading_time DESC").
		Limit(toInt(limit, 1000)).Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func toInt(s string, def int) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	if n == 0 {
		return def
	}
	return n
}

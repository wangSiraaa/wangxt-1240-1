package controllers

import (
	"fmt"
	"net/http"
	"time"

	"grain-management/database"
	"grain-management/middleware"
	"grain-management/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UnsealController struct{}

func NewUnsealController() *UnsealController {
	return &UnsealController{}
}

type SafeLimits map[string]float64

var DefaultSafeLimits = SafeLimits{
	"gas_ph3": 0.3,
	"gas_h2s": 10.0,
	"co2":     5000.0,
}

type UnsealCreateRequest struct {
	GranaryID        string            `json:"granary_id" binding:"required"`
	FumigationPlanID string            `json:"fumigation_plan_id"`
	UnsealType       models.UnsealType `json:"unseal_type" binding:"required"`
	StartTime        string            `json:"start_time"`
	WeatherCondition string            `json:"weather_condition"`
	Remark           string            `json:"remark"`
}

func (uc *UnsealController) Create(c *gin.Context) {
	var req UnsealCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)

	var startTime *time.Time
	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err == nil {
			startTime = &t
		}
	}
	if startTime == nil {
		now := time.Now()
		startTime = &now
	}

	tx := database.DB.Begin()

	record := models.UnsealRecord{
		ID:               uuid.New(),
		GranaryID:        uuid.MustParse(req.GranaryID),
		RecorderID:       uuid.MustParse(userID),
		UnsealType:       req.UnsealType,
		StartTime:        startTime,
		WeatherCondition: req.WeatherCondition,
		Remark:           req.Remark,
	}

	if req.FumigationPlanID != "" {
		pid := uuid.MustParse(req.FumigationPlanID)
		record.FumigationPlanID = &pid
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.UnsealType == models.UnsealTypeVentilation {
		if err := tx.Model(&models.Granary{}).
			Where("id = ?", req.GranaryID).
			Update("status", models.GranaryVentilating).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if req.FumigationPlanID != "" {
			var plan models.FumigationPlan
			if err := tx.First(&plan, "id = ?", req.FumigationPlanID).Error; err == nil {
				interval := plan.DetectionIntervalHours
				if interval <= 0 {
					interval = 4
				}
				nextTime := startTime.Add(time.Duration(interval) * time.Hour)
				if err := tx.Model(&plan).Updates(map[string]interface{}{
					"next_detection_time": nextTime,
					"safety_confirmed":    false,
				}).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
	}

	tx.Commit()
	c.JSON(http.StatusCreated, record)
}

func (uc *UnsealController) List(c *gin.Context) {
	granaryID := c.Query("granary_id")
	unsealType := c.Query("type")

	db := database.DB.Preload("Granary").Preload("Recorder")

	if granaryID != "" {
		db = db.Where("granary_id = ?", granaryID)
	}
	if unsealType != "" {
		db = db.Where("unseal_type = ?", unsealType)
	}

	var records []models.UnsealRecord
	if err := db.Order("created_at DESC").Limit(100).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (uc *UnsealController) Get(c *gin.Context) {
	id := c.Param("id")
	var record models.UnsealRecord
	if err := database.DB.Preload("Granary").Preload("Recorder").
		First(&record, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "解封记录不存在"})
		return
	}
	c.JSON(http.StatusOK, record)
}

type GasDetectionRequest struct {
	UnsealID        string  `json:"unseal_id"`
	GasType         string  `json:"gas_type" binding:"required"`
	Concentration   float64 `json:"concentration" binding:"required"`
	SafeLimit       float64 `json:"safe_limit"`
	DetectionPoints string  `json:"detection_points"`
	Remark          string  `json:"remark"`
}

func (uc *UnsealController) AddGasDetection(c *gin.Context) {
	granaryID := c.Param("id")
	var req GasDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)

	safeLimit := req.SafeLimit
	if safeLimit == 0 {
		if limit, ok := DefaultSafeLimits[req.GasType]; ok {
			safeLimit = limit
		} else {
			safeLimit = 0
		}
	}

	isSafe := req.Concentration <= safeLimit

	detection := models.GasDetectionRecord{
		ID:              uuid.New(),
		GranaryID:       uuid.MustParse(granaryID),
		DetectorID:      uuid.MustParse(userID),
		DetectionTime:   time.Now(),
		GasType:         req.GasType,
		Concentration:   req.Concentration,
		SafeLimit:       safeLimit,
		IsSafe:          isSafe,
		DetectionPoints: req.DetectionPoints,
		Remark:          req.Remark,
	}

	if req.UnsealID != "" {
		uid := uuid.MustParse(req.UnsealID)
		detection.UnsealID = &uid
	}

	if err := database.DB.Create(&detection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, detection)
}

func (uc *UnsealController) ListGasDetections(c *gin.Context) {
	granaryID := c.Query("granary_id")
	unsealID := c.Query("unseal_id")

	db := database.DB.Preload("Granary").Preload("Detector").Preload("Unseal")

	if granaryID != "" {
		db = db.Where("granary_id = ?", granaryID)
	}
	if unsealID != "" {
		db = db.Where("unseal_id = ?", unsealID)
	}

	var detections []models.GasDetectionRecord
	if err := db.Order("detection_time DESC").Limit(200).Find(&detections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detections)
}

type CompleteUnsealRequest struct {
	EndTime          string `json:"end_time"`
	FinalGasReadings string `json:"final_gas_readings"`
	IsSafe           bool   `json:"is_safe"`
	Remark           string `json:"remark"`
}

func (uc *UnsealController) CompleteUnseal(c *gin.Context) {
	id := c.Param("id")
	var req CompleteUnsealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var record models.UnsealRecord
	if err := tx.First(&record, "id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "解封记录不存在"})
		return
	}

	// 关键业务规则：安全员未确认达标不能解封
	if record.UnsealType == models.UnsealTypeUnseal && record.FumigationPlanID != nil {
		var plan models.FumigationPlan
		if err := tx.First(&plan, "id = ?", *record.FumigationPlanID).Error; err == nil {
			if !plan.SafetyConfirmed {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "安全员尚未确认气体检测达标，不能解封",
					"error_code": "SAFETY_NOT_CONFIRMED",
				})
				return
			}
		}
	}

	// 关键业务规则：气体浓度未达安全值不能解封
	if record.UnsealType == models.UnsealTypeUnseal && !req.IsSafe {
		var count int64
		tx.Model(&models.GasDetectionRecord{}).
			Where("unseal_id = ? AND is_safe = ?", id, false).
			Count(&count)

		if count > 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "气体浓度未达安全值，不能解封",
				"error_code": "GAS_NOT_SAFE",
			})
			return
		}
	}

	// 检查最新气体检测
	var latestDetection models.GasDetectionRecord
	err := tx.Where("unseal_id = ?", id).Order("detection_time DESC").First(&latestDetection).Error
	if record.UnsealType == models.UnsealTypeUnseal && err == nil && !latestDetection.IsSafe {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "气体浓度未达安全值（" + latestDetection.GasType + ": " + formatFloat(latestDetection.Concentration) + "，安全值: " + formatFloat(latestDetection.SafeLimit) + "），不能解封",
			"error_code": "GAS_NOT_SAFE",
		})
		return
	}

	var endTime *time.Time
	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err == nil {
			endTime = &t
		}
	}
	if endTime == nil {
		now := time.Now()
		endTime = &now
	}

	record.EndTime = endTime
	record.FinalGasReadings = req.FinalGasReadings
	record.IsSafe = req.IsSafe
	if req.Remark != "" {
		record.Remark = req.Remark
	}

	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新仓房状态为正常
	if record.UnsealType == models.UnsealTypeUnseal {
		if err := tx.Model(&models.Granary{}).
			Where("id = ?", record.GranaryID).
			Update("status", models.GranaryNormal).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if record.UnsealType == models.UnsealTypeVentilation {
		if err := tx.Model(&models.Granary{}).
			Where("id = ?", record.GranaryID).
			Update("status", models.GranarySealed).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, record)
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.4f", f)
}

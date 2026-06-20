package controllers

import (
	"net/http"
	"time"

	"grain-management/database"
	"grain-management/middleware"
	"grain-management/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (tsc *TurnoverSuggestionController) List(c *gin.Context) {
	status := c.Query("status")
	priority := c.Query("priority")
	granaryID := c.Query("granary_id")

	db := database.DB.Preload("Granary").Preload("Handler").Preload("SourceRecord")

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if priority != "" {
		db = db.Where("priority = ?", priority)
	}
	if granaryID != "" {
		db = db.Where("granary_id = ?", granaryID)
	}

	var suggestions []models.GrainTurnoverSuggestion
	if err := db.Order("priority DESC, created_at DESC").Limit(100).Find(&suggestions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, suggestions)
}

func (tsc *TurnoverSuggestionController) Get(c *gin.Context) {
	id := c.Param("id")
	var suggestion models.GrainTurnoverSuggestion
	if err := database.DB.Preload("Granary").Preload("Handler").Preload("SourceRecord").
		First(&suggestion, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "翻仓建议不存在"})
		return
	}
	c.JSON(http.StatusOK, suggestion)
}

type HandleSuggestionRequest struct {
	Status       models.SuggestionStatus `json:"status" binding:"required"`
	HandleRemark string                  `json:"handle_remark"`
}

func (tsc *TurnoverSuggestionController) Handle(c *gin.Context) {
	id := c.Param("id")
	var req HandleSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)
	handlerID := uuid.MustParse(userID)
	now := time.Now()

	var suggestion models.GrainTurnoverSuggestion
	if err := database.DB.First(&suggestion, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "翻仓建议不存在"})
		return
	}

	suggestion.Status = req.Status
	suggestion.HandlerID = &handlerID
	suggestion.HandledAt = &now
	if req.HandleRemark != "" {
		suggestion.HandleRemark = req.HandleRemark
	}

	if err := database.DB.Save(&suggestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

func (tsc *TurnoverSuggestionController) GetDashboardStats(c *gin.Context) {
	var stats struct {
		TotalGranaries      int64 `json:"total_granaries"`
		NormalGranaries     int64 `json:"normal_granaries"`
		FumigatingGranaries int64 `json:"fumigating_granaries"`
		VentilatingGranaries int64 `json:"ventilating_granaries"`
		SealedGranaries     int64 `json:"sealed_granaries"`
		AbnormalGranaries   int64 `json:"abnormal_granaries"`

		PendingSuggestions   int64 `json:"pending_suggestions"`
		ProcessingSuggestions int64 `json:"processing_suggestions"`
		UrgentSuggestions    int64 `json:"urgent_suggestions"`
		HighSuggestions      int64 `json:"high_suggestions"`

		PendingFumigation    int64 `json:"pending_fumigation"`
		InProgressFumigation int64 `json:"in_progress_fumigation"`
		TodayRecords         int64 `json:"today_records"`

		TodayAvgTemp     float64 `json:"today_avg_temp"`
		TodayMaxTemp     float64 `json:"today_max_temp"`
		AbnormalTempCount int64  `json:"abnormal_temp_count"`
	}

	db := database.DB

	db.Model(&models.Granary{}).Count(&stats.TotalGranaries)
	db.Model(&models.Granary{}).Where("status = ?", models.GranaryNormal).Count(&stats.NormalGranaries)
	db.Model(&models.Granary{}).Where("status = ?", models.GranaryFumigating).Count(&stats.FumigatingGranaries)
	db.Model(&models.Granary{}).Where("status = ?", models.GranaryVentilating).Count(&stats.VentilatingGranaries)
	db.Model(&models.Granary{}).Where("status = ?", models.GranarySealed).Count(&stats.SealedGranaries)
	db.Model(&models.Granary{}).Where("status = ?", models.GranaryAbnormal).Count(&stats.AbnormalGranaries)

	db.Model(&models.GrainTurnoverSuggestion{}).Where("status = ?", models.SuggestionPending).Count(&stats.PendingSuggestions)
	db.Model(&models.GrainTurnoverSuggestion{}).Where("status = ?", models.SuggestionProcessing).Count(&stats.ProcessingSuggestions)
	db.Model(&models.GrainTurnoverSuggestion{}).Where("priority = ?", models.PriorityUrgent).Count(&stats.UrgentSuggestions)
	db.Model(&models.GrainTurnoverSuggestion{}).Where("priority = ?", models.PriorityHigh).Count(&stats.HighSuggestions)

	db.Model(&models.FumigationPlan{}).Where("status IN ?", []string{string(models.FumigationDraft), string(models.FumigationPendingApproval)}).Count(&stats.PendingFumigation)
	db.Model(&models.FumigationPlan{}).Where("status = ?", models.FumigationInProgress).Count(&stats.InProgressFumigation)

	todayStart := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.GrainConditionRecord{}).Where("record_time >= ?", todayStart).Count(&stats.TodayRecords)

	type TempResult struct {
		AvgTemp float64
		MaxTemp float64
	}
	var tempRes TempResult
	db.Model(&models.GrainConditionRecord{}).
		Select("COALESCE(AVG(avg_temperature), 0) as avg_temp, COALESCE(MAX(max_temperature), 0) as max_temp").
		Where("record_time >= ?", todayStart).
		Scan(&tempRes)
	stats.TodayAvgTemp = tempRes.AvgTemp
	stats.TodayMaxTemp = tempRes.MaxTemp

	db.Model(&models.SensorReading{}).
		Where("is_abnormal = ? AND reading_time >= ?", true, todayStart.AddDate(0, 0, -1)).
		Count(&stats.AbnormalTempCount)

	c.JSON(http.StatusOK, stats)
}

func CheckGasSafe(tx *gorm.DB, unsealID uuid.UUID) (bool, string) {
	var latestDetections []models.GasDetectionRecord
	tx.Where("unseal_id = ?", unsealID).
		Order("gas_type ASC, detection_time DESC").
		Find(&latestDetections)

	seen := make(map[string]models.GasDetectionRecord)
	for _, d := range latestDetections {
		if _, ok := seen[d.GasType]; !ok {
			seen[d.GasType] = d
		}
	}

	for gasType, d := range seen {
		if !d.IsSafe {
			return false, "气体 " + gasType + " 浓度超标"
		}
	}

	return true, ""
}

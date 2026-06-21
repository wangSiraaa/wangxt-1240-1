package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"grain-management/database"
	"grain-management/middleware"
	"grain-management/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FumigationController struct{}

func NewFumigationController() *FumigationController {
	return &FumigationController{}
}

type FumigationCreateRequest struct {
	GranaryID           string  `json:"granary_id" binding:"required"`
	PlanTitle           string  `json:"plan_title" binding:"required"`
	ChemicalType        string  `json:"chemical_type"`
	ChemicalName        string  `json:"chemical_name"`
	Dosage              float64 `json:"dosage"`
	DosageUnit          string  `json:"dosage_unit"`
	TargetConcentration float64 `json:"target_concentration"`
	PlanStartTime       string  `json:"plan_start_time"`
	PlanEndTime         string  `json:"plan_end_time"`
	ExpectedSealHours   int     `json:"expected_seal_hours"`
	Reason              string  `json:"reason"`
}

func (fc *FumigationController) Create(c *gin.Context) {
	var req FumigationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)

	planNo := fmt.Sprintf("FM%s", time.Now().Format("20060102150405"))

	var planStartTime, planEndTime *time.Time
	if req.PlanStartTime != "" {
		t, err := time.Parse(time.RFC3339, req.PlanStartTime)
		if err == nil {
			planStartTime = &t
		}
	}
	if req.PlanEndTime != "" {
		t, err := time.Parse(time.RFC3339, req.PlanEndTime)
		if err == nil {
			planEndTime = &t
		}
	}

	plan := models.FumigationPlan{
		ID:                  uuid.New(),
		GranaryID:           uuid.MustParse(req.GranaryID),
		PlanNo:              planNo,
		PlanTitle:           req.PlanTitle,
		CreatorID:           uuid.MustParse(userID),
		ChemicalType:        req.ChemicalType,
		ChemicalName:        req.ChemicalName,
		Dosage:              req.Dosage,
		DosageUnit:          req.DosageUnit,
		TargetConcentration: req.TargetConcentration,
		PlanStartTime:       planStartTime,
		PlanEndTime:         planEndTime,
		ExpectedSealHours:   req.ExpectedSealHours,
		Reason:              req.Reason,
		Status:              models.FumigationDraft,
	}

	if err := database.DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

func (fc *FumigationController) List(c *gin.Context) {
	status := c.Query("status")
	granaryID := c.Query("granary_id")

	db := database.DB.Preload("Granary").Preload("Creator").Preload("Approver")

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if granaryID != "" {
		db = db.Where("granary_id = ?", granaryID)
	}

	var plans []models.FumigationPlan
	if err := db.Order("created_at DESC").Limit(100).Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plans)
}

func (fc *FumigationController) Get(c *gin.Context) {
	id := c.Param("id")
	var plan models.FumigationPlan
	if err := database.DB.Preload("Granary").Preload("Creator").Preload("Approver").
		First(&plan, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "熏蒸方案不存在"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (fc *FumigationController) SubmitForApproval(c *gin.Context) {
	id := c.Param("id")
	var plan models.FumigationPlan
	if err := database.DB.First(&plan, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "熏蒸方案不存在"})
		return
	}

	if plan.Status != models.FumigationDraft && plan.Status != models.FumigationRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许提交审批"})
		return
	}

	plan.Status = models.FumigationPendingApproval
	if err := database.DB.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

type ApprovalRequest struct {
	Approved      bool   `json:"approved" binding:"required"`
	ApprovalRemark string `json:"approval_remark"`
}

func (fc *FumigationController) Approve(c *gin.Context) {
	id := c.Param("id")
	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)
	approverID := uuid.MustParse(userID)
	now := time.Now()

	tx := database.DB.Begin()

	var plan models.FumigationPlan
	if err := tx.First(&plan, "id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "熏蒸方案不存在"})
		return
	}

	if plan.Status != models.FumigationPendingApproval {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许审批"})
		return
	}

	if req.Approved {
		plan.Status = models.FumigationApproved
		plan.ApproverID = &approverID
		plan.ApprovalRemark = req.ApprovalRemark
		plan.ApprovedAt = &now
	} else {
		plan.Status = models.FumigationRejected
		plan.ApproverID = &approverID
		plan.ApprovalRemark = req.ApprovalRemark
		plan.ApprovedAt = &now
	}

	if err := tx.Save(&plan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, plan)
}

type PeopleClearRequest struct {
	Cleared bool `json:"cleared" binding:"required"`
}

func (fc *FumigationController) MarkPeopleCleared(c *gin.Context) {
	id := c.Param("id")
	var req PeopleClearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)
	clearedBy := uuid.MustParse(userID)
	now := time.Now()

	var plan models.FumigationPlan
	if err := database.DB.First(&plan, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "熏蒸方案不存在"})
		return
	}

	if plan.Status != models.FumigationApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有已批准的方案才能进行清场确认"})
		return
	}

	if req.Cleared {
		plan.PeopleCleared = true
		plan.PeopleClearedTime = &now
		plan.PeopleClearedBy = &clearedBy
	} else {
		plan.PeopleCleared = false
		plan.PeopleClearedTime = nil
		plan.PeopleClearedBy = nil
	}

	if err := database.DB.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func (fc *FumigationController) StartExecution(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetCurrentUserID(c)

	tx := database.DB.Begin()

	var plan models.FumigationPlan
	if err := tx.First(&plan, "id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "熏蒸方案不存在"})
		return
	}

	if plan.Status != models.FumigationApproved {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有已批准的方案才能开始执行"})
		return
	}

	// 关键业务规则：未清场不能开始投药
	if !plan.PeopleCleared {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "仓内人员未清场，不能开始投药",
			"error_code": "PEOPLE_NOT_CLEARED",
		})
		return
	}

	now := time.Now()
	plan.Status = models.FumigationInProgress
	if err := tx.Save(&plan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新仓房状态
	if err := tx.Model(&models.Granary{}).
		Where("id = ?", plan.GranaryID).
		Update("status", models.GranaryFumigating).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建执行记录
	execution := models.FumigationExecution{
		ID:              uuid.New(),
		PlanID:          plan.ID,
		GranaryID:       plan.GranaryID,
		OperatorID:      uuid.MustParse(userID),
		ActualStartTime: &now,
	}
	if err := tx.Create(&execution).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"plan":      plan,
		"execution": execution,
	})
}

type CompleteExecutionRequest struct {
	ActualDosage        float64 `json:"actual_dosage"`
	ConcentrationReadings string  `json:"concentration_readings"`
	WeatherDuring       string  `json:"weather_during"`
	Remark              string  `json:"remark"`
}

func (fc *FumigationController) CompleteExecution(c *gin.Context) {
	id := c.Param("id")
	var req CompleteExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var plan models.FumigationPlan
	if err := tx.First(&plan, "id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "熏蒸方案不存在"})
		return
	}

	if plan.Status != models.FumigationInProgress {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有执行中的方案才能完成"})
		return
	}

	now := time.Now()
	plan.Status = models.FumigationCompleted
	if err := tx.Save(&plan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新仓房状态为密封
	if err := tx.Model(&models.Granary{}).
		Where("id = ?", plan.GranaryID).
		Update("status", models.GranarySealed).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新执行记录
	var execution models.FumigationExecution
	err := tx.Where("plan_id = ? AND actual_end_time IS NULL", id).
		First(&execution).Error
	if err == nil {
		execution.ActualEndTime = &now
		execution.ChemicalActualDosage = req.ActualDosage
		execution.ConcentrationReadings = req.ConcentrationReadings
		execution.WeatherDuring = req.WeatherDuring
		execution.Remark = req.Remark
		if err := tx.Save(&execution).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, plan)
}

func (fc *FumigationController) ListExecutions(c *gin.Context) {
	planID := c.Query("plan_id")
	granaryID := c.Query("granary_id")

	db := database.DB.Preload("Plan").Preload("Granary").Preload("Operator")

	if planID != "" {
		db = db.Where("plan_id = ?", planID)
	}
	if granaryID != "" {
		db = db.Where("granary_id = ?", granaryID)
	}

	var executions []models.FumigationExecution
	if err := db.Order("created_at DESC").Limit(100).Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, executions)
}

type SafetyConfirmRequest struct {
	Confirmed bool   `json:"confirmed" binding:"required"`
	Remark    string `json:"remark"`
}

func (fc *FumigationController) SafetyConfirm(c *gin.Context) {
	id := c.Param("id")
	var req SafetyConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetCurrentUserID(c)
	confirmerID := uuid.MustParse(userID)
	now := time.Now()

	var plan models.FumigationPlan
	if err := database.DB.First(&plan, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "熏蒸方案不存在"})
		return
	}

	if plan.Status != models.FumigationCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有已完成的熏蒸方案才能进行达标确认"})
		return
	}

	if req.Confirmed {
		plan.SafetyConfirmed = true
		plan.SafetyConfirmedAt = &now
		plan.SafetyConfirmedBy = &confirmerID
		plan.SafetyConfirmRemark = req.Remark
	} else {
		plan.SafetyConfirmed = false
		plan.SafetyConfirmedAt = nil
		plan.SafetyConfirmedBy = nil
		plan.SafetyConfirmRemark = req.Remark
	}

	if err := database.DB.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func CheckFumigationRules(plan *models.FumigationPlan) error {
	if plan == nil {
		return errors.New("熏蒸方案不存在")
	}

	if !plan.PeopleCleared {
		return errors.New("仓内人员未清场，不能开始投药")
	}

	if plan.Status != models.FumigationApproved {
		return errors.New("熏蒸方案未获审批，不能执行")
	}

	return nil
}

package controllers

import (
	"net/http"

	"grain-management/database"
	"grain-management/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GranaryController struct{}

func NewGranaryController() *GranaryController {
	return &GranaryController{}
}

func (gc *GranaryController) List(c *gin.Context) {
	status := c.Query("status")
	keyword := c.Query("keyword")

	db := database.DB.Preload("Keeper")

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		db = db.Where("code ILIKE ? OR name ILIKE ? OR location ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var granaries []models.Granary
	if err := db.Order("code ASC").Find(&granaries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, granaries)
}

func (gc *GranaryController) Get(c *gin.Context) {
	id := c.Param("id")
	var granary models.Granary
	if err := database.DB.Preload("Keeper").Preload("Sensors").
		First(&granary, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "仓房不存在"})
		return
	}
	c.JSON(http.StatusOK, granary)
}

type GranaryCreateRequest struct {
	Code         string      `json:"code" binding:"required"`
	Name         string      `json:"name" binding:"required"`
	Location     string      `json:"location"`
	Capacity     float64     `json:"capacity"`
	GrainType    string      `json:"grain_type"`
	GrainVariety string      `json:"grain_variety"`
	GrainWeight  float64     `json:"grain_weight"`
	KeeperID     *uuid.UUID  `json:"keeper_id"`
	Remark       string      `json:"remark"`
}

func (gc *GranaryController) Create(c *gin.Context) {
	var req GranaryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	granary := models.Granary{
		ID:           uuid.New(),
		Code:         req.Code,
		Name:         req.Name,
		Location:     req.Location,
		Capacity:     req.Capacity,
		GrainType:    req.GrainType,
		GrainVariety: req.GrainVariety,
		GrainWeight:  req.GrainWeight,
		Status:       models.GranaryNormal,
		KeeperID:     req.KeeperID,
		Remark:       req.Remark,
	}

	if err := database.DB.Create(&granary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, granary)
}

type GranaryUpdateRequest struct {
	Name         string            `json:"name"`
	Location     string            `json:"location"`
	Capacity     float64           `json:"capacity"`
	GrainType    string            `json:"grain_type"`
	GrainVariety string            `json:"grain_variety"`
	GrainWeight  float64           `json:"grain_weight"`
	Status       models.GranaryStatus `json:"status"`
	KeeperID     *uuid.UUID        `json:"keeper_id"`
	Remark       string            `json:"remark"`
}

func (gc *GranaryController) Update(c *gin.Context) {
	id := c.Param("id")
	var granary models.Granary
	if err := database.DB.First(&granary, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "仓房不存在"})
		return
	}

	var req GranaryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Capacity > 0 {
		updates["capacity"] = req.Capacity
	}
	if req.GrainType != "" {
		updates["grain_type"] = req.GrainType
	}
	if req.GrainVariety != "" {
		updates["grain_variety"] = req.GrainVariety
	}
	if req.GrainWeight >= 0 {
		updates["grain_weight"] = req.GrainWeight
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.KeeperID != nil {
		updates["keeper_id"] = req.KeeperID
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	if err := database.DB.Model(&granary).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	database.DB.Preload("Keeper").First(&granary, "id = ?", id)
	c.JSON(http.StatusOK, granary)
}

func (gc *GranaryController) Delete(c *gin.Context) {
	id := c.Param("id")
	var granary models.Granary
	if err := database.DB.First(&granary, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "仓房不存在"})
		return
	}

	if err := database.DB.Delete(&granary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (gc *GranaryController) GetSensors(c *gin.Context) {
	id := c.Param("id")
	var sensors []models.Sensor
	if err := database.DB.Where("granary_id = ? AND is_active = ?", id, true).
		Order("code ASC").Find(&sensors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensors)
}

type SensorCreateRequest struct {
	Code         string          `json:"code" binding:"required"`
	Type         models.SensorType `json:"type" binding:"required"`
	LocationDesc string          `json:"location_desc"`
	PositionX    float64         `json:"position_x"`
	PositionY    float64         `json:"position_y"`
	PositionZ    float64         `json:"position_z"`
	Unit         string          `json:"unit"`
}

func (gc *GranaryController) AddSensor(c *gin.Context) {
	granaryID := c.Param("id")
	var req SensorCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensor := models.Sensor{
		ID:           uuid.New(),
		GranaryID:    uuid.MustParse(granaryID),
		Code:         req.Code,
		Type:         req.Type,
		LocationDesc: req.LocationDesc,
		PositionX:    req.PositionX,
		PositionY:    req.PositionY,
		PositionZ:    req.PositionZ,
		Unit:         req.Unit,
		IsActive:     true,
	}

	if err := database.DB.Create(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sensor)
}

func (gc *GranaryController) ListKeepers(c *gin.Context) {
	var keepers []models.User
	if err := database.DB.Where("role = ?", models.RoleKeeper).
		Order("full_name ASC").Find(&keepers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keepers)
}

func (gc *GranaryController) AddSensorReading(c *gin.Context) {
	type ReadingRequest struct {
		Value      float64 `json:"value" binding:"required"`
		IsAbnormal bool    `json:"is_abnormal"`
	}
	var req ReadingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensorID := c.Param("sensorId")
	var sensor models.Sensor
	if err := database.DB.First(&sensor, "id = ?", sensorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "传感器不存在"})
		return
	}

	reading := models.SensorReading{
		SensorID:   sensor.ID,
		GranaryID:  sensor.GranaryID,
		Value:      req.Value,
		IsAbnormal: req.IsAbnormal,
	}

	if err := database.DB.Create(&reading).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reading)
}

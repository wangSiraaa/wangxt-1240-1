package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"grain-management/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TempThresholdWarning = 25.0
	TempThresholdDanger  = 30.0
	TempDiffThreshold    = 5.0
)

type TurnoverSuggestionController struct{}

func NewTurnoverSuggestionController() *TurnoverSuggestionController {
	return &TurnoverSuggestionController{}
}

func CheckAndGenerateTurnoverSuggestion(tx *gorm.DB, record *models.GrainConditionRecord) error {
	needGenerate := false
	var priority models.SuggestionPriority = models.PriorityNormal
	var abnormalDesc string
	var tempAnomalyData map[string]interface{} = make(map[string]interface{})

	if record.MaxTemperature >= TempThresholdDanger {
		needGenerate = true
		priority = models.PriorityUrgent
		abnormalDesc = fmt.Sprintf("最高温度 %.2f°C 超过危险阈值 (%.1f°C)", record.MaxTemperature, TempThresholdDanger)
		tempAnomalyData["reason"] = "max_temp_danger"
	} else if record.MaxTemperature >= TempThresholdWarning {
		needGenerate = true
		priority = models.PriorityHigh
		abnormalDesc = fmt.Sprintf("最高温度 %.2f°C 超过预警阈值 (%.1f°C)", record.MaxTemperature, TempThresholdWarning)
		tempAnomalyData["reason"] = "max_temp_warning"
	}

	if record.MaxTemperature-record.MinTemperature >= TempDiffThreshold {
		needGenerate = true
		if priority == models.PriorityNormal {
			priority = models.PriorityHigh
		}
		abnormalDesc += fmt.Sprintf("；温差 %.2f°C 超过阈值 (%.1f°C)", record.MaxTemperature-record.MinTemperature, TempDiffThreshold)
		tempAnomalyData["temp_diff"] = record.MaxTemperature - record.MinTemperature
	}

	if record.PestFound {
		needGenerate = true
		if priority != models.PriorityUrgent {
			priority = models.PriorityHigh
		}
		abnormalDesc += "；发现虫害迹象"
		tempAnomalyData["pest_found"] = true
	}

	if record.MoldFound {
		needGenerate = true
		priority = models.PriorityUrgent
		abnormalDesc += "；发现霉变迹象"
		tempAnomalyData["mold_found"] = true
	}

	if record.AbnormalAreas != "" {
		needGenerate = true
		abnormalDesc += "；存在粮情异常区域"
		var areas interface{}
		if err := json.Unmarshal([]byte(record.AbnormalAreas), &areas); err == nil {
			tempAnomalyData["abnormal_areas"] = areas
		}
	}

	if !needGenerate {
		return nil
	}

	tempAnomalyData["avg_temp"] = record.AvgTemperature
	tempAnomalyData["max_temp"] = record.MaxTemperature
	tempAnomalyData["min_temp"] = record.MinTemperature
	tempAnomalyData["avg_humidity"] = record.AvgHumidity

	anomalyJSON, _ := json.Marshal(tempAnomalyData)

	suggestionContent := generateSuggestionContent(record, priority, tempAnomalyData)

	suggestionNo := fmt.Sprintf("TC%s", time.Now().Format("20060102150405"))

	suggestion := models.GrainTurnoverSuggestion{
		ID:               uuid.New(),
		GranaryID:        record.GranaryID,
		SourceRecordID:   &record.ID,
		SuggestionNo:     suggestionNo,
		AbnormalAreaDesc: abnormalDesc,
		TemperatureAnomaly: string(anomalyJSON),
		SuggestionContent: suggestionContent,
		Priority:         priority,
		Status:           models.SuggestionPending,
	}

	return tx.Create(&suggestion).Error
}

func generateSuggestionContent(record *models.GrainConditionRecord, priority models.SuggestionPriority, anomalyData map[string]interface{}) string {
	content := "根据粮情记录分析，建议采取以下措施：\n\n"

	content += fmt.Sprintf("【基本情况】仓房粮温均值 %.2f°C，最高 %.2f°C，最低 %.2f°C，湿度 %.2f%%RH\n\n",
		record.AvgTemperature, record.MaxTemperature, record.MinTemperature, record.AvgHumidity)

	switch priority {
	case models.PriorityUrgent:
		content += "⚠️ 【紧急处置建议】\n"
		content += "1. 立即安排翻仓作业，将发热、霉变粮食彻底清理\n"
		content += "2. 检查异常区域粮食品质，对霉变粮食进行隔离处理\n"
		content += "3. 对虫害区域进行药剂熏蒸处理\n"
		content += "4. 加强通风换气，降低仓内温湿度\n"
		content += "5. 增加粮情检测频次，每日至少检测3次\n"
	case models.PriorityHigh:
		content += "🔴 【重点处置建议】\n"
		content += "1. 近期安排翻仓作业，重点处理高温区域粮食\n"
		content += "2. 检查异常区域粮食品质状况\n"
		content += "3. 启动通风降温程序\n"
		content += "4. 如发现虫害及时熏蒸处理\n"
		content += "5. 加密粮情检测，每日检测1-2次\n"
	case models.PriorityNormal:
		content += "🟡 【常规处置建议】\n"
		content += "1. 制定翻仓计划，近期安排翻仓\n"
		content += "2. 对高温区域进行局部通风降温\n"
		content += "3. 加强日常巡查，密切关注温湿度变化\n"
		content += "4. 做好熏蒸准备工作，必要时进行预防性熏蒸\n"
	case models.PriorityLow:
		content += "🟢 【日常管理建议】\n"
		content += "1. 维持正常粮情检测频次\n"
		content += "2. 关注异常区域温湿度变化趋势\n"
		content += "3. 适时安排通风换气\n"
	}

	content += "\n【翻仓技术要求】\n"
	content += "- 翻仓过程中彻底清理杂质和破碎粒\n"
	content += "- 翻仓后粮堆高度均匀，平整度误差不超过10cm\n"
	content += "- 翻仓作业记录要完整，包括作业时间、人员、粮情变化等\n"

	return content
}

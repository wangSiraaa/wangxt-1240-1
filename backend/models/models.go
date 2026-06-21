package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin         UserRole = "admin"
	RoleKeeper        UserRole = "keeper"
	RoleSafetyOfficer UserRole = "safety_officer"
	RoleDutyOfficer   UserRole = "duty_officer"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Username     string    `gorm:"size:50;unique;not null" json:"username"`
	FullName     string    `gorm:"size:100;not null" json:"full_name"`
	Role         UserRole  `gorm:"size:20;not null" json:"role"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Phone        string    `gorm:"size:20" json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GranaryStatus string

const (
	GranaryNormal      GranaryStatus = "normal"
	GranaryFumigating  GranaryStatus = "fumigating"
	GranaryVentilating GranaryStatus = "ventilating"
	GranarySealed      GranaryStatus = "sealed"
	GranaryAbnormal    GranaryStatus = "abnormal"
)

type Granary struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Code         string         `gorm:"size:50;unique;not null" json:"code"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Location     string         `gorm:"size:200" json:"location"`
	Capacity     float64        `gorm:"type:decimal(12,2)" json:"capacity"`
	GrainType    string         `gorm:"size:50" json:"grain_type"`
	GrainVariety string         `gorm:"size:100" json:"grain_variety"`
	GrainWeight  float64        `gorm:"type:decimal(12,2);default:0" json:"grain_weight"`
	Status       GranaryStatus  `gorm:"size:20;not null;default:'normal'" json:"status"`
	KeeperID     *uuid.UUID     `gorm:"type:uuid" json:"keeper_id"`
	Keeper       *User          `gorm:"foreignKey:KeeperID" json:"keeper,omitempty"`
	Remark       string         `gorm:"type:text" json:"remark"`
	Sensors      []Sensor       `gorm:"foreignKey:GranaryID" json:"sensors,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type SensorType string

const (
	SensorTemperature SensorType = "temperature"
	SensorHumidity    SensorType = "humidity"
	SensorGasPH3      SensorType = "gas_ph3"
	SensorGasH2S      SensorType = "gas_h2s"
	SensorCO2         SensorType = "co2"
	SensorO2          SensorType = "o2"
)

type Sensor struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GranaryID   uuid.UUID  `gorm:"type:uuid;not null" json:"granary_id"`
	Code        string     `gorm:"size:50;unique;not null" json:"code"`
	Type        SensorType `gorm:"size:30;not null" json:"type"`
	LocationDesc string    `gorm:"size:200" json:"location_desc"`
	PositionX   float64    `gorm:"type:decimal(8,2)" json:"position_x"`
	PositionY   float64    `gorm:"type:decimal(8,2)" json:"position_y"`
	PositionZ   float64    `gorm:"type:decimal(8,2)" json:"position_z"`
	Unit        string     `gorm:"size:20" json:"unit"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type GrainConditionRecord struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GranaryID       uuid.UUID  `gorm:"type:uuid;not null" json:"granary_id"`
	Granary         *Granary   `gorm:"foreignKey:GranaryID" json:"granary,omitempty"`
	RecorderID      uuid.UUID  `gorm:"type:uuid;not null" json:"recorder_id"`
	Recorder        *User      `gorm:"foreignKey:RecorderID" json:"recorder,omitempty"`
	RecordTime      time.Time  `gorm:"not null" json:"record_time"`
	AvgTemperature  float64    `gorm:"type:decimal(6,2)" json:"avg_temperature"`
	MaxTemperature  float64    `gorm:"type:decimal(6,2)" json:"max_temperature"`
	MinTemperature  float64    `gorm:"type:decimal(6,2)" json:"min_temperature"`
	AvgHumidity     float64    `gorm:"type:decimal(6,2)" json:"avg_humidity"`
	GrainLevel      float64    `gorm:"type:decimal(8,2)" json:"grain_level"`
	PestFound       bool       `gorm:"default:false" json:"pest_found"`
	MoldFound       bool       `gorm:"default:false" json:"mold_found"`
	AbnormalAreas   string     `gorm:"type:jsonb" json:"abnormal_areas"`
	WeatherCondition string    `gorm:"size:100" json:"weather_condition"`
	Remark          string     `gorm:"type:text" json:"remark"`
	CreatedAt       time.Time  `json:"created_at"`
}

type SensorReading struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	SensorID     uuid.UUID `gorm:"type:uuid;not null" json:"sensor_id"`
	GranaryID    uuid.UUID `gorm:"type:uuid;not null;index" json:"granary_id"`
	ReadingTime  time.Time `gorm:"not null;index:idx_sensor_readings_granary_time,desc;index:idx_sensor_readings_sensor_time,desc" json:"reading_time"`
	Value        float64   `gorm:"type:decimal(12,4);not null" json:"value"`
	IsAbnormal   bool      `gorm:"default:false" json:"is_abnormal"`
}

type FumigationStatus string

const (
	FumigationDraft          FumigationStatus = "draft"
	FumigationPendingApproval FumigationStatus = "pending_approval"
	FumigationApproved       FumigationStatus = "approved"
	FumigationRejected       FumigationStatus = "rejected"
	FumigationInProgress     FumigationStatus = "in_progress"
	FumigationCompleted      FumigationStatus = "completed"
	FumigationCancelled      FumigationStatus = "cancelled"
)

type FumigationPlan struct {
	ID                uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GranaryID         uuid.UUID        `gorm:"type:uuid;not null" json:"granary_id"`
	Granary           *Granary         `gorm:"foreignKey:GranaryID" json:"granary,omitempty"`
	PlanNo            string           `gorm:"size:50;unique;not null" json:"plan_no"`
	PlanTitle         string           `gorm:"size:200;not null" json:"plan_title"`
	CreatorID         uuid.UUID        `gorm:"type:uuid;not null" json:"creator_id"`
	Creator           *User            `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	ChemicalType      string           `gorm:"size:50" json:"chemical_type"`
	ChemicalName      string           `gorm:"size:100" json:"chemical_name"`
	Dosage            float64          `gorm:"type:decimal(10,2)" json:"dosage"`
	DosageUnit        string           `gorm:"size:20" json:"dosage_unit"`
	TargetConcentration float64        `gorm:"type:decimal(10,4)" json:"target_concentration"`
	PlanStartTime     *time.Time       `json:"plan_start_time"`
	PlanEndTime       *time.Time       `json:"plan_end_time"`
	ExpectedSealHours int              `json:"expected_seal_hours"`
	Reason            string           `gorm:"type:text" json:"reason"`
	PeopleCleared     bool             `gorm:"default:false" json:"people_cleared"`
	PeopleClearedTime *time.Time       `json:"people_cleared_time"`
	PeopleClearedBy   *uuid.UUID       `gorm:"type:uuid" json:"people_cleared_by"`
	Status            FumigationStatus `gorm:"size:30;not null;default:'draft'" json:"status"`
	ApproverID        *uuid.UUID       `gorm:"type:uuid" json:"approver_id"`
	Approver          *User            `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	ApprovalRemark    string           `gorm:"type:text" json:"approval_remark"`
	ApprovedAt        *time.Time       `json:"approved_at"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

type FumigationExecution struct {
	ID                  uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PlanID              uuid.UUID          `gorm:"type:uuid;not null" json:"plan_id"`
	Plan                *FumigationPlan    `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	GranaryID           uuid.UUID          `gorm:"type:uuid;not null" json:"granary_id"`
	Granary             *Granary           `gorm:"foreignKey:GranaryID" json:"granary,omitempty"`
	OperatorID          uuid.UUID          `gorm:"type:uuid;not null" json:"operator_id"`
	Operator            *User              `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	ActualStartTime     *time.Time         `json:"actual_start_time"`
	ActualEndTime       *time.Time         `json:"actual_end_time"`
	ChemicalActualDosage float64            `gorm:"type:decimal(10,2)" json:"chemical_actual_dosage"`
	ConcentrationReadings string           `gorm:"type:jsonb" json:"concentration_readings"`
	WeatherDuring       string             `gorm:"size:100" json:"weather_during"`
	Remark              string             `gorm:"type:text" json:"remark"`
	CreatedAt           time.Time          `json:"created_at"`
}

type UnsealType string

const (
	UnsealTypeVentilation UnsealType = "ventilation"
	UnsealTypeUnseal      UnsealType = "unseal"
)

type UnsealRecord struct {
	ID                uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GranaryID         uuid.UUID   `gorm:"type:uuid;not null" json:"granary_id"`
	Granary           *Granary    `gorm:"foreignKey:GranaryID" json:"granary,omitempty"`
	FumigationPlanID  *uuid.UUID  `gorm:"type:uuid" json:"fumigation_plan_id"`
	RecorderID        uuid.UUID   `gorm:"type:uuid;not null" json:"recorder_id"`
	Recorder          *User       `gorm:"foreignKey:RecorderID" json:"recorder,omitempty"`
	UnsealType        UnsealType  `gorm:"size:20;not null" json:"unseal_type"`
	StartTime         *time.Time  `json:"start_time"`
	EndTime           *time.Time  `json:"end_time"`
	WeatherCondition  string      `gorm:"size:100" json:"weather_condition"`
	IsSafe            bool        `gorm:"default:false" json:"is_safe"`
	FinalGasReadings  string      `gorm:"type:jsonb" json:"final_gas_readings"`
	Remark            string      `gorm:"type:text" json:"remark"`
	CreatedAt         time.Time   `json:"created_at"`
}

type GasDetectionRecord struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GranaryID       uuid.UUID       `gorm:"type:uuid;not null" json:"granary_id"`
	Granary         *Granary        `gorm:"foreignKey:GranaryID" json:"granary,omitempty"`
	UnsealID        *uuid.UUID      `gorm:"type:uuid" json:"unseal_id"`
	Unseal          *UnsealRecord   `gorm:"foreignKey:UnsealID" json:"unseal,omitempty"`
	DetectorID      uuid.UUID       `gorm:"type:uuid;not null" json:"detector_id"`
	Detector        *User           `gorm:"foreignKey:DetectorID" json:"detector,omitempty"`
	DetectionTime   time.Time       `gorm:"not null" json:"detection_time"`
	GasType         string          `gorm:"size:30;not null" json:"gas_type"`
	Concentration   float64         `gorm:"type:decimal(10,4);not null" json:"concentration"`
	SafeLimit       float64         `gorm:"type:decimal(10,4);not null" json:"safe_limit"`
	IsSafe          bool            `gorm:"default:false" json:"is_safe"`
	DetectionPoints string          `gorm:"type:jsonb" json:"detection_points"`
	Remark          string          `gorm:"type:text" json:"remark"`
	CreatedAt       time.Time       `json:"created_at"`
}

type SuggestionPriority string

const (
	PriorityLow    SuggestionPriority = "low"
	PriorityNormal SuggestionPriority = "normal"
	PriorityHigh   SuggestionPriority = "high"
	PriorityUrgent SuggestionPriority = "urgent"
)

type SuggestionStatus string

const (
	SuggestionPending    SuggestionStatus = "pending"
	SuggestionProcessing SuggestionStatus = "processing"
	SuggestionCompleted  SuggestionStatus = "completed"
	SuggestionIgnored    SuggestionStatus = "ignored"
)

type GrainTurnoverSuggestion struct {
	ID               uuid.UUID              `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	GranaryID        uuid.UUID              `gorm:"type:uuid;not null" json:"granary_id"`
	Granary          *Granary               `gorm:"foreignKey:GranaryID" json:"granary,omitempty"`
	SourceRecordID   *uuid.UUID             `gorm:"type:uuid" json:"source_record_id"`
	SourceRecord     *GrainConditionRecord  `gorm:"foreignKey:SourceRecordID" json:"source_record,omitempty"`
	SuggestionNo     string                 `gorm:"size:50;unique" json:"suggestion_no"`
	AbnormalAreaDesc string                 `gorm:"type:text" json:"abnormal_area_desc"`
	TemperatureAnomaly string               `gorm:"type:jsonb" json:"temperature_anomaly"`
	SuggestionContent string                 `gorm:"type:text;not null" json:"suggestion_content"`
	Priority         SuggestionPriority     `gorm:"size:20;default:'normal'" json:"priority"`
	Status           SuggestionStatus       `gorm:"size:20;default:'pending'" json:"status"`
	HandlerID        *uuid.UUID             `gorm:"type:uuid" json:"handler_id"`
	Handler          *User                  `gorm:"foreignKey:HandlerID" json:"handler,omitempty"`
	HandledAt        *time.Time             `json:"handled_at"`
	HandleRemark     string                 `gorm:"type:text" json:"handle_remark"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type OperationLog struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid" json:"user_id"`
	Action    string    `gorm:"size:100;not null" json:"action"`
	Module    string    `gorm:"size:50" json:"module"`
	TargetID  uuid.UUID `gorm:"type:uuid" json:"target_id"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IPAddress string    `gorm:"size:50" json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

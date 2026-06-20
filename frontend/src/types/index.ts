export type UserRole = 'admin' | 'keeper' | 'safety_officer' | 'duty_officer'

export interface User {
  id: string
  username: string
  full_name: string
  role: UserRole
  phone?: string
  created_at: string
  updated_at: string
}

export type GranaryStatus = 'normal' | 'fumigating' | 'ventilating' | 'sealed' | 'abnormal'

export interface Granary {
  id: string
  code: string
  name: string
  location?: string
  capacity?: number
  grain_type?: string
  grain_variety?: string
  grain_weight?: number
  status: GranaryStatus
  keeper_id?: string
  keeper?: User
  remark?: string
  sensors?: Sensor[]
  created_at: string
  updated_at: string
}

export type SensorType = 'temperature' | 'humidity' | 'gas_ph3' | 'gas_h2s' | 'co2' | 'o2'

export interface Sensor {
  id: string
  granary_id: string
  code: string
  type: SensorType
  location_desc?: string
  position_x?: number
  position_y?: number
  position_z?: number
  unit?: string
  is_active: boolean
  created_at: string
}

export interface SensorReading {
  id: number
  sensor_id: string
  granary_id: string
  reading_time: string
  value: number
  is_abnormal: boolean
  sensor_code?: string
  sensor_type?: string
  location_desc?: string
  unit?: string
}

export interface GrainConditionRecord {
  id: string
  granary_id: string
  granary?: Granary
  recorder_id: string
  recorder?: User
  record_time: string
  avg_temperature?: number
  max_temperature?: number
  min_temperature?: number
  avg_humidity?: number
  grain_level?: number
  pest_found: boolean
  mold_found: boolean
  abnormal_areas?: string
  weather_condition?: string
  remark?: string
  created_at: string
}

export type FumigationStatus = 'draft' | 'pending_approval' | 'approved' | 'rejected' | 'in_progress' | 'completed' | 'cancelled'

export interface FumigationPlan {
  id: string
  granary_id: string
  granary?: Granary
  plan_no: string
  plan_title: string
  creator_id: string
  creator?: User
  chemical_type?: string
  chemical_name?: string
  dosage?: number
  dosage_unit?: string
  target_concentration?: number
  plan_start_time?: string
  plan_end_time?: string
  expected_seal_hours?: number
  reason?: string
  people_cleared: boolean
  people_cleared_time?: string
  people_cleared_by?: string
  status: FumigationStatus
  approver_id?: string
  approver?: User
  approval_remark?: string
  approved_at?: string
  created_at: string
  updated_at: string
}

export interface FumigationExecution {
  id: string
  plan_id: string
  plan?: FumigationPlan
  granary_id: string
  operator_id: string
  actual_start_time?: string
  actual_end_time?: string
  chemical_actual_dosage?: number
  concentration_readings?: string
  weather_during?: string
  remark?: string
  created_at: string
}

export type UnsealType = 'ventilation' | 'unseal'

export interface UnsealRecord {
  id: string
  granary_id: string
  fumigation_plan_id?: string
  recorder_id: string
  recorder?: User
  granary?: Granary
  unseal_type: UnsealType
  start_time?: string
  end_time?: string
  weather_condition?: string
  is_safe: boolean
  final_gas_readings?: string
  remark?: string
  created_at: string
}

export interface GasDetectionViolation {
  gas_type: string
  actual: number
  limit: number
}

export interface GasDetectionRecord {
  id: string
  granary_id: string
  granary?: Granary
  unseal_id?: string
  detector_id: string
  detector?: User
  detection_time: string
  gas_type: string
  concentration: number
  safe_limit: number
  is_safe: boolean
  violations?: GasDetectionViolation[]
  detection_points?: string
  readings?: Record<string, number>
  remark?: string
  created_at: string
}

export type SuggestionPriority = 'low' | 'normal' | 'medium' | 'high' | 'urgent'
export type SuggestionStatus = 'pending' | 'in_progress' | 'processing' | 'completed' | 'resolved' | 'ignored'

export interface AnomalyZone {
  location?: string
  temperature?: number
  description?: string
  zone_type?: string
  [key: string]: any
}

export interface GrainTurnoverSuggestion {
  id: string
  granary_id: string
  granary?: Granary
  source_record_id?: string
  condition_id?: string
  suggestion_no?: string
  abnormal_area_desc?: string
  temperature_anomaly?: string
  trigger_reasons?: string[]
  max_temp?: number
  min_temp?: number
  temp_diff?: number
  avg_humidity?: number
  anomaly_zones?: AnomalyZone[]
  suggestion_content: string
  suggestion_text?: string
  priority: SuggestionPriority
  status: SuggestionStatus
  handler_id?: string
  handler?: User
  handled_at?: string
  handle_remark?: string
  process_note?: string
  processed_by?: string
  processed_at?: string
  created_at: string
  updated_at: string
}

export type TurnoverSuggestion = GrainTurnoverSuggestion

export interface DashboardStats {
  total_granaries: number
  normal_granaries: number
  fumigating_granaries: number
  ventilating_granaries: number
  sealed_granaries: number
  abnormal_granaries: number
  pending_suggestions: number
  processing_suggestions: number
  urgent_suggestions: number
  high_suggestions: number
  pending_fumigation: number
  in_progress_fumigation: number
  today_records: number
  today_avg_temp: number
  today_max_temp: number
  abnormal_temp_count: number
}

export const GranaryStatusLabels: Record<string, string> = {
  normal: '正常',
  fumigating: '熏蒸中',
  ventilating: '通风中',
  sealed: '已密封',
  abnormal: '异常'
}

export const GranaryStatusColors: Record<string, string> = {
  normal: 'success',
  fumigating: 'danger',
  ventilating: 'warning',
  sealed: 'info',
  abnormal: 'danger'
}

export const FumigationStatusLabels: Record<string, string> = {
  draft: '草稿',
  pending_approval: '待审批',
  approved: '已批准',
  rejected: '已驳回',
  in_progress: '执行中',
  completed: '已完成',
  cancelled: '已取消'
}

export const FumigationStatusColors: Record<string, string> = {
  draft: 'info',
  pending_approval: 'warning',
  approved: 'success',
  rejected: 'danger',
  in_progress: 'danger',
  completed: 'success',
  cancelled: 'info'
}

export const PriorityLabels: Record<string, string> = {
  low: '低',
  normal: '普通',
  medium: '中',
  high: '高',
  urgent: '紧急'
}

export const PriorityColors: Record<string, string> = {
  low: 'info',
  normal: 'success',
  medium: 'primary',
  high: 'warning',
  urgent: 'danger'
}

export const SuggestionStatusLabels: Record<string, string> = {
  pending: '待处理',
  in_progress: '处理中',
  processing: '处理中',
  completed: '已完成',
  resolved: '已解决',
  ignored: '已忽略'
}

export const SuggestionStatusColors: Record<string, string> = {
  pending: 'warning',
  in_progress: 'primary',
  processing: 'primary',
  completed: 'success',
  resolved: 'success',
  ignored: 'info'
}

export const RoleLabels: Record<string, string> = {
  admin: '系统管理员',
  keeper: '保管员',
  safety_officer: '安全员',
  duty_officer: '值班员'
}

export const SensorTypeLabels: Record<string, string> = {
  temperature: '温度',
  humidity: '湿度',
  gas_ph3: 'PH₃浓度',
  gas_h2s: 'H₂S浓度',
  co2: 'CO₂浓度',
  o2: 'O₂浓度'
}

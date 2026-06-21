import request from '@/utils/request'
import type {
  User,
  Granary,
  Sensor,
  SensorReading,
  GrainConditionRecord,
  FumigationPlan,
  FumigationExecution,
  UnsealRecord,
  GasDetectionRecord,
  GrainTurnoverSuggestion,
  DashboardStats
} from '@/types'

export const authApi = {
  login: (data: { username: string; password: string }) =>
    request.post<any, { token: string; user: User }>('/auth/login', data),
  getCurrentUser: () => request.get<any, User>('/auth/me')
}

export const dashboardApi = {
  getStats: () => request.get<any, DashboardStats>('/dashboard/stats')
}

export const granaryApi = {
  list: (params?: { status?: string; keyword?: string }) =>
    request.get<any, Granary[]>('/granaries', { params }),
  get: (id: string) => request.get<any, Granary>(`/granaries/${id}`),
  create: (data: any) => request.post<any, Granary>('/granaries', data),
  update: (id: string, data: any) => request.put<any, Granary>(`/granaries/${id}`, data),
  delete: (id: string) => request.delete<any, void>(`/granaries/${id}`),
  listKeepers: () => request.get<any, User[]>('/granaries/keepers'),
  getSensors: (id: string) => request.get<any, Sensor[]>(`/granaries/${id}/sensors`),
  addSensor: (id: string, data: any) => request.post<any, Sensor>(`/granaries/${id}/sensors`, data),
  addSensorReading: (id: string, sensorId: string, data: { value: number; is_abnormal?: boolean }) =>
    request.post<any, SensorReading>(`/granaries/${id}/sensors/${sensorId}/readings`, data),
  getReadings: (id: string, params?: { type?: string; start_time?: string; end_time?: string; limit?: number }) =>
    request.get<any, SensorReading[]>(`/granaries/${id}/readings`, { params })
}

export const grainConditionApi = {
  list: (params?: { granary_id?: string; start_date?: string; end_date?: string }) =>
    request.get<any, GrainConditionRecord[]>('/grain-conditions', { params }),
  get: (id: string) => request.get<any, GrainConditionRecord>(`/grain-conditions/${id}`),
  create: (data: any) => request.post<any, GrainConditionRecord>('/grain-conditions', data)
}

export const fumigationApi = {
  listPlans: (params?: { status?: string; granary_id?: string }) =>
    request.get<any, FumigationPlan[]>('/fumigation/plans', { params }),
  getPlan: (id: string) => request.get<any, FumigationPlan>(`/fumigation/plans/${id}`),
  createPlan: (data: any) => request.post<any, FumigationPlan>('/fumigation/plans', data),
  submitPlan: (id: string) => request.post<any, FumigationPlan>(`/fumigation/plans/${id}/submit`),
  approvePlan: (id: string, data: { approved: boolean; approval_remark?: string }) =>
    request.post<any, FumigationPlan>(`/fumigation/plans/${id}/approve`, data),
  clearPeople: (id: string, data: { cleared: boolean }) =>
    request.post<any, FumigationPlan>(`/fumigation/plans/${id}/clear-people`, data),
  startExecution: (id: string) =>
    request.post<any, { plan: FumigationPlan; execution: FumigationExecution }>(`/fumigation/plans/${id}/start`),
  completeExecution: (id: string, data: any) =>
    request.post<any, FumigationPlan>(`/fumigation/plans/${id}/complete`, data),
  safetyConfirm: (id: string, data: { confirmed: boolean; remark?: string }) =>
    request.post<any, FumigationPlan>(`/fumigation/plans/${id}/safety-confirm`, data),
  listExecutions: (params?: { plan_id?: string; granary_id?: string }) =>
    request.get<any, FumigationExecution[]>('/fumigation/executions', { params })
}

export const unsealApi = {
  list: (params?: { granary_id?: string; type?: string }) =>
    request.get<any, UnsealRecord[]>('/unseal', { params }),
  get: (id: string) => request.get<any, UnsealRecord>(`/unseal/${id}`),
  create: (data: any) => request.post<any, UnsealRecord>('/unseal', data),
  complete: (id: string, data: any) => request.post<any, UnsealRecord>(`/unseal/${id}/complete`, data),
  listGasDetections: (params?: { granary_id?: string; unseal_id?: string; gas_type?: string; date?: string }) =>
    request.get<any, GasDetectionRecord[]>('/unseal/gas-detections', { params }),
  addGasDetection: (granaryId: string, data: any) =>
    request.post<any, GasDetectionRecord>(`/unseal/${granaryId}/gas-detections`, data)
}

export const suggestionApi = {
  list: (params?: { status?: string; priority?: string; granary_id?: string }) =>
    request.get<any, GrainTurnoverSuggestion[]>('/turnover-suggestions', { params }),
  get: (id: string) => request.get<any, GrainTurnoverSuggestion>(`/turnover-suggestions/${id}`),
  handle: (id: string, data: { status: string; handle_remark?: string }) =>
    request.post<any, GrainTurnoverSuggestion>(`/turnover-suggestions/${id}/handle`, data)
}

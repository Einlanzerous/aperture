// ─── Shared enums ────────────────────────────────────────────────────────────

export type ServiceStatus = 'healthy' | 'degraded' | 'unhealthy' | 'unknown'
export type WidgetSize    = 's' | 'm' | 'l'
export type ServiceType   = 'http' | 'docker'

// ─── API response shapes ─────────────────────────────────────────────────────

export interface ServiceStatusData {
  name:         string
  type:         ServiceType
  url?:         string
  container?:   string
  status:       ServiceStatus
  statusCode?:  number
  responseTime?: number   // ms
  message?:     string
  checkedAt:    string    // ISO-8601
  icon?:        string
  category?:    string
  href?:        string
  size?:        WidgetSize
  detailDefault?: boolean
}

export interface ServicesResponse {
  services:  ServiceStatusData[]
  updatedAt: string
}

export interface CPUStats {
  percent: number
  cores:   number
}

export interface MemoryStats {
  total:   number  // bytes
  used:    number
  free:    number
  percent: number  // 0–100
}

export interface LoadStats {
  load1:  number
  load5:  number
  load15: number
}

export interface SystemResources {
  cpu:       CPUStats
  memory:    MemoryStats
  load:      LoadStats
  updatedAt: string
}

export interface OllamaModelDetails {
  format:             string
  family:             string
  parameter_size:     string
  quantization_level: string
}

export interface OllamaModel {
  name:        string
  size:        number  // bytes
  modified_at: string
  digest?:     string
  details?:    OllamaModelDetails
}

export interface OllamaModelsResponse {
  models: OllamaModel[]
}

export interface DashboardConfig {
  title:          string
  checkInterval:  number
  ollamaEnabled:  boolean
  systemEnabled:  boolean
  actionsEnabled: boolean
  storageEnabled: boolean
}

// ─── History types ──────────────────────────────────────────────────────────

export interface CheckRecord {
  serviceName:  string
  serviceType:  string
  status:       string
  statusCode?:  number
  responseTime?: number  // ms
  message?:     string
  checkedAt:    string   // ISO-8601
}

export interface DailySummary {
  serviceName:     string
  date:            string  // ISO-8601
  totalChecks:     number
  healthyChecks:   number
  unhealthyChecks: number
  degradedChecks:  number
  avgResponseMs:   number
  minResponseMs:   number
  maxResponseMs:   number
  uptimePct:       number
}

export interface HistoryResponse {
  service: string
  period:  string
  records: CheckRecord[]
}

export interface UptimeResponse {
  service:   string
  days:      number
  summaries: DailySummary[]
}

// ─── Action types ───────────────────────────────────────────────────────────

export type ActionStatus = 'idle' | 'waiting' | 'starting' | 'running' | 'success' | 'error' | 'stopped'

export interface ActionState {
  name:         string
  projectId:    number
  templateId:   number
  category?:    string
  icon?:        string
  size?:        WidgetSize
  taskId?:      number
  taskStatus:   ActionStatus
  triggeredAt?: string
}

export interface ActionsResponse {
  actions: ActionState[]
}

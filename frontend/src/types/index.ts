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
  title:         string
  checkInterval: number
  ollamaEnabled: boolean
  systemEnabled: boolean
}

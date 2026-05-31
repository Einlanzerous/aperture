// ─── Shared enums ────────────────────────────────────────────────────────────

export type ServiceStatus = 'healthy' | 'degraded' | 'unhealthy' | 'unknown'
export type ServiceType   = 'http' | 'docker'

// ConfigSize is the per-service/-action sizing knob in the backend YAML config.
// It maps onto the slot-based WidgetSize used by the grid (s→small, m→large,
// l→xl) — see SIZE_CLASS in DraggableGrid.vue.
export type ConfigSize = 's' | 'm' | 'l'

// WidgetSize is the slot-based footprint a widget occupies in the grid, where
// one slot is a single tiny tile and the grid uses fixed-height rows so every
// widget is an exact slot multiple in both dimensions (see DraggableGrid):
//   tiny   = 1×1   CPU/GPU/Memory + status-only tiles
//   small  = 1×2   standard service tiles, Load
//   large  = 2×2   service size "m"
//   xl     = 3×2   service size "l" (full three-column width)
//   ollama = 3×4   the Ollama widget (full width, four rows tall)
export type WidgetSize = 'tiny' | 'small' | 'large' | 'xl' | 'ollama'

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
  size?:        ConfigSize
  detailDefault?: boolean
  statusOnly?:  boolean
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

export interface GpuStats {
  available: boolean
  vendor:    'amd' | 'nvidia' | ''
  name:      string
  percent:   number
  vramUsed:  number  // bytes
  vramTotal: number  // bytes
  tempC:     number | null  // null when temperature unreadable
}

export interface SystemHistory {
  cpu:    number[]  // cpu.percent samples, oldest->newest
  memory: number[]  // memory.percent samples
  load1:  number[]  // load1 samples
  gpu:    number[]  // gpu.percent samples
}

export interface SystemResources {
  cpu:       CPUStats | null     // null when cpu disabled in config
  memory:    MemoryStats | null  // null when memory disabled in config
  load:      LoadStats | null    // null when load disabled in config
  gpu:       GpuStats | null     // null when gpu disabled in config
  history:   SystemHistory | null  // present only when ?history=n>0
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

export interface SystemMetricFlags {
  cpu:    boolean
  memory: boolean
  load:   boolean
  gpu:    boolean
}

export interface DashboardConfig {
  title:          string
  checkInterval:  number
  ollamaEnabled:  boolean
  systemEnabled:  boolean
  system:         SystemMetricFlags  // per-metric enable flags
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
  size?:        ConfigSize
  taskId?:      number
  taskStatus:   ActionStatus
  triggeredAt?: string
}

export interface ActionsResponse {
  actions: ActionState[]
}

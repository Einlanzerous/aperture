export const API = {
  config:        '/api/config',
  services:      '/api/services',
  resources:     '/api/system/resources',
  ollama:        '/api/ollama/models',
  actions:       '/api/actions',
  actionTrigger: (name: string) => `/api/actions/${encodeURIComponent(name)}/trigger`,
  actionStatus:  (name: string) => `/api/actions/${encodeURIComponent(name)}/status`,
} as const

export async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(path)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export async function apiPost<T>(path: string): Promise<T> {
  const res = await fetch(path, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export const API = {
  config:    '/api/config',
  services:  '/api/services',
  resources: '/api/system/resources',
  ollama:    '/api/ollama/models',
} as const

export async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(path)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

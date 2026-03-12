const BYTES_PER_MB = 1024 ** 2
const BYTES_PER_GB = 1024 ** 3

export function fmtBytes(bytes: number): string {
  const gb = bytes / BYTES_PER_GB
  return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / BYTES_PER_MB).toFixed(0)} MB`
}

export function getErrorMessage(e: unknown): string {
  return e instanceof Error ? e.message : 'Unknown error'
}

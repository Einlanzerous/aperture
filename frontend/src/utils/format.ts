const BYTES_PER_MB = 1024 ** 2
const BYTES_PER_GB = 1024 ** 3

export function fmtBytes(bytes: number): string {
  const gb = bytes / BYTES_PER_GB
  return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / BYTES_PER_MB).toFixed(0)} MB`
}

// High-level GB, always one decimal — for the compact resource tiles where a
// rough figure ("4.4 GB") is all that's wanted, regardless of magnitude.
export function fmtGB(bytes: number): string {
  return `${(bytes / BYTES_PER_GB).toFixed(1)} GB`
}

export function getErrorMessage(e: unknown): string {
  return e instanceof Error ? e.message : 'Unknown error'
}

import { computed, ref, watch, type Ref } from 'vue'
import type { WidgetSize } from '@/types'

// v2: APTR-47 reworked widget ids (tiny tiles now pack into TinyStack wrappers)
// and the size vocabulary (slot-based tiny/small/large/xl), so stale v1 entries
// are dropped rather than migrated.
const STORAGE_VERSION = 2
const STORAGE_PREFIX  = 'aperture:layout:'

interface LayoutEntry {
  order?:        number
  sizeOverride?: WidgetSize
}

interface StoredLayout {
  v:       number
  entries: Record<string, LayoutEntry>
}

interface LayoutAware {
  id:   string
  size: WidgetSize
}

function storageKey(title: string): string {
  return `${STORAGE_PREFIX}${title}`
}

function readStored(title: string): Record<string, LayoutEntry> {
  try {
    const raw = localStorage.getItem(storageKey(title))
    if (!raw) return {}
    const parsed = JSON.parse(raw) as Partial<StoredLayout>
    if (parsed.v !== STORAGE_VERSION || !parsed.entries) return {}
    return parsed.entries
  } catch {
    return {}
  }
}

function writeStored(title: string, entries: Record<string, LayoutEntry>): void {
  const payload: StoredLayout = { v: STORAGE_VERSION, entries }
  try {
    localStorage.setItem(storageKey(title), JSON.stringify(payload))
  } catch {
    /* quota exceeded or storage unavailable — drop silently */
  }
}

export function useLayout(title: Ref<string>) {
  const entries = ref<Record<string, LayoutEntry>>(readStored(title.value))

  watch(title, (next) => {
    entries.value = readStored(next)
  })

  function persist(): void {
    writeStored(title.value, entries.value)
  }

  function applyLayout<T extends LayoutAware>(widgets: T[]): T[] {
    const stored = entries.value
    const ordered:   T[] = []
    const unordered: T[] = []

    for (const w of widgets) {
      if (stored[w.id]?.order != null) ordered.push(w)
      else unordered.push(w)
    }

    ordered.sort((a, b) => stored[a.id].order! - stored[b.id].order!)

    return [...ordered, ...unordered].map((w) => {
      const override = stored[w.id]?.sizeOverride
      return override ? { ...w, size: override } : w
    })
  }

  function setOrder(orderedIds: string[]): void {
    const next: Record<string, LayoutEntry> = {}
    orderedIds.forEach((id, order) => {
      const prev = entries.value[id]
      next[id] = prev ? { ...prev, order } : { order }
    })
    entries.value = next
    persist()
  }

  function setSizeOverride(id: string, size: WidgetSize | undefined): void {
    const next = { ...entries.value }
    const prev = next[id] ?? {}
    if (size) {
      next[id] = { ...prev, sizeOverride: size }
    } else {
      const { sizeOverride: _omit, ...rest } = prev
      next[id] = rest
    }
    entries.value = next
    persist()
  }

  function reset(): void {
    entries.value = {}
    try {
      localStorage.removeItem(storageKey(title.value))
    } catch {
      /* no-op */
    }
  }

  const isCustomized = computed(() => Object.keys(entries.value).length > 0)

  return { applyLayout, setOrder, setSizeOverride, reset, isCustomized }
}

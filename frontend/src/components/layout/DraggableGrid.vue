<script
  setup
  lang="ts"
  generic="T extends { id: string; size?: WidgetSize }"
>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { WidgetSize } from '@/types'

const props = defineProps<{
  items: readonly T[]
}>()

const emit = defineEmits<{
  reorder: [orderedIds: string[]]
}>()

const SIZE_CLASS: Record<WidgetSize, string> = {
  s: 'col-span-1',
  m: 'col-span-1 md:col-span-2',
  l: 'col-span-1 md:col-span-3',
}

// ─── Drag state ─────────────────────────────────────────────────────────────
// The source tile stays at its original DOM position throughout the drag —
// only its opacity changes. The projected landing slot is rendered as a
// separate "extra cell" in the grid, which causes neighbouring tiles to
// reflow naturally around it. Keeping the source DOM intact preserves the
// browser's native drag image (snapshot of the source follows the cursor)
// AND keeps its `@dragend` listener bound, so we don't get stuck state when
// the drop fires outside the window.

const dragId         = ref<string | null>(null)
const projectedIndex = ref<number | null>(null)

const sourceIndex = computed<number>(() =>
  dragId.value ? props.items.findIndex((i) => i.id === dragId.value) : -1,
)

const draggedSize = computed<WidgetSize>(() => {
  if (!dragId.value) return 's'
  return props.items.find((i) => i.id === dragId.value)?.size ?? 's'
})

const isActive = computed(
  () => dragId.value !== null && projectedIndex.value !== null,
)

// A projection that lands the source at its current slot (its own index, or
// the slot immediately after — same thing visually) is a no-op drop. We hide
// the placeholder in that case so the user sees a clean "return to start"
// state rather than a redundant outline next to the dimmed source.
const isNoopProjection = computed(() => {
  if (projectedIndex.value === null || sourceIndex.value < 0) return false
  return projectedIndex.value === sourceIndex.value
      || projectedIndex.value === sourceIndex.value + 1
})

function reset(): void {
  dragId.value         = null
  projectedIndex.value = null
}

function commitDrop(): void {
  const sourceId = dragId.value
  const target   = projectedIndex.value
  const fromIdx  = sourceIndex.value
  reset()
  if (!sourceId || target === null || fromIdx < 0) return

  // target is an index in the N+1 cell layout (source + placeholder both in).
  // Translate to the without-source insert index.
  const insertIdx = target > fromIdx ? target - 1 : target
  if (insertIdx === fromIdx) return  // no-op drop

  const ids = props.items.map((i) => i.id)
  ids.splice(fromIdx, 1)
  ids.splice(insertIdx, 0, sourceId)
  emit('reorder', ids)
}

// ─── Projection ─────────────────────────────────────────────────────────────
// Translate the cursor's position over a hovered tile into a slot index in
// the items array (range 0..items.length). Size-l sources snap to row
// boundaries; smaller sources use before/after of the hovered tile.

function projectFromPointer(
  hoveredId: string,
  hoveredEl: HTMLElement,
  clientX: number,
  clientY: number,
): number | null {
  if (!dragId.value || hoveredId === dragId.value) return null
  const idx = props.items.findIndex((i) => i.id === hoveredId)
  if (idx < 0) return null

  const rect = hoveredEl.getBoundingClientRect()

  if (draggedSize.value === 'l') {
    const bounds = rowBoundsOf(hoveredEl)
    if (!bounds) return idx
    const before = clientY < rect.top + rect.height / 2
    return before ? bounds.startIdx : bounds.endIdx + 1
  }

  const before = clientX < rect.left + rect.width / 2
  return before ? idx : idx + 1
}

function rowBoundsOf(
  hoveredEl: HTMLElement,
): { startIdx: number; endIdx: number } | null {
  const grid = hoveredEl.parentElement
  if (!grid) return null
  const hoveredTop = Math.round(hoveredEl.getBoundingClientRect().top)
  const sameRowIdx: number[] = []
  for (const el of grid.querySelectorAll<HTMLElement>('[data-grid-item-id]')) {
    if (Math.round(el.getBoundingClientRect().top) !== hoveredTop) continue
    const id = el.dataset.gridItemId
    if (!id || id === dragId.value) continue
    const i = props.items.findIndex((it) => it.id === id)
    if (i >= 0) sameRowIdx.push(i)
  }
  if (!sameRowIdx.length) return null
  sameRowIdx.sort((a, b) => a - b)
  return { startIdx: sameRowIdx[0], endIdx: sameRowIdx[sameRowIdx.length - 1] }
}

// ─── Native HTML5 drag-and-drop (mouse) ─────────────────────────────────────

function onDragStart(e: DragEvent, id: string): void {
  dragId.value = id
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', id)
  }
}

function onDragOver(e: DragEvent, id: string): void {
  if (!dragId.value) return
  // Always preventDefault so the cursor reads "drop allowed" — including over
  // the source itself. Without this, dragging back to the starting slot shows
  // a no-drop cursor and the user can't release there.
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  if (id === dragId.value) {
    projectedIndex.value = sourceIndex.value
    return
  }
  projectedIndex.value = projectFromPointer(
    id,
    e.currentTarget as HTMLElement,
    e.clientX,
    e.clientY,
  )
}

function onDrop(e: DragEvent): void {
  e.preventDefault()
  commitDrop()
}

// The placeholder cell is `pointer-events-none`, so when the user releases
// over it the drop event lands on the grid container rather than a tile.
// We catch it here so the drop still goes through commitDrop. Per-tile
// drops bubble up too — the second commitDrop is a no-op because reset()
// already cleared dragId.
function onGridDragOver(e: DragEvent): void {
  if (!dragId.value) return
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
}

function onGridDrop(e: DragEvent): void {
  if (!dragId.value) return
  e.preventDefault()
  commitDrop()
}

// Safety net: if a drop happens outside any tile (window edge, blank space),
// dragend on the source still fires — but only if the source still owns the
// listener. We bind globally too so a stale state can never linger.
onMounted(() => {
  document.addEventListener('dragend', reset)
})
onBeforeUnmount(() => {
  document.removeEventListener('dragend', reset)
})

// ─── Touch fallback via pointer events ──────────────────────────────────────

const LONG_PRESS_MS = 250
let pressTimer: ReturnType<typeof setTimeout> | null = null
let touchActive = false

function clearPressTimer(): void {
  if (pressTimer) {
    clearTimeout(pressTimer)
    pressTimer = null
  }
}

function onPointerDown(e: PointerEvent, id: string): void {
  if (e.pointerType !== 'touch') return
  clearPressTimer()
  pressTimer = setTimeout(() => {
    pressTimer = null
    dragId.value = id
    touchActive  = true
    document.addEventListener('pointermove', onTouchMove, { passive: false })
    document.addEventListener('pointerup', onTouchEnd)
    document.addEventListener('pointercancel', onTouchCancel)
  }, LONG_PRESS_MS)
}

function onPointerCancelBeforePress(): void {
  if (!touchActive) clearPressTimer()
}

function onTouchMove(e: PointerEvent): void {
  if (!touchActive) return
  e.preventDefault()
  const el = document
    .elementFromPoint(e.clientX, e.clientY)
    ?.closest<HTMLElement>('[data-grid-item-id]')
  if (!el) return
  const id = el.dataset.gridItemId
  if (!id || id === dragId.value) return
  projectedIndex.value = projectFromPointer(id, el, e.clientX, e.clientY)
}

function onTouchEnd(): void {
  if (touchActive && projectedIndex.value !== null) commitDrop()
  else reset()
  endTouch()
}

function onTouchCancel(): void {
  reset()
  endTouch()
}

function endTouch(): void {
  touchActive = false
  document.removeEventListener('pointermove', onTouchMove)
  document.removeEventListener('pointerup', onTouchEnd)
  document.removeEventListener('pointercancel', onTouchCancel)
}
</script>

<template>
  <div
    class="grid grid-cols-1 gap-4 md:grid-cols-3"
    @dragover="onGridDragOver"
    @drop="onGridDrop"
  >
    <slot name="before" />

    <template v-for="(item, index) in items" :key="item.id">
      <!-- Dashed placeholder cell at the projected landing slot. -->
      <div
        v-if="isActive && projectedIndex === index && !isNoopProjection"
        :class="[
          SIZE_CLASS[draggedSize],
          'pointer-events-none min-h-24 rounded-xl border-2 border-dashed border-gray-600 bg-gray-900/40',
        ]"
        aria-hidden="true"
      />

      <div
        :data-grid-item-id="item.id"
        :class="[
          SIZE_CLASS[item.size ?? 's'],
          'group relative transition-opacity',
          dragId === item.id ? 'opacity-30' : '',
        ]"
        :style="touchActive ? { touchAction: 'none' } : undefined"
        draggable="true"
        @dragstart="onDragStart($event, item.id)"
        @dragover="onDragOver($event, item.id)"
        @drop="onDrop($event)"
        @dragend="reset"
        @pointerdown="onPointerDown($event, item.id)"
        @pointerup="onPointerCancelBeforePress"
        @pointermove="onPointerCancelBeforePress"
        @pointercancel="onPointerCancelBeforePress"
      >
        <span
          v-if="dragId !== item.id"
          class="pointer-events-none absolute left-1 top-1 z-10 text-gray-400
                 opacity-0 transition-opacity group-hover:opacity-100"
          aria-hidden="true"
          title="Drag to reorder"
        >
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
            <circle cx="9"  cy="6"  r="1.5" />
            <circle cx="15" cy="6"  r="1.5" />
            <circle cx="9"  cy="12" r="1.5" />
            <circle cx="15" cy="12" r="1.5" />
            <circle cx="9"  cy="18" r="1.5" />
            <circle cx="15" cy="18" r="1.5" />
          </svg>
        </span>

        <slot :item="item" :index="index" />
      </div>

      <slot name="after-item" :item="item" :index="index" />
    </template>

    <!-- Trailing placeholder when dropping after the last tile. -->
    <div
      v-if="isActive && projectedIndex === items.length && !isNoopProjection"
      :class="[
        SIZE_CLASS[draggedSize],
        'pointer-events-none min-h-24 rounded-xl border-2 border-dashed border-gray-600 bg-gray-900/40',
      ]"
      aria-hidden="true"
    />
  </div>
</template>

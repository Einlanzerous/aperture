# Aperture — Frontend

Vue 3 + TypeScript single-page application. Polls the backend API and renders a responsive grid dashboard.

## Requirements

- Node 20 or later
- npm 10 or later

## Development

```bash
npm install
npm run dev        # starts Vite dev server on http://localhost:4000
```

The Vite dev server proxies all `/api/*` requests to `http://localhost:8888`, so the backend must be running separately. See [`backend/README.md`](../backend/README.md).

## Building for production

```bash
npm run build      # outputs to dist/
npm run preview    # preview the production build locally
```

## Tech stack

| Tool | Purpose |
|------|---------|
| Vue 3 (Composition API) | Reactive UI |
| TypeScript | Type safety across components and composables |
| Tailwind CSS v3 | Utility-first styling, dark theme |
| Vite | Dev server and bundler |
| Lucide Vue Next | Icon set (available, not yet wired to service icons) |

## Widget system

Widgets are standard Vue components placed inside `DashboardGrid`. Each widget occupies a number of grid columns determined by its `size` prop:

| Size | Columns | Width |
|------|---------|-------|
| `s`  | 1 of 3  | ~33%  |
| `m`  | 2 of 3  | ~66%  |
| `l`  | 3 of 3  | 100%  |

On mobile all widgets collapse to full width regardless of size.

### Adding a new widget

1. Create `src/components/widgets/MyWidget.vue`.
2. Add a composable in `src/composables/` if the widget needs its own polling logic.
3. Import and render the widget in `App.vue` inside `<DashboardGrid>`, wrapping it in a `<div :class="widgetSizeClass('m')">`.

### Existing widgets

| Widget | Source | Data source | Poll interval |
|--------|--------|-------------|---------------|
| `ServiceWidget` | `components/widgets/ServiceWidget.vue` | `/api/services` | 30 s (via `useServices`) |
| `ResourceWidget` | `components/widgets/ResourceWidget.vue` | `/api/system/resources` | 5 s (via `useResources`) |
| `OllamaWidget` | `components/widgets/OllamaWidget.vue` | `/api/ollama/models` | 60 s (self-contained) |

## Project layout

```
frontend/
├── index.html
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
├── nginx.conf               Used by the production Docker image
└── src/
    ├── main.ts
    ├── style.css            Tailwind directives + scrollbar styles
    ├── App.vue              Root component — header, grid, config fetch
    ├── types/index.ts       All shared TypeScript interfaces
    ├── composables/
    │   ├── useServices.ts   Polls /api/services, exposes reactive state
    │   └── useResources.ts  Polls /api/system/resources, exposes reactive state
    └── components/
        ├── layout/
        │   └── DashboardGrid.vue   3-column CSS grid + widgetSizeClass export
        └── widgets/
            ├── ServiceWidget.vue
            ├── ResourceWidget.vue
            └── OllamaWidget.vue
```

## Status colour conventions

| Status | Colour |
|--------|--------|
| `healthy` | Emerald (`#34d399`) — pulses to indicate live data |
| `degraded` | Amber (`#fbbf24`) |
| `unhealthy` | Red (`#f87171`) |
| `unknown` | Gray (`#6b7280`) |

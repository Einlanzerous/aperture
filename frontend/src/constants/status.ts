export const STATUS_COLORS = {
  healthy: {
    dot:   'bg-emerald-400',
    badge: 'bg-emerald-400/10 text-emerald-400 ring-1 ring-emerald-400/20',
    msg:   'text-emerald-400',
    label: 'Healthy',
    pulse: true,
  },
  degraded: {
    dot:   'bg-amber-400',
    badge: 'bg-amber-400/10 text-amber-400 ring-1 ring-amber-400/20',
    msg:   'text-amber-400',
    label: 'Degraded',
    pulse: false,
  },
  unhealthy: {
    dot:   'bg-red-400',
    badge: 'bg-red-400/10 text-red-400 ring-1 ring-red-400/20',
    msg:   'text-red-400',
    label: 'Unhealthy',
    pulse: false,
  },
  unknown: {
    dot:   'bg-gray-500',
    badge: 'bg-gray-500/10 text-gray-400 ring-1 ring-gray-500/20',
    msg:   'text-gray-400',
    label: 'Unknown',
    pulse: false,
  },
} as const

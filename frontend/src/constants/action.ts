export const ACTION_STATUS_COLORS = {
  idle: {
    dot:   'bg-gray-500',
    badge: 'bg-gray-500/10 text-gray-400 ring-1 ring-gray-500/20',
    label: 'Idle',
    pulse: false,
  },
  waiting: {
    dot:   'bg-blue-400',
    badge: 'bg-blue-400/10 text-blue-400 ring-1 ring-blue-400/20',
    label: 'Waiting',
    pulse: true,
  },
  starting: {
    dot:   'bg-blue-400',
    badge: 'bg-blue-400/10 text-blue-400 ring-1 ring-blue-400/20',
    label: 'Starting',
    pulse: true,
  },
  running: {
    dot:   'bg-amber-400',
    badge: 'bg-amber-400/10 text-amber-400 ring-1 ring-amber-400/20',
    label: 'Running',
    pulse: true,
  },
  success: {
    dot:   'bg-emerald-400',
    badge: 'bg-emerald-400/10 text-emerald-400 ring-1 ring-emerald-400/20',
    label: 'Success',
    pulse: false,
  },
  error: {
    dot:   'bg-red-400',
    badge: 'bg-red-400/10 text-red-400 ring-1 ring-red-400/20',
    label: 'Error',
    pulse: false,
  },
  stopped: {
    dot:   'bg-gray-400',
    badge: 'bg-gray-400/10 text-gray-400 ring-1 ring-gray-400/20',
    label: 'Stopped',
    pulse: false,
  },
} as const

export function toLocalDatetime(iso: string) {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formatDuration(mins: number) {
  if (mins <= 0) return '0min'
  if (mins < 60) return `${mins}min`
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0 ? `${h}h${m}min` : `${h}h`
}

export function durationParts(mins: number) {
  if (mins <= 0) return [{ val: '0', unit: 'min' as string }]
  if (mins < 60) return [{ val: `${mins}`, unit: 'min' }]
  const h = Math.floor(mins / 60)
  const m = mins % 60
  if (m === 0) return [{ val: `${h}`, unit: 'h' }]
  return [{ val: `${h}`, unit: 'h' }, { val: `${m}`, unit: 'min' }]
}

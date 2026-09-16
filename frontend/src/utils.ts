export function toLocalDatetime(iso: string) {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formatDuration(mins: number) {
  if (mins <= 0) return '0分钟'
  if (mins < 60) return `${mins}分钟`
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0 ? `${h}小时${m}分钟` : `${h}小时`
}

export function durationParts(mins: number) {
  if (mins <= 0) return [{ val: '0', unit: '分钟' as string }]
  if (mins < 60) return [{ val: `${mins}`, unit: '分钟' }]
  const h = Math.floor(mins / 60)
  const m = mins % 60
  if (m === 0) return [{ val: `${h}`, unit: '小时' }]
  return [{ val: `${h}`, unit: '小时' }, { val: `${m}`, unit: '分钟' }]
}

/* ============================================================
   通用格式化工具
   ============================================================ */

const pad2 = (n: number) => String(n).padStart(2, '0')

function toDate(v: string | Date): Date {
  return v instanceof Date ? v : new Date(v)
}

/** 转为 <input type="datetime-local"> 所需的本地时间 YYYY-MM-DDTHH:mm */
export function toLocalDatetime(iso: string) {
  const d = new Date(iso)
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}T${pad2(d.getHours())}:${pad2(d.getMinutes())}`
}

/** 当前本地时间 YYYY-MM-DDTHH:mm（datetime-local 默认值） */
export function nowLocalDatetime() {
  return toLocalDatetime(new Date().toISOString())
}

/** HH:mm */
export function formatClock(v: string | Date) {
  const d = toDate(v)
  return `${pad2(d.getHours())}:${pad2(d.getMinutes())}`
}

/** 时间段：HH:mm~HH:mm；无结束时间时仅返回 HH:mm */
export function formatTimeRange(startIso: string, endIso?: string | null) {
  const start = formatClock(startIso)
  return endIso ? `${start}~${formatClock(endIso)}` : start
}

/** 列表时间标签：今天 HH:mm / 昨天 HH:mm / M-D HH:mm；withDate=false 时仅 HH:mm */
export function formatDayTime(iso: string, withDate = true) {
  const d = new Date(iso)
  const hhmm = formatClock(d)
  if (!withDate) return hhmm
  const now = new Date()
  if (d.toDateString() === now.toDateString()) return `今天 ${hhmm}`
  const yesterday = new Date(now.getTime() - 86400000)
  if (d.toDateString() === yesterday.toDateString()) return `昨天 ${hhmm}`
  return `${pad2(d.getMonth() + 1)}-${pad2(d.getDate())} ${hhmm}`
}

export const WEEKDAY_SHORT = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
export const WEEKDAY_LONG = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']

/* ============================================================
   时长（统一使用紧凑的 h / m 单位，避免半宽卡片中文单位挤压）
   ============================================================ */

export type DurationPart = { val: string; unit: 'h' | 'm' }

/** 大数字 + 小单位渲染用：12h30m -> [{12,h},{30,m}]；45m -> [{45,m}]；0 -> [{0,m}] */
export function durationCompactParts(mins: number): DurationPart[] {
  if (!mins || mins <= 0) return [{ val: '0', unit: 'm' }]
  if (mins < 60) return [{ val: String(mins), unit: 'm' }]
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0
    ? [{ val: String(h), unit: 'h' }, { val: String(m), unit: 'm' }]
    : [{ val: String(h), unit: 'h' }]
}

/** 紧凑文本：12h30m / 45m / 2h / 0m */
export function formatDurationCompact(mins: number) {
  return durationCompactParts(mins).map(p => p.val + p.unit).join('')
}

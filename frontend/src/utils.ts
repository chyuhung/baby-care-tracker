/* ============================================================
   通用格式化工具
   ============================================================ */

const pad2 = (n: number) => String(n).padStart(2, '0')

function toDate(v: string | Date): Date {
  return v instanceof Date ? v : new Date(v)
}

/** 转为 <input type="datetime-local"> 所需的本地时间 YYYY-MM-DDTHH:mm；
    date-only（如出生日期）原样保留并补齐 00:00，避免经 UTC 解析跨日偏移 */
export function toLocalDatetime(iso: string) {
  if (/^\d{4}-\d{2}-\d{2}$/.test(iso)) return `${iso}T00:00`
  const d = new Date(iso)
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}T${pad2(d.getHours())}:${pad2(d.getMinutes())}`
}

/** 解析日历日：date-only（YYYY-MM-DD）直接按本地日历日构造，不做时区换算；
    RFC3339（旧数据/时刻）按本地时区转回本地时刻 */
export function parseLocalDate(s: string): Date | null {
  if (!s) return null
  if (/^\d{4}-\d{2}-\d{2}$/.test(s)) {
    const [y, m, d] = s.split('-').map(Number)
    const t = new Date(y, m - 1, d)
    return isNaN(t.getTime()) ? null : t
  }
  const t = new Date(s)
  return isNaN(t.getTime()) ? null : t
}

/** 日历日 → 中文：2026年9月1日（无前导零；仅取本地年/月/日） */
export function formatDateCN(s: string): string {
  const d = parseLocalDate(s)
  if (!d) return ''
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
}

/** 出生到某日的年龄（精确到天）：20天 / 3月5天 / 1年3月5天 / 1年 / 0天
    仅年月日计算，不做时区换算；任意一端为空或该日早于出生日 → 空串 */
export function measureAgeText(birthDate: string, measuredAt: string): string {
  const birth = parseLocalDate(birthDate)
  const at = parseLocalDate(measuredAt)
  if (!birth || !at) return ''
  if (at.getTime() < birth.getTime()) return ''
  let y = at.getFullYear() - birth.getFullYear()
  let m = at.getMonth() - birth.getMonth()
  let d = at.getDate() - birth.getDate()
  if (d < 0) {
    const prev = new Date(at.getFullYear(), at.getMonth(), 0)
    d += prev.getDate()
    m--
  }
  if (m < 0) {
    m += 12
    y--
  }
  const parts: string[] = []
  if (y > 0) parts.push(`${y}年`)
  if (m > 0) parts.push(`${m}月`)
  if (d > 0) parts.push(`${d}天`)
  return parts.length ? parts.join('') : '0天'
}

/** 两个日历日之间的整天数（按本地日历日算，不受时分秒/夏令时影响） */
function calendarDaysBetween(birth: Date, at: Date): number {
  const utc = (d: Date) => Date.UTC(d.getFullYear(), d.getMonth(), d.getDate())
  return Math.round((utc(at) - utc(birth)) / 86400000)
}

/** 宝宝当前年龄（分段口径）：
    ≤100 天 → 天数 `86天`；满 100 天至 1 岁 → 月龄 `7个月`；
    1 岁至满 3 岁 → 年+月 `2岁5个月`；超过 3 岁 → 仅年 `4岁`。
    边界按「满 N 岁」判定：满 3 岁当天即不再展示月份（3岁0月 与 3岁1月 同显示 `3岁`）。
    at 缺省取今天；无出生日期或该日早于出生日 → 空串 */
export function babyAgeText(birthDate: string, at?: string): string {
  const birth = parseLocalDate(birthDate)
  const ref = at ? parseLocalDate(at) : new Date()
  if (!birth || !ref) return ''
  if (ref.getTime() < birth.getTime()) return ''
  const days = calendarDaysBetween(birth, ref)
  // 出生 100 天内按天更直观（婴儿期变化以天计）
  if (days <= 100) return `${days}天`
  let y = ref.getFullYear() - birth.getFullYear()
  let m = ref.getMonth() - birth.getMonth()
  if (ref.getDate() < birth.getDate()) m--
  if (m < 0) { m += 12; y-- }
  // 满 3 岁起不再展示月份（「超过三岁只展示年龄」）
  if (y >= 3) return `${y}岁`
  if (y >= 1) return m > 0 ? `${y}岁${m}个月` : `${y}岁`
  return `${m}个月`
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

/** 日历日标签：今天 / 昨天 / M-D；markToday=false 时今天返回空串 */
export function formatDayTag(v: string | Date, markToday = true) {
  const d = toDate(v)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) return markToday ? '今天' : ''
  const yesterday = new Date(now.getTime() - 86400000)
  if (d.toDateString() === yesterday.toDateString()) return '昨天'
  return `${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`
}

/** 列表时间标签：今天 HH:mm / 昨天 HH:mm / M-D HH:mm；withDate=false 时仅 HH:mm；
    markToday=false（首页）时今天不标，仅 "HH:mm" */
export function formatDayTime(iso: string, withDate = true, markToday = true) {
  const d = new Date(iso)
  const hhmm = formatClock(d)
  if (!withDate) return hhmm
  const tag = formatDayTag(d, markToday)
  return tag ? `${tag} ${hhmm}` : hhmm
}

/** 区间时间标签：
    同日 + withDate → "12:00~13:00"（今天）/ "昨天 12:00~13:00"；同日 + !withDate → "12:00~13:00"
    跨天 → 两端点日期必带（含今天），消除歧义："昨天 23:00~今天 08:00"/"23:00~今天 08:00"
    无结束 → 进行中，仅显示起点 */
export function formatTimeRangeDay(startIso: string, endIso?: string | null, withDate = true) {
  const start = toDate(startIso)
  const tag = withDate ? formatDayTag(start, false) : ''
  const startLabel = tag ? `${tag} ${formatClock(start)}` : formatClock(start)
  if (!endIso) return startLabel
  const end = toDate(endIso)
  const endLabel = `${formatDayTag(end)} ${formatClock(end)}`
  if (start.toDateString() === end.toDateString()) return `${startLabel}~${formatClock(end)}`
  return `${startLabel}~${endLabel}`
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

/** 中文时长：2小时30分钟 / 2小时 / 15分钟 / 0分钟 */
export function formatDurationCN(mins: number) {
  if (!mins || mins <= 0) return '0分钟'
  if (mins < 60) return `${mins}分钟`
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0 ? `${h}小时${m}分钟` : `${h}小时`
}

/* ============================================================
   头像字色：按底色亮度自动取墨色/白字（WCAG 对比度）
   ============================================================ */

/** sRGB 相对亮度（WCAG 2.1 定义）。输入 0–1 的线性化前通道值 */
function channelLum(v: number) {
  return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4)
}

/** 解析 #RGB / #RRGGBB → {r,g,b}（0–255）；解析失败返回 null */
function parseHex(hex: string): { r: number, g: number, b: number } | null {
  const h = (hex || '').trim().replace(/^#/, '')
  if (h.length === 3) {
    return {
      r: parseInt(h[0] + h[0], 16),
      g: parseInt(h[1] + h[1], 16),
      b: parseInt(h[2] + h[2], 16),
    }
  }
  if (h.length === 6) {
    return {
      r: parseInt(h.slice(0, 2), 16),
      g: parseInt(h.slice(2, 4), 16),
      b: parseInt(h.slice(4, 6), 16),
    }
  }
  return null
}

/** 底色相对亮度 0–1；无法解析时按中性灰兜底（倾向返回低亮度 → 白字） */
export function colorLuminance(hex: string): number {
  const c = parseHex(hex)
  if (!c) return 0.2
  return 0.2126 * channelLum(c.r / 255) + 0.7152 * channelLum(c.g / 255) + 0.0722 * channelLum(c.b / 255)
}

/**
 * 头像首字母的字色：在「墨色」与「白」之间取对比度更高的一方。
 *
 * 为什么需要：宝宝头像底色是用户从 8 色调色板选的，8 个色配白字有 5 个
 * 不达 WCAG AA（最低 #FFB300 仅 1.79:1 —— 浅黄底白字几乎不可见）。
 * 逐个把底色加深会破坏「选色器所见即所得」且要迁移存量数据，
 * 改为按亮度自动换墨色：8 色实测全部落到墨色，最低对比 5.08:1（全部达 AA）。
 *
 * 墨色是**主题无关**的深墨，不用 --text-primary —— 该 token 在暗色模式是
 * 浅色（#F2F2F7），会在饱和色块上反过来失效。
 */
export function avatarInk(hex: string): string {
  const l = colorLuminance(hex)
  // 白字对比 = 1.05 / (l + 0.05)；墨色(≈#000，对比度 ≈20)对比 = (l + 0.05) / 0.05
  const vsWhite = 1.05 / (l + 0.05)
  const vsInk = (l + 0.05) / 0.05
  return vsInk >= vsWhite ? '#000000' : '#FFFFFF'
}

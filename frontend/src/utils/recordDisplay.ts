/* ============================================================
   recordDisplay — 记录卡片的展示元数据（供上下文菜单、骨架屏等复用）
   ============================================================ */

export interface RecordDisplay {
  emoji: string
  title: string
  subtitle: string
}

const FEEDING: Record<string, string> = { breast: '母乳亲喂', bottle: '母乳瓶喂', formula: '配方奶' }
const DIAPER: Record<string, string> = { pee: '小便', poop: '大便', mixed: '混合' }

function hhmm(iso?: string | null) {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

export function recordDisplay(record: any): RecordDisplay {
  const rd = record?.data || {}
  const type = record?.record_type
  switch (type) {
    case 'feeding': {
      const parts: string[] = []
      if (rd.amount_ml > 0) parts.push(`${rd.amount_ml} ml`)
      if (rd.duration_minutes > 0) parts.push(`${rd.duration_minutes} 分钟`)
      if (rd.type === 'breast' && rd.side) parts.push(rd.side === 'left' ? '左侧' : rd.side === 'right' ? '右侧' : '双边')
      parts.push(hhmm(record.occurred_at))
      return { emoji: '🍼', title: FEEDING[rd.type] || '喂奶', subtitle: parts.filter(Boolean).join(' · ') }
    }
    case 'diaper': {
      const parts = [hhmm(record.occurred_at), rd.note].filter(Boolean)
      return { emoji: '🩲', title: DIAPER[rd.type] || '尿布', subtitle: parts.join(' · ') }
    }
    case 'sleep': {
      const parts = [`${hhmm(rd.started_at)} – ${rd.ended_at ? hhmm(rd.ended_at) : '进行中'}`]
      return { emoji: '😴', title: '睡眠', subtitle: parts.join(' · ') }
    }
    case 'temperature':
      return { emoji: '🌡️', title: '体温', subtitle: [`${rd.temperature ?? '--'}°C`, rd.location, hhmm(record.occurred_at)].filter(Boolean).join(' · ') }
    case 'supplement':
      return { emoji: '💊', title: rd.name || '补剂', subtitle: [rd.dosage_value > 0 ? `${rd.dosage_value}${rd.dosage_unit || ''}` : '', hhmm(record.occurred_at)].filter(Boolean).join(' · ') }
    default:
      return { emoji: '🌳', title: '户外活动', subtitle: [`${hhmm(rd.started_at)} – ${rd.ended_at ? hhmm(rd.ended_at) : '进行中'}`].join('') }
  }
}

// 长按菜单现在只有「删除」一项（编辑走卡片点按），edit/copy 图标随之下线
export const CONTEXT_ICONS = {
  delete: 'M4 7h16M9 7V5a1 1 0 011-1h4a1 1 0 011 1v2m-8 0v12a1 1 0 001 1h6a1 1 0 001-1V7',
}

/* ============================================================
   useUndoDelete — iOS 风格「优化删除 + 撤销」
   ------------------------------------------------------------
   删除时先从列表移除（即时反馈），后端 DELETE 延迟到撤销窗口
   结束后才真正发出；窗口内点「撤销」则取消定时器并把记录放回。
   ============================================================ */
import { useAppStore } from '@/stores/app'
import { recordAPI } from '@/api'

const UNDO_MS = 5000

export interface DeletableRecord {
  id: number
  record_type: string
  [k: string]: unknown
}

export function useUndoDelete<T extends DeletableRecord>(
  list: { value: T[] },
  opts: { onRemoved?: (r: T) => void; onRestored?: (r: T) => void } = {},
) {
  const app = useAppStore()

  function softDelete(r: T) {
    const index = list.value.findIndex(x => x.id === r.id && x.record_type === r.record_type)
    if (index < 0) return
    // 1) 即时移除
    list.value.splice(index, 1)
    opts.onRemoved?.(r)

    // 2) 延迟真正删除（撤销窗口内可取消）
    let cancelled = false
    const timer = setTimeout(async () => {
      if (cancelled) return
      try {
        await recordAPI.delete(r.id, r.record_type)
        window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id: r.id, type: r.record_type } }))
      } catch {
        // 删除失败 → 回滚
        list.value.splice(Math.min(index, list.value.length), 0, r)
        opts.onRestored?.(r)
        app.showToast('删除失败，已恢复', 'error')
      }
    }, UNDO_MS)

    // 3) 带「撤销」的 toast
    app.showToast(`已删除${typeLabel(r.record_type)}`, 'success', {
      label: '撤销',
      handler: () => {
        cancelled = true
        clearTimeout(timer)
        list.value.splice(Math.min(index, list.value.length), 0, r)
        opts.onRestored?.(r)
      },
    }, UNDO_MS + 500)
  }

  return { softDelete }
}

export function typeLabel(t: string): string {
  return ({
    feeding: '喂奶记录',
    diaper: '尿布记录',
    sleep: '睡眠记录',
    temperature: '体温记录',
    outdoor: '户外记录',
    supplement: '补剂记录',
  } as Record<string, string>)[t] || '记录'
}

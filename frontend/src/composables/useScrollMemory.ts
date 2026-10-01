/* ============================================================
   useScrollMemory — 页面滚动位置记忆
   ------------------------------------------------------------
   iOS HIG 要求：从列表进详情再返回，必须回到离开时的位置；Material 3 同理。

   按 URL（fullPath，含 query）存键 —— 与浏览器原生 scrollRestoration
   「按 history entry 存」的模型一致，形态同 Flutter 的 PageStorage + PageStorageKey。
   刻意不绑组件实例：某页即使不在 keep-alive 里、被销毁后重挂载，位置照样能恢复。

   为什么必须手写（其他框架都是内建、白送的）：
   - UIKit 保留 view controller、Compose 的 rememberLazyListState 内部是
     rememberSaveable、RN native-stack 保留原生 scroll offset、Flutter 直接有 PageStorage；
   - 本项目列表页的滚动容器是 PullRefresh 自己的 overflow-y-auto，不是 window，
     所以浏览器原生滚动恢复 / history.scrollRestoration 都不起作用；
   - Vue keep-alive 停用页面时会把 DOM move 进一个「脱离文档」的 storage 容器，
     浏览器随即清零该容器的 scrollTop 且不会自行恢复。

   两处时序是关键，均读 runtime-core KeepAlive 源码确认：
   1. 保存用路由守卫 onBeforeRouteLeave —— 它在导航确认阶段触发，此时 DOM 仍在文档中，
      scrollTop 是真值。若挪到 onDeactivated 里读，元素已被摘出文档，读到的永远是 0
      （deactivate 里 move 早于 invokeArrayFns(da)）。
   2. 恢复必须等 DOM 重新挂回 —— keep-alive 的 activate 先 move 回文档、post-render 才
      触发 onActivated；router.afterEach 早于组件更新，故要 await nextTick()。
   ============================================================ */
import { nextTick, onUnmounted } from 'vue'
import { onBeforeRouteLeave, useRouter } from 'vue-router'

/* URL → 滚动偏移。全模块共享一份，等价 Flutter 的 PageStorage。 */
const positions = new Map<string, number>()

/** 清空全部记忆（登出时调用，避免换账号后沿用上个账号的滚动位置） */
export function clearScrollMemory() {
  positions.clear()
}

/**
 * 在持有滚动容器的组件里调用，自动完成「离开时保存 / 回来时恢复」。
 *
 * 恢复统一走 router.afterEach 一个入口，因此「切 tab」与「深页 push/pop 后返回」
 * 共用同一条路径，无需区分 keep-alive 的嵌套层级（App.vue 缓存 MainLayout，
 * MainLayout 内层又缓存 4 个标签页，两层的激活钩子行为并不相同）。
 *
 * @param getEl 取滚动容器元素，通常直接传模板 ref 的 getter，如 () => rootRef.value
 */
export function useScrollMemory(getEl: () => HTMLElement | null) {
  const router = useRouter()

  // 离开本路由时保存。from 就是当前要离开的 URL，DOM 此刻仍在文档中，值可信。
  onBeforeRouteLeave((_to, from) => {
    const el = getEl()
    if (el) positions.set(from.fullPath, el.scrollTop)
  })

  const stop = router.afterEach(async (to) => {
    // 等 keep-alive 把 DOM 挂回文档后再判断，否则 isConnected 恒为 false
    await nextTick()
    const el = getEl()
    // 未激活的缓存页 DOM 停在 keep-alive 的 storage 容器里（脱离文档），据此自我排除，
    // 只有当前真正可见的那一页会被恢复
    if (!el || !el.isConnected) return
    // 该 URL 从未存过位置（如首页带筛选跳来的 /timeline?filter=xx）→ 保持顶部。
    // 这是对的：换了筛选就是换了个视图，等同于原生 app 推入一个新页面。
    const top = positions.get(to.fullPath) ?? 0
    if (top > 0) el.scrollTop = top
  })

  // keep-alive 停用不触发 onUnmounted，故缓存中的页面仍保有自己的钩子（正确）；
  // 真正卸载时才摘除，避免钩子残留。
  onUnmounted(stop)
}
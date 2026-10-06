import { ref, computed } from 'vue'

/**
 * 后端可达性（HTTP 口径）。
 *
 * 为什么不用 wsConnected：WS 是纯接收通道，不参与任何读写判定；运营商 NAT 会静默
 * 回收空闲连接，此时 HTTP 完全正常。若拿 WS 状态门控提交，会在网络其实可用时
 * 误封记录提交——比误报离线严重得多。
 *
 * 为什么不用 navigator.onLine：连上 Wi-Fi 但无外网时它仍报 true。
 *
 * 判定规则：任何 HTTP 响应（含 4xx/5xx）都算「可达」——后端确实答话了；
 * 只有「完全没收到响应」才算不可达，且要探针二次确认，避免单次抖动被误判。
 * 后者正是启动时那次瞬时重连被当成离线的原因（探针一发即通 → 不报警、不封提交）。
 */
const HEALTH_URL = '/api/health'
const PROBE_DELAY_MS = 500
const PROBE_TIMEOUT_MS = 3000
const RETRY_INTERVAL_MS = 10000
// 冷启动宽限期：打开页面时后端连接建立本来就需要几秒（服务未就绪/排队），
// 宽限期内探测失败只重探、不落离线，否则会把「还在连」误报成「已离线」弹 toast。
export const BOOT_GRACE_MS = 10000
const BOOT_RETRY_MS = 2000

export const backendReachable = ref(true)
export const offline = computed(() => !backendReachable.value)

let probeTimer: number | null = null
let retryTimer: number | null = null
let abort: AbortController | null = null
let started = false
let bootAt = 0

function stopPolling() {
  if (retryTimer !== null) {
    clearInterval(retryTimer)
    retryTimer = null
  }
}

function startPolling() {
  if (retryTimer !== null) return
  // 恢复检测：离线时每 10s 探一次，前台才探（后台交给 online/visibility 事件）。
  // 没有它，一旦离线就再没有请求发出去，状态将永远停在不可达、提交再也解不开。
  retryTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible') runProbe()
  }, RETRY_INTERVAL_MS)
}

function withinBootGrace(): boolean {
  return bootAt !== 0 && Date.now() - bootAt < BOOT_GRACE_MS
}

function setReachability(v: boolean) {
  if (backendReachable.value === v) return
  if (!v && withinBootGrace()) {
    // 宽限期内的失败不落离线：改排一次重探（复用 probeTimer 互斥槽，防探针风暴）。
    // 探针失败本身就已确认过一次网络错误，宽限期满后的下一次失败才真正翻转。
    if (probeTimer === null) probeTimer = window.setTimeout(() => { void runProbe() }, BOOT_RETRY_MS)
    return
  }
  backendReachable.value = v
  if (v) stopPolling()
  else startPolling()
}

/** 距冷启动的毫秒数；未初始化时返回 0（调用方据此套用完整延迟） */
export function sinceBootMs(): number {
  return bootAt === 0 ? 0 : Date.now() - bootAt
}

async function runProbe(): Promise<boolean> {
  if (probeTimer !== null) {
    clearTimeout(probeTimer)
    probeTimer = null
  }
  const controller = new AbortController()
  abort = controller
  const timer = window.setTimeout(() => controller.abort(), PROBE_TIMEOUT_MS)
  let ok = false
  try {
    const res = await fetch(HEALTH_URL, { cache: 'no-store', signal: controller.signal })
    ok = res.ok
  } catch {
    ok = false
  } finally {
    window.clearTimeout(timer)
    if (abort === controller) abort = null
  }
  setReachability(ok)
  return ok
}

/** 收到任意 HTTP 响应（含 4xx/5xx）→ 后端确定可达，顺带撤掉待发探针 */
export function noteResponse() {
  if (probeTimer !== null) {
    clearTimeout(probeTimer)
    probeTimer = null
  }
  setReachability(true)
}

/**
 * 请求完全没拿到响应。**不立即**判定离线——单次抖动是常态（尤其页面刚打开、
 * 并行请求抢占连接额度时）。延迟后探一次 /api/health：通了说明只是抖动，
 * 不报警也不封提交；不通才落不可达。
 */
export function noteNetworkError() {
  if (!backendReachable.value) return
  if (probeTimer !== null) return
  probeTimer = window.setTimeout(() => { void runProbe() }, PROBE_DELAY_MS)
}

/** 主动复查（app 启动 / 网络恢复 / 回前台） */
export function probeBackendNow(): Promise<boolean> {
  return runProbe()
}

/** 挂上 `online` / `visibilitychange` 恢复探测。幂等。 */
export function initReachability() {
  if (started) return
  started = true
  bootAt = Date.now()
  window.addEventListener('online', () => { void runProbe() })
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') void runProbe()
  })
  // 冷启动就探一次：健康时立即转在线；失败则进入宽限期重探（见 setReachability），
  // 约 10s 后仍连不上才落离线——给连接建立留出时间，不误报「打开即离线」。
  void runProbe()
}

/** 请求拦截器拒绝写入时抛出的对象；形状对齐 `!e.response` 以复用 writeErrorMessage */
export function offlineWriteError() {
  return { offline: true, message: '当前离线，本次操作未生效' }
}
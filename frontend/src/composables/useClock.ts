import { ref } from 'vue'

// 全站唯一 10s 时钟。
// 消费者在 computed 里读一次 clockTick.value 即可订阅。
// 用于两处必须随停留时间自己走动的显示：
// RecordCard 进行中记录的已持续时长、HomePage 距上次 N 分钟前。
// 不按订阅数启停，因为这些组件活在 keep-alive 里、永不 unmount，
// 订阅计数与 onUnmounted 清理都不可靠。
// 一个 10s 定时器开销可忽略，换来全站只有一只钟、两处显示不会各自漂移。
export const clockTick = ref(0)

window.setInterval(() => { clockTick.value++ }, 10000)

<template>
  <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
    <div class="px-4 pt-3 pb-1 flex items-center justify-between">
      <h2 class="text-sm font-semibold text-text-secondary">提醒</h2>
      <button v-if="!permissionGranted" type="button" @click="enable"
        class="text-xs text-primary-deep font-medium py-2 px-2 -mr-2 min-h-[44px]">开启通知</button>
    </div>

    <p v-if="!supported" class="px-4 pb-3 text-xs text-text-secondary leading-relaxed">
      当前环境不支持系统通知。可将本应用添加到主屏幕（PWA）后再开启提醒。
    </p>
    <p v-else-if="permission === 'denied'" class="px-4 pb-3 text-xs text-text-secondary leading-relaxed">
      通知权限已被拒绝，请在浏览器设置中允许本网站发送通知。
    </p>

    <button v-for="r in list" :key="r.id" type="button" @click="toggle(r)"
      class="w-full px-4 py-3 flex items-center justify-between border-t border-border-color/60 min-h-[44px] text-left btn-press">
      <span class="flex items-center gap-2.5 text-text-primary text-sm">
        <span class="text-lg leading-none">{{ r.emoji }}</span>
        {{ r.label }}
      </span>
      <span class="flex items-center gap-2">
        <span v-if="r.enabled" class="text-xs text-text-secondary">每 {{ formatInterval(r.intervalMin) }}</span>
        <span :class="['relative w-[51px] h-[31px] rounded-full transition-colors duration-200 shrink-0', r.enabled ? 'bg-success' : 'bg-border-color']">
          <span :class="['absolute top-[2px] w-[27px] h-[27px] rounded-full bg-white shadow transition-transform duration-200', r.enabled ? 'translate-x-[22px]' : 'translate-x-[2px]']"></span>
        </span>
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import {
  Reminder, loadReminders, saveReminders, notificationsSupported,
  notificationPermission, requestNotificationPermission,
} from '@/utils/reminders'
import { hapticSelection, hapticSuccess } from '@/utils/haptic'

const app = useAppStore()
const supported = notificationsSupported()
const permission = ref(notificationPermission())
const list = ref<Reminder[]>(loadReminders())

const permissionGranted = computed(() => permission.value === 'granted')

function formatInterval(min: number) {
  if (min % 60 === 0) return `${min / 60} 小时`
  return `${min} 分钟`
}

async function enable() {
  const ok = await requestNotificationPermission()
  permission.value = notificationPermission()
  if (ok) {
    hapticSuccess()
    app.showToast('通知已开启', 'success')
  } else {
    app.showToast('未能开启通知', 'error')
  }
}

async function toggle(r: Reminder) {
  if (!r.enabled && !permissionGranted.value) {
    const ok = await requestNotificationPermission()
    permission.value = notificationPermission()
    if (!ok) {
      app.showToast('请先允许通知权限', 'error')
      return
    }
  }
  r.enabled = !r.enabled
  hapticSelection()
  saveReminders(list.value)
  app.showToast(r.enabled ? `已开启${r.label}` : `已关闭${r.label}`, 'success')
}

onMounted(() => {
  list.value = loadReminders()
})
</script>

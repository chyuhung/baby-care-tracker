<template>
  <div class="flex flex-col h-dvh">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="refreshAll">
    <template #header>
      <NavBar title="我的" />
    </template>

      <!-- 用户信息（微信「我」页：头像+昵称直接落在页面底色上，不套白卡） -->
      <div class="flex items-center gap-4">
        <div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
          <svg class="w-7 h-7 text-primary-deep" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2M12 11a4 4 0 100-8 4 4 0 000 8z" />
          </svg>
        </div>
        <div>
          <div class="text-xl font-semibold text-text-primary tracking-tight">{{ auth.user?.username }}</div>
          <div class="text-[13px] text-text-secondary mt-0.5">家庭成员</div>
        </div>
      </div>

      <!-- 宝宝：切换 + 档案合并（微信分组：单卡多行，行 tap=切换、行尾「编辑」） -->
      <div>
        <div class="pb-1.5 flex items-center justify-between">
          <h2 class="text-[13px] text-text-secondary self-start">宝宝</h2>
          <!-- 触控区 min-h-11 靠负 margin 外扩，不参与父行高度计算：
               否则 items-center 会把 20px 的 h2 沉到 44px 行正中，标题比另两组低 12px -->
          <router-link to="/baby/new" aria-label="添加宝宝"
            class="text-primary-deep text-sm font-medium flex items-center gap-1 min-h-[44px] -my-3 -mr-2 px-2 btn-press">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
            添加
          </router-link>
        </div>
        <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
          <div v-if="app.babies.length === 0" class="px-4 py-4 text-sm text-text-secondary">还没有宝宝档案，点击右上角添加</div>
          <div v-else class="divide-y divide-border-color/60">
            <div v-for="baby in app.babies" :key="baby.id" role="button" tabindex="0"
              :aria-current="isCurrentBaby(baby) ? 'true' : undefined"
              class="w-full px-4 py-3.5 flex items-center gap-3 min-h-[44px] text-left press-card"
              @click="switchBaby(baby)">
              <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-sm font-bold"
                :style="{ background: baby.avatar_color, color: avatarInk(baby.avatar_color) }">{{ baby.name[0] }}</span>
              <span class="flex-1 min-w-0">
                <span class="block font-medium text-text-primary truncate">
                  {{ baby.name }}
                  <svg v-if="isCurrentBaby(baby)" class="inline-block h-4 w-4 -mt-0.5 ml-0.5 text-primary-deep"
                    fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M5 13l4 4L19 7" />
                  </svg>
                </span>
                <span class="block text-xs text-text-secondary mt-0.5 truncate">
                  <template v-if="babyAgeText(baby.birth_date)">{{ babyAgeText(baby.birth_date) }} · </template>{{ formatBirthDate(baby.birth_date) }}
                </span>
              </span>
              <!-- 当前宝宝也必须有编辑入口：此前此处是 v-else，单宝宝时全 App 无资料编辑入口 -->
              <button type="button" @click.stop="router.push(`/baby/${baby.id}/edit`)"
                class="-mr-1 shrink-0 text-xs font-medium text-primary-deep py-2 px-1 min-h-[44px]">编辑</button>
            </div>
          </div>
        </div>
      </div>

      <!-- 我的家庭（折叠：默认一行，点击展开管理；无家庭时直接显示加入表单） -->
      <div>
        <h2 class="pb-1.5 text-[13px] text-text-secondary">我的家庭</h2>
        <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
          <!-- 无家庭：直接加入 -->
          <div v-if="!family" class="p-4 space-y-2">
            <p class="text-xs text-text-secondary">加入家庭后，你创建的宝宝会跟随你，可与家人共同记录</p>
            <div class="flex gap-2 pt-1">
              <input v-model="joinCode" placeholder="输入对方的邀请码" maxlength="6"
                aria-label="邀请码" inputmode="text" autocapitalize="characters" autocomplete="off" enterkeyhint="done"
                class="flex-1 min-h-[44px] px-3 py-2.5 bg-muted border border-border-color rounded-xl text-base focus:border-primary focus:outline-none transition-colors uppercase" />
              <button @click="joinFamily" :disabled="!joinCode.trim()"
                class="px-4 py-3 bg-primary/10 text-primary-deep text-sm font-medium rounded-xl btn-press min-h-[44px] disabled:opacity-40 disabled:cursor-not-allowed">加入</button>
            </div>
          </div>

          <!-- 有家庭：默认一行，点击展开 -->
          <template v-else>
            <div role="button" tabindex="0" :aria-expanded="famOpen" :aria-controls="'fam-detail'" @keydown.enter.prevent="famOpen = !famOpen"
              class="w-full px-4 py-3.5 flex items-center gap-3 min-h-[44px] text-left press-card"
              @click="famOpen = !famOpen">
              <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10">
                <svg class="w-5 h-5 text-primary-deep" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M4 10.5V20h16V10.5M3 10.5L12 4l9 6.5" />
                </svg>
              </span>
              <span class="flex-1 min-w-0">
                <span class="block font-medium text-text-primary">我的家庭</span>
                <span class="block text-xs text-text-secondary mt-0.5">{{ family.members.length }} 位成员</span>
              </span>
              <svg class="w-5 h-5 shrink-0 text-text-secondary/50 transition-transform" :class="famOpen ? 'rotate-180' : ''"
                fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </div>

            <div v-if="famOpen" id="fam-detail" class="divide-y divide-border-color/60">
              <!-- 邀请码 + 重置/复制 并列 -->
              <div class="px-4 py-3.5 flex items-center justify-between min-h-[44px]">
                <span class="text-sm text-text-secondary">邀请码</span>
                <span class="flex items-center gap-2">
                  <span class="text-base font-bold tracking-widest text-primary-deep select-all">{{ family.invite_code }}</span>
                  <button @click="requestRegenerateCode" class="-mr-2 text-xs font-medium text-danger py-2 px-2 min-h-[44px] flex items-center btn-press">重置</button>
                  <button @click="copyCode" class="-mr-2 text-xs font-medium text-primary-deep py-2 px-2 min-h-[44px] flex items-center">复制</button>
                </span>
              </div>

              <!-- 成员 -->
              <div class="px-4 py-3.5 flex items-center justify-between gap-3 min-h-[44px]">
                <span class="text-sm text-text-secondary shrink-0">家庭成员</span>
                <span class="truncate text-right text-sm text-text-primary">
                  {{ memberText }}
                </span>
              </div>

              <!-- 加入其他家庭（单行：标签 + 邀请码 + 加入，不再折叠） -->
              <div class="px-4 py-3.5 flex items-center gap-3 min-h-[44px]">
                <span class="text-sm text-text-secondary shrink-0">加入其他家庭</span>
                <input v-model="joinCode" placeholder="输入对方邀请码" maxlength="6"
                  aria-label="邀请码" inputmode="text" autocapitalize="characters" autocomplete="off" enterkeyhint="done"
                  class="flex-1 min-w-0 min-h-[44px] px-3 py-2.5 bg-muted border border-border-color rounded-xl text-base focus:border-primary focus:outline-none transition-colors uppercase" />
                <button @click="joinFamily" :disabled="!joinCode.trim()"
                  class="shrink-0 px-4 py-3 bg-primary/10 text-primary-deep text-sm font-medium rounded-xl btn-press min-h-[44px] disabled:opacity-40 disabled:cursor-not-allowed">加入</button>
              </div>

              <!-- 退出家庭（破坏性操作挂所属语境，微信退出群聊式卡底居中红字） -->
              <button v-if="family.members.length > 1" type="button" @click="leaveFamily"
                class="w-full py-3.5 text-sm font-medium text-danger text-center min-h-[44px] btn-press">
                退出家庭
              </button>
            </div>
          </template>
        </div>
      </div>

      <!-- 数据（成长记录 / 导出 / 版本，合并单卡多行） -->
      <div>
        <h2 class="pb-1.5 text-[13px] text-text-secondary">数据</h2>
        <div class="bg-surface rounded-2xl shadow-card overflow-hidden divide-y divide-border-color/60">
          <router-link to="/growth" class="px-4 py-3.5 flex items-center gap-3 min-h-[44px] press-card block">
            <span class="w-9 h-9 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
              <svg class="w-5 h-5 text-primary-deep" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M4 20V10M10 20V4M16 20v-8M4 20h16" />
              </svg>
            </span>
            <span class="flex-1 min-w-0">
              <span class="block font-medium text-text-primary">成长记录</span>
              <span class="block text-xs text-text-secondary mt-0.5">身高 · 体重 · 头围与生长标准百分位</span>
            </span>
            <svg class="w-5 h-5 text-text-secondary/50 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
          </router-link>

          <button type="button" @click="exportData" :disabled="exporting || app.babies.length === 0"
            class="w-full px-4 py-3.5 flex items-center gap-3 min-h-[44px] text-left press-card disabled:opacity-40">
            <span class="w-9 h-9 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
              <svg class="w-5 h-5 text-primary-deep" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" />
              </svg>
            </span>
            <span class="flex-1 min-w-0">
              <span class="block font-medium text-text-primary">导出全部记录（CSV）</span>
              <span class="block text-xs text-text-secondary mt-0.5">Excel / 医生可直接打开，用于就诊或备份</span>
            </span>
            <ActivityIndicator v-if="exporting" :size="18" class="text-text-secondary" />
          </button>
        </div>
      </div>

      <!-- 退出登录（页底唯一危险操作，微信「我」页式全宽居中红字） -->
      <button type="button" @click="logout"
        class="w-full py-3.5 text-sm font-medium text-danger text-center min-h-[44px] bg-surface rounded-2xl shadow-card mt-2 btn-press">
        退出登录
      </button>
    </PullRefresh>

    <!-- 加入 / 退出家庭确认（iOS 底部操作表） -->
    <ConfirmSheet :open="sheet.open" :title="sheet.title" :message="sheet.message"
      :confirm-text="sheet.confirmText" :danger="sheet.danger" @confirm="onSheetConfirm" @cancel="sheetMode = ''" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

// 显式命名：MainLayout 内层 <keep-alive include="ProfilePage,..."> 命中缓存
defineOptions({ name: 'ProfilePage' })

import { familyAPI, recordAPI } from '@/api'
import PullRefresh from '@/components/PullRefresh.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'

import NavBar from '@/components/NavBar.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import { parseLocalDate, babyAgeText, avatarInk } from '@/utils'

interface FamilyMember {
  id: number
  username: string
}

interface Family {
  id: number
  invite_code: string
  members: FamilyMember[]
}

const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()
const exporting = ref(false)

async function exportData() {
  const baby = app.currentBaby
  if (!baby) { app.showToast('请先添加宝宝', 'error'); return }
  exporting.value = true
  try {
    const res = await recordAPI.exportRecords(baby.id)
    const blob = new Blob([res.data], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const d = new Date()
    const p2 = (n: number) => String(n).padStart(2, '0')
    a.href = url
    a.download = `${baby.name}-${d.getFullYear()}${p2(d.getMonth() + 1)}${p2(d.getDate())}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    setTimeout(() => URL.revokeObjectURL(url), 1000)
    app.showToast('已导出', 'success')
  } catch {
    app.showToast('导出失败', 'error')
  } finally {
    exporting.value = false
  }
}

const family = ref<Family | null>(null)
const joinCode = ref('')
const sheetMode = ref<'' | 'join' | 'leave' | 'reset'>('')

// 我的家庭折叠展开状态
const famOpen = ref(false)

const memberText = computed(() =>
  family.value
    ? family.value.members.map(m => (m.id === auth.user?.id ? `${m.username}(我)` : m.username)).join('、')
    : ''
)

const sheet = computed(() => ({
  open: sheetMode.value !== '',
  title: sheetMode.value === 'join' ? '加入新家庭'
    : sheetMode.value === 'leave' ? '退出当前家庭' : '重置邀请码',
  message: sheetMode.value === 'join'
    ? '加入新家庭后，你将退出当前家庭。\n\n仅你创建的宝宝会跟随你，原家庭成员将无法看到它们；其他成员创建的宝宝和记录则留在原家庭。确定继续？'
    : sheetMode.value === 'leave'
      ? '退出后，你创建的宝宝将随你离开，原家庭无法再看到它们；其他成员创建的宝宝和记录留在原家庭，你将无法查看。确定继续？'
      : '重置后旧邀请码立即失效，已拿到旧码的家人将无法加入，需要把新码重新发给他们。确定继续？',
  confirmText: sheetMode.value === 'join' ? '加入' : sheetMode.value === 'leave' ? '退出家庭' : '重置',
  // 加入新家庭不是破坏性操作（默认 danger=true 会把「加入」按钮染成红色）
  danger: sheetMode.value === 'leave' || sheetMode.value === 'reset',
}))

async function loadFamily() {
  try {
    const res = await familyAPI.getMyFamily()
    family.value = { ...res.data.family, members: res.data.members }
  } catch {
    // family not available
  }
}

async function refreshAll() {
  await Promise.all([loadFamily(), app.loadBabies()])
}

function joinFamily() {
  if (!joinCode.value.trim()) return
  // 如果用户已有家庭，先二次确认
  if (family.value) {
    sheetMode.value = 'join'
    return
  }
  void doJoin()
}

async function doJoin() {
  const code = joinCode.value.trim().toUpperCase()
  try {
    await familyAPI.join(code)
    joinCode.value = ''
    await loadFamily()
    await app.loadBabies()
    app.showToast('已加入家庭', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '加入失败', 'error')
  } finally {
    sheetMode.value = ''
  }
}

function leaveFamily() {
  sheetMode.value = 'leave'
}
async function doLeave() {
  try {
    await familyAPI.leave()
    family.value = null
    await app.loadBabies()
    app.showToast('已退出家庭', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '退出失败', 'error')
  } finally {
    sheetMode.value = ''
  }
}

function onSheetConfirm() {
  if (sheetMode.value === 'join') void doJoin()
  else if (sheetMode.value === 'reset') void doRegenerateCode()
  else void doLeave()
}

// 重置是破坏性操作（旧码立即失效，家人可能正在用），先走 ConfirmSheet 确认
function requestRegenerateCode() {
  if (!family.value) return
  sheetMode.value = 'reset'
}

async function doRegenerateCode() {
  try {
    const res = await familyAPI.regenerateCode()
    family.value!.invite_code = res.data.invite_code
    app.showToast('邀请码已重置，旧码已失效', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '操作失败', 'error')
  } finally {
    sheetMode.value = ''
  }
}

function copyCode() {
  if (!family.value) return
  const text = family.value.invite_code
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text).then(() => {
      app.showToast('已复制邀请码', 'success')
    }).catch(fallbackCopy)
  } else {
    fallbackCopy()
  }
  function fallbackCopy() {
    const el = document.createElement('textarea')
    el.value = text
    el.style.position = 'fixed'
    el.style.opacity = '0'
    document.body.appendChild(el)
    el.select()
    try {
      document.execCommand('copy')
      app.showToast('已复制邀请码', 'success')
    } catch {
      app.showToast('复制失败，请手动复制', 'error')
    }
    document.body.removeChild(el)
  }
}

function formatBirthDate(bd: string) {
  if (!bd) return ''
  const d = parseLocalDate(bd)
  if (!d) return bd
  const p2 = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p2(d.getMonth() + 1)}-${p2(d.getDate())}`
}

function isCurrentBaby(baby: { id: number }) {
  return app.currentBaby?.id === baby.id
}

function switchBaby(baby: { id: number, name: string }) {
  app.setCurrentBaby(baby.id)
  app.showToast(`已切换为 ${baby.name}`, 'success')
}

onMounted(() => {
  loadFamily()
})

function logout() {
  app.disconnectWebSocket()
  auth.logout()
  router.push('/login')
}
</script>

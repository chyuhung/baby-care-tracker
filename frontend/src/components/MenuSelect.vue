<template>
  <!-- 触发按钮（iOS 胶囊 + 下拉箭头） -->
  <button type="button" @click="open = true"
    class="inline-flex items-center gap-1.5 h-11 pl-3.5 pr-3 rounded-full bg-surface border border-border-color text-text-primary text-sm font-semibold btn-press shadow-sm whitespace-nowrap focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
    :aria-haspopup="true" :aria-expanded="open" :aria-label="ariaLabel">
    <span v-if="current?.emoji" class="text-base leading-none">{{ current.emoji }}</span>
    <span>{{ current?.label ?? placeholder }}</span>
    <svg class="w-3 h-3 text-text-secondary transition-transform duration-200" :class="open ? 'rotate-180' : ''"
      viewBox="0 0 10 6" fill="none">
      <path d="M1 1l4 4 4-4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
    </svg>
  </button>

  <Teleport to="body">
    <transition name="sheet-mask">
      <div v-if="open" class="fixed inset-0 z-[80] bg-black/35" @click.self="close">
        <transition name="sheet-panel" appear>
          <div v-if="open" ref="panelRef" tabindex="-1"
            class="absolute bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[480px] px-3 pb-safe space-y-2 outline-none"
            role="dialog" aria-modal="true" :aria-label="title">
            <!-- 选项组 -->
            <div class="bg-surface rounded-2xl overflow-hidden">
              <div class="pt-2.5 flex justify-center">
                <span class="w-9 h-1 rounded-full bg-border-color"></span>
              </div>
              <p v-if="title" class="px-4 pt-2 pb-2 text-[13px] font-semibold text-text-secondary text-center">{{ title }}</p>
              <button v-for="opt in options" :key="String(opt.value)" type="button" @click="select(opt.value)"
                class="w-full flex items-center gap-3 px-4 min-h-[52px] text-[17px] text-text-primary border-t border-border-color/70 btn-press active:bg-muted/60">
                <span v-if="opt.emoji" class="text-xl w-6 text-center leading-none">{{ opt.emoji }}</span>
                <span class="flex-1 text-left">{{ opt.label }}</span>
                <svg v-if="opt.value === modelValue" class="w-5 h-5 text-primary-deep" viewBox="0 0 24 24" fill="none">
                  <path d="M5 13l4 4L19 7" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </button>
            </div>
            <!-- 取消（独立成组） -->
            <button type="button" @click="close"
              class="w-full py-4 bg-surface rounded-2xl text-[17px] font-semibold text-text-primary btn-press">
              取消
            </button>
          </div>
        </transition>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'

export interface MenuOption {
  label: string
  value: string | number
  emoji?: string
}

const props = withDefaults(defineProps<{
  modelValue: string | number
  options: MenuOption[]
  title?: string
  placeholder?: string
  ariaLabel?: string
}>(), {
  title: '',
  placeholder: '请选择',
  ariaLabel: '选择',
})

const emit = defineEmits<{ (e: 'update:modelValue', v: string | number): void }>()

const open = ref(false)
const panelRef = ref<HTMLElement | null>(null)
const current = computed(() => props.options.find(o => o.value === props.modelValue))

function select(v: string | number) {
  emit('update:modelValue', v)
  close()
}
function close() { open.value = false }

/* 焦点陷阱（a11y）+ Esc 关闭 */
function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') { e.preventDefault(); close(); return }
  if (e.key !== 'Tab' || !panelRef.value) return
  const focusables = Array.from(panelRef.value.querySelectorAll<HTMLElement>('button:not([disabled])'))
  if (!focusables.length) return
  const first = focusables[0]
  const last = focusables[focusables.length - 1]
  const active = document.activeElement as HTMLElement | null
  if (e.shiftKey && (active === first || !panelRef.value.contains(active))) {
    e.preventDefault(); last.focus()
  } else if (!e.shiftKey && active === last) {
    e.preventDefault(); first.focus()
  }
}

// 打开时锁定背景滚动 + 监听 Esc/焦点 + 初始聚焦
watch(open, async (v) => {
  if (typeof document === 'undefined') return
  document.body.style.overflow = v ? 'hidden' : ''
  if (v) {
    window.addEventListener('keydown', onKey)
    await nextTick()
    panelRef.value?.focus()
  } else {
    window.removeEventListener('keydown', onKey)
  }
})
onUnmounted(() => {
  if (typeof document !== 'undefined') {
    document.body.style.overflow = ''
    window.removeEventListener('keydown', onKey)
  }
})
</script>

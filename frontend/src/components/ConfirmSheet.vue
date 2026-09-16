<template>
  <Teleport to="body">
    <transition name="sheet-mask">
      <div v-if="open" class="fixed inset-0 z-[80] flex items-end justify-center bg-black/35" @click.self="cancel">
        <transition name="sheet-panel" appear>
          <div v-if="open" class="w-full max-w-[480px] px-3 pb-safe space-y-2" role="alertdialog" aria-modal="true"
            :aria-label="title || confirmText">
            <!-- 上组：说明 + 确认按钮 -->
            <div class="bg-white rounded-2xl overflow-hidden text-center">
              <div class="pt-2.5 flex justify-center">
                <span class="w-9 h-1 rounded-full bg-border-color"></span>
              </div>
              <div class="px-4 pt-2.5 pb-4">
                <p v-if="title" class="text-[15px] font-semibold text-text-primary">{{ title }}</p>
                <p v-if="message" class="text-[13px] text-text-secondary mt-1 leading-relaxed whitespace-pre-line">{{ message }}</p>
              </div>
              <button type="button" @click="onConfirm" :disabled="loading"
                class="w-full py-4 px-4 text-[17px] font-medium border-t border-border-color/70 btn-press disabled:opacity-60 flex items-center justify-center gap-2"
                :class="danger ? 'text-danger' : 'text-primary-deep'">
                <ActivityIndicator v-if="loading" :size="18" />
                <span>{{ confirmText }}</span>
              </button>
            </div>
            <!-- 取消（独立成组） -->
            <button type="button" @click="cancel"
              class="w-full py-4 bg-white rounded-2xl text-[17px] font-semibold text-text-primary btn-press">
              {{ cancelText }}
            </button>
          </div>
        </transition>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import ActivityIndicator from './ActivityIndicator.vue'

withDefaults(defineProps<{
  open: boolean
  title?: string
  message?: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
  loading?: boolean
}>(), {
  title: '',
  message: '',
  confirmText: '确认',
  cancelText: '取消',
  danger: true,
  loading: false,
})

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
  (e: 'update:open', v: boolean): void
}>()

function onConfirm() {
  emit('confirm')
}

function cancel() {
  emit('cancel')
  emit('update:open', false)
}
</script>

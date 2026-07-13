<template>
  <div class="min-h-screen bg-bg-main flex flex-col" :data-theme="app.theme">
    <Toast />
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<script setup lang="ts">
import Toast from '@/components/Toast.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { onMounted } from 'vue'

const auth = useAuthStore()
const app = useAppStore()
onMounted(() => {
  auth.restoreSession()
})
</script>

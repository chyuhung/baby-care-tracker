<template>
  <div class="min-h-screen bg-bg-main flex flex-col">
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
import { onMounted } from 'vue'

const auth = useAuthStore()
onMounted(() => {
  auth.restoreSession()
})
</script>

<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe px-4 py-3 flex items-center gap-3">
      <button aria-label="返回" @click="router.back()" class="p-2 -ml-2 flex items-center justify-center min-w-[44px] min-h-[44px] btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ isEdit ? '编辑宝宝' : '添加宝宝' }}</h1>
    </header>

    <main class="flex-1 px-4 py-6 space-y-5 pb-[calc(6.5rem+env(safe-area-inset-bottom))]">
      <!-- 头像颜色 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">头像颜色</label>
        <div class="flex flex-wrap gap-3">
          <button v-for="c in avatarColors" :key="c" type="button" @click="form.avatar_color = c"
            :aria-label="'选择头像颜色'"
            :class="['w-11 h-11 rounded-full flex items-center justify-center text-white font-bold transition-all btn-press',
              form.avatar_color === c ? 'ring-2 ring-offset-2 ring-text-primary scale-110' : '']"
            :style="{ background: c }">
            {{ form.name ? form.name[0] : '👶' }}
          </button>
        </div>
      </div>

      <!-- 姓名 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">姓名 *</label>
        <input v-model="form.name" type="text" placeholder="宝宝的名字"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 出生日期 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">出生日期 *</label>
        <input v-model="form.birth_date" type="datetime-local"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 性别（iOS 分段控件） -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">性别</label>
        <Segmented :model-value="form.gender" :options="genderOptions" @update:model-value="selectGender" />
      </div>

      <div v-if="error" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">
        {{ error }}
      </div>

      <!-- 删除（编辑态，留在内容区） -->
      <button v-if="isEdit" type="button" @click="showDelete = true"
        class="btn-press w-full py-3 bg-white text-danger font-medium rounded-xl border border-danger/25 min-h-[44px]">
        删除宝宝
      </button>
    </main>

    <!-- 固定底部保存栏 -->
    <FormBar>
      <button type="button" @click="save" :disabled="saving"
        class="btn-press w-full py-3.5 bg-primary-deep text-white font-semibold rounded-xl shadow-card disabled:opacity-50 flex items-center justify-center gap-2">
        <ActivityIndicator v-if="saving" :size="20" class="text-white" />
        <span>{{ saving ? '保存中...' : '保存' }}</span>
      </button>
    </FormBar>

    <ConfirmSheet :open="showDelete" :loading="deleting" message="删除宝宝将同时删除其所有护理记录，且无法恢复。"
      confirm-text="删除宝宝" @confirm="doDelete" @cancel="showDelete = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { babyAPI } from '@/api'
import { nowLocalDatetime, toLocalDatetime } from '@/utils'
import Segmented from '@/components/Segmented.vue'
import FormBar from '@/components/FormBar.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const deleting = ref(false)
const showDelete = ref(false)
const error = ref('')

const avatarColors = ['#F25C8C', '#348EED', '#2DB84F', '#F78A06', '#FFB300', '#FF6B53', '#AF52DE', '#2BAEDF']

const genderOptions = [
  { value: 'male', label: '男孩', emoji: '👦' },
  { value: 'female', label: '女孩', emoji: '👧' },
  { value: '', label: '保密', emoji: '🤷' },
]

const form = reactive({
  name: '',
  birth_date: nowLocalDatetime(),
  gender: '',
  avatar_color: '#F25C8C',
})

function selectGender(g: string) {
  form.gender = g
  if (g === 'male') form.avatar_color = '#348EED'
  else if (g === 'female') form.avatar_color = '#F25C8C'
}

async function loadBaby() {
  if (!route.params.id) return
  try {
    const res = await babyAPI.get(Number(route.params.id))
    const baby = res.data
    form.name = baby.name
    form.birth_date = baby.birth_date ? toLocalDatetime(baby.birth_date) : ''
    form.gender = baby.gender || ''
    form.avatar_color = baby.avatar_color || '#F25C8C'
  } catch {
    app.showToast('加载失败', 'error')
    router.back()
  }
}

async function save() {
  error.value = ''
  if (!form.name.trim()) { error.value = '请输入姓名'; return }
  if (!form.birth_date) { error.value = '请选择出生日期'; return }
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      birth_date: new Date(form.birth_date).toISOString(),
      gender: form.gender,
      avatar_color: form.avatar_color,
    }
    if (isEdit.value) {
      await babyAPI.update(Number(route.params.id), payload)
      app.showToast('已保存', 'success')
    } else {
      await babyAPI.create(payload)
      app.showToast('已添加宝宝', 'success')
    }
    await app.loadBabies()
    router.push('/')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '保存失败', 'error')
  } finally {
    saving.value = false
  }
}

async function doDelete() {
  if (!route.params.id || deleting.value) return
  deleting.value = true
  try {
    await babyAPI.delete(Number(route.params.id))
    await app.loadBabies()
    app.showToast('已删除', 'success')
    router.push('/')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '删除失败', 'error')
    showDelete.value = false
  } finally {
    deleting.value = false
  }
}

onMounted(loadBaby)
</script>

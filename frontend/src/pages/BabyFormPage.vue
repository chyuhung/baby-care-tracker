<template>
  <div class="min-h-screen bg-bg-main">
    <header class="pt-safe bg-white px-4 py-3 border-b border-border-color flex items-center gap-3">
      <button @click="router.back()" class="p-1 -ml-1 btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ isEdit ? '编辑宝宝' : '添加宝宝' }}</h1>
    </header>

    <main class="px-4 py-6 space-y-5">
      <!-- 头像颜色选择 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">头像颜色</label>
        <div class="flex gap-3 flex-wrap">
          <button v-for="color in colors" :key="color"
            @click="form.avatar_color = color"
            :class="['w-10 h-10 rounded-full btn-press transition-transform', form.avatar_color === color ? 'ring-2 ring-offset-2 ring-primary scale-110' : '']"
            :style="{ background: color }">
          </button>
        </div>
      </div>

      <div>
        <label class="text-sm text-text-secondary block mb-2">宝宝姓名 <span class="text-red-500">*</span></label>
        <input v-model="form.name" type="text" placeholder="输入宝宝姓名"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
      </div>

      <div>
        <label class="text-sm text-text-secondary block mb-2">出生日期 <span class="text-red-500">*</span></label>
        <input v-model="form.birth_date" type="datetime-local"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
      </div>

      <div>
        <label class="text-sm text-text-secondary block mb-3">性别</label>
        <div class="flex gap-3">
          <button v-for="g in genders" :key="g.value"
            @click="selectGender(g.value)"
            :class="['flex-1 py-3 rounded-xl font-medium text-sm transition-colors btn-press border',
              form.gender === g.value ? 'bg-primary text-white border-primary shadow-card' : 'bg-white border-border-color text-text-secondary']">
            {{ g.emoji }} {{ g.label }}
          </button>
        </div>
      </div>

      <div v-if="error" class="bg-red-50 text-red-500 text-sm px-4 py-2 rounded-xl text-center">{{ error }}</div>

      <div class="space-y-3 pt-2">
        <button @click="submit" :disabled="loading"
          class="btn-press w-full py-3 bg-primary text-white font-semibold rounded-xl shadow-card disabled:opacity-50">
          {{ loading ? '保存中...' : '保存' }}
        </button>
        <button v-if="isEdit" @click="confirmDelete"
          class="btn-press w-full py-3 bg-white text-red-500 font-medium rounded-xl border border-red-200">
          删除宝宝
        </button>
      </div>
    </main>

    <!-- 删除确认 -->
    <div v-if="showDelete" class="fixed inset-0 bg-black/30 flex items-end z-50" @click.self="showDelete = false">
      <div class="bg-white w-full rounded-t-2xl p-6 space-y-4 pb-safe animate-slide-up">
        <p class="text-text-secondary text-sm text-center">删除后所有记录将无法恢复</p>
        <div class="flex gap-3">
          <button @click="showDelete = false" class="flex-1 py-3 bg-gray-100 text-text-primary rounded-xl font-medium btn-press">取消</button>
          <button @click="doDelete" class="flex-1 py-3 bg-red-500 text-white rounded-xl font-medium btn-press">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { babyAPI } from '@/api'

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const isEdit = computed(() => !!route.params.id)
const loading = ref(false)
const error = ref('')
const showDelete = ref(false)

const colors = ['#FF7EB3', '#4D9DFD', '#7C6CFF', '#43C59E', '#FFD93D', '#FF9F68', '#74B9FF', '#FDA7DF']
const genders = [
  { value: 'male', label: '男孩', emoji: '👦' },
  { value: 'female', label: '女孩', emoji: '👧' },
  { value: '', label: '保密', emoji: '🤷' },
]

const form = reactive({
  name: '',
  birth_date: '',
  gender: '',
  avatar_color: '#7C6CFF',
})

// 选择性别时自动套用对应主题默认头像色
function selectGender(v: string) {
  form.gender = v
  form.avatar_color = app.defaultAvatarColor(v)
}

async function loadBaby() {
  if (!isEdit.value) return
  const id = Number(route.params.id)
  const baby = app.babies.find(b => b.id === id)
  if (baby) {
    form.name = baby.name
    form.birth_date = baby.birth_date ? baby.birth_date.replace('Z', '').slice(0, 16) : ''
    form.gender = baby.gender || ''
    form.avatar_color = baby.avatar_color || '#7C6CFF'
  }
}

async function submit() {
  error.value = ''
  if (!form.name.trim()) { error.value = '请输入宝宝姓名'; return }
  if (!form.birth_date) { error.value = '请选择出生日期'; return }
  // datetime-local 返回格式 YYYY-MM-DDTHH:mm，补上 :00 确保标准 ISO 格式
  const birthDate = form.birth_date.length === 16 ? form.birth_date + ':00' : form.birth_date
  const payload = { ...form, birth_date: birthDate }
  loading.value = true
  try {
    if (isEdit.value) {
      await babyAPI.update(Number(route.params.id), payload)
    } else {
      await babyAPI.create(payload)
    }
    await app.loadBabies()
    app.showToast('✅ 保存成功', 'success')
    router.back()
  } catch (e: any) {
    error.value = e.response?.data?.error || '保存失败'
  } finally {
    loading.value = false
  }
}

function confirmDelete() { showDelete.value = true }

async function doDelete() {
  try {
    await babyAPI.delete(Number(route.params.id))
    await app.loadBabies()
    if (app.currentBabyId === Number(route.params.id)) {
      app.setCurrentBaby(app.babies[0]?.id || 0)
    }
    app.showToast('✅ 已删除', 'success')
    router.replace('/')
  } catch {
    app.showToast('删除失败', 'error')
  }
}

onMounted(loadBaby)
</script>

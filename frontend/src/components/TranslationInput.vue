<script setup lang="ts">
import { ref } from 'vue'
import { useTranslationStore } from '@/stores/translation'

const store = useTranslationStore()
const text = ref('')
const targetLang = ref('es')

const emit = defineEmits<{
  (e: 'submit', text: string, targetLang: string): void
}>()

function handleSubmit() {
  if (!text.value.trim()) return
  emit('submit', text.value, targetLang.value)
  text.value = ''
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="bg-white rounded-lg shadow p-6 space-y-4">
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Text to translate</label>
      <textarea
        v-model="text"
        rows="4"
        class="w-full border rounded-md p-3 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
        placeholder="Enter English text..."
        required
      />
    </div>

    <div class="flex items-center gap-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Target language</label>
        <select v-model="targetLang" class="border rounded-md p-2">
          <option value="es">🇪🇸 Spanish</option>
          <option value="it">🇮🇹 Italian</option>
        </select>
      </div>
      <button
        type="submit"
        class="mt-auto bg-indigo-600 text-white px-6 py-2 rounded-md hover:bg-indigo-700 transition"
        :disabled="store.loading"
      >
        {{ store.loading ? 'Translating...' : 'Translate' }}
      </button>
    </div>
  </form>
</template>

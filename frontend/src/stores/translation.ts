import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export interface VocabularyItem {
  id?: string
  word: string
  translation: string
  partOfSpeech?: string
  difficulty?: string
  example?: string
}

export interface TranslateResponse {
  translatedText: string
  vocabulary: VocabularyItem[]
  cached: boolean
}

export const useTranslationStore = defineStore('translation', () => {
  const result = ref<TranslateResponse | null>(null)
  const vocabulary = ref<VocabularyItem[]>([])
  const loading = ref(false)
  const lastSourceText = ref('')
  const lastTargetLang = ref('')

  async function translate(text: string, targetLang: string) {
    loading.value = true
    lastSourceText.value = text
    lastTargetLang.value = targetLang
    try {
      const response = await axios.post('/api/translate', {
        text,
        target_lang: targetLang,
      })
      const data = response.data
      result.value = {
        translatedText: data.translated_text,
        vocabulary: data.vocabulary,
        cached: data.cached,
      }
      vocabulary.value = data.vocabulary
    } catch (err) {
      console.error('Translation failed:', err)
    } finally {
      loading.value = false
    }
  }

  function clear() {
    result.value = null
    vocabulary.value = []
    lastSourceText.value = ''
    lastTargetLang.value = ''
  }

  return { result, vocabulary, loading, translate, clear, lastSourceText, lastTargetLang }
})
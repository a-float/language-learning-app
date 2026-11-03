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

export interface Lesson {
  id: string
  title: string
  sourceText: string
  translatedText: string
  language: string
  difficulty: string
  sourceType: string
  vocabulary: VocabularyItem[]
  createdAt: string
}

export const useLessonsStore = defineStore('lessons', () => {
  const lessons = ref<Lesson[]>([])
  const currentLesson = ref<Lesson | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get<Lesson[]>('/api/lessons')
      lessons.value = response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load lessons'
      console.error('Failed to load lessons:', err)
    } finally {
      loading.value = false
    }
  }

  async function getLesson(id: string) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get<Lesson>(`/api/lessons/${id}`)
      currentLesson.value = response.data
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load lesson'
      console.error('Failed to load lesson:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createLesson(data: {
    title: string
    sourceText: string
    targetLang: string
    difficulty?: string
    sourceType?: string
  }) {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post<Lesson>('/api/lessons', data)
      lessons.value.unshift(response.data)
      return response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create lesson'
      console.error('Failed to create lesson:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteLesson(id: string) {
    try {
      await axios.delete(`/api/lessons/${id}`)
      lessons.value = lessons.value.filter(l => l.id !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete lesson'
      console.error('Failed to delete lesson:', err)
      throw err
    }
  }

  function clearCurrentLesson() {
    currentLesson.value = null
  }

  return {
    lessons,
    currentLesson,
    loading,
    error,
    load,
    getLesson,
    createLesson,
    deleteLesson,
    clearCurrentLesson,
  }
})
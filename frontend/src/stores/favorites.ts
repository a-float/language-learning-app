import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export interface Favorite {
  id: string
  vocabularyId: string
  word: string
  translation: string
  language: string
  partOfSpeech?: string
  difficulty?: string
  example?: string
  createdAt: string
}

export const useFavoritesStore = defineStore('favorites', () => {
  const items = ref<Favorite[]>([])

  async function load() {
    try {
      const response = await axios.get<Favorite[]>('/api/favorites')
      items.value = response.data
    } catch (err) {
      console.error('Failed to load favorites:', err)
    }
  }

  async function add(vocabularyId: string) {
    try {
      await axios.post('/api/favorites', { vocabulary_id: vocabularyId })
      await load()
    } catch (err) {
      console.error('Failed to add favorite:', err)
    }
  }

  async function remove(id: string) {
    try {
      await axios.delete(`/api/favorites/${id}`)
      await load()
    } catch (err) {
      console.error('Failed to remove favorite:', err)
    }
  }

  return { items, load, add, remove }
})
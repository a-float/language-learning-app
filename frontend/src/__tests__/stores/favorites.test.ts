import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useFavoritesStore } from '@/stores/favorites'

// Mock axios
vi.mock('axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

import axios from 'axios'

describe('useFavoritesStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes with empty items', () => {
    const store = useFavoritesStore()
    expect(store.items).toEqual([])
  })

  it('load fetches favorites from API', async () => {
    const store = useFavoritesStore()
    const mockFavorites = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', createdAt: '2024-01-01T00:00:00Z' },
      { id: 'fav2', vocabularyId: 'vocab2', word: 'world', translation: 'mundo', language: 'es', createdAt: '2024-01-02T00:00:00Z' },
    ]
    ;(axios.get as vi.Mock).mockResolvedValue({ data: mockFavorites })

    await store.load()

    expect(store.items).toEqual(mockFavorites)
    expect(axios.get).toHaveBeenCalledWith('/api/favorites')
  })

  it('load handles API error', async () => {
    const store = useFavoritesStore()
    ;(axios.get as vi.Mock).mockRejectedValue(new Error('Network error'))

    await store.load()

    expect(store.items).toEqual([])
  })

  it('add posts to API and reloads', async () => {
    const store = useFavoritesStore()
    ;(axios.post as vi.Mock).mockResolvedValue({})
    ;(axios.get as vi.Mock).mockResolvedValue({ data: [] })

    await store.add('vocab123')

    expect(axios.post).toHaveBeenCalledWith('/api/favorites', { vocabulary_id: 'vocab123' })
    expect(axios.get).toHaveBeenCalledWith('/api/favorites')
  })

  it('add handles API error', async () => {
    const store = useFavoritesStore()
    ;(axios.post as vi.Mock).mockRejectedValue(new Error('Network error'))

    await store.add('vocab123')

    // Should not throw, just log error
    expect(axios.post).toHaveBeenCalled()
  })

  it('remove deletes from API and reloads', async () => {
    const store = useFavoritesStore()
    ;(axios.delete as vi.Mock).mockResolvedValue({})
    ;(axios.get as vi.Mock).mockResolvedValue({ data: [] })

    await store.remove('fav123')

    expect(axios.delete).toHaveBeenCalledWith('/api/favorites/fav123')
    expect(axios.get).toHaveBeenCalledWith('/api/favorites')
  })

  it('remove handles API error', async () => {
    const store = useFavoritesStore()
    ;(axios.delete as vi.Mock).mockRejectedValue(new Error('Network error'))

    await store.remove('fav123')

    expect(axios.delete).toHaveBeenCalled()
  })

  it('load returns empty array on 404', async () => {
    const store = useFavoritesStore()
    ;(axios.get as vi.Mock).mockRejectedValue({ response: { status: 404 } })

    await store.load()

    expect(store.items).toEqual([])
  })
})
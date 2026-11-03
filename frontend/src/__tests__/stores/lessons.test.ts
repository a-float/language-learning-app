import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useLessonsStore } from '@/stores/lessons'

// Mock axios
vi.mock('axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

import axios from 'axios'

describe('useLessonsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes with correct default state', () => {
    const store = useLessonsStore()
    expect(store.lessons).toEqual([])
    expect(store.currentLesson).toBeNull()
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('load fetches lessons from API', async () => {
    const store = useLessonsStore()
    const mockLessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
      { id: 'lesson2', title: 'Lesson 2', sourceText: 'World', translatedText: 'Mundo', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-02T00:00:00Z' },
    ]
    ;(axios.get as vi.Mock).mockResolvedValue({ data: mockLessons })

    await store.load()

    expect(store.lessons).toEqual(mockLessons)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
    expect(axios.get).toHaveBeenCalledWith('/api/lessons')
  })

  it('load handles API error', async () => {
    const store = useLessonsStore()
    ;(axios.get as vi.Mock).mockRejectedValue(new Error('Network error'))

    await store.load()

    expect(store.lessons).toEqual([])
    expect(store.loading).toBe(false)
    expect(store.error).toBe('Network error')
  })

  it('getLesson fetches single lesson', async () => {
    const store = useLessonsStore()
    const mockLesson = { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' }
    ;(axios.get as vi.Mock).mockResolvedValue({ data: mockLesson })

    const result = await store.getLesson('lesson1')

    expect(result).toEqual(mockLesson)
    expect(store.currentLesson).toEqual(mockLesson)
    expect(axios.get).toHaveBeenCalledWith('/api/lessons/lesson1')
  })

  it('getLesson handles API error', async () => {
    const store = useLessonsStore()
    ;(axios.get as vi.Mock).mockRejectedValue(new Error('Not found'))

    await expect(store.getLesson('lesson1')).rejects.toThrow('Not found')
    expect(store.error).toBe('Not found')
    expect(store.loading).toBe(false)
  })

  it('createLesson posts to API and adds to list', async () => {
    const store = useLessonsStore()
    const newLesson = { id: 'lesson3', title: 'Lesson 3', sourceText: 'Test', translatedText: 'Prueba', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-03T00:00:00Z' }
    ;(axios.post as vi.Mock).mockResolvedValue({ data: newLesson })

    const result = await store.createLesson({
      title: 'Lesson 3',
      sourceText: 'Test',
      targetLang: 'es',
      difficulty: 'beginner',
      sourceType: 'text',
    })

    expect(result).toEqual(newLesson)
    expect(store.lessons).toHaveLength(1)
    expect(store.lessons[0]).toEqual(newLesson)
    expect(axios.post).toHaveBeenCalledWith('/api/lessons', {
      title: 'Lesson 3',
      source_text: 'Test',
      target_lang: 'es',
      difficulty: 'beginner',
      source_type: 'text',
    })
  })

  it('createLesson handles API error', async () => {
    const store = useLessonsStore()
    ;(axios.post as vi.Mock).mockRejectedValue(new Error('Failed to create'))

    await expect(store.createLesson({
      title: 'Lesson 3',
      sourceText: 'Test',
      targetLang: 'es',
    })).rejects.toThrow('Failed to create')
    expect(store.error).toBe('Failed to create')
    expect(store.loading).toBe(false)
  })

  it('deleteLesson removes from list', async () => {
    const store = useLessonsStore()
    store.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
      { id: 'lesson2', title: 'Lesson 2', sourceText: 'World', translatedText: 'Mundo', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-02T00:00:00Z' },
    ]
    ;(axios.delete as vi.Mock).mockResolvedValue({})

    await store.deleteLesson('lesson1')

    expect(store.lessons).toHaveLength(1)
    expect(store.lessons[0].id).toBe('lesson2')
    expect(axios.delete).toHaveBeenCalledWith('/api/lessons/lesson1')
  })

  it('deleteLesson handles API error', async () => {
    const store = useLessonsStore()
    store.lessons = [{ id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' }]
    ;(axios.delete as vi.Mock).mockRejectedValue(new Error('Failed to delete'))

    await expect(store.deleteLesson('lesson1')).rejects.toThrow('Failed to delete')
    expect(store.error).toBe('Failed to delete')
    // Lesson should still be in list on error
    expect(store.lessons).toHaveLength(1)
  })

  it('clearCurrentLesson resets current lesson', () => {
    const store = useLessonsStore()
    store.currentLesson = { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' }

    store.clearCurrentLesson()

    expect(store.currentLesson).toBeNull()
  })

  it('load sets loading state correctly', async () => {
    const store = useLessonsStore()
    let resolveLoad: (value: any) => void
    const loadPromise = new Promise(resolve => { resolveLoad = resolve })
    ;(axios.get as vi.Mock).mockReturnValue(loadPromise)

    const loadPromise2 = store.load()
    expect(store.loading).toBe(true)

    resolveLoad!({ data: [] })
    await loadPromise2

    expect(store.loading).toBe(false)
  })
})
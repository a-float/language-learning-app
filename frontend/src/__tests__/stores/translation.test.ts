import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useTranslationStore } from '@/stores/translation'

// Mock axios
vi.mock('axios', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
  },
}))

import axios from 'axios'

describe('useTranslationStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes with correct default state', () => {
    const store = useTranslationStore()
    expect(store.result).toBeNull()
    expect(store.vocabulary).toEqual([])
    expect(store.loading).toBe(false)
    expect(store.lastSourceText).toBe('')
    expect(store.lastTargetLang).toBe('')
  })

  it('translate sets loading and calls API', async () => {
    const store = useTranslationStore()
    const mockResponse = {
      data: {
        translated_text: 'Hola mundo',
        vocabulary: [
          { word: 'hello', translation: 'hola', part_of_speech: 'interjection', difficulty: 'beginner', example: 'Hello world' },
          { word: 'world', translation: 'mundo', part_of_speech: 'noun', difficulty: 'beginner', example: 'World peace' },
        ],
        cached: false,
      },
    }
    ;(axios.post as vi.Mock).mockResolvedValue(mockResponse)

    await store.translate('Hello world', 'es')

    expect(store.loading).toBe(false)
    expect(store.result).toEqual({
      translatedText: 'Hola mundo',
      vocabulary: mockResponse.data.vocabulary,
      cached: false,
    })
    expect(store.vocabulary).toEqual(mockResponse.data.vocabulary)
    expect(store.lastSourceText).toBe('Hello world')
    expect(store.lastTargetLang).toBe('es')
    expect(axios.post).toHaveBeenCalledWith('/api/translate', {
      text: 'Hello world',
      target_lang: 'es',
    })
  })

  it('translate handles API error', async () => {
    const store = useTranslationStore()
    ;(axios.post as vi.Mock).mockRejectedValue(new Error('Network error'))

    await store.translate('Hello world', 'es')

    expect(store.loading).toBe(false)
    expect(store.result).toBeNull()
    expect(store.vocabulary).toEqual([])
  })

  it('translate with cached response', async () => {
    const store = useTranslationStore()
    const mockResponse = {
      data: {
        translated_text: 'Hola mundo',
        vocabulary: [{ word: 'hello', translation: 'hola' }],
        cached: true,
      },
    }
    ;(axios.post as vi.Mock).mockResolvedValue(mockResponse)

    await store.translate('Hello world', 'es')

    expect(store.result?.cached).toBe(true)
  })

  it('clear resets all state', async () => {
    const store = useTranslationStore()
    const mockResponse = {
      data: {
        translated_text: 'Hola mundo',
        vocabulary: [{ word: 'hello', translation: 'hola' }],
        cached: false,
      },
    }
    ;(axios.post as vi.Mock).mockResolvedValue(mockResponse)

    await store.translate('Hello world', 'es')
    store.clear()

    expect(store.result).toBeNull()
    expect(store.vocabulary).toEqual([])
    expect(store.lastSourceText).toBe('')
    expect(store.lastTargetLang).toBe('')
  })

  it('translate with Italian target language', async () => {
    const store = useTranslationStore()
    const mockResponse = {
      data: {
        translated_text: 'Ciao mondo',
        vocabulary: [{ word: 'hello', translation: 'ciao' }],
        cached: false,
      },
    }
    ;(axios.post as vi.Mock).mockResolvedValue(mockResponse)

    await store.translate('Hello world', 'it')

    expect(store.lastTargetLang).toBe('it')
    expect(axios.post).toHaveBeenCalledWith('/api/translate', {
      text: 'Hello world',
      target_lang: 'it',
    })
  })
})
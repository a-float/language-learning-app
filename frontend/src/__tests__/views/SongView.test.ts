import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import SongView from '@/views/SongView.vue'
import { useLessonsStore } from '@/stores/lessons'

vi.mock('@/stores/lessons', () => ({
  useLessonsStore: vi.fn(),
}))

// Mock speechSynthesis
Object.defineProperty(window, 'speechSynthesis', {
  value: {
    speak: vi.fn(),
    cancel: vi.fn(),
    pause: vi.fn(),
    resume: vi.fn(),
    getVoices: vi.fn().mockReturnValue([]),
    onvoiceschanged: null,
    paused: false,
  },
  writable: true,
})

describe('SongView', () => {
  let mockLessonsStore: any
  let router: any

  beforeEach(() => {
    setActivePinia(createPinia())
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/song/:id', name: 'song', component: { template: '<div />' } },
        { path: '/', name: 'home', component: { template: '<div />' } },
        { path: '/quiz', name: 'quiz', component: { template: '<div />' } },
      ],
    })

    mockLessonsStore = {
      getLesson: vi.fn().mockResolvedValue({
        id: 'lesson1',
        title: 'Test Lesson',
        sourceText: 'Hello world',
        translatedText: 'Hola mundo',
        language: 'es',
        difficulty: 'beginner',
        sourceType: 'text',
        vocabulary: [
          { word: 'hello', translation: 'hola', partOfSpeech: 'interjection', difficulty: 'beginner', example: 'Hello world' },
          { word: 'world', translation: 'mundo', partOfSpeech: 'noun', difficulty: 'beginner', example: 'World peace' },
        ],
        createdAt: '2024-01-01T00:00:00Z',
      }),
    }

    ;(useLessonsStore as vi.Mock).mockReturnValue(mockLessonsStore)
  })

  it('renders header with back button and app name', () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('button:contains("Back")').exists()).toBe(true)
    expect(wrapper.find('h1').text()).toBe('🐾 Mówik')
  })

  it('shows loading state initially', () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('.animate-spin').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading lesson...')
  })

  it('shows error state on load failure', async () => {
    mockLessonsStore.getLesson.mockRejectedValue(new Error('Not found'))

    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Not found')
    expect(wrapper.find('button:contains("Retry")').exists()).toBe(true)
  })

  it('renders lesson content after load', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    expect(wrapper.find('h2').text()).toBe('Test Lesson')
    expect(wrapper.text()).toContain('🇪🇸')
    expect(wrapper.text()).toContain('Spanish')
    expect(wrapper.text()).toContain('beginner')
    expect(wrapper.text()).toContain('text')
  })

  it('shows playback controls', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    expect(wrapper.find('button:contains("Play")').exists()).toBe(true)
    expect(wrapper.find('button:contains("Stop")').exists()).toBe(true)
    expect(wrapper.find('input[type="range"]').exists()).toBe(true)
    expect(wrapper.find('input[type="checkbox"]').exists()).toBe(true)
  })

  it('renders source text with word tokens', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Hello')
    expect(wrapper.text()).toContain('world')
  })

  it('shows vocabulary section', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Key Vocabulary')
    expect(wrapper.text()).toContain('hello')
    expect(wrapper.text()).toContain('hola')
    expect(wrapper.text()).toContain('world')
    expect(wrapper.text()).toContain('mundo')
  })

  it('shows Start Quiz button', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    expect(wrapper.find('button:contains("Start Quiz")').exists()).toBe(true)
  })

  it('navigates to quiz with lesson source on Start Quiz', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    await wrapper.find('button:contains("Start Quiz")').trigger('click')

    expect(router.currentRoute.value.name).toBe('quiz')
    expect(router.currentRoute.value.query.source).toBe('lesson')
    expect(router.currentRoute.value.query.sourceId).toBe('lesson1')
  })

  it('navigates back on Back button', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    await wrapper.find('button:contains("Back")').trigger('click')

    expect(router.currentRoute.value.name).toBe('home')
  })

  it('calls getLesson on mount with route param', () => {
    mount(SongView, {
      global: { plugins: [router] },
    })

    expect(mockLessonsStore.getLesson).toHaveBeenCalledWith('lesson1')
  })

  it('cancels speech on unmount', async () => {
    const wrapper = mount(SongView, {
      global: { plugins: [router] },
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    wrapper.unmount()

    expect(window.speechSynthesis.cancel).toHaveBeenCalled()
  })
})
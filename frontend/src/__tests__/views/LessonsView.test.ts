import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import LessonsView from '@/views/LessonsView.vue'
import { useLessonsStore } from '@/stores/lessons'

vi.mock('@/stores/lessons', () => ({
  useLessonsStore: vi.fn(),
}))

describe('LessonsView', () => {
  let mockLessonsStore: any
  let router: any

  beforeEach(() => {
    setActivePinia(createPinia())
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/lessons', name: 'lessons', component: { template: '<div />' } },
        { path: '/', name: 'home', component: { template: '<div />' } },
        { path: '/song/:id', name: 'song', component: { template: '<div />' } },
        { path: '/quiz', name: 'quiz', component: { template: '<div />' } },
      ],
    })

    mockLessonsStore = {
      lessons: [],
      loading: false,
      error: null,
      load: vi.fn().mockResolvedValue({}),
      deleteLesson: vi.fn().mockResolvedValue({}),
    }

    ;(useLessonsStore as vi.Mock).mockReturnValue(mockLessonsStore)
  })

  it('renders header with app name and back link', () => {
    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('h1').text()).toBe('🐾 Mówik')
    expect(wrapper.find('router-link').text()).toContain('← Home')
  })

  it('shows loading state', () => {
    mockLessonsStore.loading = true

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('.animate-spin').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading lessons...')
  })

  it('shows error state', () => {
    mockLessonsStore.error = 'Failed to load'

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('Failed to load')
    expect(wrapper.find('button:contains("Retry")').exists()).toBe(true)
  })

  it('shows empty state when no lessons', () => {
    mockLessonsStore.lessons = []

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('No lessons yet')
    expect(wrapper.find('router-link:contains("Go Translate")').exists()).toBe(true)
  })

  it('renders lesson list', () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello world', translatedText: 'Hola mundo', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [{}, {}], createdAt: '2024-01-01T00:00:00Z' },
      { id: 'lesson2', title: 'Lesson 2', sourceText: 'Good morning', translatedText: 'Buenos días', language: 'es', difficulty: 'intermediate', sourceType: 'song', vocabulary: [{}], createdAt: '2024-01-02T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    const lessonCards = wrapper.findAll('.bg-white.rounded-lg.shadow.p-6')
    expect(lessonCards).toHaveLength(2)
    expect(lessonCards[0].text()).toContain('Lesson 1')
    expect(lessonCards[1].text()).toContain('Lesson 2')
  })

  it('shows language flag and name', () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
      { id: 'lesson2', title: 'Lesson 2', sourceText: 'Hello', translatedText: 'Ciao', language: 'it', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('🇪🇸')
    expect(wrapper.text()).toContain('Spanish')
    expect(wrapper.text()).toContain('🇮🇹')
    expect(wrapper.text()).toContain('Italian')
  })

  it('shows difficulty and source type badges', () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'advanced', sourceType: 'article', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('advanced')
    expect(wrapper.text()).toContain('article')
  })

  it('shows vocabulary count and date', () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello world', translatedText: 'Hola mundo', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [{}, {}, {}], createdAt: '2024-01-15T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('3 words')
    expect(wrapper.text()).toContain('Jan 15, 2024')
  })

  it('navigates to song view on View button', async () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("View")').trigger('click')
    expect(router.currentRoute.value.name).toBe('song')
    expect(router.currentRoute.value.params.id).toBe('lesson1')
  })

  it('navigates to quiz with lesson source on Quiz button', async () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("Quiz")').trigger('click')
    expect(router.currentRoute.value.name).toBe('quiz')
    expect(router.currentRoute.value.query.source).toBe('lesson')
    expect(router.currentRoute.value.query.sourceId).toBe('lesson1')
  })

  it('deletes lesson on Delete button', async () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    // Mock confirm
    vi.spyOn(window, 'confirm').mockReturnValue(true)

    await wrapper.find('button:contains("Delete")').trigger('click')

    expect(mockLessonsStore.deleteLesson).toHaveBeenCalledWith('lesson1')
  })

  it('does not delete when confirm cancelled', async () => {
    mockLessonsStore.lessons = [
      { id: 'lesson1', title: 'Lesson 1', sourceText: 'Hello', translatedText: 'Hola', language: 'es', difficulty: 'beginner', sourceType: 'text', vocabulary: [], createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(LessonsView, {
      global: { plugins: [router] },
    })

    vi.spyOn(window, 'confirm').mockReturnValue(false)

    await wrapper.find('button:contains("Delete")').trigger('click')

    expect(mockLessonsStore.deleteLesson).not.toHaveBeenCalled()
  })

  it('calls load on mount', () => {
    mount(LessonsView, {
      global: { plugins: [router] },
    })

    expect(mockLessonsStore.load).toHaveBeenCalled()
  })
})
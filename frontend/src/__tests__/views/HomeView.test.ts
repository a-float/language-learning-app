import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import { useTranslationStore } from '@/stores/translation'
import { useLessonsStore } from '@/stores/lessons'
import { useFavoritesStore } from '@/stores/favorites'

// Mock child components
vi.mock('@/components/TranslationInput.vue', () => ({
  default: {
    name: 'TranslationInput',
    template: '<div data-test="translation-input" @submit="onSubmit"><slot /></div>',
    emits: ['submit'],
    props: [],
  },
}))

vi.mock('@/components/TranslationOutput.vue', () => ({
  default: {
    name: 'TranslationOutput',
    template: '<div data-test="translation-output" />',
    props: ['result'],
  },
}))

vi.mock('@/components/VocabularyList.vue', () => ({
  default: {
    name: 'VocabularyList',
    template: '<div data-test="vocabulary-list" @add-favorite="onAddFavorite"><slot /></div>',
    emits: ['add-favorite'],
    props: ['items'],
  },
}))

vi.mock('@/stores/translation', () => ({
  useTranslationStore: vi.fn(),
}))

vi.mock('@/stores/lessons', () => ({
  useLessonsStore: vi.fn(),
}))

vi.mock('@/stores/favorites', () => ({
  useFavoritesStore: vi.fn(),
}))

describe('HomeView', () => {
  let mockTranslationStore: any
  let mockLessonsStore: any
  let mockFavoritesStore: any
  let router: any

  beforeEach(() => {
    setActivePinia(createPinia())
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', name: 'home', component: { template: '<div />' } },
        { path: '/song/:id', name: 'song', component: { template: '<div />' } },
      ],
    })

    mockTranslationStore = {
      result: null,
      vocabulary: [],
      loading: false,
      lastSourceText: '',
      lastTargetLang: '',
      translate: vi.fn(),
    }
    mockLessonsStore = {
      createLesson: vi.fn().mockResolvedValue({ id: 'lesson1' }),
    }
    mockFavoritesStore = {
      add: vi.fn().mockResolvedValue({}),
    }

    ;(useTranslationStore as vi.Mock).mockReturnValue(mockTranslationStore)
    ;(useLessonsStore as vi.Mock).mockReturnValue(mockLessonsStore)
    ;(useFavoritesStore as vi.Mock).mockReturnValue(mockFavoritesStore)
  })

  it('renders header with app name and navigation', () => {
    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('h1').text()).toBe('🐾 Mówik')
    expect(wrapper.find('nav').text()).toContain('Lessons')
    expect(wrapper.find('nav').text()).toContain('Favorites')
    expect(wrapper.find('nav').text()).toContain('Quiz')
  })

  it('renders TranslationInput component', () => {
    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('[data-test="translation-input"]').exists()).toBe(true)
  })

  it('does not show translation result when none', () => {
    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('[data-test="translation-output"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="vocabulary-list"]').exists()).toBe(false)
  })

  it('shows translation result when available', async () => {
    mockTranslationStore.result = {
      translatedText: 'Hola mundo',
      vocabulary: [{ word: 'hello', translation: 'hola' }],
      cached: false,
    }
    mockTranslationStore.vocabulary = [{ word: 'hello', translation: 'hola' }]

    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('[data-test="translation-output"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="vocabulary-list"]').exists()).toBe(true)
  })

  it('shows save as lesson button when result exists', async () => {
    mockTranslationStore.result = {
      translatedText: 'Hola mundo',
      vocabulary: [],
      cached: false,
    }
    mockTranslationStore.lastSourceText = 'Hello world'
    mockTranslationStore.lastTargetLang = 'es'

    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('button:contains("Save as Lesson")').exists()).toBe(true)
  })

  it('opens save dialog when save button clicked', async () => {
    mockTranslationStore.result = {
      translatedText: 'Hola mundo',
      vocabulary: [],
      cached: false,
    }
    mockTranslationStore.lastSourceText = 'Hello world'
    mockTranslationStore.lastTargetLang = 'es'

    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("Save as Lesson")').trigger('click')
    expect(wrapper.find('.fixed.inset-0').exists()).toBe(true)
  })

  it('calls createLesson when saving lesson', async () => {
    mockTranslationStore.result = {
      translatedText: 'Hola mundo',
      vocabulary: [],
      cached: false,
    }
    mockTranslationStore.lastSourceText = 'Hello world'
    mockTranslationStore.lastTargetLang = 'es'

    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("Save as Lesson")').trigger('click')
    await wrapper.find('input').setValue('My Lesson')
    await wrapper.find('button:contains("Save")').trigger('click')

    expect(mockLessonsStore.createLesson).toHaveBeenCalledWith({
      title: 'My Lesson',
      sourceText: 'Hello world',
      targetLang: 'es',
      difficulty: 'beginner',
      sourceType: 'text',
    })
  })

  it('uses default title when empty', async () => {
    mockTranslationStore.result = {
      translatedText: 'Hola mundo',
      vocabulary: [],
      cached: false,
    }
    mockTranslationStore.lastSourceText = 'Hello world'
    mockTranslationStore.lastTargetLang = 'es'

    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("Save as Lesson")').trigger('click')
    await wrapper.find('button:contains("Save")').trigger('click')

    expect(mockLessonsStore.createLesson).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'Untitled Lesson' })
    )
  })

  it('calls favoritesStore.add when adding favorite', async () => {
    mockTranslationStore.result = {
      translatedText: 'Hola mundo',
      vocabulary: [{ id: 'vocab1', word: 'hello', translation: 'hola' }],
      cached: false,
    }
    mockTranslationStore.vocabulary = [{ id: 'vocab1', word: 'hello', translation: 'hola' }]

    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    // Emit add-favorite from VocabularyList
    wrapper.find('[data-test="vocabulary-list"]').vm.$emit('add-favorite', 'vocab1', { word: 'hello', translation: 'hola' })

    expect(mockFavoritesStore.add).toHaveBeenCalledWith('vocab1')
  })

  it('navigates to song view after saving lesson', async () => {
    mockTranslationStore.result = {
      translatedText: 'Hola mundo',
      vocabulary: [],
      cached: false,
    }
    mockTranslationStore.lastSourceText = 'Hello world'
    mockTranslationStore.lastTargetLang = 'es'

    const wrapper = mount(HomeView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("Save as Lesson")').trigger('click')
    await wrapper.find('button:contains("Save")').trigger('click')

    await wrapper.vm.$nextTick()
    expect(router.currentRoute.value.name).toBe('song')
    expect(router.currentRoute.value.params.id).toBe('lesson1')
  })
})
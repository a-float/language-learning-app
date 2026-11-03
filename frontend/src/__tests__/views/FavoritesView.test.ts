import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import FavoritesView from '@/views/FavoritesView.vue'
import { useFavoritesStore } from '@/stores/favorites'

vi.mock('@/stores/favorites', () => ({
  useFavoritesStore: vi.fn(),
}))

describe('FavoritesView', () => {
  let mockFavoritesStore: any
  let router: any

  beforeEach(() => {
    setActivePinia(createPinia())
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/favorites', name: 'favorites', component: { template: '<div />' } },
        { path: '/', name: 'home', component: { template: '<div />' } },
        { path: '/quiz', name: 'quiz', component: { template: '<div />' } },
      ],
    })

    mockFavoritesStore = {
      items: [],
      loading: false,
      error: null,
      load: vi.fn().mockResolvedValue({}),
      remove: vi.fn().mockResolvedValue({}),
    }

    ;(useFavoritesStore as vi.Mock).mockReturnValue(mockFavoritesStore)
  })

  it('renders header with app name and back link', () => {
    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('h1').text()).toBe('🐾 Mówik')
    expect(wrapper.find('router-link').text()).toContain('← Home')
  })

  it('shows loading state', () => {
    mockFavoritesStore.loading = true

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('.animate-spin').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading favorites...')
  })

  it('shows error state', () => {
    mockFavoritesStore.error = 'Failed to load'

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('Failed to load')
    expect(wrapper.find('button:contains("Retry")').exists()).toBe(true)
  })

  it('shows empty state when no favorites', () => {
    mockFavoritesStore.items = []

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('No favorites yet')
    expect(wrapper.find('router-link:contains("Start Translating")').exists()).toBe(true)
  })

  it('renders favorites list', () => {
    mockFavoritesStore.items = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', partOfSpeech: 'interjection', difficulty: 'beginner', example: 'Hello world', createdAt: '2024-01-01T00:00:00Z' },
      { id: 'fav2', vocabularyId: 'vocab2', word: 'world', translation: 'mundo', language: 'es', partOfSpeech: 'noun', difficulty: 'beginner', example: 'World peace', createdAt: '2024-01-02T00:00:00Z' },
    ]

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    const favCards = wrapper.findAll('.bg-white.rounded-lg.shadow.p-4')
    expect(favCards).toHaveLength(2)
    expect(favCards[0].text()).toContain('hello')
    expect(favCards[1].text()).toContain('world')
  })

  it('shows language flag and name', () => {
    mockFavoritesStore.items = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', createdAt: '2024-01-01T00:00:00Z' },
      { id: 'fav2', vocabularyId: 'vocab2', word: 'world', translation: 'mondo', language: 'it', createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('🇪🇸')
    expect(wrapper.text()).toContain('Spanish')
    expect(wrapper.text()).toContain('🇮🇹')
    expect(wrapper.text()).toContain('Italian')
  })

  it('shows part of speech and difficulty when present', () => {
    mockFavoritesStore.items = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', partOfSpeech: 'interjection', difficulty: 'advanced', example: 'Hello world', createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('interjection')
    expect(wrapper.text()).toContain('advanced')
  })

  it('shows example when present', () => {
    mockFavoritesStore.items = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', example: 'Hello world', createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.text()).toContain('"Hello world"')
  })

  it('shows Quiz Favorites button when favorites exist', () => {
    mockFavoritesStore.items = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('button:contains("Quiz Favorites")').exists()).toBe(true)
  })

  it('hides Quiz Favorites button when no favorites', () => {
    mockFavoritesStore.items = []

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('button:contains("Quiz Favorites")').exists()).toBe(false)
  })

  it('navigates to quiz with favorites source on Quiz Favorites button', async () => {
    mockFavoritesStore.items = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("Quiz Favorites")').trigger('click')
    expect(router.currentRoute.value.name).toBe('quiz')
    expect(router.currentRoute.value.query.source).toBe('favorites')
  })

  it('removes favorite on heart button click', async () => {
    mockFavoritesStore.items = [
      { id: 'fav1', vocabularyId: 'vocab1', word: 'hello', translation: 'hola', language: 'es', createdAt: '2024-01-01T00:00:00Z' },
    ]

    const wrapper = mount(FavoritesView, {
      global: { plugins: [router] },
    })

    await wrapper.find('button:contains("♡")').trigger('click')

    expect(mockFavoritesStore.remove).toHaveBeenCalledWith('fav1')
  })

  it('calls load on mount', () => {
    mount(FavoritesView, {
      global: { plugins: [router] },
    })

    expect(mockFavoritesStore.load).toHaveBeenCalled()
  })
})
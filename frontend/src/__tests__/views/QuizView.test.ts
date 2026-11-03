import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import QuizView from '@/views/QuizView.vue'
import { useFavoritesStore } from '@/stores/favorites'
import { useLessonsStore } from '@/stores/lessons'

vi.mock('@/stores/favorites', () => ({
  useFavoritesStore: vi.fn(),
}))

vi.mock('@/stores/lessons', () => ({
  useLessonsStore: vi.fn(),
}))

// Mock fetch
global.fetch = vi.fn()

describe('QuizView', () => {
  let mockFavoritesStore: any
  let mockLessonsStore: any
  let router: any

  beforeEach(() => {
    setActivePinia(createPinia())
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/quiz', name: 'quiz', component: { template: '<div />' } },
        { path: '/', name: 'home', component: { template: '<div />' } },
      ],
    })

    mockFavoritesStore = {
      items: [],
    }
    mockLessonsStore = {
      lessons: [
        { id: 'lesson1', title: 'Lesson 1', language: 'es' },
        { id: 'lesson2', title: 'Lesson 2', language: 'it' },
      ],
      load: vi.fn().mockResolvedValue({}),
    }

    ;(useFavoritesStore as vi.Mock).mockReturnValue(mockFavoritesStore)
    ;(useLessonsStore as vi.Mock).mockReturnValue(mockLessonsStore)

    vi.clearAllMocks()
    ;(global.fetch as vi.Mock).mockReset()
  })

  it('renders header with back button and app name', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('button:contains("Back")').exists()).toBe(true)
    expect(wrapper.find('h1').text()).toBe('🐾 Mówik Quiz')
  })

  it('shows setup screen initially', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    expect(wrapper.find('h2').text()).toBe('Configure Quiz')
    expect(wrapper.find('select[v-model="quizConfig.source"]').exists()).toBe(true)
    expect(wrapper.find('button:contains("Start Quiz")').exists()).toBe(true)
  })

  it('has all source options', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    const options = wrapper.findAll('option')
    const sourceOptions = options.filter(o => o.element.value === 'mixed' || o.element.value === 'favorites' || o.element.value === 'progress' || o.element.value === 'lesson')
    expect(sourceOptions).toHaveLength(4)
  })

  it('shows lesson selector when source is lesson', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    await wrapper.find('select[v-model="quizConfig.source"]').setValue('lesson')

    expect(wrapper.find('select[v-model="quizConfig.sourceId"]').exists()).toBe(true)
    expect(wrapper.find('option:contains("Lesson 1")').exists()).toBe(true)
    expect(wrapper.find('option:contains("Lesson 2")').exists()).toBe(true)
  })

  it('has target language options', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    const langSelect = wrapper.find('select[v-model="quizConfig.targetLang"]')
    const options = langSelect.findAll('option')
    expect(options).toHaveLength(3)
    expect(options[0].element.value).toBe('')
    expect(options[1].element.value).toBe('es')
    expect(options[2].element.value).toBe('it')
  })

  it('has question count input with min/max', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    const input = wrapper.find('input[type="number"]')
    expect(input.attributes('min')).toBe('5')
    expect(input.attributes('max')).toBe('50')
  })

  it('has question type options', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    const typeSelect = wrapper.find('select[v-model="quizConfig.questionType"]')
    const options = typeSelect.findAll('option')
    expect(options).toHaveLength(3)
    expect(options[0].element.value).toBe('mixed')
    expect(options[1].element.value).toBe('multiple_choice')
    expect(options[2].element.value).toBe('type_answer')
  })

  it('has direction options', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    const dirSelect = wrapper.find('select[v-model="quizConfig.direction"]')
    const options = dirSelect.findAll('option')
    expect(options).toHaveLength(3)
    expect(options[0].element.value).toBe('mixed')
    expect(options[1].element.value).toBe('en_to_target')
    expect(options[2].element.value).toBe('target_to_en')
  })

  it('generates quiz on button click', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        questions: [
          { id: 'q1', vocabulary_id: 'v1', question: 'hello', question_lang: 'en', answer: 'hola', answer_lang: 'es', options: ['hola', 'mundo', 'bueno', 'malo'], question_type: 'multiple_choice', direction: 'en_to_target' },
        ],
      }),
    })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')

    expect(global.fetch).toHaveBeenCalledWith('/api/quiz', expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }))
  })

  it('shows active quiz screen after generation', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        questions: [
          { id: 'q1', vocabulary_id: 'v1', question: 'hello', question_lang: 'en', answer: 'hola', answer_lang: 'es', options: ['hola', 'mundo', 'bueno', 'malo'], question_type: 'multiple_choice', direction: 'en_to_target' },
        ],
      }),
    })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('.bg-white.rounded-lg.shadow.p-8').exists()).toBe(true)
    expect(wrapper.find('div:contains("Question 1 of 1")').exists()).toBe(true)
    expect(wrapper.find('div:contains("hello")').exists()).toBe(true)
  })

  it('shows multiple choice options', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        questions: [
          { id: 'q1', vocabulary_id: 'v1', question: 'hello', question_lang: 'en', answer: 'hola', answer_lang: 'es', options: ['hola', 'mundo', 'bueno', 'malo'], question_type: 'multiple_choice', direction: 'en_to_target' },
        ],
      }),
    })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')
    await wrapper.vm.$nextTick()

    const optionButtons = wrapper.findAll('button')
    expect(optionButtons.length).toBeGreaterThanOrEqual(4)
  })

  it('shows type answer input for type_answer questions', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        questions: [
          { id: 'q1', vocabulary_id: 'v1', question: 'hello', question_lang: 'en', answer: 'hola', answer_lang: 'es', options: [], question_type: 'type_answer', direction: 'en_to_target' },
        ],
      }),
    })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('input[type="text"]').exists()).toBe(true)
    expect(wrapper.find('button:contains("Submit")').exists()).toBe(true)
  })

  it('shows feedback after answer', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        questions: [
          { id: 'q1', vocabulary_id: 'v1', question: 'hello', question_lang: 'en', answer: 'hola', answer_lang: 'es', options: ['hola', 'mundo'], question_type: 'multiple_choice', direction: 'en_to_target' },
        ],
      }),
    })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')
    await wrapper.vm.$nextTick()

    // Click first option
    await wrapper.findAll('button')[0].trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('div:contains("Correct!")').exists()).toBe(true)
    expect(wrapper.find('div:contains("The answer was: hola")').exists()).toBe(true)
  })

  it('shows results screen after quiz completion', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          questions: [
            { id: 'q1', vocabulary_id: 'v1', question: 'hello', question_lang: 'en', answer: 'hola', answer_lang: 'es', options: ['hola', 'mundo'], question_type: 'multiple_choice', direction: 'en_to_target' },
          ],
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          score: 1,
          total: 1,
          percentage: 100,
          results: [{ question_id: 'q1', vocabulary_id: 'v1', user_answer: 'hola', correct_answer: 'hola', is_correct: true, time_taken_ms: 1000 }],
          xp_earned: 15,
        }),
      })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')
    await wrapper.vm.$nextTick()

    await wrapper.findAll('button')[0].trigger('click')
    await wrapper.vm.$nextTick()

    // Wait for auto-advance and finish
    await new Promise(r => setTimeout(r, 1600))
    await wrapper.vm.$nextTick()

    expect(wrapper.find('h2:contains("Quiz Complete")').exists()).toBe(true)
    expect(wrapper.text()).toContain('100%')
    expect(wrapper.text()).toContain('XP earned')
  })

  it('restarts quiz on Try Another Quiz', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          questions: [{ id: 'q1', vocabulary_id: 'v1', question: 'hello', question_lang: 'en', answer: 'hola', answer_lang: 'es', options: ['hola', 'mundo'], question_type: 'multiple_choice', direction: 'en_to_target' }],
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          score: 1, total: 1, percentage: 100, results: [], xp_earned: 15,
        }),
      })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')
    await wrapper.vm.$nextTick()
    await wrapper.findAll('button')[0].trigger('click')
    await new Promise(r => setTimeout(r, 1600))
    await wrapper.vm.$nextTick()

    await wrapper.find('button:contains("Try Another Quiz")').trigger('click')

    expect(wrapper.find('h2').text()).toBe('Configure Quiz')
  })

  it('navigates back on Back button', () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    wrapper.find('button:contains("Back")').trigger('click')

    expect(router.currentRoute.value.name).toBe('home')
  })

  it('shows error on quiz generation failure', async () => {
    const wrapper = mount(QuizView, {
      global: { plugins: [router] },
    })

    ;(global.fetch as vi.Mock).mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ message: 'No vocabulary available' }),
    })

    await wrapper.find('button:contains("Start Quiz")').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('No vocabulary available')
  })

  it('loads lessons on mount', () => {
    mount(QuizView, {
      global: { plugins: [router] },
    })

    expect(mockLessonsStore.load).toHaveBeenCalled()
  })
})
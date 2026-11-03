import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import TranslationInput from '@/components/TranslationInput.vue'
import { createPinia, setActivePinia } from 'pinia'
import { useTranslationStore } from '@/stores/translation'

vi.mock('@/stores/translation', () => ({
  useTranslationStore: vi.fn(),
}))

describe('TranslationInput', () => {
  let mockStore: any

  beforeEach(() => {
    setActivePinia(createPinia())
    mockStore = {
      loading: false,
      translate: vi.fn(),
    }
    ;(useTranslationStore as vi.Mock).mockReturnValue(mockStore)
  })

  it('renders form with textarea and select', () => {
    const wrapper = mount(TranslationInput)

    expect(wrapper.find('textarea').exists()).toBe(true)
    expect(wrapper.find('select').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })

  it('has correct placeholder and label', () => {
    const wrapper = mount(TranslationInput)

    expect(wrapper.find('textarea').attributes('placeholder')).toBe('Enter English text...')
    expect(wrapper.find('label').text()).toContain('Text to translate')
  })

  it('has Spanish and Italian options', () => {
    const wrapper = mount(TranslationInput)

    const options = wrapper.findAll('option')
    expect(options).toHaveLength(2)
    expect(options[0].attributes('value')).toBe('es')
    expect(options[0].text()).toContain('Spanish')
    expect(options[1].attributes('value')).toBe('it')
    expect(options[1].text()).toContain('Italian')
  })

  it('defaults to Spanish', () => {
    const wrapper = mount(TranslationInput)

    expect(wrapper.find('select').element.value).toBe('es')
  })

  it('emits submit event with text and targetLang on submit', async () => {
    const wrapper = mount(TranslationInput)

    await wrapper.find('textarea').setValue('Hello world')
    await wrapper.find('select').setValue('it')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.emitted('submit')).toBeTruthy()
    expect(wrapper.emitted('submit')?.[0]).toEqual(['Hello world', 'it'])
  })

  it('clears textarea after submit', async () => {
    const wrapper = mount(TranslationInput)

    await wrapper.find('textarea').setValue('Hello world')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.find('textarea').element.value).toBe('')
  })

  it('does not emit submit when text is empty', async () => {
    const wrapper = mount(TranslationInput)

    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.emitted('submit')).toBeFalsy()
  })

  it('does not emit submit when text is only whitespace', async () => {
    const wrapper = mount(TranslationInput)

    await wrapper.find('textarea').setValue('   ')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.emitted('submit')).toBeFalsy()
  })

  it('disables button when store.loading is true', async () => {
    mockStore.loading = true
    const wrapper = mount(TranslationInput)

    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
    expect(wrapper.find('button').text()).toBe('Translating...')
  })

  it('enables button when store.loading is false', () => {
    mockStore.loading = false
    const wrapper = mount(TranslationInput)

    expect(wrapper.find('button').attributes('disabled')).toBeUndefined()
    expect(wrapper.find('button').text()).toBe('Translate')
  })

  it('shows loading state in button text', async () => {
    const wrapper = mount(TranslationInput)
    expect(wrapper.find('button').text()).toBe('Translate')

    mockStore.loading = true
    await wrapper.vm.$nextTick()
    expect(wrapper.find('button').text()).toBe('Translating...')
  })
})
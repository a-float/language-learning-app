import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import TranslationOutput from '@/components/TranslationOutput.vue'

describe('TranslationOutput', () => {
  const mockResult = {
    translatedText: 'Hola mundo',
    vocabulary: [
      { word: 'hello', translation: 'hola' },
      { word: 'world', translation: 'mundo' },
    ],
    cached: false,
  }

  it('renders translation result', () => {
    const wrapper = mount(TranslationOutput, {
      props: { result: mockResult },
    })

    expect(wrapper.find('h2').text()).toBe('Translation')
    expect(wrapper.find('p.text-lg').text()).toBe('Hola mundo')
  })

  it('shows cached indicator when cached is true', () => {
    const wrapper = mount(TranslationOutput, {
      props: { result: { ...mockResult, cached: true } },
    })

    expect(wrapper.find('div.text-amber-600').text()).toBe('⚡ Cached result')
  })

  it('hides cached indicator when cached is false', () => {
    const wrapper = mount(TranslationOutput, {
      props: { result: mockResult },
    })

    expect(wrapper.find('div.text-amber-600').exists()).toBe(false)
  })

  it('shows source label', () => {
    const wrapper = mount(TranslationOutput, {
      props: { result: mockResult },
    })

    expect(wrapper.find('p.text-sm.text-gray-500').text()).toBe('Source')
  })

  it('renders with different translated text', () => {
    const wrapper = mount(TranslationOutput, {
      props: { result: { ...mockResult, translatedText: 'Ciao mondo' } },
    })

    expect(wrapper.find('p.text-lg').text()).toBe('Ciao mondo')
  })

  it('renders empty translated text', () => {
    const wrapper = mount(TranslationOutput, {
      props: { result: { ...mockResult, translatedText: '' } },
    })

    expect(wrapper.find('p.text-lg').text()).toBe('')
  })

  it('has correct structure', () => {
    const wrapper = mount(TranslationOutput, {
      props: { result: mockResult },
    })

    expect(wrapper.find('div.bg-white.rounded-lg.shadow.p-6').exists()).toBe(true)
    expect(wrapper.find('div.space-y-2').exists()).toBe(true)
  })
})
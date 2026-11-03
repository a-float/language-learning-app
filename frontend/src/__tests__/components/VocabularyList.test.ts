import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import VocabularyList from '@/components/VocabularyList.vue'

describe('VocabularyList', () => {
  const mockItems = [
    { id: 'vocab1', word: 'hello', translation: 'hola', partOfSpeech: 'interjection', example: 'Hello world' },
    { id: 'vocab2', word: 'world', translation: 'mundo', partOfSpeech: 'noun', example: 'World peace' },
    { id: 'vocab3', word: 'good', translation: 'bueno', partOfSpeech: 'adjective' },
  ]

  it('renders vocabulary items', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: mockItems },
    })

    expect(wrapper.find('h2').text()).toBe('Extracted Vocabulary')
    expect(wrapper.findAll('li')).toHaveLength(3)
  })

  it('displays word and translation', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: mockItems },
    })

    const items = wrapper.findAll('li')
    expect(items[0].find('span.font-medium').text()).toBe('hello')
    expect(items[0].find('span.text-indigo-600').text()).toBe('hola')
    expect(items[1].find('span.font-medium').text()).toBe('world')
    expect(items[1].find('span.text-indigo-600').text()).toBe('mundo')
  })

  it('displays part of speech when present', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: mockItems },
    })

    const items = wrapper.findAll('li')
    expect(items[0].find('p.text-xs.text-gray-500').text()).toBe('interjection')
    expect(items[1].find('p.text-xs.text-gray-500').text()).toBe('noun')
  })

  it('displays example when present', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: mockItems },
    })

    const items = wrapper.findAll('li')
    expect(items[0].find('p.text-sm.italic.text-gray-600').text()).toBe('"Hello world"')
    expect(items[1].find('p.text-sm.italic.text-gray-600').text()).toBe('"World peace"')
  })

  it('does not display part of speech when absent', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: [{ id: 'vocab3', word: 'good', translation: 'bueno' }] },
    })

    expect(wrapper.find('p.text-xs.text-gray-500').exists()).toBe(false)
  })

  it('does not display example when absent', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: [{ id: 'vocab3', word: 'good', translation: 'bueno' }] },
    })

    expect(wrapper.find('p.text-sm.italic.text-gray-600').exists()).toBe(false)
  })

  it('shows favorite button when item has id', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: mockItems },
    })

    const buttons = wrapper.findAll('button')
    expect(buttons).toHaveLength(3)
    expect(buttons[0].text()).toBe('♡')
    expect(buttons[1].text()).toBe('♡')
    expect(buttons[2].text()).toBe('♡')
  })

  it('emits add-favorite event when button clicked', async () => {
    const wrapper = mount(VocabularyList, {
      props: { items: mockItems },
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')

    expect(wrapper.emitted('add-favorite')).toBeTruthy()
    expect(wrapper.emitted('add-favorite')?.[0]).toEqual(['vocab1', { word: 'hello', translation: 'hola' }])
  })

  it('does not render when items array is empty', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: [] },
    })

    expect(wrapper.find('div.bg-white').exists()).toBe(false)
    expect(wrapper.find('h2').exists()).toBe(false)
  })

  it('renders single item correctly', () => {
    const wrapper = mount(VocabularyList, {
      props: { items: [{ id: 'vocab1', word: 'hello', translation: 'hola' }] },
    })

    expect(wrapper.findAll('li')).toHaveLength(1)
    expect(wrapper.find('span.font-medium').text()).toBe('hello')
    expect(wrapper.find('span.text-indigo-600').text()).toBe('hola')
  })
})
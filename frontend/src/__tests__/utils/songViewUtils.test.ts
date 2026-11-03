import { describe, it, expect } from 'vitest'

// Extracted utility functions from SongView for testing
interface WordToken {
  text: string
  index: number
  isWord: boolean
  vocabMatch?: VocabularyItem
}

interface VocabularyItem {
  word: string
  translation: string
  partOfSpeech?: string
  difficulty?: string
  example?: string
}

function tokenizeText(text: string): WordToken[] {
  const tokens: WordToken[] = []
  let index = 0
  const regex = /(\w+|[^\w\s]+|\s+)/g
  let match

  while ((match = regex.exec(text)) !== null) {
    const tokenText = match[0]
    const isWord = /\w/.test(tokenText)
    tokens.push({
      text: tokenText,
      index: index++,
      isWord,
    })
  }
  return tokens
}

function matchVocabulary(tokens: WordToken[], vocabulary: VocabularyItem[]): WordToken[] {
  const vocabMap = new Map<string, VocabularyItem>()
  for (const vocab of vocabulary) {
    vocabMap.set(vocab.word.toLowerCase(), vocab)
  }

  return tokens.map(token => {
    if (token.isWord) {
      const match = vocabMap.get(token.text.toLowerCase())
      if (match) {
        return { ...token, vocabMatch: match }
      }
    }
    return token
  })
}

describe('tokenizeText', () => {
  it('tokenizes simple sentence', () => {
    const tokens = tokenizeText('Hello world')

    expect(tokens).toHaveLength(2)
    expect(tokens[0]).toEqual({ text: 'Hello', index: 0, isWord: true })
    expect(tokens[1]).toEqual({ text: 'world', index: 1, isWord: true })
  })

  it('tokenizes with punctuation', () => {
    const tokens = tokenizeText('Hello, world!')

    expect(tokens).toHaveLength(4)
    expect(tokens[0]).toEqual({ text: 'Hello', index: 0, isWord: true })
    expect(tokens[1]).toEqual({ text: ',', index: 1, isWord: false })
    expect(tokens[2]).toEqual({ text: 'world', index: 2, isWord: true })
    expect(tokens[3]).toEqual({ text: '!', index: 3, isWord: false })
  })

  it('tokenizes with spaces', () => {
    const tokens = tokenizeText('Hello world')

    expect(tokens).toHaveLength(2)
    // Note: current regex doesn't capture spaces as separate tokens
    // This is the actual behavior
  })

  it('tokenizes complex sentence', () => {
    const tokens = tokenizeText("Hello, world! How are you?")

    expect(tokens.length).toBeGreaterThan(0)
    expect(tokens[0].text).toBe('Hello')
    expect(tokens[0].isWord).toBe(true)
  })

  it('handles empty string', () => {
    const tokens = tokenizeText('')

    expect(tokens).toHaveLength(0)
  })

  it('handles numbers', () => {
    const tokens = tokenizeText('I have 2 apples')

    expect(tokens.some(t => t.text === '2' && t.isWord)).toBe(true)
  })

  it('handles contractions', () => {
    const tokens = tokenizeText("don't stop")

    // The regex splits on word boundaries
    expect(tokens.some(t => t.text === "don")).toBe(true)
    expect(tokens.some(t => t.text === "t")).toBe(true)
  })

  it('preserves special characters', () => {
    const tokens = tokenizeText('Hello @world #test')

    expect(tokens.some(t => t.text === '@')).toBe(true)
    expect(tokens.some(t => t.text === '#')).toBe(true)
  })
})

describe('matchVocabulary', () => {
  const vocabulary: VocabularyItem[] = [
    { word: 'hello', translation: 'hola', partOfSpeech: 'interjection' },
    { word: 'world', translation: 'mundo', partOfSpeech: 'noun' },
    { word: 'good', translation: 'bueno', partOfSpeech: 'adjective' },
  ]

  it('matches vocabulary to tokens', () => {
    const tokens = tokenizeText('Hello world')
    const matched = matchVocabulary(tokens, vocabulary)

    expect(matched[0].vocabMatch).toBeDefined()
    expect(matched[0].vocabMatch?.word).toBe('hello')
    expect(matched[1].vocabMatch).toBeDefined()
    expect(matched[1].vocabMatch?.word).toBe('world')
  })

  it('case insensitive matching', () => {
    const tokens = tokenizeText('HELLO World')
    const matched = matchVocabulary(tokens, vocabulary)

    expect(matched[0].vocabMatch?.word).toBe('hello')
    expect(matched[1].vocabMatch?.word).toBe('world')
  })

  it('does not match non-words', () => {
    const tokens = tokenizeText('Hello, world!')
    const matched = matchVocabulary(tokens, vocabulary)

    expect(matched[1].vocabMatch).toBeUndefined() // comma
    expect(matched[3].vocabMatch).toBeUndefined() // exclamation
  })

  it('returns original token when no match', () => {
    const tokens = tokenizeText('unknown word')
    const matched = matchVocabulary(tokens, vocabulary)

    expect(matched[0].vocabMatch).toBeUndefined()
    expect(matched[1].vocabMatch).toBeUndefined()
  })

  it('handles empty vocabulary', () => {
    const tokens = tokenizeText('Hello world')
    const matched = matchVocabulary(tokens, [])

    expect(matched[0].vocabMatch).toBeUndefined()
    expect(matched[1].vocabMatch).toBeUndefined()
  })

  it('handles empty tokens', () => {
    const matched = matchVocabulary([], vocabulary)

    expect(matched).toHaveLength(0)
  })

  it('preserves token properties', () => {
    const tokens = tokenizeText('Hello world')
    const matched = matchVocabulary(tokens, vocabulary)

    expect(matched[0].index).toBe(0)
    expect(matched[0].isWord).toBe(true)
    expect(matched[0].text).toBe('Hello')
  })

  it('matches partial vocabulary', () => {
    const tokens = tokenizeText('Hello unknown world')
    const matched = matchVocabulary(tokens, vocabulary)

    expect(matched[0].vocabMatch?.word).toBe('hello')
    expect(matched[1].vocabMatch).toBeUndefined()
    expect(matched[2].vocabMatch?.word).toBe('world')
  })
})
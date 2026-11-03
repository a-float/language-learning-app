import { describe, it, expect } from 'vitest'

// Extracted utility functions from QuizView for testing
interface QuizQuestion {
  id: string
  vocabulary_id: string
  question: string
  question_lang: string
  answer: string
  answer_lang: string
  options: string[]
  question_type: 'multiple_choice' | 'type_answer'
  direction: 'en_to_target' | 'target_to_en'
}

function checkAnswer(userAnswer: string, correctAnswer: string): boolean {
  return userAnswer.trim().toLowerCase() === correctAnswer.trim().toLowerCase()
}

function getLangFlag(lang: string): string {
  if (lang === 'es') return '🇪🇸'
  if (lang === 'it') return '🇮🇹'
  if (lang === 'en') return '🇬🇧'
  return '🌐'
}

function getDirectionLabel(direction: string): string {
  if (direction === 'en_to_target') return 'English → Target'
  if (direction === 'target_to_en') return 'Target → English'
  return 'Mixed'
}

function calculateProgress(currentIndex: number, total: number): number {
  if (total === 0) return 0
  return ((currentIndex + 1) / total) * 100
}

function isLastQuestion(currentIndex: number, total: number): boolean {
  return currentIndex === total - 1
}

describe('checkAnswer', () => {
  it('returns true for exact match', () => {
    expect(checkAnswer('hola', 'hola')).toBe(true)
  })

  it('returns true for case insensitive match', () => {
    expect(checkAnswer('HOLA', 'hola')).toBe(true)
    expect(checkAnswer('Hola', 'hola')).toBe(true)
    expect(checkAnswer('hola', 'HOLA')).toBe(true)
  })

  it('returns true for whitespace trimmed match', () => {
    expect(checkAnswer('  hola  ', 'hola')).toBe(true)
    expect(checkAnswer('hola', '  hola  ')).toBe(true)
    expect(checkAnswer('  hola  ', '  hola  ')).toBe(true)
  })

  it('returns false for different answer', () => {
    expect(checkAnswer('hola', 'mundo')).toBe(false)
    expect(checkAnswer('hello', 'hola')).toBe(false)
  })

  it('returns false for partial match', () => {
    expect(checkAnswer('hol', 'hola')).toBe(false)
    expect(checkAnswer('hola mundo', 'hola')).toBe(false)
  })

  it('handles empty strings', () => {
    expect(checkAnswer('', '')).toBe(true)
    expect(checkAnswer(' ', '')).toBe(true)
    expect(checkAnswer('', 'hola')).toBe(false)
  })

  it('handles unicode', () => {
    expect(checkAnswer('niño', 'niño')).toBe(true)
    expect(checkAnswer('NIÑO', 'niño')).toBe(true)
  })
})

describe('getLangFlag', () => {
  it('returns Spanish flag for es', () => {
    expect(getLangFlag('es')).toBe('🇪🇸')
  })

  it('returns Italian flag for it', () => {
    expect(getLangFlag('it')).toBe('🇮🇹')
  })

  it('returns British flag for en', () => {
    expect(getLangFlag('en')).toBe('🇬🇧')
  })

  it('returns globe for unknown language', () => {
    expect(getLangFlag('fr')).toBe('🌐')
    expect(getLangFlag('de')).toBe('🌐')
    expect(getLangFlag('')).toBe('🌐')
  })
})

describe('getDirectionLabel', () => {
  it('returns English → Target for en_to_target', () => {
    expect(getDirectionLabel('en_to_target')).toBe('English → Target')
  })

  it('returns Target → English for target_to_en', () => {
    expect(getDirectionLabel('target_to_en')).toBe('Target → English')
  })

  it('returns Mixed for mixed', () => {
    expect(getDirectionLabel('mixed')).toBe('Mixed')
  })

  it('returns Mixed for unknown direction', () => {
    expect(getDirectionLabel('unknown')).toBe('Mixed')
    expect(getDirectionLabel('')).toBe('Mixed')
  })
})

describe('calculateProgress', () => {
  it('calculates progress for first question', () => {
    expect(calculateProgress(0, 10)).toBe(10)
  })

  it('calculates progress for last question', () => {
    expect(calculateProgress(9, 10)).toBe(100)
  })

  it('calculates progress for middle question', () => {
    expect(calculateProgress(4, 10)).toBe(50)
  })

  it('handles single question', () => {
    expect(calculateProgress(0, 1)).toBe(100)
  })

  it('returns 0 for zero total', () => {
    expect(calculateProgress(0, 0)).toBe(0)
  })

  it('handles fractional progress', () => {
    expect(calculateProgress(1, 3)).toBeCloseTo(66.67, 1)
  })
})

describe('isLastQuestion', () => {
  it('returns true for last question', () => {
    expect(isLastQuestion(9, 10)).toBe(true)
  })

  it('returns false for non-last question', () => {
    expect(isLastQuestion(0, 10)).toBe(false)
    expect(isLastQuestion(5, 10)).toBe(false)
  })

  it('returns true for single question', () => {
    expect(isLastQuestion(0, 1)).toBe(true)
  })

  it('returns false for zero total', () => {
    expect(isLastQuestion(0, 0)).toBe(false)
  })
})
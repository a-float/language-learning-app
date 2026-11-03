<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLessonsStore } from '@/stores/lessons'

const route = useRoute()
const router = useRouter()
const lessonsStore = useLessonsStore()

const lesson = ref<Lesson | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const currentWordIndex = ref(0)
const isPlaying = ref(false)
const speechRate = ref(1)
const selectedVoice = ref<SpeechSynthesisVoice | null>(null)
const voices = ref<SpeechSynthesisVoice[]>([])
const utterance = ref<SpeechSynthesisUtterance | null>(null)
const highlightMode = ref<'word' | 'sentence'>('word')
const showTranslation = ref(false)

interface Lesson {
  id: string
  title: string
  sourceText: string
  translatedText: string
  language: string
  difficulty: string
  sourceType: string
  vocabulary: VocabularyItem[]
  createdAt: string
}

interface VocabularyItem {
  word: string
  translation: string
  partOfSpeech?: string
  difficulty?: string
  example?: string
}

interface WordToken {
  text: string
  index: number
  isWord: boolean
  vocabMatch?: VocabularyItem
}

const sourceWords = computed(() => {
  if (!lesson.value) return []
  const tokens = tokenizeText(lesson.value.sourceText)
  return matchVocabulary(tokens, lesson.value.vocabulary)
})

const translatedWords = computed(() => {
  if (!lesson.value) return []
  return tokenizeText(lesson.value.translatedText)
})

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

async function loadLesson() {
  const id = route.params.id as string
  if (!id) {
    error.value = 'Lesson ID required'
    loading.value = false
    return
  }

  try {
    loading.value = true
    const data = await lessonsStore.getLesson(id)
    lesson.value = data
    loadVoices()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load lesson'
  } finally {
    loading.value = false
  }
}

function loadVoices() {
  if ('speechSynthesis' in window) {
    const updateVoices = () => {
      voices.value = speechSynthesis.getVoices().filter(v => 
        lesson.value && (v.lang.startsWith(lesson.value.language === 'es' ? 'es' : 'it') || v.lang.startsWith('en'))
      )
      if (voices.value.length && !selectedVoice.value) {
        selectedVoice.value = voices.value[0]
      }
    }
    updateVoices()
    speechSynthesis.onvoiceschanged = updateVoices
  }
}

function speakText(text: string, lang: string) {
  if (!('speechSynthesis' in window)) return

  speechSynthesis.cancel()
  
  utterance.value = new SpeechSynthesisUtterance(text)
  utterance.value.lang = lang === 'es' ? 'es-ES' : 'it-IT'
  utterance.value.rate = speechRate.value
  if (selectedVoice.value) {
    utterance.value.voice = selectedVoice.value
  }

  utterance.value.onend = () => {
    isPlaying.value = false
    currentWordIndex.value = 0
  }

  utterance.value.onboundary = (event) => {
    if (event.name === 'word' && highlightMode.value === 'word') {
      const charIndex = event.charIndex
      let wordIdx = 0
      let charCount = 0
      for (const token of sourceWords.value) {
        if (token.isWord) {
          charCount += token.text.length
          if (charCount >= charIndex) break
          wordIdx++
        }
      }
      currentWordIndex.value = Math.min(wordIdx, sourceWords.value.length - 1)
    }
  }

  isPlaying.value = true
  speechSynthesis.speak(utterance.value)
}

function playPause() {
  if (isPlaying.value) {
    speechSynthesis.pause()
    isPlaying.value = false
  } else if (utterance.value && speechSynthesis.paused) {
    speechSynthesis.resume()
    isPlaying.value = true
  } else if (lesson.value) {
    speakText(lesson.value.sourceText, lesson.value.language)
  }
}

function stop() {
  speechSynthesis.cancel()
  isPlaying.value = false
  currentWordIndex.value = 0
}

function speakWord(word: string) {
  if (!('speechSynthesis' in window)) return
  const u = new SpeechSynthesisUtterance(word)
  u.lang = lesson.value?.language === 'es' ? 'es-ES' : 'it-IT'
  u.rate = 0.8
  speechSynthesis.speak(u)
}

function getWordClass(token: WordToken, index: number): string {
  const base = 'px-1 py-0.5 rounded transition-colors cursor-pointer'
  if (!token.isWord) return base
  if (index === currentWordIndex.value && isPlaying.value) {
    return `${base} bg-indigo-200 text-indigo-900 font-medium`
  }
  if (token.vocabMatch) {
    return `${base} hover:bg-indigo-100 border-b-2 border-indigo-400`
  }
  return `${base} hover:bg-gray-100`
}

function getTranslatedWordClass(token: WordToken, index: number): string {
  const base = 'px-1 py-0.5 rounded transition-colors'
  if (!token.isWord) return base
  if (index === currentWordIndex.value && isPlaying.value) {
    return `${base} bg-indigo-200 text-indigo-900 font-medium`
  }
  return `${base} hover:bg-gray-100`
}

async function addToFavorites(vocab: VocabularyItem) {
  if (!vocab.id) return
  try {
    await fetch('/api/favorites', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ vocabulary_id: vocab.id })
    })
  } catch (err) {
    console.error('Failed to add favorite:', err)
  }
}

function goBack() {
  router.push('/')
}

onMounted(() => {
  loadLesson()
})

onUnmounted(() => {
  speechSynthesis.cancel()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-indigo-600 text-white p-4">
      <div class="max-w-4xl mx-auto flex items-center justify-between">
        <button @click="goBack" class="text-white hover:opacity-80">
          ← Back
        </button>
        <h1 class="text-xl font-bold">🐾 Mówik</h1>
        <div class="w-20"></div>
      </div>
    </header>

    <main class="max-w-4xl mx-auto p-6 space-y-6">
      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        <p class="mt-2 text-gray-500">Loading lesson...</p>
      </div>

      <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 p-4 rounded-lg">
        {{ error }}
        <button @click="loadLesson" class="ml-4 text-indigo-600 hover:underline">Retry</button>
      </div>

      <div v-else-if="lesson" class="space-y-6">
        <!-- Lesson Header -->
        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h2 class="text-2xl font-bold text-gray-900">{{ lesson.title }}</h2>
              <div class="flex items-center gap-4 mt-2 text-sm text-gray-500">
                <span class="flex items-center gap-1">
                  <span v-if="lesson.language === 'es'">🇪🇸</span>
                  <span v-else>🇮🇹</span>
                  {{ lesson.language === 'es' ? 'Spanish' : 'Italian' }}
                </span>
                <span class="px-2 py-0.5 bg-gray-100 rounded">{{ lesson.difficulty || 'beginner' }}</span>
                <span class="px-2 py-0.5 bg-gray-100 rounded capitalize">{{ lesson.sourceType || 'text' }}</span>
              </div>
            </div>
          </div>

          <!-- Playback Controls -->
          <div class="flex items-center gap-4 p-4 bg-gray-50 rounded-lg">
            <button
              @click="playPause"
              :disabled="isPlaying && speechSynthesis.paused === false"
              class="bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50"
            >
              {{ isPlaying && !speechSynthesis.paused ? '⏸ Pause' : '▶ Play' }}
            </button>
            <button
              @click="stop"
              :disabled="!isPlaying && currentWordIndex === 0"
              class="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-300 transition disabled:opacity-50"
            >
              ⏹ Stop
            </button>

            <div class="flex-1 max-w-xs">
              <label class="block text-sm text-gray-600 mb-1">Speed</label>
              <input
                type="range"
                v-model.number="speechRate"
                min="0.5"
                max="2"
                step="0.1"
                class="w-full"
              />
              <span class="text-sm text-gray-500">{{ speechRate.toFixed(1) }}x</span>
            </div>

            <div v-if="voices.length" class="flex-1 max-w-xs">
              <label class="block text-sm text-gray-600 mb-1">Voice</label>
              <select v-model="selectedVoice" class="w-full border rounded-md p-2 text-sm">
                <option v-for="voice in voices" :key="voice.name" :value="voice">
                  {{ voice.name }} ({{ voice.lang }})
                </option>
              </select>
            </div>

            <div class="flex items-center gap-2">
              <label class="flex items-center gap-2 text-sm">
                <input type="checkbox" v-model="highlightMode" true-value="word" false-value="sentence" class="rounded" />
                Word highlight
              </label>
              <label class="flex items-center gap-2 text-sm">
                <input type="checkbox" v-model="showTranslation" class="rounded" />
                Show translation
              </label>
            </div>
          </div>
        </div>

        <!-- Source Text with Word Highlighting -->
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <span v-if="lesson.language === 'es'">🇪🇸</span>
            <span v-else>🇮🇹</span>
            Original Text
          </h3>
          <div class="prose max-w-none leading-relaxed text-lg" style="word-break: break-word;">
            <span
              v-for="(token, index) in sourceWords"
              :key="index"
              :class="getWordClass(token, index)"
              @click="token.isWord && speakWord(token.text)"
              title="Click to hear pronunciation"
            >
              {{ token.text }}
            </span>
          </div>
        </div>

        <!-- Translation -->
        <div v-if="showTranslation" class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <span>🇬🇧</span>
            Translation
          </h3>
          <div class="prose max-w-none leading-relaxed text-lg" style="word-break: break-word;">
            <span
              v-for="(token, index) in translatedWords"
              :key="index"
              :class="getTranslatedWordClass(token, index)"
            >
              {{ token.text }}
            </span>
          </div>
        </div>

        <!-- Vocabulary -->
        <div v-if="lesson.vocabulary.length" class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-semibold mb-4">Key Vocabulary</h3>
          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="(vocab, i) in lesson.vocabulary"
              :key="i"
              class="border rounded-lg p-4 hover:shadow-md transition"
            >
              <div class="flex justify-between items-start mb-2">
                <span class="font-medium text-indigo-700">{{ vocab.word }}</span>
                <button
                  @click="addToFavorites(vocab)"
                  class="text-gray-400 hover:text-red-500 transition"
                  title="Add to favorites"
                >
                  ♡
                </button>
              </div>
              <p class="text-indigo-600 font-medium">{{ vocab.translation }}</p>
              <p v-if="vocab.partOfSpeech" class="text-xs text-gray-500 mt-1">{{ vocab.partOfSpeech }}</p>
              <p v-if="vocab.example" class="text-sm italic text-gray-600 mt-2">"{{ vocab.example }}"</p>
            </div>
          </div>
        </div>

        <!-- Quiz Button -->
        <div class="text-center">
          <button
            @click="$router.push({ name: 'quiz', query: { source: 'lesson', sourceId: lesson.id } })"
            class="bg-indigo-600 text-white px-8 py-3 rounded-lg hover:bg-indigo-700 transition text-lg"
          >
            🎯 Start Quiz
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.prose {
  line-height: 1.8;
}
</style>
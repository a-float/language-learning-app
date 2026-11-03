<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useFavoritesStore } from '@/stores/favorites'
import { useLessonsStore } from '@/stores/lessons'

const route = useRoute()
const router = useRouter()
const favoritesStore = useFavoritesStore()
const lessonsStore = useLessonsStore()

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

interface QuizAnswer {
  question_id: string
  user_answer: string
  is_correct: boolean
  time_taken_ms: number
}

interface QuizResult {
  question_id: string
  vocabulary_id: string
  user_answer: string
  correct_answer: string
  is_correct: boolean
  time_taken_ms: number
}

interface QuizResultResponse {
  score: number
  total: number
  percentage: number
  results: QuizResult[]
  xp_earned: number
}

const quizQuestions = ref<QuizQuestion[]>([])
const currentQuestionIndex = ref(0)
const userAnswers = ref<QuizAnswer[]>([])
const quizState = ref<'setup' | 'active' | 'review' | 'results'>('setup')
const loading = ref(false)
const error = ref<string | null>(null)
const showFeedback = ref(false)
const feedbackCorrect = ref(false)
const questionStartTime = ref<number>(0)

const quizConfig = ref({
  source: 'mixed' as 'favorites' | 'progress' | 'lesson' | 'mixed',
  sourceId: '',
  questionCount: 10,
  questionType: 'mixed' as 'multiple_choice' | 'type_answer' | 'mixed',
  direction: 'mixed' as 'en_to_target' | 'target_to_en' | 'mixed',
  targetLang: '' as 'es' | 'it' | '',
})

const currentQuestion = computed(() => quizQuestions.value[currentQuestionIndex.value])
const progress = computed(() => ((currentQuestionIndex.value + 1) / quizQuestions.value.length) * 100)
const isLastQuestion = computed(() => currentQuestionIndex.value === quizQuestions.value.length - 1)
const answeredQuestions = computed(() => userAnswers.value.length)

async function generateQuiz() {
  loading.value = true
  error.value = null

  try {
    const params = new URLSearchParams()
    params.append('source', quizConfig.value.source)
    if (quizConfig.value.sourceId) params.append('source_id', quizConfig.value.sourceId)
    params.append('question_count', quizConfig.value.questionCount.toString())
    params.append('question_type', quizConfig.value.questionType)
    params.append('direction', quizConfig.value.direction)
    if (quizConfig.value.targetLang) params.append('target_lang', quizConfig.value.targetLang)

    const response = await fetch('/api/quiz', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        source: quizConfig.value.source,
        source_id: quizConfig.value.sourceId,
        question_count: quizConfig.value.questionCount,
        question_type: quizConfig.value.questionType,
        direction: quizConfig.value.direction,
        target_lang: quizConfig.value.targetLang,
      }),
    })

    if (!response.ok) {
      const err = await response.json()
      throw new Error(err.message || 'Failed to generate quiz')
    }

    const data = await response.json()
    quizQuestions.value = data.questions
    currentQuestionIndex.value = 0
    userAnswers.value = []
    quizState.value = 'active'
    questionStartTime.value = Date.now()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to generate quiz'
  } finally {
    loading.value = false
  }
}

function submitAnswer(answer: string) {
  if (!currentQuestion.value) return

  const timeTaken = Date.now() - questionStartTime.value
  const isCorrect = checkAnswer(answer, currentQuestion.value.answer)

  userAnswers.value.push({
    question_id: currentQuestion.value.id,
    user_answer: answer,
    is_correct: isCorrect,
    time_taken_ms: timeTaken,
  })

  feedbackCorrect.value = isCorrect
  showFeedback.value = true

  // Auto-advance after short delay for multiple choice
  if (currentQuestionType.value === 'multiple_choice') {
    setTimeout(() => {
      showFeedback.value = false
      nextQuestion()
    }, 1500)
  } else {
    // For type_answer, wait for user to click next
  }
}

function checkAnswer(userAnswer: string, correctAnswer: string): boolean {
  return userAnswer.trim().toLowerCase() === correctAnswer.trim().toLowerCase()
}

function nextQuestion() {
  if (isLastQuestion.value) {
    finishQuiz()
  } else {
    currentQuestionIndex.value++
    questionStartTime.value = Date.now()
    showFeedback.value = false
  }
}

async function finishQuiz() {
  loading.value = true
  try {
    const response = await fetch('/api/quiz/submit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ answers: userAnswers.value }),
    })

    if (!response.ok) throw new Error('Failed to submit quiz')

    const result: QuizResultResponse = await response.json()
    quizResult.value = result
    quizState.value = 'results'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to submit quiz'
    quizState.value = 'results' // Show local results anyway
  } finally {
    loading.value = false
  }
}

function restartQuiz() {
  quizState.value = 'setup'
  quizQuestions.value = []
  currentQuestionIndex.value = 0
  userAnswers.value = []
  quizResult.value = null
}

function goBack() {
  router.push('/')
}

const quizResult = ref<QuizResultResponse | null>(null)
const typeAnswer = ref('')

const getLangFlag = (lang: string) => {
  if (lang === 'es') return '🇪🇸'
  if (lang === 'it') return '🇮🇹'
  if (lang === 'en') return '🇬🇧'
  return '🌐'
}

const getDirectionLabel = (direction: string) => {
  if (direction === 'en_to_target') return 'English → Target'
  if (direction === 'target_to_en') return 'Target → English'
  return 'Mixed'
}

const currentQuestionLang = computed(() => currentQuestion.value?.question_lang || '')
const currentQuestionType = computed(() => currentQuestion.value?.question_type || '')
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-indigo-600 text-white p-4">
      <div class="max-w-3xl mx-auto flex items-center justify-between">
        <button @click="goBack" class="text-white hover:opacity-80">
          ← Back
        </button>
        <h1 class="text-xl font-bold">🐾 Mówik Quiz</h1>
        <div class="w-20"></div>
      </div>
    </header>

    <main class="max-w-3xl mx-auto p-6 space-y-6">
      <!-- Setup Screen -->
      <div v-if="quizState === 'setup'" class="bg-white rounded-lg shadow p-6 space-y-6">
        <h2 class="text-2xl font-bold text-center">Configure Quiz</h2>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Source</label>
            <select v-model="quizConfig.source" class="w-full border rounded-md p-3 focus:ring-2 focus:ring-indigo-500">
              <option value="mixed">🎲 Mixed (all vocabulary)</option>
              <option value="favorites">❤️ Favorites only</option>
              <option value="progress">📚 Study progress (due for review)</option>
              <option value="lesson">📖 Specific lesson</option>
            </select>
          </div>

          <div v-if="quizConfig.source === 'lesson'">
            <label class="block text-sm font-medium text-gray-700 mb-1">Lesson</label>
            <select v-model="quizConfig.sourceId" class="w-full border rounded-md p-3 focus:ring-2 focus:ring-indigo-500">
              <option value="">Select a lesson...</option>
              <option v-for="lesson in lessonsStore.lessons" :key="lesson.id" :value="lesson.id">
                {{ lesson.title }} ({{ lesson.language }})
              </option>
            </select>
            <button @click="lessonsStore.load()" class="text-sm text-indigo-600 hover:underline mt-1">
              Refresh lessons
            </button>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Target Language</label>
            <select v-model="quizConfig.targetLang" class="w-full border rounded-md p-3 focus:ring-2 focus:ring-indigo-500">
              <option value="">All languages</option>
              <option value="es">🇪🇸 Spanish</option>
              <option value="it">🇮🇹 Italian</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Number of Questions</label>
            <input
              type="number"
              v-model.number="quizConfig.questionCount"
              min="5"
              max="50"
              class="w-full border rounded-md p-3 focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Question Type</label>
            <select v-model="quizConfig.questionType" class="w-full border rounded-md p-3 focus:ring-2 focus:ring-indigo-500">
              <option value="mixed">🎲 Mixed</option>
              <option value="multiple_choice">☑️ Multiple Choice</option>
              <option value="type_answer">⌨️ Type Answer</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Direction</label>
            <select v-model="quizConfig.direction" class="w-full border rounded-md p-3 focus:ring-2 focus:ring-indigo-500">
              <option value="mixed">🎲 Mixed</option>
              <option value="en_to_target">English → Target</option>
              <option value="target_to_en">Target → English</option>
            </select>
          </div>
        </div>

        <button
          @click="generateQuiz"
          :disabled="loading"
          class="w-full bg-indigo-600 text-white py-3 rounded-lg hover:bg-indigo-700 transition text-lg disabled:opacity-50"
        >
          {{ loading ? 'Generating...' : 'Start Quiz' }}
        </button>

        <p v-if="error" class="text-red-600 text-center">{{ error }}</p>
      </div>

      <!-- Active Quiz Screen -->
      <div v-else-if="quizState === 'active'" class="space-y-6">
        <!-- Progress Bar -->
        <div class="bg-white rounded-lg shadow p-4">
          <div class="flex justify-between text-sm mb-2">
            <span>Question {{ currentQuestionIndex + 1 }} of {{ quizQuestions.length }}</span>
            <span>{{ Math.round(progress) }}%</span>
          </div>
          <div class="h-2 bg-gray-200 rounded-full overflow-hidden">
            <div
              class="h-full bg-indigo-600 transition-all duration-300"
              :style="{ width: progress + '%' }"
            ></div>
          </div>
        </div>

        <!-- Question Card -->
        <div class="bg-white rounded-lg shadow p-8">
          <div class="mb-6">
            <span class="text-2xl">{{ getLangFlag(currentQuestionLang) }}</span>
            <span class="ml-2 text-sm text-gray-500">{{ getDirectionLabel(currentQuestion.value?.direction) }}</span>
          </div>

          <div class="text-3xl font-medium text-gray-900 mb-8 text-center">
            {{ currentQuestion.value?.question }}
          </div>

          <!-- Multiple Choice -->
          <div v-if="currentQuestionType === 'multiple_choice' && currentQuestion.value?.options?.length" class="space-y-3">
            <button
              v-for="(option, i) in currentQuestion.options"
              :key="i"
              @click="submitAnswer(option)"
              :disabled="showFeedback"
              :class="[
                'w-full text-left p-4 rounded-lg border-2 transition',
                showFeedback
                  ? (option === currentQuestion.answer ? 'border-green-500 bg-green-50' : 'border-red-500 bg-red-50')
                  : 'border-gray-200 hover:border-indigo-300 hover:bg-indigo-50'
              ]"
            >
              {{ option }}
            </button>
          </div>

          <!-- Type Answer -->
          <div v-else-if="currentQuestionType === 'type_answer'" class="space-y-4">
            <input
              v-model="typeAnswer"
              @keyup.enter="submitAnswer(typeAnswer)"
              :disabled="showFeedback"
              type="text"
              class="w-full p-4 border-2 rounded-lg text-xl focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Type your answer..."
              autofocus
            />
            <button
              @click="submitAnswer(typeAnswer)"
              :disabled="showFeedback || !typeAnswer.trim()"
              class="w-full bg-indigo-600 text-white py-3 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50"
            >
              Submit
            </button>
          </div>

          <!-- Feedback -->
          <div v-if="showFeedback" class="mt-6 p-4 rounded-lg" :class="feedbackCorrect ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'">
            <div class="flex items-center gap-3">
              <span class="text-2xl">{{ feedbackCorrect ? '✅' : '❌' }}</span>
              <div>
                <p class="font-medium" :class="feedbackCorrect ? 'text-green-800' : 'text-red-800'">
                  {{ feedbackCorrect ? 'Correct!' : 'Not quite right' }}
                </p>
                <p class="text-sm mt-1" :class="feedbackCorrect ? 'text-green-700' : 'text-red-700'">
                  The answer was: <strong>{{ currentQuestion?.answer }}</strong>
                </p>
              </div>
            </div>
            <button
              v-if="!isLastQuestion"
              @click="nextQuestion"
              class="mt-4 w-full bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 transition"
            >
              Next Question
            </button>
            <button
              v-else
              @click="finishQuiz"
              class="mt-4 w-full bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 transition"
            >
              See Results
            </button>
          </div>
        </div>
      </div>

      <!-- Results Screen -->
      <div v-else-if="quizState === 'results'" class="space-y-6">
        <div class="bg-white rounded-lg shadow p-8 text-center">
          <h2 class="text-3xl font-bold mb-4">Quiz Complete!</h2>
          
          <div class="grid grid-cols-3 gap-4 mb-8">
            <div class="p-4 bg-indigo-50 rounded-lg">
              <p class="text-3xl font-bold text-indigo-600">{{ quizResult?.score || userAnswers.filter(a => a.is_correct).length }}</p>
              <p class="text-sm text-indigo-700">Correct</p>
            </div>
            <div class="p-4 bg-gray-50 rounded-lg">
              <p class="text-3xl font-bold text-gray-600">{{ quizResult?.total || quizQuestions.length }}</p>
              <p class="text-sm text-gray-600">Total</p>
            </div>
            <div class="p-4 bg-green-50 rounded-lg">
              <p class="text-3xl font-bold text-green-600">{{ quizResult?.percentage?.toFixed(1) || ((userAnswers.filter(a => a.is_correct).length / quizQuestions.length) * 100).toFixed(1) }}%</p>
              <p class="text-sm text-green-700">Score</p>
            </div>
          </div>

          <div v-if="quizResult?.xp_earned" class="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg">
            <p class="text-amber-800 font-medium">+{{ quizResult.xp_earned }} XP earned! 🎉</p>
          </div>

          <button @click="restartQuiz" class="bg-indigo-600 text-white px-8 py-3 rounded-lg hover:bg-indigo-700 transition">
            Try Another Quiz
          </button>
        </div>

        <!-- Detailed Results -->
        <div v-if="quizResult?.results?.length || userAnswers.length" class="bg-white rounded-lg shadow">
          <div class="p-4 border-b">
            <h3 class="font-semibold">Question Review</h3>
          </div>
          <div class="divide-y">
            <div
              v-for="(result, i) in (quizResult?.results || userAnswers)"
              :key="i"
              class="p-4 hover:bg-gray-50"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium">{{ result.question_id }}</span>
                <span :class="result.is_correct ? 'text-green-600' : 'text-red-600'" class="font-medium">
                  {{ result.is_correct ? '✅ Correct' : '❌ Incorrect' }}
                </span>
              </div>
              <div class="text-sm text-gray-600 space-y-1">
                <p>Your answer: <span :class="result.is_correct ? 'text-green-700' : 'text-red-700'">{{ result.user_answer }}</span></p>
                <p v-if="!result.is_correct">Correct: <span class="text-green-700">{{ result.correct_answer }}</span></p>
                <p class="text-gray-400">{{ result.time_taken_ms }}ms</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
input:focus {
  outline: none;
}
</style>
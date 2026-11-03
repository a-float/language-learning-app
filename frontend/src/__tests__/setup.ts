import { vi } from 'vitest'

// Mock axios globally
vi.mock('axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
  },
}))

// Mock router globally
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    go: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    currentRoute: { value: { name: 'home', params: {}, query: {} } },
  }),
  useRoute: () => ({
    params: {},
    query: {},
    path: '/',
    name: 'home',
  }),
  createRouter: vi.fn(),
  createWebHistory: vi.fn(),
}))

// Mock Pinia
vi.mock('pinia', () => ({
  createPinia: vi.fn(),
  setActivePinia: vi.fn(),
  defineStore: vi.fn((id, setup) => {
    const store = setup()
    return () => store
  }),
  storeToRefs: vi.fn((store) => store),
}))

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))
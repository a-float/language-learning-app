# Language Learning App — System Design

## 1. Overview

A web-based language learning application with a SPA frontend and a lightweight backend. Users submit blocks of text (up to a length limit); the app translates it, extracts vocabulary, and generates lessons. Includes favorites, speech/listening practice via browser TTS, and LLM-powered translation with DB caching.

---

## 2. Requirements

### 2.1 Functional Requirements

| ID | Requirement |
|----|-------------|
| FR-1 | User can browse/select words and sentences and add them to a personal favorites list |
| FR-2 | User can submit any block of text (up to a length limit); the app translates it and extracts vocabulary into structured lessons |
| FR-3 | User can enter a listening/speech mode that uses browser TTS APIs for pronunciation practice |
| FR-4 | User can review favorites as flashcards or quiz sessions |
| FR-5 | Lessons are tagged by source media, difficulty, and vocabulary |
| FR-6 | User progress (favorites, completed lessons, scores) is persisted per user |

### 2.2 Non-Functional Requirements

| ID | Requirement | Rationale |
|----|-------------|-----------|
| NFR-1 | Backend memory footprint < 100 MB under normal load | User explicitly wants lightweight |
| NFR-2 | Backend startup time < 2 s | Fast iteration |
| NFR-3 | SPA loads in < 2 s on broadband | UX baseline |
| NFR-4 | System should support 1–100 concurrent users initially | Solo dev → small community scale |
| NFR-5 | Containerizable via Docker | Consistent dev/prod environments |
| NFR-6 | Type-safe across frontend and shared contracts | Reduce runtime errors |

---

## 3. Architecture

```
┌──────────────────────────────────────────────────────────┐
│  Browser (SPA Frontend)                                  │
│  ┌────────────────────────────────────────────────────┐  │
│  │ UI Framework (TBD — React / Vue / Svelte / Lit)   │  │
│  │ ┌──────────┐ ┌──────────┐ ┌────────────────────┐  │  │
│  │ │ Favorites│ │ Lesson   │ │ Listening/Speech   │  │  │
│  │ │ View     │ │ Viewer   │ │ Mode (TTS)         │  │  │
│  │ └──────────┘ └──────────┘ └────────────────────┘  │  │
│  └────────────────────────┬───────────────────────────┘  │
│                           │ REST / JSON over HTTPS      │
└───────────────────────────┼──────────────────────────────┘
                            │
┌───────────────────────────▼──────────────────────────────┐
│  Backend (Go or Rust — lightweight HTTP server)          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐  │
│  │ Auth     │ │ Lesson   │ │ Favorite │ │ Translation│  │
│  │ Service  │ │ Generator│ │ Service  │ │ Service    │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └─────┬──────┘  │
│       │            │            │              │          │
│  ┌────▼────────────▼────────────▼──────────────▼──────┐  │
│  │ Data Layer: SQLite (dev) / PostgreSQL (prod)       │  │
│  │ + Translation Cache                                │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│       ┌────────────────────────────┐                      │
│       │ LLM API (external)         │                      │
│       └────────────────────────────┘                      │
└──────────────────────────────────────────────────────────┘
```

### 3.1 Key Design Decisions (provisional)

| Decision | Proposed | Status |
|----------|----------|--------|
| Backend language | Go (Gin) or Rust (Axum) | See §8 for both options |
| Frontend UI framework | React 19 + TypeScript + Vite | See §8 |
| Database | SQLite (dev), PostgreSQL (prod) | Proposed |
| API style | REST/JSON | Proposed |
| TTS | Browser-native `SpeechSynthesis` API | Decided (no server TTS needed) |
| Lesson generation | Rule-based extraction + LLM for definitions/examples | See §8 |
| Auth | JWT-based session tokens | Confirmed — see §9 |
| Target languages | English → Spanish, English → Italian | Confirmed — see §9 |
| LLM provider | OpenAI GPT-4o-mini (default), Ollama (optional) | See §9-B |
| Cache key | FNV-1a hash (non-cryptographic) | See §10 |

---

## 4. Feature Breakdown

### 4.1 Favorites Mode

- User selects any word/phrase from any lesson
- POST `/api/favorites` — add vocabulary item to user's favorites
- GET `/api/favorites` — list with pagination
- DELETE `/api/favorites/:id` — remove
- Favorites feed into review/quiz mode
- Backed by `favorites` table referencing `vocabulary.id`

### 4.2 Text Input & Translation

- User submits any block of text (up to a length limit)
- Backend flow:
  1. Check translation cache — if cached, return immediately
  2. Otherwise, call LLM API to translate text (EN → ES/IT)
  3. Extract vocabulary: for each word/phrase, create or update `vocabulary` row with translation
  4. Create `lesson` row + `lesson_vocabulary` junction rows
  5. Store full-text translation in `translation_cache` for reuse
- Core model: every word/phrase is a `vocabulary` row (word → translation)

### 4.3 Translation Cache

- Full-text translations cached in `translation_cache` table
- Cache key: FNV-1a hash of (source text + target lang)
- Avoids re-calling LLM for identical text submissions
- Word/phrase translations live in `vocabulary` table (no separate cache needed)
- TTL or manual invalidation

### 4.4 Listening / Speech Mode

- Renders lesson text or favorite sentences in the browser
- Uses `SpeechSynthesisUtterance` (browser TTS API) for pronunciation
- User can:
  - Play/pause/stop TTS
  - Adjust speed and voice
  - Highlight current word being spoken (word-level sync if supported)
- No server-side TTS required — all client-side
- Data model: `session_id`, `content_text`, `settings` (speed, voice, locale), `progress` (word index)

---

## 5. Data Model (Simplified)

```
users
  id (PK)
  username (unique)
  email (unique)
  password_hash
  created_at

Core unit: **word/phrase → translation** (every row is a vocabulary entry).

```
users
  id (PK)
  username (unique)
  email (unique)
  password_hash
  created_at

vocabulary  ← core table: every word/phrase is a row
  id (PK)
  word (TEXT — source language)
  translation (TEXT — target language)
  language (target: es|it)
  part_of_speech
  frequency_rank
  difficulty (beginner|intermediate|advanced)
  example_sentence (TEXT — context)
  created_at

favorites
  id (PK)
  user_id (FK → users)
  vocabulary_id (FK → vocabulary)
  created_at

lessons
  id (PK)
  title
  source_text (TEXT — original text submitted)
  translated_text (TEXT — LLM translation)
  language (target: es|it)
  difficulty (beginner|intermediate|advanced)
  created_at
  user_id (FK → users, nullable — shared lessons?)

lesson_vocabulary  ← junction: which words belong to which lesson
  id (PK)
  lesson_id (FK → lessons)
  vocabulary_id (FK → vocabulary)
  position_in_lesson (INT)

translation_cache  ← full-text LLM cache (avoids re-calling for same text)
  id (PK)
  source_hash (FNV-1a of source text + target lang)
  source_text (TEXT)
  target_lang (es|it)
  translated_text (TEXT)
  tokens_used (INT — LLM cost tracking)
  created_at
  accessed_at

speech_sessions
  id (PK)
  user_id (FK → users)
  content_text
  language
  speed
  voice_name
  completed_at
```

---

## 6. Backend Language Comparison: Go vs Rust

| Criterion | Go | Rust |
|-----------|----|----|
| Memory usage | ~20–50 MB baseline | ~5–20 MB baseline |
| Startup time | ~50–200 ms | ~10–100 ms |
| Ecosystem for web | Excellent (std net/http, Gin, Echo) | Good (Axum, Actix, Rocket) |
| Learning curve | Low | High |
| Async model | Goroutines (green threads) | async/await (zero-cost) |
| JSON handling | Built-in `encoding/json` | Serde (excellent) |
| SQLite support | mattn/go-sqlite3 (CGO dep) | rusqlite (pure Rust) |
| Dev speed | Fast | Slower (borrow checker) |
| Binary size | ~10–20 MB | ~2–5 MB (static) |

**Recommendation**: Go if developer velocity matters most; Rust if absolute minimal footprint and binary size are priorities. Both are viable.

---

## 7. API Surface (Draft)

```
# Auth
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh

# Favorites
GET    /api/favorites?page=1&limit=20
POST   /api/favorites
DELETE /api/favorites/:id

# Lessons
GET    /api/lessons
POST   /api/lessons              # submit text → translate → generate lesson
GET    /api/lessons/:id
DELETE /api/lessons/:id

# Translation Cache
GET    /api/cache/stats          # cache hit rate, tokens used (admin)
DELETE /api/cache/entries        # clear stale translations

# Speech
GET    /api/voices?lang=en-US    # available TTS voices (proxy to browser or cache)

# User
GET    /api/users/me
PUT    /api/users/me
```

---

## 8. Recommended Technology Stack

### Backend

| Layer | Go Option | Rust Option |
|-------|-----------|-------------|
| HTTP framework | **Gin** (minimal, fast) | **Axum** (Tokio-native, type-safe) |
| Database | **sqlx** (compile-time checked queries) | **sqlx** (same, works great) |
| SQLite driver | **mattn/go-sqlite3** (CGO) | **rusqlite** (pure Rust) |
| Auth/JWT | **golang-jwt/jwt/v5** | **jsonwebtoken** crate |
| Config | **dotenvy** or env vars | **config** crate or env vars |
| LLM client | **reqwest** (HTTP to OpenAI API) | **reqwest** + **serde** |
| Logging | **slog** or **zap** | **tracing** |
| CLI/entry | Standard `main.go` | Standard `main.rs` |

### Frontend

| Layer | Recommendation | Rationale |
|-------|---------------|-----------|
| Framework | **React 19 + TypeScript** | Largest ecosystem, you know TS |
| Build tool | **Vite** | Fast HMR, you already use it |
| Styling | **Tailwind CSS** or **Pico CSS** | Rapid UI, minimal config |
| State management | **Zustand** or React Context | Lightweight, fits SPA scale |
| HTTP client | **fetch** or **axios** | Standard |
| TTS integration | Browser `SpeechSynthesis` API | Native, no deps |

#### Frontend Alternatives (if not React)

| Framework | Pros | Cons |
|-----------|------|------|
| **Vue 3 + Vite** | Gentle learning curve, great docs, Composition API | Smaller ecosystem than React |
| **Svelte 5 + SvelteKit** | Minimal runtime, compiler-based, fast | Smaller community, fewer libs |
| **Lit** | Lightweight web components, framework-agnostic | Less opinionated, more boilerplate |
| **SolidJS** | Fine-grained reactivity, no virtual DOM | Smaller ecosystem, JSX-only |
| **HTMX** | Minimal JS, server-driven UI | Less interactive, not a full SPA |
| **Qwik** | Resumable, instant startup | Niche, steep learning curve |

### Database

| Environment | Choice | Rationale |
|-------------|--------|-----------|
| Dev | **SQLite** | Zero-config, file-based, fast iteration |
| Prod | **PostgreSQL** | JSONB for vocabulary, concurrent users |
| Migrations | **golang-migrate** (Go) or **sqlx-cli** (Rust) | Version-controlled schema |

### LLM

| Option | When to use |
|--------|-------------|
| **OpenAI API** (GPT-4o-mini) | Simplest, pay-per-token, good quality |
| **Ollama** (local) | No API costs, privacy, higher latency |
| **Google Gemini API** | Alternative, competitive pricing |

### Infrastructure

| Concern | Tool |
|---------|------|
| Containerization | **Docker** + **Docker Compose** |
| Reverse proxy | **Nginx** or **Caddy** |
| CI/CD | **GitHub Actions** (you already use it) |

---

## 9. Deployment

- **Dev**: Docker Compose — backend + SQLite, frontend served via dev server
- **Prod**: Docker Compose or single container — backend + PostgreSQL + Nginx reverse proxy for SPA
- Backend exposes port 8080 (configurable)
- Frontend built to static files, served by Nginx or CDN

---

## 9. Clarifying Questions

**Answered:**
- ✅ **Auth**: Simple JWT-based, multi-user
- ✅ **LLM**: Used for translation (EN→ES/IT) + vocabulary enrichment; translations cached in DB
- ✅ **Target languages**: English → Spanish, English → Italian
- ✅ **LLM caching**: Translations stored in `translation_cache` table; cache key = FNV-1a hash (fast, non-cryptographic)
- ✅ **Backend language**: Go (Gin) or Rust (Axum) — see §8
- ✅ **Frontend**: React 19 + TypeScript + Vite — see §8

**Still open:**

### A. Text Length Limit
What's a reasonable max input length? (e.g., 500 chars, 2000 chars, 5000 chars?) Affects caching granularity and LLM cost.

### B. LLM Provider
OpenAI (GPT-4o-mini), Ollama (local), or other?

### C. Scope for Now
Should I focus on:
1. **Finalize system design** — update doc with your answers
2. **Backend scaffold** — project structure, DB schema, API stubs
3. **Frontend scaffold** — SPA skeleton with routing and basic layout
4. **All of the above**

---

## 10. LLM Integration Notes

LLM will be used for:
- **Translation**: EN → ES/IT for submitted text blocks (primary LLM usage)
- **Vocabulary enrichment**: Definitions, example sentences, difficulty ratings
- **Optional**: Grammar explanations, spaced repetition scheduling hints

### Translation Caching Strategy
- **Full-text translations** cached in `translation_cache` table (FNV-1a hash key) — avoids re-calling LLM for identical text blocks
- **Word/phrase translations** live in `vocabulary` table — no separate cache needed; each word is a row
- On cache hit: return instantly, no LLM call
- On cache miss: call LLM, store result, extract vocabulary, return
- Benefits: repeated study sessions don't re-incur LLM costs
- `tokens_used` column tracks LLM usage for cost monitoring

### LLM Call Guidelines
- Async where possible (don't block HTTP requests on LLM latency)
- Configurable provider (OpenAI GPT-4o-mini default, Ollama optional)
- Per-word caching for definitions (same word → same explanation)
- Rate limiting to prevent abuse

---

*Document version: 0.4 — 7 of 9 questions answered.*

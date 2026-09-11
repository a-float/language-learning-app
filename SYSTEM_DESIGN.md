# Language Learning App — System Design

## 1. Overview

A web-based language learning application with a SPA frontend and a lightweight backend. The app supports multiple learning modes centered around vocabulary acquisition from real media (songs, movies), personal favorites, and speech/listening practice via browser TTS.

---

## 2. Requirements

### 2.1 Functional Requirements

| ID | Requirement |
|----|-------------|
| FR-1 | User can browse/select words and sentences and add them to a personal favorites list |
| FR-2 | User can select a song or movie; the app extracts vocabulary and generates structured lessons |
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
│  │ Auth     │ │ Lesson   │ │ Favorite │ │ Media      │  │
│  │ Service  │ │ Generator│ │ Service  │ │ Parser     │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └─────┬──────┘  │
│       │            │            │              │         │
│  ┌────▼────────────▼────────────▼──────────────▼──────┐  │
│  │ Data Layer: SQLite (dev) / PostgreSQL (prod)       │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

### 3.1 Key Design Decisions (provisional)

| Decision | Proposed | Status |
|----------|----------|--------|
| Backend language | Go **or** Rust | **Open** — see §6 |
| Frontend UI framework | TBD | **Open** |
| Database | SQLite (dev), PostgreSQL (prod) | Proposed |
| API style | REST/JSON | Proposed |
| TTS | Browser-native `SpeechSynthesis` API | Decided (no server TTS needed) |
| Lesson generation | Rule-based extraction + LLM for definitions/examples | **Open** — see §9-D |
| Auth | JWT-based session tokens | Confirmed — see §9 |
| Target languages | English → Spanish, English → Italian | Confirmed — see §9 |

---

## 4. Feature Breakdown

### 4.1 Favorites Mode

- User selects any word/sentence from any lesson or media source
- POST `/api/favorites` — add item
- GET `/api/favorites` — list with pagination
- DELETE `/api/favorites/:id` — remove
- Favorites feed into review/quiz mode
- Data model: `favorite_id`, `user_id`, `word/sentence text`, `source_media`, `language`, `added_at`

### 4.2 Song/Movie Lesson Generator

- User provides a URL or paste text from a song/movie
- Backend extracts vocabulary:
  - Tokenize text, filter by known word lists (common vocab baseline)
  - Group into lesson units (by scene/verse/paragraph)
  - Tag each word with: frequency rank, part of speech, difficulty
- Optional: call external LLM API for richer explanations/examples
- Lesson data model:
  - `lesson_id`, `title`, `source_type` (song|movie), `source_ref`, `language`, `difficulty`, `vocabulary_items` (JSON array), `created_at`

### 4.3 Listening / Speech Mode

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

favorites
  id (PK)
  user_id (FK → users)
  text
  language
  source_type (word|sentence)
  source_media_id (nullable FK → lessons)
  created_at

lessons
  id (PK)
  title
  source_type (song|movie|manual)
  source_ref (URL or text snippet)
  language
  difficulty (beginner|intermediate|advanced)
  vocabulary (JSONB — array of {word, pos, frequency, definition})
  full_text (TEXT — original media text)
  created_at
  user_id (FK → users, nullable — shared lessons?)

lesson_words
  id (PK)
  lesson_id (FK → lessons)
  word
  part_of_speech
  frequency_rank
  difficulty
  definition

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
POST   /api/lessons              # generate from media text/URL
GET    /api/lessons/:id
DELETE /api/lessons/:id

# Speech
GET    /api/voices?lang=en-US    # available TTS voices (proxy to browser or cache)

# User
GET    /api/users/me
PUT    /api/users/me
```

---

## 8. Deployment

- **Dev**: Docker Compose — backend + SQLite, frontend served via dev server
- **Prod**: Docker Compose or single container — backend + PostgreSQL + Nginx reverse proxy for SPA
- Backend exposes port 8080 (configurable)
- Frontend built to static files, served by Nginx or CDN

---

## 9. Clarifying Questions

**Answered:**
- ✅ **Auth**: Simple JWT-based, multi-user
- ✅ **LLM**: Will be used for lesson generation / vocabulary enrichment
- ✅ **Target languages**: English → Spanish, English → Italian

**Still open:**

### A. Backend Language
Go or Rust? Go = faster to build; Rust = leaner binaries.

### B. UI Framework
Which direction are you leaning? Options:
- **React** — largest ecosystem, most resources
- **Vue** — gentle learning curve, great docs
- **Svelte** — minimal runtime, compiler-based
- **Lit** — lightweight web components (you've used it before)

### C. Media Input — How does text get into the app?
When you say "select song or movie," how would the user provide the content?
1. **Paste text** — user copies & pastes lyrics/dialogue into a textarea
2. **Upload file** — .txt, .pdf, .srt subtitle file upload
3. **URL** — app fetches from a link (e.g., YouTube transcript, subtitle site)
4. **Built-in library** — pre-loaded sample songs/movies in the app

### D. Lesson Generation — How much LLM vs rule-based?
- **Full LLM** — send text to LLM, get structured lessons back
- **Hybrid** — rule-based extraction (word frequency, segmentation) + LLM for definitions/examples
- **Tiered** — simple extraction by default, LLM enrichment on demand

### E. Scope for Now
Should I focus on:
1. **Finalize system design** — update doc with your answers
2. **Backend scaffold** — project structure, DB schema, API stubs
3. **Frontend scaffold** — SPA skeleton with routing and basic layout
4. **All of the above**

---

## 10. LLM Integration Notes

LLM will be used for at least:
- **Lesson generation**: Given extracted vocabulary from media text, LLM generates definitions, example sentences, and difficulty ratings
- **Vocabulary enrichment**: Contextual explanations, collocations, usage notes
- **Optional**: Grammar explanations, spaced repetition scheduling hints

LLM calls should be:
- Async where possible (don't block HTTP requests on LLM latency)
- Cacheable (same word/phrase → same explanation)
- Configurable provider (OpenAI, local LLM, etc.)

---

*Document version: 0.2 — 3 of 5 questions answered.*

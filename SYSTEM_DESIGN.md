# Mówik — Language Learning App — System Design

## 1. Overview

A web-based language learning SPA (Vue 3) with a PocketBase backend and custom Go handlers. Users submit blocks of text; the app translates via OpenRouter, extracts vocabulary, and generates lessons. Includes favorites, speech/listening practice via browser TTS, and LLM-powered translation with DB caching.

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
| NFR-1 | Backend memory footprint < 100 MB under normal load | PocketBase + Go handlers are lightweight |
| NFR-2 | Backend startup time < 2 s | Fast iteration |
| NFR-3 | SPA loads in < 2 s on broadband | UX baseline |
| NFR-4 | System should support 1–100 concurrent users initially | Solo dev → small community scale |
| NFR-5 | Containerizable via Docker | Consistent dev/prod environments |
| NFR-6 | Type-safe across frontend and shared contracts | Reduce runtime errors |

---

## 3. Architecture

```
┌──────────────────────────────────────────────────────────┐
│  Browser (Vue 3 SPA Frontend)                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │ Vue 3 + TypeScript + Vite                        │  │
│  │ ┌──────────┐ ┌──────────┐ ┌────────────────────┐  │  │
│  │ │ Favorites│ │ Lesson   │ │ Listening/Speech   │  │  │
│  │ │ View     │ │ Viewer   │ │ Mode (TTS)         │  │  │
│  │ └──────────┘ └──────────┘ └────────────────────┘  │  │
│  └─────────────────┬──────────────────────────────────┘  │
│                    │ REST / JSON over HTTPS             │
└────────────────────┼────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│  PocketBase (Go) + Custom Handlers                     │
│  ┌────────────┐  ┌────────────┐  ┌────────────────┐   │
│  │ Auth (JWT) │  │ CRUD API │  │ Custom Handlers│   │
│  │ Built-in   │  │ Collections│ │ (Translate,    │   │
│  │            │  │            │ │  Vocabulary)   │   │
│  └────────────┘  └────────────┘  └───────┬────────┘   │
│                                         │             │
│  ┌──────────────────────────────────────▼──────────┐  │
│  │ Data Layer: SQLite (dev) / PostgreSQL (prod)    │  │
│  └─────────────────────────────────────────────────┘  │
│                                                       │
│         ┌────────────────────────────┐                │
│         │ OpenRouter (LLM API)     │                │
│         └────────────────────────────┘                │
└───────────────────────────────────────────────────────┘
```

### 3.1 Key Design Decisions

| Decision | Chosen | Status |
|----------|--------|--------|
| App name | **Mówik** | ✅ |
| Backend | PocketBase (Go) + custom Go handlers | ✅ |
| Frontend | Vue 3 + TypeScript + Vite | ✅ |
| Database | SQLite (dev), PostgreSQL (prod) | ✅ |
| API style | REST/JSON | ✅ |
| TTS | Browser-native `SpeechSynthesis` API | ✅ |
| Auth | PocketBase built-in JWT | ✅ |
| Target languages | English → Spanish, English → Italian | ✅ |
| LLM provider | OpenRouter (OpenAI-compatible API) | ✅ |
| Cache key | FNV-1a hash (non-cryptographic) | ✅ |

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
  2. Otherwise, custom handler calls OpenRouter to translate text (EN → ES/IT)
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

## 5. Data Model

```sql
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

## 6. Backend: PocketBase + Custom Go Handlers

PocketBase provides:
- Built-in JWT auth (users, sessions, tokens)
- REST API for all collections (CRUD, list, filter)
- SQLite (dev) / PostgreSQL (prod) data layer
- File storage (not used yet — speech is browser-side)
- Admin UI for data management

Custom Go handlers extend PocketBase for:
- **Translation**: call OpenRouter, cache results, extract vocabulary
- **Vocabulary enrichment**: definitions, examples, difficulty ratings via LLM

### 6.1 Project Structure

```
backend/
├── go.mod
├── main.go              # PocketBase instance + custom route registration
├── handlers/
│   └── translate.go     # OpenRouter call + cache read/write + vocab extraction
├── collections/
│   ├── vocabulary.json
│   ├── translation_cache.json
│   └── lesson_vocabulary.json
└── Dockerfile
```

### 6.2 Translation Flow

```
POST /api/translate { text, target_lang }
  │
  ├─ FNV-1a(source_text + target_lang) → cache key
  │
  ├─ Cache HIT → return cached translation
  │
  └─ Cache MISS →
       ├─ Call OpenRouter (OpenAI-compatible SDK)
       ├─ Parse response → extract vocabulary items
       ├─ Upsert vocabulary rows
       ├─ Store in translation_cache
       └─ Return translation + vocabulary
```

---

## 7. API Surface

```
# Auth (PocketBase built-in)
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

# Translation (custom handler)
POST   /api/translate             # { text, target_lang } → { translated_text, vocabulary, cached }

# Translation Cache (admin)
GET    /api/cache/stats           # cache hit rate, tokens used
DELETE /api/cache/entries         # clear stale translations

# User
GET    /api/users/me
PUT    /api/users/me
```

---

## 8. Technology Stack

### Backend

| Layer | Choice | Notes |
|-------|--------|-------|
| Backend framework | **PocketBase** (Go) | Built-in auth, REST, DB, admin UI |
| Custom handlers | **Go** (stdlib + PocketBase SDK) | OpenRouter, vocabulary extraction |
| LLM client | **openai-go** (via OpenRouter) | OpenAI-compatible API at `https://openrouter.ai/api/v1` |
| Database | **PocketBase** (SQLite dev / Postgres prod) | Zero-config, single binary |
| Auth | **PocketBase** built-in JWT | Tokens: `@base`/`@user` |
| Logging | **slog** | Stdlib-adjacent |

### Frontend

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Framework | **Vue 3 + TypeScript** | Composition API, great docs |
| Build tool | **Vite** | Fast HMR |
| Routing | **Vue Router 4** | SPA navigation |
| State | **Pinia** | Lightweight, TypeScript-first |
| HTTP client | **fetch** or **axios** | Standard |
| Styling | **Tailwind CSS** or **Pico CSS** | Rapid UI |
| TTS integration | Browser `SpeechSynthesis` API | Native, no deps |

### Database

| Environment | Choice | Rationale |
|-------------|--------|-----------|
| Dev | **SQLite** | Zero-config, file-based |
| Prod | **PostgreSQL** | Concurrent users, JSONB |

### Infrastructure

| Concern | Tool |
|---------|------|
| Containerization | **Docker** + **Docker Compose** |
| Reverse proxy | **Nginx** or **Caddy** |
| CI/CD | **GitHub Actions** |

---

## 9. Deployment

- **Dev**: Docker Compose — PocketBase (SQLite) + frontend dev server
- **Prod**: Docker Compose — PocketBase (PostgreSQL) + Nginx for static SPA
- PocketBase exposes port 8080 (configurable)
- Frontend built to static files, served by Nginx

### Docker Compose

```yaml
services:
  pocketbase:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - POCKETBASE_ADMIN_PASSWORD=${PB_ADMIN_PASSWORD}
      - POCKETBASE_DB=${PB_DB:-./data.db}
    volumes:
      - pb_data:/pocketbase/data

  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    depends_on:
      - pocketbase

volumes:
  pb_data:
```

---

## 10. LLM Integration (OpenRouter)

LLM is used for:
- **Translation**: EN → ES/IT for submitted text blocks (primary usage)
- **Vocabulary enrichment**: Definitions, example sentences, difficulty ratings

### OpenRouter Setup

- **SDK**: `github.com/openai/openai-go` (OpenAI-compatible Go client)
- **Base URL**: `https://openrouter.ai/api/v1`
- **API Key**: `OPENROUTER_API_KEY` environment variable
- **Model**: `openai/gpt-4o-mini` (default)

### Translation Caching Strategy

- **Full-text translations** cached in `translation_cache` table (FNV-1a hash key)
- **Word/phrase translations** live in `vocabulary` table — no separate cache needed
- On cache hit: return instantly, no LLM call
- On cache miss: call OpenRouter, store result, extract vocabulary, return
- `tokens_used` column tracks LLM usage for cost monitoring

---

## 11. Clarifying Questions

**Answered:**
- ✅ **App name**: Mówik
- ✅ **Backend**: PocketBase (Go) + custom handlers
- ✅ **Frontend**: Vue 3 + TypeScript + Vite
- ✅ **LLM**: OpenRouter (OpenAI-compatible SDK)
- ✅ **TTS**: Browser-native `SpeechSynthesis` API — no server-side
- ✅ **Auth**: PocketBase built-in JWT
- ✅ **Target languages**: English → Spanish, English → Italian
- ✅ **LLM caching**: FNV-1a hash key in `translation_cache` table
- ✅ **Database**: SQLite (dev), PostgreSQL (prod)

**Still open:**

### A. Text Length Limit
What's a reasonable max input length? (e.g., 500 chars, 2000 chars, 5000 chars?)

### B. Scope for Now
1. **Scaffold backend** — PocketBase + translate handler + collections
2. **Scaffold frontend** — Vue SPA with routing, translation UI, favorites
3. **Both**

---

*Document version: 1.0 — All key decisions finalized.*

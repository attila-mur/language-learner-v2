# Language Learner MVP - Project Plan & Architecture

## 🎯 **Technology Stack (RECOMMENDED)**

### Backend: **Go**
- **Why:** Fast development, single binary, minimal resources, great for DevOps showcase
- **Framework:** Gin (lightweight, fast, good for REST APIs)
- **Database:** SQLite (for MVP speed, can migrate to PostgreSQL later)

### Frontend: **HTML/CSS/Vanilla JS** (minimal, ~200 lines)
- Simple, fast to build
- Shows you understand frontend basics without distraction
- Can be replaced later without backend changes

### Deployment: **Fly.io** (or Railway/Render as backup)
- Free tier generous (~3 shared-cpu-1x-256mb VMs)
- GitHub integration for auto-deploy
- Easy Docker setup

---

## 📁 **Project Structure**

```
language-learner/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── handlers/
│   │   ├── quiz.go         (GET /api/topics, GET /api/quiz/:id)
│   │   └── submit.go       (POST /api/submit)
│   ├── models/
│   │   ├── word.go
│   │   ├── quiz.go
│   │   └── score.go
│   ├── db/
│   │   ├── db.go           (SQLite connection)
│   │   └── seed.go         (load words.json into DB)
│   ├── data/
│   │   └── words.json      (word pool)
│   ├── Dockerfile
│   ├── fly.toml            (Fly.io config)
│   └── .gitignore
│
├── frontend/
│   ├── index.html
│   ├── style.css
│   ├── app.js              (vanilla JS, fetch calls)
│   └── favicon.ico
│
└── README.md
```

---

## 🗄️ **Database Schema (SQLite)**

```sql
-- Seeded topics (immutable)
CREATE TABLE topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seeded words (immutable)
CREATE TABLE words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id INTEGER NOT NULL,
    hungarian TEXT NOT NULL,
    english TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(id)
);

-- Anonymous user sessions (localStorage-based)
CREATE TABLE user_sessions (
    session_id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    custom_words_count INTEGER DEFAULT 0
);

-- User-added custom words (max 100 per session)
CREATE TABLE custom_words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    topic_id INTEGER,           -- Can be null for "Personal" words
    hungarian TEXT NOT NULL,
    english TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(session_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id)
);

-- Quiz results history (optional, for future leaderboard)
CREATE TABLE scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    topic_id INTEGER,
    score INTEGER,
    total INTEGER,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(session_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id)
);
```

---

## 📊 **Word Pool Data Format** (`data/words.json`)

```json
{
  "topics": [
    {
      "id": 1,
      "name": "Animals",
      "description": "Common Hungarian animals",
      "words": [
        { "hungarian": "macska", "english": "cat" },
        { "hungarian": "kutya", "english": "dog" },
        { "hungarian": "madár", "english": "bird" },
        { "hungarian": "hal", "english": "fish" },
        { "hungarian": "ló", "english": "horse" }
      ]
    },
    {
      "id": 2,
      "name": "Food",
      "description": "Hungarian food vocabulary",
      "words": [
        { "hungarian": "kenyér", "english": "bread" },
        { "hungarian": "vaj", "english": "butter" },
        { "hungarian": "sajt", "english": "cheese" }
      ]
    },
    {
      "id": 3,
      "name": "Colors",
      "description": "Hungarian colors",
      "words": [
        { "hungarian": "piros", "english": "red" },
        { "hungarian": "kék", "english": "blue" },
        { "hungarian": "zöld", "english": "green" }
      ]
    }
  ]
}
```

---

## 🔌 **API Endpoints**

### **Session Management**

#### 1. **POST /api/session**
Creates an anonymous session. Called on first visit.

**Request:**
```json
{}  // Empty body
```

**Response:**
```json
{
  "sessionId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Note:** Frontend stores this in `localStorage.setItem('sessionId', sessionId)` for persistence across page reloads.

---

#### 2. **GET /api/session/validate**
Checks if session exists and is valid. Optional heartbeat endpoint.

**Query Params:** `sessionId=uuid-here`

**Response:**
```json
{
  "valid": true,
  "customWordsCount": 5,
  "createdAt": "2025-03-14T10:00:00Z"
}
```

---

### **Quiz & Word Endpoints**

#### 3. **GET /api/topics**
Returns all available seeded topics.

**Response:**
```json
[
  { "id": 1, "name": "Animals", "description": "Common Hungarian animals" },
  { "id": 2, "name": "Food", "description": "Hungarian food vocabulary" },
  { "id": 3, "name": "Colors", "description": "Hungarian colors" }
]
```

---

#### 4. **GET /api/quiz/:topicId**
Returns 10 random cards from a topic (seeded + custom words mixed).

**Query Params (optional):** `?includeCustom=true` (default: true)

**Response:**
```json
{
  "topicId": 1,
  "topicName": "Animals",
  "totalCards": 10,
  "cards": [
    { "id": 1, "hungarian": "macska", "english": null, "isCustom": false },
    { "id": 2, "hungarian": "kutya", "english": null, "isCustom": false },
    { "id": 150, "hungarian": "tigris", "english": null, "isCustom": true }
  ]
}
```

---

#### 5. **POST /api/submit**
User submits answers. Backend validates and returns detailed results.

**Request:**
```json
{
  "sessionId": "550e8400-e29b-41d4-a716-446655440000",
  "topicId": 1,
  "answers": [
    { "cardId": 1, "userAnswer": "cat" },
    { "cardId": 2, "userAnswer": "dog" },
    { "cardId": 150, "userAnswer": "tiger" }
  ]
}
```

**Response:**
```json
{
  "score": 8,
  "totalCards": 10,
  "percentage": 80,
  "results": [
    {
      "cardId": 1,
      "hungarian": "macska",
      "correct": "cat",
      "userAnswer": "cat",
      "isCorrect": true
    },
    {
      "cardId": 2,
      "hungarian": "kutya",
      "correct": "dog",
      "userAnswer": "dog",
      "isCorrect": true
    },
    {
      "cardId": 150,
      "hungarian": "tigris",
      "correct": "tiger",
      "userAnswer": "tiger",
      "isCorrect": true
    }
  ]
}
```

**Side effects:**
- Stores result in `scores` table
- Updates `user_sessions.last_accessed`

---

### **Custom Words Management**

#### 6. **POST /api/custom-words**
User adds a custom word. Limited to 100 per session.

**Request:**
```json
{
  "sessionId": "550e8400-e29b-41d4-a716-446655440000",
  "hungarian": "tigris",
  "english": "tiger",
  "topicId": 1
}
```

**Response (Success - 201):**
```json
{
  "id": 150,
  "hungarian": "tigris",
  "english": "tiger",
  "topicId": 1,
  "createdAt": "2025-03-14T12:30:00Z"
}
```

**Response (Limit Reached - 400):**
```json
{
  "error": "Custom words limit reached (100/100)",
  "customWordsCount": 100
}
```

---

#### 7. **GET /api/custom-words**
Retrieve all custom words for a session.

**Query Params:** `?sessionId=uuid-here&topicId=1` (topicId optional)

**Response:**
```json
{
  "count": 5,
  "words": [
    {
      "id": 150,
      "hungarian": "tigris",
      "english": "tiger",
      "topicId": 1,
      "createdAt": "2025-03-14T12:30:00Z"
    },
    {
      "id": 151,
      "hungarian": "elefánt",
      "english": "elephant",
      "topicId": 1,
      "createdAt": "2025-03-14T12:35:00Z"
    }
  ]
}
```

---

#### 8. **DELETE /api/custom-words/:wordId**
User deletes a custom word.

**Query Params:** `?sessionId=uuid-here`

**Response (Success - 200):**
```json
{
  "message": "Word deleted",
  "wordId": 150
}
```

**Response (Unauthorized - 403):**
```json
{
  "error": "Not authorized to delete this word"
}
```

---

### **Validation Rules**

| Endpoint | Validation |
|----------|-----------|
| **POST /api/custom-words** | session_id must exist, custom_words_count < 100 |
| **DELETE /api/custom-words/:id** | session_id must own the word |
| **GET /api/quiz/:topicId** | Mix seeded + custom words, return 10 cards |
| **POST /api/submit** | All cardIds must exist (seeded or custom) |

---

## ⚡ **1-2 Week MVP Timeline**

| Week | Task | Time |
|------|------|------|
| **Day 1** | Set up Go project, create main.go + handlers skeleton | 1-2 hrs |
| **Day 2** | SQLite DB setup, seed with words.json | 1-2 hrs |
| **Day 3** | Implement GET /api/topics, GET /api/quiz/:id | 2-3 hrs |
| **Day 4** | Implement POST /api/submit with scoring logic | 2 hrs |
| **Day 5** | Test API with Postman/curl | 1 hr |
| **Day 6** | Build simple HTML/CSS/JS frontend | 3-4 hrs |
| **Day 7** | Dockerize, set up Fly.io deployment | 1-2 hrs |
| **Day 8** | Polish, add README, GitHub setup | 2 hrs |
| **Spare** | Bug fixes, extra word data, nice-to-haves | |

---

## 🐳 **Deployment Strategy**

### Step 1: Dockerfile
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/app .
COPY --from=builder /app/data ./data
EXPOSE 8080
CMD ["./app"]
```

### Step 2: Fly.io Setup
```bash
flyctl auth login
flyctl launch --name my-language-learner
# Follow prompts, deploys automatically from GitHub
```

### Step 3: GitHub Actions (Optional but Impressive)
Auto-deploy on push to main:
```yaml
name: Deploy to Fly
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: superfly/flyctl-actions/setup-flyctl@master
      - run: flyctl deploy --remote-only
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

---

---

## 🎨 **Frontend Architecture - Session & UX Flow**

### **Session Management (Vanilla JS)**

```javascript
// On page load
document.addEventListener('DOMContentLoaded', async () => {
  let sessionId = localStorage.getItem('sessionId');
  
  if (!sessionId) {
    // Create new session
    const response = await fetch('/api/session', { method: 'POST' });
    const data = await response.json();
    sessionId = data.sessionId;
    localStorage.setItem('sessionId', sessionId);
  }
  
  // Store globally for API calls
  window.sessionId = sessionId;
  
  // Load topics and render home page
  loadTopics();
});
```

### **API Call Pattern**

```javascript
// Example: submit quiz answers
async function submitQuiz(topicId, answers) {
  const response = await fetch('/api/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      sessionId: window.sessionId,
      topicId: topicId,
      answers: answers
    })
  });
  return await response.json();
}
```

### **Frontend Pages**

1. **Home Page** (`/`)
   - Displays available topics
   - Button to start quiz for each topic
   - Link to custom words (stretch goal)

2. **Quiz Page** (`/?topic=1`)
   - Shows Hungarian word
   - Input field for English answer
   - Progress indicator (Card 3/10)
   - No submit until all 10 filled

3. **Results Page**
   - Big score display (8/10 = 80%)
   - List of results (correct/wrong breakdown)
   - Option to retake quiz
   - Button to add custom words (optional)

4. **Custom Words Page** (optional)
   - List of user's custom words
   - Form to add new word
   - Delete button for each word
   - Word count (5/100)

### **Local Storage Structure**

```javascript
{
  sessionId: "550e8400-e29b-41d4-a716-446655440000",
  recentTopic: 1,              // For UX: default to last topic
  darkMode: false              // User preference (nice-to-have)
}
```

---

## 🐳 **Deployment Architecture**

### **Single Docker Container Strategy**

```
┌─────────────────────────────────────────────┐
│         Docker Container (Multi-stage)      │
├─────────────────────────────────────────────┤
│                                             │
│  Stage 1: Builder                           │
│  ├─ Go 1.22 Alpine                          │
│  ├─ Compile backend binary                  │
│  └─ Size: ~800MB (discarded after build)    │
│                                             │
│  Stage 2: Runtime                           │
│  ├─ Alpine Linux (minimal)                  │
│  ├─ ~200MB total                            │
│  ├─ Backend binary (from Stage 1)           │
│  ├─ /frontend static files                  │
│  ├─ /data/words.json                        │
│  └─ SQLite with volume mount for DB         │
│                                             │
└─────────────────────────────────────────────┘
        │
        ↓ (Deployed to)
   Fly.io / Railway / Render
   (Free tier: 3 shared-cpu-1x VMs)
```

### **Dockerfile**

```dockerfile
# Stage 1: Build Go backend
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o app .

---

# Stage 2: Runtime
FROM alpine:3.19

WORKDIR /app

# Install only runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy backend binary from builder
COPY --from=builder /build/app .

# Copy frontend static files
COPY frontend ./frontend

# Copy seed data
COPY data ./data

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD [ -f /app/app ] || exit 1

EXPOSE 8080

CMD ["./app"]
```

### **Volume Mounts (Persistent Storage)**

In `docker-compose.yml` or deployment config:

```yaml
volumes:
  - ./data:/app/data          # Seed data (read-only)
  - quiz-db:/app              # SQLite DB persists here
  
volumes:
  quiz-db:
```

This ensures the database survives container restarts.

---

## 🚀 **Hosting & Deployment Options**

### **Recommended: Fly.io**

**Why:**
- Free tier: 3 shared-cpu-1x-256mb VMs
- GitHub integration (auto-deploy on push)
- Built-in Docker support
- Easy scaling if needed

**Setup:**
```bash
# Install flyctl
brew install flyctl

# Login
flyctl auth login

# Launch app
flyctl launch --name language-learner-app

# Deploy
git push origin main  # Auto-deploys

# View logs
flyctl logs

# Scale (future)
flyctl scale count=2
```

**fly.toml config:**
```toml
app = "language-learner-app"
primary_region = "ams"

[build]
  image = "language-learner:latest"

[[services]]
  http_checks = []
  internal_port = 8080
  processes = ["app"]
  protocol = "tcp"
  
  [services.concurrency]
    hard_limit = 25
    soft_limit = 20
  
  [[services.ports]]
    handlers = ["http"]
    port = 80
  
  [[services.ports]]
    handlers = ["tls", "http"]
    port = 443

[env]
  PORT = "8080"
  GIN_MODE = "release"

[mounts]
  source = "quiz_db"
  destination = "/app/quiz_db"
```

---

### **Alternatives**

| Host | Free Tier | Setup Effort | Notes |
|------|-----------|--------------|-------|
| **Fly.io** | Generous | Easy | Recommended for this project |
| **Railway** | $5/month credit | Very easy | Great UX, good for quick deploy |
| **Render** | Limited | Easy | Good alternative, database support |
| **Heroku** | Paid now | Was easy | Not recommended (paid-only now) |

---

## 🔄 **CI/CD Pipeline (Optional but Impressive)**

### **GitHub Actions Workflow**

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Fly.io

on:
  push:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest

      - name: Deploy to Fly.io
        uses: superfly/flyctl-actions@v1
        with:
          args: "deploy --remote-only"
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

**What this does:**
- Runs on every push to main
- Builds Docker image
- Deploys to Fly.io automatically
- Your app goes live in ~2 minutes

---

## 📊 **Project Structure (Final)**

```
language-learner/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── handlers/
│   │   ├── session.go        (POST /api/session, GET /api/session/validate)
│   │   ├── quiz.go           (GET /api/topics, GET /api/quiz/:id)
│   │   ├── submit.go         (POST /api/submit)
│   │   └── custom_words.go   (POST/GET/DELETE /api/custom-words)
│   ├── models/
│   │   ├── session.go
│   │   ├── word.go
│   │   ├── quiz.go
│   │   └── custom_word.go
│   ├── db/
│   │   ├── db.go             (SQLite connection, initialization)
│   │   └── seed.go           (Load words.json on startup)
│   ├── middleware/
│   │   ├── cors.go
│   │   └── auth.go           (Session validation middleware)
│   ├── data/
│   │   └── words.json        (Seed data - 50-100 words, 3-5 topics)
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── .dockerignore
│   ├── go.mod
│   └── .gitignore
│
├── frontend/
│   ├── index.html            (Single-page app shell)
│   ├── style.css             (Simple, responsive CSS)
│   ├── app.js                (Session + API logic)
│   ├── pages.js              (Home, Quiz, Results, CustomWords)
│   ├── utils.js              (Fetch helpers, validation)
│   └── favicon.ico
│
├── .github/
│   └── workflows/
│       └── deploy.yml        (Auto-deploy on push)
│
├── Dockerfile                 (Points to backend)
├── docker-compose.yml        (Local dev: backend + volume)
├── .gitignore
├── README.md
├── ARCHITECTURE.md           (This document)
└── LICENSE
```

---

## ⚡ **Timeline with Anonymous Sessions**

| Day | Tasks | Estimate |
|-----|-------|----------|
| **1** | Project setup, Go skeleton, models | 2-3 hrs |
| **2** | Database setup, seeding from words.json | 2 hrs |
| **3** | Session endpoints, middleware | 2-3 hrs |
| **4** | Quiz endpoints (GET topics, GET quiz) | 2-3 hrs |
| **5** | Submit endpoint with scoring | 2 hrs |
| **6** | Custom words endpoints (POST/GET/DELETE) | 2-3 hrs |
| **7** | Frontend: home, quiz, results pages | 4-5 hrs |
| **8** | Docker, Fly.io setup, testing | 2-3 hrs |
| **9** | Polish, README, GitHub setup | 2 hrs |

**Total:** ~18-24 hours over 1-2 weeks (part-time)

---

---

## 📝 **Go Code Highlights** (What to Showcase)

1. **Clean handler functions** — separates concerns
2. **Database abstraction** — seed function, query functions
3. **Error handling** — proper HTTP status codes
4. **CORS middleware** — frontend ↔ backend communication
5. **Graceful shutdown** — shows production thinking
6. **Environment variables** — PORT, DATABASE_URL flexibility
7. **Structured logging** — not just `log.Println`

---

## 🚀 **GitHub README Structure**

```markdown
# Language Learner

A simple, fast language learning quiz app. Built with Go + SQLite + Vanilla JS.

## Features
- 10-card quizzes by topic (Hungarian → English)
- Instant scoring & feedback
- Multiple topics (Animals, Food, Colors, etc.)

## Architecture
- **Backend:** Gin (Go REST API)
- **Database:** SQLite (embedded)
- **Frontend:** Vanilla HTML/CSS/JS
- **Hosting:** Fly.io

## Quick Start

### Local Development
```bash
go run main.go
# Visit http://localhost:8080
```

### Deploy to Fly.io
```bash
flyctl launch
flyctl deploy
```

## Project Structure
[explain folder layout]

## API Docs
[link to endpoints]

## What This Showcases
✅ Clean Go code & REST API design
✅ Database design & seeding strategy
✅ Full-stack development
✅ Docker containerization
✅ CI/CD deployment
```

---

## 💡 **Nice-to-Haves** (If Time Allows)

- Add difficulty levels (easier/harder words)
- Timer for each card (5 sec per word)
- Dark mode toggle
- Multiple word formats (flashcard view before quiz)
- Fuzzy matching for answers ("kat" ≈ "cat")

---

## ⚠️ **Gotchas to Watch**

1. **SQLite concurrency:** Fine for MVP, upgrade to PostgreSQL if needed later
2. **Word data:** Start with ~50-100 words (3-5 topics × 10-20 words each)
3. **CORS:** Enable it for frontend on `localhost:3000` during dev, `*.fly.dev` in prod
4. **Frontend assets:** Serve from Go static handler (not separate server)
5. **Testing:** Write basic tests for quiz logic, submission scoring

---

## 🔮 **Future Expansion Path**

```
MVP (now)
  ↓
+ User auth (Firebase or simple JWT)
+ Score persistence per user
+ More languages (German, French, Italian)
+ Spaced repetition algorithm
+ Mobile app (React Native)
```

This structure makes it easy to add these without rewriting the core.


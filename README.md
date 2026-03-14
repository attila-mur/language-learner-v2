# Language Learner

A flashcard-style web app for learning vocabulary. Quiz yourself on built-in word sets or add your own custom vocabulary and topics.

## Features

- **6 built-in topics**: Animals, Food, Colors, Numbers, Greetings, Body Parts
- **Custom topics**: Create up to 4 named topics of your own (shown on the home screen alongside built-in topics)
- **Custom words**: Add your own word pairs to any custom topic (up to 100 per session)
- **Quiz mode**: One card at a time, type the English translation
- **Results page**: Score breakdown showing correct/incorrect answers
- **Anonymous sessions**: No login required — sessions stored in localStorage

## Tech Stack

- **Backend**: Go 1.22, Gin, SQLite (via go-sqlite3)
- **Frontend**: Vanilla HTML/CSS/JavaScript (no frameworks)
- **Database**: SQLite with foreign keys enabled
- **Hosting**: Render (auto-deploys from GitHub)

## Running with Docker

```bash
docker compose up --build
```

Then open http://localhost:8080

## Running locally (requires Go 1.22+)

```bash
cd backend
go mod tidy
go run main.go
```

The server will start on port 8080 and serve the frontend from `../frontend`.

## Environment Variables

| Variable         | Default               | Description                     |
|------------------|-----------------------|---------------------------------|
| `PORT`           | `8080`                | HTTP port                       |
| `DATABASE_PATH`  | `./quiz.db`           | Path to SQLite database file    |
| `FRONTEND_PATH`  | `../frontend`         | Path to frontend static files   |
| `DATA_PATH`      | `./data/words.json`   | Path to seed data JSON          |
| `GIN_MODE`       | (not set)             | Set to `release` for production |

## API Endpoints

| Method | Path                        | Description                         |
|--------|-----------------------------|-------------------------------------|
| POST   | `/api/session`              | Create anonymous session            |
| GET    | `/api/session/validate`     | Validate existing session           |
| GET    | `/api/topics`               | List seeded + session's custom topics |
| POST   | `/api/topics`               | Create a custom topic (max 4)       |
| GET    | `/api/quiz/:topicId`        | Get quiz cards for a topic          |
| POST   | `/api/submit`               | Submit quiz answers                 |
| POST   | `/api/custom-words`         | Add a custom word                   |
| GET    | `/api/custom-words`         | List custom words                   |
| DELETE | `/api/custom-words/:wordId` | Delete a custom word                |

## Design Notes

**Custom topic/word ID offset:** Custom word IDs are offset by 1,000,000 in API responses (e.g. DB `id=5` → `1000005`). The submit handler uses `cardId >= 1,000,000` to decide which table to query.

**Topics table:** A single `topics` table holds both seeded and custom topics. Seeded topics have `session_id = NULL`; custom topics store the owner's session ID. A partial unique index enforces name uniqueness among seeded topics only, so users can name their custom topics freely.

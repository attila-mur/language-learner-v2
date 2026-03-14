# Hungarian Language Learner

A flashcard-style web app for learning Hungarian vocabulary. Quiz yourself on built-in word sets or add your own custom vocabulary.

## Features

- **5 built-in topics**: Animals, Food, Colors, Numbers, Greetings
- **Quiz mode**: One card at a time, type the English translation
- **Custom words**: Add your own Hungarian-English pairs (up to 100 per session)
- **Results page**: Score breakdown showing correct/incorrect answers
- **Anonymous sessions**: No login required — sessions stored in localStorage

## Tech Stack

- **Backend**: Go 1.22, Gin, SQLite (via go-sqlite3)
- **Frontend**: Vanilla HTML/CSS/JavaScript (no frameworks)
- **Database**: SQLite with foreign keys enabled

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

| Variable         | Default                        | Description                        |
|------------------|--------------------------------|------------------------------------|
| `PORT`           | `8080`                         | HTTP port                          |
| `DATABASE_PATH`  | `./quiz.db`                    | Path to SQLite database file       |
| `FRONTEND_PATH`  | `../frontend`                  | Path to frontend static files      |
| `DATA_PATH`      | `./data/words.json`            | Path to seed data JSON             |
| `GIN_MODE`       | (not set)                      | Set to `release` for production    |

## API Endpoints

| Method | Path                        | Description                    |
|--------|-----------------------------|--------------------------------|
| POST   | `/api/session`              | Create anonymous session       |
| GET    | `/api/session/validate`     | Validate existing session      |
| GET    | `/api/topics`               | List all topics                |
| GET    | `/api/quiz/:topicId`        | Get quiz cards for a topic     |
| POST   | `/api/submit`               | Submit quiz answers            |
| POST   | `/api/custom-words`         | Add a custom word              |
| GET    | `/api/custom-words`         | List custom words              |
| DELETE | `/api/custom-words/:wordId` | Delete a custom word           |

## Custom Word ID Design

Custom word IDs are offset by 1,000,000 in API responses. A custom word with database `id=5` is returned as `id=1000005`. The submit handler checks `cardId >= 1,000,000` to determine whether to look up in the `custom_words` or `words` table.

package models

type QuizCard struct {
	ID        int     `json:"id"`
	Hungarian string  `json:"hungarian"`
	English   *string `json:"english"`
	IsCustom  bool    `json:"isCustom"`
}

type QuizResponse struct {
	TopicID    int        `json:"topicId"`
	TopicName  string     `json:"topicName"`
	TotalCards int        `json:"totalCards"`
	Cards      []QuizCard `json:"cards"`
}

type AnswerSubmission struct {
	CardID     int    `json:"cardId"`
	UserAnswer string `json:"userAnswer"`
}

type SubmitRequest struct {
	SessionID string             `json:"sessionId"`
	TopicID   int                `json:"topicId"`
	Answers   []AnswerSubmission `json:"answers"`
}

type CardResult struct {
	CardID     int    `json:"cardId"`
	Hungarian  string `json:"hungarian"`
	Correct    string `json:"correct"`
	UserAnswer string `json:"userAnswer"`
	IsCorrect  bool   `json:"isCorrect"`
}

type SubmitResponse struct {
	Score      int          `json:"score"`
	TotalCards int          `json:"totalCards"`
	Percentage int          `json:"percentage"`
	Results    []CardResult `json:"results"`
}

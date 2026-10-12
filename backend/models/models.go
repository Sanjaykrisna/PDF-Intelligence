package models

import "time"

type Document struct {
	ID               string      `json:"id"`
	UserID           string      `json:"user_id"`
	Title            string      `json:"title"`
	Filename         string      `json:"filename"`
	StoredFilename   string      `json:"stored_filename,omitempty"`
	FileHash         string      `json:"file_hash,omitempty"`
	FileSize         int64       `json:"file_size,omitempty"`
	MimeType         string      `json:"mime_type,omitempty"`
	PageCount        int         `json:"page_count,omitempty"`
	TextPageCount    int         `json:"text_page_count,omitempty"`
	Status           string      `json:"status"`
	ValidationStatus string      `json:"validation_status,omitempty"`
	ProcessingStatus string      `json:"processing_status,omitempty"`
	PIIDetected      bool        `json:"pii_detected"`
	TotalChunks      int         `json:"total_chunks,omitempty"`
	Summary          interface{} `json:"summary,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at,omitempty"`
}

type Chunk struct {
	ID                 string    `json:"id"`
	DocumentID         string    `json:"document_id"`
	PageNumber         int       `json:"page_number"`
	SectionName        string    `json:"section_name"`
	ChunkIndex         int       `json:"chunk_index"`
	Text               string    `json:"text"`
	ContentHash        string    `json:"content_hash,omitempty"`
	EmbeddingProvider  string    `json:"embedding_provider,omitempty"`
	EmbeddingModel     string    `json:"embedding_model,omitempty"`
	EmbeddingDimension int       `json:"embedding_dimension,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
}

type ChunkSearchResult struct {
	Chunk
	VectorSimilarity float64 `json:"vector_similarity"`
	KeywordRank      float64 `json:"keyword_rank"`
	CombinedScore    float64 `json:"combined_score"`
}

type DocumentSummary struct {
	Title        string   `json:"title"`
	Objective    string   `json:"objective"`
	Problem      string   `json:"problem"`
	Solution     string   `json:"solution"`
	Technologies []string `json:"technologies"`
	Algorithms   []string `json:"algorithms"`
	Methodology  string   `json:"methodology"`
	Results      string   `json:"results"`
	Advantages   []string `json:"advantages"`
	Limitations  []string `json:"limitations"`
	FutureScope  string   `json:"future_scope"`
}

type DocumentValidationAudit struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	CheckName  string    `json:"check_name"`
	Status     string    `json:"status"` // PASS, FAIL, WARN
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type QueryLog struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	DocumentID       string    `json:"document_id"`
	OriginalQuery    string    `json:"original_query"`
	NormalizedQuery  string    `json:"normalized_query"`
	QueryType        string    `json:"query_type"` // basic, reasoning, direct
	RetrievalAttempt int       `json:"retrieval_attempt"`
	TopSimilarity    float64   `json:"top_similarity"`
	ContextValid     bool      `json:"context_valid"`
	LLMProvider      string    `json:"llm_provider"`
	LLMAttempt       int       `json:"llm_attempt"`
	AnswerValid      bool      `json:"answer_valid"`
	FallbackUsed     bool      `json:"fallback_used"`
	CreatedAt        time.Time `json:"created_at"`
}

type QueryAnalysis struct {
	OriginalQuery      string   `json:"original_query"`
	NormalizedQuery    string   `json:"normalized_query"`
	AlternativeQueries []string `json:"alternative_queries,omitempty"`
	Intent             string   `json:"intent"` // basic, reasoning
	Confidence         float64  `json:"confidence"`
}

type SourceReference struct {
	Page    int    `json:"page"`
	Section string `json:"section"`
}

type AnswerResponse struct {
	Answer     string            `json:"answer"`
	Level      int               `json:"level"`
	QueryType  string            `json:"query_type"` // direct, rag
	Confidence float64           `json:"confidence,omitempty"`
	Sources    []SourceReference `json:"sources"`
}

type DocumentInsights struct {
	DocumentID       string                    `json:"document_id"`
	Title            string                    `json:"title"`
	Filename         string                    `json:"filename"`
	FileSize         int64                     `json:"file_size"`
	MimeType         string                    `json:"mime_type"`
	PageCount        int                       `json:"page_count"`
	TextPageCount    int                       `json:"text_page_count"`
	TotalCharacters  int                       `json:"total_characters"`
	TotalWords       int                       `json:"total_words"`
	TotalChunks      int                       `json:"total_chunks"`
	Status           string                    `json:"status"`
	ValidationStatus string                    `json:"validation_status"`
	ProcessingStatus string                    `json:"processing_status"`
	PIIDetected      bool                      `json:"pii_detected"`
	ValidationChecks []DocumentValidationAudit `json:"validation_checks"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

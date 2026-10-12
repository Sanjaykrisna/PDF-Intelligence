package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server     ServerConfig     `toml:"server"`
	Database   DatabaseConfig   `toml:"database"`
	AI         AIConfig         `toml:"ai"`
	Fallback   FallbackConfig   `toml:"fallback"`
	Storage    StorageConfig    `toml:"storage"`
	Validation ValidationConfig `toml:"validation"`
	Privacy    PrivacyConfig    `toml:"privacy"`
	Retrieval  RetrievalConfig  `toml:"retrieval"`
}

type ServerConfig struct {
	Port         int    `toml:"port"`
	JWTSecret    string `toml:"jwt_secret"`
	AllowOrigins string `toml:"allow_origins"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
}

type AIConfig struct {
	Provider       string  `toml:"provider"`
	APIKey         string  `toml:"api_key"`
	Model          string  `toml:"model"`
	EmbedModel     string  `toml:"embed_model"`
	EmbedDimension int     `toml:"embed_dimension"`
	BaseURL        string  `toml:"base_url"`
	Temperature    float64 `toml:"temperature"`
	TimeoutSeconds int     `toml:"timeout_seconds"`
}

type FallbackConfig struct {
	Enabled        bool    `toml:"enabled"`
	Provider       string  `toml:"provider"`
	APIKey         string  `toml:"api_key"`
	Model          string  `toml:"model"`
	BaseURL        string  `toml:"base_url"`
	Temperature    float64 `toml:"temperature"`
	TimeoutSeconds int     `toml:"timeout_seconds"`
}

type StorageConfig struct {
	UploadDir string `toml:"upload_dir"`
}

type ValidationConfig struct {
	MaxFileSizeMB int `toml:"max_file_size_mb"`
	MaxPages      int `toml:"max_pages"`
	MaxAttempts   int `toml:"max_attempts"`
}

type PrivacyConfig struct {
	RemoveEmail      bool `toml:"remove_email"`
	RemovePhone      bool `toml:"remove_phone"`
	RemoveRollNumber bool `toml:"remove_roll_number"`
	RemoveStudentID  bool `toml:"remove_student_id"`
	RemoveAadhaar    bool `toml:"remove_aadhaar"`
	RemoveAddress    bool `toml:"remove_address"`
}

type RetrievalConfig struct {
	TopK                int     `toml:"top_k"`
	SimilarityThreshold float64 `toml:"similarity_threshold"`
	HybridSearch        bool    `toml:"hybrid_search"`
	VectorWeight        float64 `toml:"vector_weight"`
	KeywordWeight       float64 `toml:"keyword_weight"`
}

var AppConfig Config

func LoadConfig(path string) error {
	resolvedPath := path
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		candidates := []string{
			"toml/" + path,
			"toml/config.toml",
			"config.toml",
			"../toml/config.toml",
			"../config.toml",
			"backend/toml/config.toml",
			"backend/config.toml",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				resolvedPath = c
				break
			}
		}
	}

	if _, err := toml.DecodeFile(resolvedPath, &AppConfig); err != nil {
		log.Printf("Error decoding config: %v", err)
		return err
	}
	
	// Apply sensible defaults if not set
	if AppConfig.Validation.MaxFileSizeMB <= 0 {
		AppConfig.Validation.MaxFileSizeMB = 20
	}
	if AppConfig.Validation.MaxPages <= 0 {
		AppConfig.Validation.MaxPages = 100
	}
	if AppConfig.Validation.MaxAttempts <= 0 {
		AppConfig.Validation.MaxAttempts = 2
	}
	if AppConfig.Retrieval.TopK <= 0 {
		AppConfig.Retrieval.TopK = 5
	}
	if AppConfig.Retrieval.SimilarityThreshold <= 0 {
		AppConfig.Retrieval.SimilarityThreshold = 0.50
	}
	if AppConfig.Retrieval.VectorWeight <= 0 && AppConfig.Retrieval.KeywordWeight <= 0 {
		AppConfig.Retrieval.VectorWeight = 0.70
		AppConfig.Retrieval.KeywordWeight = 0.30
	}

	return nil
}

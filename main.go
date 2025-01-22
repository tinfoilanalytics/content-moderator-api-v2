package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	systemPrompt = `You are a content moderator. Analyze the following text and respond with 'safe' if the content is safe, or 'unsafe' followed by the category codes (e.g., 'unsafe\nS1,S2') if any violations are detected.`
	modelName    = "llama-guard3:1b"
)

type analyzeRequest struct {
	Messages []string `json:"messages"`
}

type analysisScores struct {
	ThreatOfHarm           float64 `json:"threat_of_harm"`
	CommercialSolicitation float64 `json:"commercial_solicitation"`
}

type analysisResult struct {
	Content string         `json:"content"`
	Scores  analysisScores `json:"scores"`
	IsSafe  bool           `json:"is_safe"`
}

type ollamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message Message `json:"message"`
}

type server struct {
	ollamaURL string
}

func newServer(ollamaURL string) *server {
	return &server{ollamaURL: ollamaURL}
}

func (s *server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Messages array cannot be empty", http.StatusBadRequest)
		return
	}

	results := make([]analysisResult, 0, len(req.Messages))
	for _, message := range req.Messages {
		result, err := s.analyzeMessage(r.Context(), message)
		if err != nil {
			log.Printf("Error analyzing message '%s': %v", message, err)
			continue
		}
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *server) analyzeMessage(ctx context.Context, message string) (analysisResult, error) {
	ollamaReq := ollamaRequest{
		Model: modelName,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: message},
		},
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return analysisResult{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.ollamaURL+"/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return analysisResult{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return analysisResult{}, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return analysisResult{}, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return analysisResult{}, fmt.Errorf("decoding response: %w", err)
	}

	outputStr := ollamaResp.Message.Content
	violations := parseViolations(outputStr)
	result := analysisResult{
		Content: message,
		Scores:  calculateScores(violations),
		IsSafe:  !containsUnsafe(outputStr),
	}

	log.Printf("Analysis Result - Safe: %v, Threat Score: %.1f, Commercial Score: %.1f",
		result.IsSafe,
		result.Scores.ThreatOfHarm,
		result.Scores.CommercialSolicitation)

	return result, nil
}

func parseViolations(output string) []string {
	parts := strings.Split(output, "\n")
	if len(parts) <= 1 {
		return nil
	}
	return strings.Split(strings.TrimSpace(parts[1]), ",")
}

func calculateScores(violations []string) analysisScores {
	var scores analysisScores
	for _, v := range violations {
		switch strings.TrimSpace(v) {
		case "S1":
			scores.ThreatOfHarm = 1.0
		case "S2", "S8":
			scores.CommercialSolicitation = 1.0
		}
	}
	return scores
}

func containsUnsafe(output string) bool {
	return strings.Contains(strings.ToLower(output), "unsafe")
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	srv := newServer(ollamaURL)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/analyze", corsMiddleware(srv.handleAnalyze))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("Server starting on %s, connecting to Ollama at %s", addr, ollamaURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/replicate/replicate-go"
)

const (
	systemPrompt = `You are a content moderator. Analyze the following text and respond with 'safe' if the content is safe, or 'unsafe' followed by the category codes (e.g., 'unsafe\nS1,S2') if any violations are detected.`
	modelID      = "meta/llama-guard-3-8b:146d1220d447cdcc639bc17c5f6137416042abee6ae153a2615e6ef5749205c8"
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

type server struct {
	client *replicate.Client
}

func newServer(client *replicate.Client) *server {
	return &server{client: client}
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
	input := replicate.PredictionInput{
		"prompt":        message,
		"system_prompt": systemPrompt,
	}

	output, err := s.client.Run(ctx, modelID, input, nil)
	if err != nil {
		return analysisResult{}, fmt.Errorf("running model: %w", err)
	}

	outputStr, ok := output.(string)
	if !ok {
		return analysisResult{}, fmt.Errorf("unexpected output type: %T", output)
	}

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
	client, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		log.Fatalf("Error creating Replicate client: %v", err)
	}

	srv := newServer(client)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/analyze", corsMiddleware(srv.handleAnalyze))

	addr := ":" + os.Getenv("PORT")
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

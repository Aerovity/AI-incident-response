package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"incident-ai/models"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Analyzer uses AI to analyze incidents and suggest fixes
type Analyzer struct {
	client *openai.Client
	model  string
}

// NewAnalyzer creates a new AI analyzer
func NewAnalyzer(apiKey string) *Analyzer {
	client := openai.NewClient(apiKey)
	return &Analyzer{
		client: client,
		model:  openai.GPT4, // Use GPT-4 for better reasoning, or GPT3Dot5Turbo for faster/cheaper
	}
}

// AnalyzeIncident sends incident details to OpenAI and gets back a fix
func (a *Analyzer) AnalyzeIncident(ctx context.Context, incident *models.Incident) (*models.AIResponse, error) {
	log.Printf("[AI] Analyzing incident: %s (Type: %s)\n", incident.ID, incident.Type)

	prompt := a.buildPrompt(incident)

	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: a.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: a.getSystemPrompt(),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3, // Lower temperature for more focused/deterministic responses
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	log.Printf("[AI] Received response from OpenAI\n")

	// Parse the JSON response
	aiResponse, err := a.parseResponse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	log.Printf("[AI] Diagnosis: %s\n", aiResponse.Diagnosis)
	log.Printf("[AI] Fix Type: %s\n", aiResponse.FixType)

	return aiResponse, nil
}

func (a *Analyzer) getSystemPrompt() string {
	return `You are an expert Site Reliability Engineer and DevOps specialist. Your job is to analyze system incidents and provide actionable fixes.

When analyzing an incident, you should:
1. Carefully examine all symptoms, logs, and configuration details
2. Identify the root cause
3. Provide a clear, step-by-step remediation plan
4. Consider the safest and most effective approach

You must respond ONLY with valid JSON in this exact format:
{
  "diagnosis": "Clear explanation of the root cause",
  "fix_type": "restart|config|code",
  "fix_steps": ["Step 1", "Step 2", ...],
  "code": "Any Go code needed (only if fix_type is code)",
  "confidence": 0.95
}

Rules:
- fix_type must be one of: "restart", "config", "code"
- For restart: service just needs to be restarted
- For config: configuration needs to be corrected (provide correct values in fix_steps)
- For code: actual code changes needed (provide Go code in "code" field)
- Be concise but complete
- Only respond with JSON, no additional text`
}

func (a *Analyzer) buildPrompt(incident *models.Incident) string {
	var sb strings.Builder

	sb.WriteString("# INCIDENT ANALYSIS REQUEST\n\n")
	sb.WriteString("## Service Information\n")
	sb.WriteString("- Service Type: HTTP REST API\n")
	sb.WriteString("- Language: Go\n")
	sb.WriteString("- Port: 8080\n\n")

	sb.WriteString("## Incident Details\n")
	sb.WriteString(fmt.Sprintf("- Incident ID: %s\n", incident.ID))
	sb.WriteString(fmt.Sprintf("- Type: %s\n", incident.Type))
	sb.WriteString(fmt.Sprintf("- Detected At: %s\n\n", incident.DetectedAt.Format("2006-01-02 15:04:05")))

	sb.WriteString("## Symptoms\n")
	if len(incident.Symptoms) > 0 {
		for i, symptom := range incident.Symptoms {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, symptom))
		}
	} else {
		sb.WriteString("No specific symptoms recorded\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Recent Logs\n")
	if len(incident.Logs) > 0 {
		sb.WriteString("```\n")
		for _, log := range incident.Logs {
			sb.WriteString(log + "\n")
		}
		sb.WriteString("```\n")
	} else {
		sb.WriteString("No recent logs available\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Current Configuration\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString(`  "database_url": "localhost:5432",` + "\n")
	sb.WriteString(`  "timeout": "30s",` + "\n")
	sb.WriteString(`  "max_retries": "3"` + "\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Your Task\n")
	sb.WriteString("Analyze this incident and provide a JSON response with:\n")
	sb.WriteString("1. Root cause diagnosis\n")
	sb.WriteString("2. Fix type (restart/config/code)\n")
	sb.WriteString("3. Detailed fix steps\n")
	sb.WriteString("4. Any code needed\n")
	sb.WriteString("5. Your confidence level (0-1)\n\n")

	sb.WriteString("Respond ONLY with valid JSON. No markdown, no explanations outside the JSON.")

	return sb.String()
}

func (a *Analyzer) parseResponse(content string) (*models.AIResponse, error) {
	// Clean up the response - remove markdown code blocks if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var response models.AIResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		// Log the problematic content for debugging
		log.Printf("[AI] Failed to parse response: %s\n", content)
		return nil, fmt.Errorf("JSON parsing error: %w", err)
	}

	// Validate the response
	if response.Diagnosis == "" {
		return nil, fmt.Errorf("missing diagnosis in AI response")
	}

	if response.FixType == "" {
		return nil, fmt.Errorf("missing fix_type in AI response")
	}

	validFixTypes := map[string]bool{"restart": true, "config": true, "code": true}
	if !validFixTypes[response.FixType] {
		return nil, fmt.Errorf("invalid fix_type: %s", response.FixType)
	}

	if len(response.FixSteps) == 0 {
		return nil, fmt.Errorf("missing fix_steps in AI response")
	}

	return &response, nil
}

// GetQuickAnalysis provides a simpler, faster analysis (useful for testing)
func (a *Analyzer) GetQuickAnalysis(incident *models.Incident) *models.AIResponse {
	// Fallback analysis based on incident type
	switch incident.Type {
	case models.ServiceDown:
		return &models.AIResponse{
			Diagnosis: "Service process has crashed or stopped responding",
			FixType:   "restart",
			FixSteps: []string{
				"Stop the service if it's still partially running",
				"Restart the service process",
				"Verify health check passes",
			},
			Confidence: 0.9,
		}

	case models.ConfigError:
		return &models.AIResponse{
			Diagnosis: "Configuration file contains invalid values",
			FixType:   "config",
			FixSteps: []string{
				"Restore database_url to 'localhost:5432'",
				"Reset timeout to '30s'",
				"Restart service to apply changes",
			},
			Confidence: 0.85,
		}

	case models.DependencyFailure:
		return &models.AIResponse{
			Diagnosis: "External dependency (database) is unreachable",
			FixType:   "config",
			FixSteps: []string{
				"Update database_url to valid host",
				"Verify database is running",
				"Restart service to reconnect",
			},
			Confidence: 0.8,
		}

	case models.ResourceExhaustion:
		return &models.AIResponse{
			Diagnosis: "System resources exhausted (port blocked or memory full)",
			FixType:   "restart",
			FixSteps: []string{
				"Stop the service",
				"Clear any blocked resources",
				"Restart service on clean port",
			},
			Confidence: 0.75,
		}

	default:
		return &models.AIResponse{
			Diagnosis: "Unknown incident type",
			FixType:   "restart",
			FixSteps: []string{
				"Attempt service restart",
				"Monitor logs for errors",
			},
			Confidence: 0.5,
		}
	}
}

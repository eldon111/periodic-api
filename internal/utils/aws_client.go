package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// AWSConfig represents the AWS configuration loaded from file
type AWSConfig struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
	SessionToken    string `json:"sessionToken"`
}

// AWSLLMClient handles interactions with AWS Bedrock LLM
type AWSLLMClient struct {
	client       *bedrockruntime.Client
	systemPrompt string
}

// NewAWSLLMClient creates a new AWS LLM client
func NewAWSLLMClient(ctx context.Context) (*AWSLLMClient, error) {
	// Load AWS configuration from file
	awsConfig, err := loadAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	log.Printf("AWS Config loaded - Region: %s, AccessKeyID: %s", awsConfig.Region, awsConfig.AccessKeyID[:8]+"...")

	// Create AWS config with custom credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsConfig.AccessKeyID,
			awsConfig.SecretAccessKey,
			awsConfig.SessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Load system prompt from file
	systemPrompt, err := loadSystemPrompt()
	if err != nil {
		return nil, fmt.Errorf("failed to load system prompt: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)
	return &AWSLLMClient{
		client:       client,
		systemPrompt: systemPrompt,
	}, nil
}

// loadAWSConfig reads the AWS configuration from the external file
func loadAWSConfig() (*AWSConfig, error) {
	content, err := os.ReadFile("aws_config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read AWS awsConfig file: %w", err)
	}

	var awsConfig AWSConfig
	if err := json.Unmarshal(content, &awsConfig); err != nil {
		return nil, fmt.Errorf("failed to parse AWS awsConfig: %w", err)
	}

	// Validate required fields
	if awsConfig.AccessKeyID == "" || awsConfig.AccessKeyID == "YOUR_AWS_ACCESS_KEY_ID" {
		return nil, fmt.Errorf("AWS Access Key ID not configured")
	}
	if awsConfig.SecretAccessKey == "" || awsConfig.SecretAccessKey == "YOUR_AWS_SECRET_ACCESS_KEY" {
		return nil, fmt.Errorf("AWS Secret Access Key not configured")
	}
	if awsConfig.Region == "" {
		return nil, fmt.Errorf("AWS Region not configured")
	}

	return &awsConfig, nil
}

// loadSystemPrompt reads the system prompt from the external file
func loadSystemPrompt() (string, error) {
	content, err := os.ReadFile("new_scheduled_item_system_prompt.txt")
	if err != nil {
		return "", fmt.Errorf("failed to read system prompt file: %w", err)
	}
	return string(content), nil
}

// GenerateScheduledItemJSON sends a prompt to AWS LLM and returns JSON response
func (c *AWSLLMClient) GenerateScheduledItemJSON(ctx context.Context, userPrompt string, userTimezone string) (string, error) {
	// Validate and load the user's timezone
	location, err := time.LoadLocation(userTimezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone '%s': %w", userTimezone, err)
	}

	// Get current time in user's timezone
	currentTime := time.Now()
	currentTimeInUserTZ := currentTime.In(location)
	formatter := "2006-01-02 15:04:05"
	currentDateTime := currentTimeInUserTZ.Format(formatter)

	// Build additional context
	additionalContext := fmt.Sprintf("Additional context:\n User's timezone: %s,\n User's current date and time: %s", userTimezone, currentDateTime)

	// Use the loaded system prompt with additional context
	fullPrompt := fmt.Sprintf("%s\n\n%s\n\nUser request: %s", c.systemPrompt, additionalContext, userPrompt)

	// Prepare the request body for Nova model
	requestBody := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"text": fullPrompt,
					},
				},
			},
		},
		"inferenceConfig": map[string]interface{}{
			"temperature": 0.0,
			"top_p":       0.0,
			"top_k":       20,
		},
	}

	log.Printf("Request body: %+v", requestBody)

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Call Bedrock with Nova model using inference profile
	// The inference profile should match the region from config
	modelId := fmt.Sprintf("us.amazon.nova-lite-v1:0")
	log.Printf("Using model ID: %s", modelId)

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        bodyBytes,
	}

	result, err := c.client.InvokeModel(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Printf("Response: %+v", response)

	// Extract the generated text for Nova model response format
	output, ok := response["output"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format - no output field")
	}

	message, ok := output["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format - no message field")
	}

	content, ok := message["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("unexpected response format - no content array")
	}

	textContent, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected content format")
	}

	generatedText, ok := textContent["text"].(string)
	if !ok {
		return "", fmt.Errorf("no text in response")
	}

	return generatedText, nil
}

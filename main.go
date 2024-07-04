package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Time   string `json:"time"`
	Detail struct {
		Result         string `json:"result"`
		RepositoryName string `json:"repository-name"`
		ImageTag       string `json:"image-tag"`
	} `json:"detail"`
}

type MySecret struct {
	EcrWebhookSecret string `json:"ecr-webhook-secret"`
}

func handleRequest(ctx context.Context, event MyEvent) (events.LambdaFunctionURLResponse, error) {
	secretName := nil           // provide the secret name
	region := nil               // provide the region
	var webhookUrl string = nil // provide the webhook URL

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		log.Fatal(err.Error())
	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	var secret MySecret
	json.Unmarshal([]byte(secretString), &secret)

	client := http.Client{
		Timeout: time.Second * 2,
	}
	jsonStr, err := json.Marshal(event)
	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(jsonStr))

	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secret.EcrWebhookSecret)
	req.Host = "registry.ecr.eu-north-1.amazonaws.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Webhook failed")
	} else {
		log.Println("Webhook success")
	}

	return events.LambdaFunctionURLResponse{Body: (event.Detail.RepositoryName + event.Detail.ImageTag), StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}

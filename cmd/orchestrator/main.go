package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"orchestrator/internal/api"
	"orchestrator/internal/authenticator"
	"orchestrator/internal/db"
	"orchestrator/internal/engine"
	"orchestrator/internal/store"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	codeBucket := os.Getenv("CODE_BUCKET")
	dynamoTable := os.Getenv("DYNAMO_TABLE")
	lambdaExecutionRole := os.Getenv("LAMBDA_EXECUTION_ROLE_ARN")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("failed loading config, %v", err))
	}

	engine := engine.NewLambdaService(
		lambda.NewFromConfig(cfg),
		codeBucket,
		lambdaExecutionRole,
	)

	store := store.NewS3Service(
		s3.NewFromConfig(cfg),
		codeBucket,
		5*time.Minute,
	)

	tokenRepo := db.NewDynamoTokenRepository(dynamodb.NewFromConfig(cfg), dynamoTable)

	authenticator := authenticator.NewTokenAuthenticator(tokenRepo)

	mux := http.NewServeMux()

	computeHandler := api.NewComputeHandler(engine, store, authenticator)
	computeHandler.RegisterRoutes(mux)

	fmt.Printf("Starting server on :%s\n", port)
	http.ListenAndServe(":"+port, mux)
}

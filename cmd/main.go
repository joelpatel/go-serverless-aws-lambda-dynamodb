package main

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/joelpatel/go-serverless-aws-lambda-dynamodb/pkg/handlers"
)

var (
	dynamodbClient dynamodbiface.DynamoDBAPI
)

func main() {
	region := os.Getenv("AWS_REGION")
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return
	}

	dynamodbClient = dynamodb.New(awsSession)
	lambda.Start(handler)
}

const TABLENAME = "go-serverless-aws-lambda-dynamodb"

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return handlers.GetUser(req, TABLENAME, dynamodbClient)
	case "POST":
		return handlers.CreateUser(req, TABLENAME, dynamodbClient)
	case "PUT":
		return handlers.UpdateUser(req, TABLENAME, dynamodbClient)
	case "DELETE":
		return handlers.DeleteUser(req, TABLENAME, dynamodbClient)
	default:
		return handlers.UnhandledMetod()
	}
}

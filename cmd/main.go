package main

import (
	"os"

	"github.com/YD-GitGang/serverless-with-go-and-aws---portfolio2---/pkg/handlers"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var dynaClient dynamodbiface.DynamoDBAPI

func main() {
	region := os.Getenv("AWS_REGION")
	awsSession, err := session.NewSession(&aws.Config{Region: aws.String(region)})

	if err != nil {
		return
	}

	dynaClient = dynamodb.New(awsSession)
	lambda.Start(handler)
}

const tableName = "go-serverless"

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	code := req.HTTPMethod
	switch {
	case code == "GET":
		return handlers.GetUser(req, tableName, dynaClient)
	case code == "POST":
		return handlers.CreateUser(req, tableName, dynaClient)
	case code == "PUT":
		return handlers.UpdateUser(req, tableName, dynaClient)
	case code == "DELETE":
		return handlers.DeleteUser(req, tableName, dynaClient)
	default:
		return handlers.UnhandledMethod()
	}
}

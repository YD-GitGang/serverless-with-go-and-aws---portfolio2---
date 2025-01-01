package handlers

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

func GetUser(req events.APIGatewayProxyRequest, tableName string, dinaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dinaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dinaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dinaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
}

func UnhandledMethod(req events.APIGatewayProxyRequest, tableName string, dinaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
}

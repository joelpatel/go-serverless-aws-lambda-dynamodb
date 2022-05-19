/*
This represents a combination of both models and controllers from MVC architecture.
Contains models that communicates to the database.
*/
package user

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/joelpatel/go-serverless-aws-lambda-dynamodb/pkg/validators"
)

const (
	ErrorFailedToFetchRecord     = "failed to fetch record(s)"
	ErrorFailedToUnmarshalRecord = "failed to unmarshal"
	ErrorInvalidUserData         = "invalid user data"
	ErrorInvalidEmail            = "invalid email"
	ErrorCouldNotMarshalItem     = "could not marshal item"
	ErrorCouldNotDeleteItem      = "could not delete item"
	ErrorCouldNotDynamoPutItem   = "could not dynamo put item"
	ErrorUserAlreadyExists       = "user.User already exists"
	ErrorUserDoesNotExist        = "user.User does not exist"
)

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

func FetchUser(email string, tablename string, dynamodbClient dynamodbiface.DynamoDBAPI) (*User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tablename),
	} // query

	result, err := dynamodbClient.GetItem(input) // actual fetching
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new(User)
	err = dynamodbattribute.UnmarshalMap(result.Item, item) // json/string -> golang struct
	if err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}
	return item, nil
}

func FetchUsers(tablename string, dynamodbClient dynamodbiface.DynamoDBAPI) (*[]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tablename),
	}

	results, err := dynamodbClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	items := new([]User)
	err = dynamodbattribute.UnmarshalListOfMaps(results.Items, items)
	return items, nil
}

func CreateUser(req events.APIGatewayProxyRequest, tablename string, dynamodbClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var user User

	err := json.Unmarshal([]byte(req.Body), &user)
	if err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	if !validators.IsEmailValid(user.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}

	// check if the use already exists
	currentUser, _ := FetchUser(user.Email, tablename, dynamodbClient)
	if currentUser != nil && len(currentUser.Email) != 0 {
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	dynaUserJSON, err := dynamodbattribute.MarshalMap(user) // dynamodb can understand it now; golang struct -> json
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      dynaUserJSON,
		TableName: aws.String(tablename),
	}

	_, err = dynamodbClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return &user, nil
}

func UpdateUser(req events.APIGatewayProxyRequest, tablename string, dynamodbClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var user User

	err := json.Unmarshal([]byte(req.Body), &user)
	if err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	currentUser, _ := FetchUser(user.Email, tablename, dynamodbClient)
	if currentUser != nil && len(currentUser.Email) == 0 {
		return nil, errors.New(ErrorUserDoesNotExist)
	}

	dynaUserJSON, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      dynaUserJSON,
		TableName: aws.String(tablename),
	}

	_, err = dynamodbClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}
	return &user, nil
}

func DeleteUser(req events.APIGatewayProxyRequest, tablename string, dynamodbClient dynamodbiface.DynamoDBAPI) (string, error) {
	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tablename),
	}

	_, err := dynamodbClient.DeleteItem(input)
	if err != nil {
		return "", errors.New(ErrorCouldNotDeleteItem)
	}

	return email, nil
}

package credstash

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
)

type dynamoDB interface {
	PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
	Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}

type decrypter interface {
	Decrypt(*kms.DecryptInput) (*kms.DecryptOutput, error)
	GenerateDataKey(*kms.GenerateDataKeyInput) (*kms.GenerateDataKeyOutput, error)
}

var _ dynamoDB = (*dynamodb.DynamoDB)(nil)
var _ decrypter = (*kms.KMS)(nil)

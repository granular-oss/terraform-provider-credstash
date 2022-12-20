package credstash

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
)

var (
	ErrSecretNotFound = errors.New("Secret Not Found")

	// ErrHmacValidationFailed returned when the hmac signature validation fails
	ErrHmacValidationFailed = errors.New("Secret HMAC validation failed")
)

type DecryptedCredential struct {
	*Credential
	Secret string
}

type DataKey struct {
	CiphertextBlob []byte
	Plaintext      []byte
}

func generateDataKey(svc decrypter, alias string, ctx *EncryptionContextValue, size int) (*DataKey, error) {

	numberOfBytes := int64(size)

	params := &kms.GenerateDataKeyInput{
		KeyId:             aws.String(alias),
		EncryptionContext: *ctx,
		GrantTokens:       []*string{},
		NumberOfBytes:     aws.Int64(numberOfBytes),
	}

	resp, err := svc.GenerateDataKey(params)

	if err != nil {
		return nil, err
	}

	return &DataKey{
		CiphertextBlob: resp.CiphertextBlob,
		Plaintext:      resp.Plaintext, // return the plain text key after generation
	}, nil
}

type keyMaterial struct {
	Name    string
	version int
	Digest  string

	Content []byte
	HMAC    []byte
	Key     []byte
}

// GetHighestVersion look up the highest version for a given name
func GetHighestVersion(svc dynamoDB, tableName *string, name string) (string, error) {
	log.Printf("[DEBUG]  Looking up highest version: %s", name)

	res, err := svc.Query(&dynamodb.QueryInput{
		TableName: tableName,
		ExpressionAttributeNames: map[string]*string{
			"#N": aws.String("name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name": {
				S: aws.String(name),
			},
		},
		KeyConditionExpression: aws.String("#N = :name"),
		Limit:                  aws.Int64(1),
		ConsistentRead:         aws.Bool(true),
		ScanIndexForward:       aws.Bool(false), // descending order
		ProjectionExpression:   aws.String("version"),
	})

	if err != nil {
		return "", err
	}
	log.Print("[DEBUG]  Got to line 315")

	if len(res.Items) == 0 {
		return "", ErrSecretNotFound
	}

	v := res.Items[0]["version"]

	if v == nil {
		return "", ErrSecretNotFound
	}

	return aws.StringValue(v.S), nil
}

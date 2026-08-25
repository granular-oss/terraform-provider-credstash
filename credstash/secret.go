package credstash

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
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

func generateDataKey(ctx context.Context, svc decrypter, alias string, encCtx *EncryptionContextValue, size int) (*DataKey, error) {

	numberOfBytes := int32(size)

	params := &kms.GenerateDataKeyInput{
		KeyId:             aws.String(alias),
		EncryptionContext: *encCtx,
		GrantTokens:       []string{},
		NumberOfBytes:     aws.Int32(numberOfBytes),
	}

	resp, err := svc.GenerateDataKey(ctx, params)

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
func GetHighestVersion(ctx context.Context, svc dynamoDB, tableName *string, name string) (string, error) {
	log.Printf("[DEBUG]  Looking up highest version: %s", name)

	res, err := svc.Query(ctx, &dynamodb.QueryInput{
		TableName: tableName,
		ExpressionAttributeNames: map[string]string{
			"#N": "name",
		},
		ExpressionAttributeValues: map[string]dbtypes.AttributeValue{
			":name": &dbtypes.AttributeValueMemberS{Value: name},
		},
		KeyConditionExpression: aws.String("#N = :name"),
		Limit:                  aws.Int32(1),
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

	v, ok := res.Items[0]["version"]
	if !ok {
		return "", ErrSecretNotFound
	}

	if member, ok := v.(*dbtypes.AttributeValueMemberS); ok {
		return member.Value, nil
	}

	return "", ErrSecretNotFound
}

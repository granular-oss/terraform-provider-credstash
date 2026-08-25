package credstash

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/smithy-go"
	"github.com/secrethub/secrethub-go/pkg/randchar"
)

type Client struct {
	table string

	dynamoDB  dynamoDB
	decrypter decrypter
}

// Credential managed credential information
type Credential struct {
	Name      string `dynamodbav:"name"`
	Version   string `dynamodbav:"version"`
	Key       string `dynamodbav:"key"`
	Contents  string `dynamodbav:"contents"`
	Hmac      []byte `dynamodbav:"hmac"`
	CreatedAt int64  `dynamodbav:"created_at"`
}

const (
	DefaultKmsKey = "alias/credstash"
)

func New(table string, dynamoClient dynamoDB, kmsClient decrypter) *Client {
	return &Client{
		table:     table,
		decrypter: kmsClient,
		dynamoDB:  dynamoClient,
	}
}

func (c *Client) decryptCredential(ctx context.Context, cred *Credential, encCtx *EncryptionContextValue) (*DecryptedCredential, error) {

	wrappedKey, err := base64.StdEncoding.DecodeString(cred.Key)

	if err != nil {
		return nil, err
	}

	dk, err := c.DecryptDataKey(ctx, wrappedKey, encCtx)
	if err != nil {
		// Create reasoned responses to assist with debugging
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "AccessDeniedException":
				err = fmt.Errorf("KMS Access Denied to decrypt: %w", err)
			case "InvalidCiphertextException":
				err = fmt.Errorf("the encryption context provided may not match the one used when the credential was stored: %w", err)
			}
		}
		return nil, err
	}

	dataKey := dk.Plaintext[:32]
	hmacKey := dk.Plaintext[32:]

	contents, err := base64.StdEncoding.DecodeString(cred.Contents)
	if err != nil {
		return nil, err
	}

	hexhmac := ComputeHmac256(contents, hmacKey)

	if !bytes.Equal(hexhmac, cred.Hmac) {
		return nil, ErrHmacValidationFailed
	}

	secret, err := Decrypt(dataKey, contents)

	if err != nil {
		return nil, err
	}

	plainText := string(secret)

	return &DecryptedCredential{Credential: cred, Secret: plainText}, nil
}

// GetHighestVersionSecret retrieves latest secret from dynamodb using the name
func (c *Client) GetHighestVersionSecret(ctx context.Context, table string, name string, encContext *EncryptionContextValue) (*DecryptedCredential, error) {
	log.Print("Getting highest version secret")
	if table == "" {
		table = c.table
	}

	res, err := c.dynamoDB.Query(ctx, &dynamodb.QueryInput{
		TableName: &table,
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
	})

	if err != nil {
		return nil, err
	}

	cred := new(Credential)

	if len(res.Items) == 0 {
		return nil, ErrSecretNotFound
	}

	err = Decode(res.Items[0], cred)

	if err != nil {
		return nil, err
	}

	return c.decryptCredential(ctx, cred, encContext)
}

func (c *Client) GetSecret(ctx context.Context, name string, table string, paddedVersion string, encCtx *EncryptionContextValue) (*DecryptedCredential, error) {
	log.Printf("Getting secret: %s", name)

	if table == "" {
		table = c.table
	}
	log.Printf("GetSecret Final Table Name: %s", table)
	params := &dynamodb.GetItemInput{
		Key: map[string]dbtypes.AttributeValue{
			"name":    &dbtypes.AttributeValueMemberS{Value: name},
			"version": &dbtypes.AttributeValueMemberS{Value: paddedVersion},
		},
		TableName: &table,
	}
	log.Printf("GetSecret Params: %v", params)
	res, err := c.dynamoDB.GetItem(ctx, params)
	if err != nil {
		return nil, err
	}

	cred := new(Credential)
	log.Printf("GetSecret Items Found: %v", res)
	if len(res.Item) == 0 {

		return nil, ErrSecretNotFound
	}

	err = Decode(res.Item, cred)

	if err != nil {
		return nil, err
	}

	return c.decryptCredential(ctx, cred, encCtx)
}

// DecryptDataKey ask kms to decrypt the supplied data key
func (c *Client) DecryptDataKey(ctx context.Context, ciphertext []byte, encCtx *EncryptionContextValue) (*DataKey, error) {

	params := &kms.DecryptInput{
		CiphertextBlob:    ciphertext,
		EncryptionContext: *encCtx,
		GrantTokens:       []string{},
	}
	resp, err := c.decrypter.Decrypt(ctx, params)

	if err != nil {
		return nil, err
	}

	return &DataKey{
		CiphertextBlob: ciphertext,
		Plaintext:      resp.Plaintext, // transfer the plain text key after decryption
	}, nil
}

func (c *Client) GenerateRandomSecret(length int, useSymbols bool, charsets []interface{}, minRuleMap map[string]interface{}) (string, error) {
	charset := randchar.Charset{}
	if len(charsets) == 0 {
		charset = randchar.Alphanumeric
	}
	if useSymbols {
		charset = charset.Add(randchar.Symbols)
	}
	for _, charsetName := range charsets {
		set, found := randchar.CharsetByName(charsetName.(string))
		if !found {
			return "", fmt.Errorf("unknown charset: %s", charsetName)
		}
		charset = charset.Add(set)
	}
	var minRules []randchar.Option
	for charset, min := range minRuleMap {
		n := min.(int)
		set, found := randchar.CharsetByName(charset)
		if !found {
			return "", fmt.Errorf("unknown charset: %s", charset)
		}
		minRules = append(minRules, randchar.Min(n, set))
	}

	var err error
	rand, err := randchar.NewRand(charset, minRules...)
	if err != nil {
		return "", err
	}
	byteValue, err := rand.Generate(length)
	if err != nil {
		return "", err
	}
	value := string(byteValue)
	return value, nil
}

func (c *Client) PutSecret(ctx context.Context, tableName string, name string, value string, paddedVersion string, encCtx *EncryptionContextValue) error {
	log.Print("Putting secret")

	kmsKey := DefaultKmsKey

	if tableName == "" {
		tableName = c.table
	}

	dk, err := generateDataKey(ctx, c.decrypter, kmsKey, encCtx, 64)
	if err != nil {
		log.Printf("[DEBUG] GenerateDataKey failed: %v", err)
		return err
	}

	dataKey := dk.Plaintext[:32]
	hmacKey := dk.Plaintext[32:]
	wrappedKey := dk.CiphertextBlob

	ctext, err := Encrypt(dataKey, []byte(value))
	if err != nil {
		log.Printf("[DEBUG] Encrypt failed: %v", err)
		return err
	}

	b64hmac := ComputeHmac256(ctext, hmacKey)

	b64ctext := base64.StdEncoding.EncodeToString(ctext)

	cred := &Credential{
		Name:      name,
		Version:   paddedVersion,
		Key:       base64.StdEncoding.EncodeToString(wrappedKey),
		Contents:  b64ctext,
		Hmac:      b64hmac,
		CreatedAt: time.Now().Unix(),
	}

	data, err := attributevalue.MarshalMap(cred)

	if err != nil {
		log.Printf("[DEBUG] failed to DynamoDB marshal Record: %v", err)
		return err
	}

	_, err = c.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      data,
		ExpressionAttributeNames: map[string]string{
			"#N": "name",
		},
		ConditionExpression: aws.String("attribute_not_exists(#N)"),
	})

	return err

}

func (c *Client) DeleteSecret(ctx context.Context, tableName string, name string) error {
	log.Print("Deleting secret")

	if tableName == "" {
		tableName = c.table
	}

	res, err := c.dynamoDB.Query(ctx, &dynamodb.QueryInput{
		TableName: &tableName,
		ExpressionAttributeNames: map[string]string{
			"#N": "name",
		},
		ExpressionAttributeValues: map[string]dbtypes.AttributeValue{
			":name": &dbtypes.AttributeValueMemberS{Value: name},
		},
		KeyConditionExpression: aws.String("#N = :name"),
		ConsistentRead:         aws.Bool(true),
		ScanIndexForward:       aws.Bool(false), // descending order
	})

	if err != nil {
		return err
	}

	for _, item := range res.Items {
		cred := new(Credential)

		err = Decode(item, cred)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Deleting name: %s version: %v", cred.Name, cred.Version)

		_, err = c.dynamoDB.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &tableName,
			Key: map[string]dbtypes.AttributeValue{
				"name":    &dbtypes.AttributeValueMemberS{Value: cred.Name},
				"version": &dbtypes.AttributeValueMemberS{Value: cred.Version},
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}

const MaxPaddingLength = 19 // Number of digits in MaxInt64

// PaddedInt returns an integer left-padded with zeroes to the max-int length
func (c *Client) PaddedInt(i int) string {
	iString := strconv.Itoa(i)
	padLength := MaxPaddingLength - len(iString)
	return strings.Repeat("0", padLength) + strconv.Itoa(i)
}

// ResolveVersion converts an integer version to a string, or if a version isn't provided (0),
// returns "1" if the secret doesn't exist or the latest version plus one (auto-increment) if it does.
func (c *Client) ResolveVersion(ctx context.Context, tableName string, name string, version int) (string, error) {
	log.Print("Resolving version")

	if version != 0 {
		return c.PaddedInt(version), nil
	}

	if tableName == "" {
		tableName = c.table
	}

	ver, err := GetHighestVersion(ctx, c.dynamoDB, &tableName, name)
	if err != nil {
		if err == ErrSecretNotFound {
			return c.PaddedInt(1), nil
		}
		return "", err
	}

	if version, err = strconv.Atoi(ver); err != nil {
		return "", err
	}

	version++

	return c.PaddedInt(version), nil
}

package credstash

import (
	"encoding/base64"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/kms"
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

func New(table string, sess *session.Session) *Client {
	return &Client{
		table:     table,
		decrypter: kms.New(sess),
		dynamoDB:  dynamodb.New(sess),
	}
}

func (c *Client) GetSecret(name, table, version string, ctx map[string]string) (string, error) {
	if table == "" {
		table = c.table
	}
	material, err := getKeyMaterial(c.dynamoDB, name, version, table)
	if err != nil {
		return "", err
	}

	dataKey, hmacKey, err := decryptKey(c.decrypter, material.Key, ctx)
	if err != nil {
		return "", err
	}

	if err := checkHMAC(material, hmacKey); err != nil {
		return "", err
	}

	return decryptData(material, dataKey)
}

func (c *Client) PutSecret(tableName string, name string, value string, version string, ctx map[string]*string) error {
	log.Print("Putting secret")

	kmsKey := DefaultKmsKey

	// if alias != "" {
	// 	kmsKey = alias
	// }

	if tableName == "" {
		tableName = c.table
	}

	if version == "" {
		version = PaddedInt(1)
	}

	dk, err := generateDataKey(c.decrypter, kmsKey, ctx, 64)
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
		Version:   version,
		Key:       base64.StdEncoding.EncodeToString(wrappedKey),
		Contents:  b64ctext,
		Hmac:      b64hmac,
		CreatedAt: time.Now().Unix(),
	}

	data, err := dynamodbattribute.MarshalMap(cred)

	if err != nil {
		log.Printf("[DEBUG] failed to DynamoDB marshal Record: %v", err)
		return err
	}

	_, err = c.dynamoDB.PutItem(&dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      data,
		ExpressionAttributeNames: map[string]*string{
			"#N": aws.String("name"),
		},
		ConditionExpression: aws.String("attribute_not_exists(#N)"),
	})

	return err

}

func (c *Client) DeleteSecret(tableName string, name string) error {
	log.Print("Deleting secret")

	if tableName == "" {
		tableName = c.table
	}

	res, err := c.dynamoDB.Query(&dynamodb.QueryInput{
		TableName: &tableName,
		ExpressionAttributeNames: map[string]*string{
			"#N": aws.String("name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name": {
				S: aws.String(name),
			},
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

		_, err = c.dynamoDB.DeleteItem(&dynamodb.DeleteItemInput{
			TableName: &tableName,
			Key: map[string]*dynamodb.AttributeValue{
				"name": {
					S: aws.String(cred.Name),
				},
				"version": {
					S: aws.String(cred.Version),
				},
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
func PaddedInt(i int) string {
	iString := strconv.Itoa(i)
	padLength := MaxPaddingLength - len(iString)
	return strings.Repeat("0", padLength) + strconv.Itoa(i)
}

// ResolveVersion converts an integer version to a string, or if a version isn't provided (0),
// returns "1" if the secret doesn't exist or the latest version plus one (auto-increment) if it does.
func (c *Client) ResolveVersion(tableName string, name string, version int) (string, error) {
	log.Print("Resolving version")

	if version != 0 {
		return PaddedInt(version), nil
	}

	if tableName == "" {
		tableName = c.table
	}

	ver, err := GetHighestVersion(c.dynamoDB, &tableName, name)
	if err != nil {
		if err == ErrSecretNotFound {
			return PaddedInt(1), nil
		}
		return "", err
	}

	if version, err = strconv.Atoi(ver); err != nil {
		return "", err
	}

	version++

	return PaddedInt(version), nil
}

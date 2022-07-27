package credstash

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"

	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func dummyItemWithAllFields() map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"name":     attrValueString("test_key"),
		"version":  attrValueString("0000000000000000001"),
		"digest":   attrValueString("SHA256"),
		"hmac":     attrValueHexString([]byte{1, 2, 3, 4}),
		"contents": attrValueB64String([]byte{1, 2, 3, 4}),
		"key":      attrValueB64String([]byte{1, 2, 3, 4}),
	}
}

func dummyItemWithBinaryHMAC(hmac string) map[string]*dynamodb.AttributeValue {
	item := dummyItemWithAllFields()
	item["hmac"] = &dynamodb.AttributeValue{B: []byte(hmac)}
	return item
}

func dummyItemWithWrongKey() map[string]*dynamodb.AttributeValue {
	item := dummyItemWithAllFields()
	item["key"] = attrValueString("not base64")
	return item
}

func dummyItemWithMissingKey() map[string]*dynamodb.AttributeValue {
	item := dummyItemWithAllFields()
	delete(item, "key")
	return item
}

func attrValueString(v string) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{S: aws.String(v)}
}

func attrValueHexString(d []byte) *dynamodb.AttributeValue {
	return attrValueString(hex.EncodeToString(d))
}

func attrValueB64String(d []byte) *dynamodb.AttributeValue {
	return attrValueString(base64.StdEncoding.EncodeToString(d))
}

func encrypt(t *testing.T, key, nonce, data []byte) []byte {
	b, err := aes.NewCipher(key)
	assertNoError(t, err)

	s := cipher.NewCTR(b, nonce)

	result := make([]byte, len(data))
	s.XORKeyStream(result, data)

	return result
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
func assertError(t *testing.T, err error) {
	if err == nil {
		t.Fatalf("should have been an error")
	}
}

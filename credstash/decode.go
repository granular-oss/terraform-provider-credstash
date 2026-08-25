package credstash

// Code From Unicreds
// https://github.com/Versent/unicreds/blob/master/LICENSE.md

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// adjustHmac will force the hmac to be a byte array if present as string
func adjustHmac(record map[string]dbtypes.AttributeValue) {
	if val, ok := record["hmac"]; ok {
		// If it's stored as a string (S), convert to binary (B)
		if member, ok := val.(*dbtypes.AttributeValueMemberS); ok {
			record["hmac"] = &dbtypes.AttributeValueMemberB{Value: []byte(member.Value)}
		}
	}
}

// Decode decode the supplied struct from the dynamodb result map
func Decode(data map[string]dbtypes.AttributeValue, rawVal interface{}) error {
	adjustHmac(data)
	return attributevalue.UnmarshalMap(data, rawVal)
}

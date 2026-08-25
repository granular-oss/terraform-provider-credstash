package credstash

// Code From Unicreds
// https://github.com/Versent/unicreds/blob/master/LICENSE.md

import (
	"fmt"
	"testing"

	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {

	cred := struct {
		Name      string `dynamodbav:"name"`
		Timestamp int64  `dynamodbav:"timestamp"`
	}{}

	data := map[string]dbtypes.AttributeValue{
		"name":      &dbtypes.AttributeValueMemberS{Value: "data"},
		"timestamp": &dbtypes.AttributeValueMemberN{Value: "1449038525717338459"},
	}

	err := Decode(data, &cred)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	assert.Equal(t, "data", cred.Name)
	assert.Equal(t, int64(1449038525717338459), cred.Timestamp)
}

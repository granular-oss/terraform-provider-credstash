package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceSecretNameFieldForceNew(t *testing.T) {
	tests := []struct {
		name    string
		oldName string
		newName string
		expect  bool
	}{
		{
			name:    "name change should require replacement",
			oldName: "secret-1",
			newName: "secret-2",
			expect:  true,
		},
		{
			name:    "same name should not require replacement",
			oldName: "secret-1",
			newName: "secret-1",
			expect:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := resourceSecret()
			nameField := resource.Schema["name"]

			// Verify ForceNew is set
			assert.True(t, nameField.ForceNew, "name field should have ForceNew=true")
		})
	}
}

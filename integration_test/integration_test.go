package test

import (
	"log"
	"strconv"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
)

var credstash *credstashCli

func TestMain(m *testing.M) {
	setup()
	m.Run()
	teardown()
}

func setup() {
	log.Println("\n-----Starting setup-----")
	n := 1
	log.Println("Testing with credstash CLI.")
	credstash = newCredstashCliCustomTable("terraform-provider-credstash-test-table")
	log.Println("Cleaning up previously added credstash keys...")
	credstash.list()
	for n < 8 {
		credstash.delete("terraform-provider-credstash-integration-test-" + strconv.Itoa(n))
		n += 1
	}
	log.Println("Clean up complete.  Adding starting values...")
	credstash.put("terraform-provider-credstash-integration-test-4", "test-4-1", 0)
	credstash.put("terraform-provider-credstash-integration-test-4", "test-4-2", 0)
	credstash.put("terraform-provider-credstash-integration-test-5", "test-5-1", 0)
	credstash.put("terraform-provider-credstash-integration-test-5", "test-5-2", 0)
	credstash.put("terraform-provider-credstash-integration-test-7", "test-7-1", 0)
	credstash.put("terraform-provider-credstash-integration-test-7", "test-7-2", 0)
	credstash.put("terraform-provider-credstash-integration-test-7", "test-7-3", 0)

	log.Println("\n-----Setup complete-----")
}

func teardown() {
	// retryable errors in terraform testing.

	log.Println("\n----Teardown complete----")
}

func TestTerraform(t *testing.T) {
	// retryable errors in terraform testing.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./tf/",
		EnvVars: map[string]string{
			"TF_CLI_CONFIG_FILE": "dev.tfrc",
		},
		Vars: map[string]interface{}{
			"secret_1_version": 1,
			"secret_1":         "IAmABadPassword1",
			"secret_3":         "IAmABadPassword2",
			"secret_5":         "test-5-2",
		},
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.Init(t, terraformOptions)

	// Import credstash key into terraform
	args := terraform.FormatArgs(terraformOptions, "import")
	args = append(args, "credstash_secret.terraform-provider-credstash-integration-test-5", "terraform-provider-credstash-integration-test-5")
	terraform.RunTerraformCommand(t, terraformOptions, args...)
	//Apply Terraform and Test Apply is Idempotent
	terraform.ApplyAndIdempotent(t, terraformOptions)
	//Parse the terraform show output
	show := terraform.Show(t, terraformOptions)
	state := parseShowJson(show)
	log.Println(state)

	// This Run will not return until its parallel subtests complete.
	// Test terraform state
	t.Run("group", func(t *testing.T) {
		t.Run("TestSecretImport", importSecretTest(state))
		t.Run("createValueAndVersionTest", createValueAndVersionTest(state))
		t.Run("createGeneratedValueOnlyTest", createGeneratedValueOnlyTest(state))
		t.Run("createGeneratedValueAndVersionTest", createGeneratedValueAndVersionTest(state))
		t.Run("createValueAndNoVersionTest", createValueAndNoVersionTest(state))
		t.Run("datablockNameOnlyTest", datablockNameOnlyTest(state))
		t.Run("datablockNameAndVersionTest", datablockNameAndVersionTest(state))
	})

	// Update Password Variables
	terraformOptions = terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./tf/",
		EnvVars: map[string]string{
			"TF_CLI_CONFIG_FILE": "dev.tfrc",
		},
		Vars: map[string]interface{}{
			"secret_1_version": 3,
			"secret_1":         "IAmABadPassword3",
			"secret_3":         "IAmABadPassword2Again",
			"secret_5":         "test-5-2",
		},
	})

	//Apply Terraform and Test Apply is Idempotent
	terraform.ApplyAndIdempotent(t, terraformOptions)

	//Parse the terraform show output
	show = terraform.Show(t, terraformOptions)
	state = parseShowJson(show)

	// This Run will not return until its parallel subtests complete.
	// Test terraform state
	t.Run("group", func(t *testing.T) {
		t.Run("createValueAndVersionTest2", createValueAndVersionTest2(state))
		t.Run("createValueAndNoVersionTest2", createValueAndNoVersionTest2(state))
	})
}

/*
	Test the imported credstash secret
*/
func importSecretTest(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-5 Imported Value Matches credstash latest version
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-5", 2)
		terraformValue := showState["credstash_secret.terraform-provider-credstash-integration-test-5"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-5 Imported Version is 0, so we can autoincrementin the future
		terraformVersion := showState["credstash_secret.terraform-provider-credstash-integration-test-5"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(0))
		// Test terraform-provider-credstash-integration-test-5 Imported Name is correct
		terraformName := showState["credstash_secret.terraform-provider-credstash-integration-test-5"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-5")
	}
}

/*
	Test credstash resource creation that specifies a value and version
*/
func createValueAndVersionTest(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-1 Imported Value Matches credstash latest version
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-1", 1)
		terraformValue := showState["credstash_secret.terraform-provider-credstash-integration-test-1"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-1 Imported Version is 1
		terraformVersion := showState["credstash_secret.terraform-provider-credstash-integration-test-1"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(1))
		// Test terraform-provider-credstash-integration-test-1 Imported Name is correct
		terraformName := showState["credstash_secret.terraform-provider-credstash-integration-test-1"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-1")
		// Test that credstash latest version is 1
		credstashVersion := credstash.getLatestVersion("terraform-provider-credstash-integration-test-1")
		assert.Equal(t, credstashVersion, "1")
	}
}

/*
	Test credstash resource creation that autogenerate a value
*/
func createGeneratedValueOnlyTest(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-2 Imported Value Matches credstash latest version
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-2", 1)
		terraformValue := showState["credstash_secret.terraform-provider-credstash-integration-test-2"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-2 Imported Version is 0, so we can autoincrementin the future
		terraformVersion := showState["credstash_secret.terraform-provider-credstash-integration-test-2"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(0))
		// Test terraform-provider-credstash-integration-test-2 Imported Name is correct
		terraformName := showState["credstash_secret.terraform-provider-credstash-integration-test-2"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-2")
		// Test that credstash latest version is 1
		credstashVersion := credstash.getLatestVersion("terraform-provider-credstash-integration-test-2")
		assert.Equal(t, credstashVersion, "1")
	}
}

/*
	Test credstash resource creation that auto generates a value and specifies a version
*/
func createGeneratedValueAndVersionTest(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-6 Imported Value Matches credstash latest version
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-6", 10)
		terraformValue := showState["credstash_secret.terraform-provider-credstash-integration-test-6"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-6 Imported Version is 10
		terraformVersion := showState["credstash_secret.terraform-provider-credstash-integration-test-6"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(10))
		// Test terraform-provider-credstash-integration-test-6 Imported Name is correct
		terraformName := showState["credstash_secret.terraform-provider-credstash-integration-test-6"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-6")
		// Test that credstash latest version is 10
		credstashVersion := credstash.getLatestVersion("terraform-provider-credstash-integration-test-6")
		assert.Equal(t, credstashVersion, "10")
	}
}

/*
	Test credstash resource creation that specifies a value without a version
*/
func createValueAndNoVersionTest(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-3 Imported Value Matches credstash latest version
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-3", 1)
		terraformValue := showState["credstash_secret.terraform-provider-credstash-integration-test-3"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-3 Imported Version is 0, so we can autoincrementin the future
		terraformVersion := showState["credstash_secret.terraform-provider-credstash-integration-test-3"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(0))
		// Test terraform-provider-credstash-integration-test-3 Imported Name is correct
		terraformName := showState["credstash_secret.terraform-provider-credstash-integration-test-3"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-3")
		// Test that credstash latest version is 1
		credstashVersion := credstash.getLatestVersion("terraform-provider-credstash-integration-test-3")
		assert.Equal(t, credstashVersion, "1")
	}
}

/*
	Test credstash data read that specifies a name only
*/
func datablockNameOnlyTest(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-4 Imported Value Matches credstash latest version
		t.Log(showState)
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-4", 2)
		terraformValue := showState["data.credstash_secret.terraform-provider-credstash-integration-test-4"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-4 Imported Version is 0, so we can autoincrementin the future
		terraformVersion := showState["data.credstash_secret.terraform-provider-credstash-integration-test-4"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(0))
		// Test terraform-provider-credstash-integration-test-4 Imported Name is correct
		terraformName := showState["data.credstash_secret.terraform-provider-credstash-integration-test-4"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-4")
	}
}

/*
	Test credstash data read that specifies a name and version
*/
func datablockNameAndVersionTest(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-7 Imported Value Matches credstash latest version
		t.Log(showState)
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-7", 2)
		terraformValue := showState["data.credstash_secret.terraform-provider-credstash-integration-test-7"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-7 Imported Version is 2
		terraformVersion := showState["data.credstash_secret.terraform-provider-credstash-integration-test-7"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(2))
		// Test terraform-provider-credstash-integration-test-7 Imported Name is correct
		terraformName := showState["data.credstash_secret.terraform-provider-credstash-integration-test-7"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-7")
	}
}

/*
	This Test Runs after we apply a new password and version. Test that the resource updates the version and value
*/
func createValueAndVersionTest2(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-1 Imported Value Matches credstash latest version
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-1", 3)
		terraformValue := showState["credstash_secret.terraform-provider-credstash-integration-test-1"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-1 Imported Version is 3
		terraformVersion := showState["credstash_secret.terraform-provider-credstash-integration-test-1"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(3))
		// Test terraform-provider-credstash-integration-test-1 Imported Name is correct
		terraformName := showState["credstash_secret.terraform-provider-credstash-integration-test-1"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-1")
		// Test that credstash latest version is 3
		credstashVersion := credstash.getLatestVersion("terraform-provider-credstash-integration-test-1")
		assert.Equal(t, credstashVersion, "3")
	}
}

/*
	This Test Runs after we apply a new password. Test that the resource updates the value
*/
func createValueAndNoVersionTest2(showState map[string]*tfjson.StateResource) func(*testing.T) {
	return func(t *testing.T) {
		//Test terraform-provider-credstash-integration-test-3 Imported Value Matches credstash latest version
		credstashValue := credstash.get("terraform-provider-credstash-integration-test-3", 2)
		terraformValue := showState["credstash_secret.terraform-provider-credstash-integration-test-3"].AttributeValues["value"].(string)
		assert.Equal(t, terraformValue, credstashValue)
		// Test terraform-provider-credstash-integration-test-3 Imported Version is 0, so we can autoincrementin the future
		terraformVersion := showState["credstash_secret.terraform-provider-credstash-integration-test-3"].AttributeValues["version"].(float64)
		assert.Equal(t, terraformVersion, float64(0))
		// Test terraform-provider-credstash-integration-test-3 Imported Name is correct
		terraformName := showState["credstash_secret.terraform-provider-credstash-integration-test-3"].AttributeValues["name"].(string)
		assert.Equal(t, terraformName, "terraform-provider-credstash-integration-test-3")
		// Test that credstash latest version is 2
		credstashVersion := credstash.getLatestVersion("terraform-provider-credstash-integration-test-3")
		assert.Equal(t, credstashVersion, "2")
	}
}

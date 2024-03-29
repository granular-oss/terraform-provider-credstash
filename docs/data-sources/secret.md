---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "credstash_secret Data Source - terraform-provider-credstash"
subcategory: ""
description: |-
  
---

# credstash_secret (Data Source)



## Example Usage

```terraform
# Read a secret named "rds_password" from credstash, will get latest version
data "credstash_secret" "rds_password" {
  name = "rds_password"
}

# Read version 1 of some_secret from credstashsome_secret
data "credstash_secret" "my_secret" {
  name    = "some_secret"
  version = "0000000000000000001"
}

# Use the rds_password password you read above, as a value in another resource.
resource "aws_db_instance" "postgres" {
  password = data.credstash_secret.rds_password.value
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) name of the secret

### Optional

- `context` (Map of String) encryption context for the secret
- `table` (String) name of DynamoDB table where the secrets are stored
- `version` (Number) version of the secrets

### Read-Only

- `id` (String) The ID of this resource.
- `value` (String, Sensitive) value of the secret



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

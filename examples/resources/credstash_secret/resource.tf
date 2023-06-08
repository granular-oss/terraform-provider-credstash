
# Generate a secret in credstash using 34 radomchars, name "test.secret"
resource "credstash_secret" "test_secret" {
  name = "test.secret"
  generate {
    length = 34
  }
}

# Create a secret in credstash using the provided value, name "my_pub_key"
resource "credstash_secret" "my_pub_key" {
  name  = "my_pub_key"
  value = file("${path.root}/id_rsa.pub")
}

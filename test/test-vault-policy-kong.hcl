# All capabilities in the Kong namespace
path "secret/edgex/edgex-kong/*" {
  capabilities = ["create", "update", "delete", "list", "read"]
}

# List/Read only for the TLS materials: private key and certificate
path "secret/edgex/pki/tls/edgex-kong" {
  capabilities = ["list", "read"]
}

# Kong can renew its own creds lease (vault lease renew <lease id>)
path "sys/leases/renew" {
  capabilities = ["create"]
}

# Kong can revoke its own creds lease (vault lease revoke <lease id>)
path "sys/leases/revoke" {
  capabilities = ["update"]
}
#!/bin/bash

CN=$1
SERVER_ID=$2

# Generate dynamic openssl.cnf
sed "s/{{SERVER_SAN}}/localhost:$CN/g; s/{{CLIENT_SAN}}/test-client/g" ./certificate/openssl.cnf.template > ./certificate/openssl.cnf

echo "Generated openssl.cnf with SERVER_SAN=$CN"

# Create a single CA for both client and server.
openssl req -x509                                     \
  -newkey rsa:4096                                    \
  -nodes                                              \
  -days 3650                                          \
  -keyout ca_key_server$SERVER_ID.pem                 \
  -out ca_cert_server$SERVER_ID.pem                   \
  -subj "/C=US/ST=CA/L=SVL/O=gRPC/CN=test-ca/"        \
  -config ./certificate/openssl.cnf                   \
  -extensions test_ca                                 \
  -sha256

# Generate a server cert signed by the same CA.
openssl genrsa -out server_key_server$SERVER_ID.pem 4096
openssl req -new                                    \
  -key server_key_server$SERVER_ID.pem              \
  -days 3650                                        \
  -out server_csr.pem                               \
  -subj "/C=US/ST=CA/L=SVL/O=gRPC/CN=$CN/"          \
  -config ./certificate/openssl.cnf                 \
  -reqexts test_server

openssl x509 -req           \
  -in server_csr.pem        \
  -CAkey ca_key_server$SERVER_ID.pem         \
  -CA ca_cert_server$SERVER_ID.pem           \
  -days 3650                \
  -set_serial 1000          \
  -out server_cert_server$SERVER_ID.pem      \
  -extfile ./certificate/openssl.cnf         \
  -extensions test_server   \
  -sha256

openssl verify -verbose -CAfile ca_cert_server$SERVER_ID.pem  server_cert_server$SERVER_ID.pem

# Generate a client cert signed by the same CA.
openssl genrsa -out client_key_server$SERVER_ID.pem 4096
openssl req -new                                    \
  -key client_key_server$SERVER_ID.pem              \
  -days 3650                                        \
  -out client_csr.pem                               \
  -subj "/C=US/ST=CA/L=SVL/O=gRPC/CN=test-client/"  \
  -config ./certificate/openssl.cnf                 \
  -reqexts test_client

openssl x509 -req           \
  -in client_csr.pem        \
  -CAkey ca_key_server$SERVER_ID.pem         \
  -CA ca_cert_server$SERVER_ID.pem           \
  -days 3650                \
  -set_serial 1001          \
  -out client_cert_server$SERVER_ID.pem      \
  -extfile ./certificate/openssl.cnf         \
  -extensions test_client   \
  -sha256

openssl verify -verbose -CAfile ca_cert_server$SERVER_ID.pem  client_cert_server$SERVER_ID.pem

rm *_csr.pem

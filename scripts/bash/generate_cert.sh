#!/bin/sh
openssl req -new -key ./keys/cert.key -subj "/CN=$1" -sha256 | openssl x509 -req -days 3650 -CA ./keys/ca.crt -CAkey ./keys/ca.key -set_serial "$2" > ./certificates/"$1".crt
#openssl req -new -nodes -newkey rsa:2048 -keyout "./keys/$1.key" -out "./keys/$1.csr" -subj "/C=US/ST=$1"
#openssl x509 -req -sha256 -days 300 -in "./keys/$1.csr" -CA ./keys/ca.pem -CAkey ca.key -CAcreateserial -extfile "cfgs/$1.ext" -out "certificates/$1.crt"
#!/bin/sh
mkdir ../../keys/
openssl genrsa -out ../../keys/ca.key 2048
openssl req -new -x509 -days 3650 -key ../../keys/ca.key -out ../../keys/ca.crt -subj "/CN=yngwie proxy CA"
openssl genrsa -out ../../keys/cert.key 2048
mkdir ../../certificates/


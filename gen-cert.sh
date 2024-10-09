#!/bin/bash

# COPIED/MODIFIED from the redis server gen-certs util

# Generate some test certificates which are used by the regression test suite:
#
#   config/cert/ca.{crt,key,txt}      Self signed CA certificate.
#   config/cert/cert.{crt,key}        A certificate with no key usage/policy restrictions.
#   config/cert/client.{crt,key}      A certificate restricted for SSL client usage.
#   config/cert/server.{crt,key}      A certificate restricted for SSL server usage.
#   config/cert/application.dh              DH Params file.

generate_cert() {
    local name=$1
    local cn="$2"
    local opts="$3"

    local keyfile=config/cert/${name}.key
    local certfile=config/cert/${name}.crt

    [ -f $keyfile ] || openssl genrsa -out $keyfile 2048
    openssl req \
        -new -sha256 \
        -subj "/O=Application Test/CN=$cn" \
        -key $keyfile | \
        openssl x509 \
            -req -sha256 \
            -CA config/cert/ca.crt \
            -CAkey config/cert/ca.key \
            -CAserial config/cert/ca.txt \
            -CAcreateserial \
            -days 365 \
            $opts \
            -out $certfile
}

rm -rf config/cert
mkdir -p config/cert
[ -f config/cert/ca.key ] || openssl genrsa -out config/cert/ca.key 4096
openssl req \
    -x509 -new -nodes -sha256 \
    -key config/cert/ca.key \
    -days 3650 \
    -subj '/O=Application Test/CN=Certificate Authority' \
    -out config/cert/ca.crt

cat > config/cert/openssl.cnf <<_END_
[ server_cert ]
keyUsage = digitalSignature, keyEncipherment
nsCertType = server
[ client_cert ]
keyUsage = digitalSignature, keyEncipherment
nsCertType = client
_END_

generate_cert server "Server-only" "-extfile config/cert/openssl.cnf -extensions server_cert"
generate_cert client "Client-only" "-extfile config/cert/openssl.cnf -extensions client_cert"
generate_cert cert "Generic-cert"

[ -f config/cert/application.dh ] || openssl dhparam -out config/cert/application.dh 2048
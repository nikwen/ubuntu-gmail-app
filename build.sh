#!/bin/bash

DIR=$(dirname $(readlink -f "$0"))
export GOPATH=$DIR

go get -d -u gopkg.in/qml.v1 code.google.com/p/google-api-go-client/gmail/v1 code.google.com/p/goauth2/oauth
go install -v -x ubuntu-gmail-app
cp bin/ubuntu-gmail-app $DIR

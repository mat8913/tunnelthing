#!/bin/sh

set -e

mkdir -p bin/

go build -o bin/ github.com/mat8913/tunnelthing/cmd/tt-connect
go build -o bin/ github.com/mat8913/tunnelthing/cmd/tt-gencert
go build -o bin/ github.com/mat8913/tunnelthing/cmd/tt-listen
go build -o bin/ github.com/mat8913/tunnelthing/cmd/tt-ping

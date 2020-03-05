#!/usr/bin/env bash

java -jar antlr.jar -Dlanguage=Go -o aparser Dandelion.g4 DandelionLex.g4
go build

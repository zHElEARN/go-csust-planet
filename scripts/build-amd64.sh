#!/bin/bash
docker buildx build --platform linux/amd64 -t go-csust-planet -o type=docker,dest=./go-csust-planet.tar .

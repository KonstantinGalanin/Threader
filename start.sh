#!/bin/bash

export ADDR=":8000"
export TOKEN_SECRET="HJFSKTEXIGTE"

docker-compose up -d --wait

until nc -z -v -w60 localhost 3306
do
  sleep 1
done

until nc -z -v -w60 localhost 27017
do
  sleep 1
done

until nc -z -v -w60 localhost 6379
do
  sleep 1
done

go build -o redditclone ./cmd/redditclone/main.go

./redditclone
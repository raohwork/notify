#!/bin/bash

tag="$1"
if [[ $tag == "" ]]
then
    tag=latest
fi

docker run -it --rm --name pgsqldrv_test -p 5432:5432 -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test "postgres:${tag}"

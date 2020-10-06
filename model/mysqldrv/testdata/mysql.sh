#!/bin/bash

tag="$1"
if [[ $tag == "" ]]
then
    tag=8
fi

docker run -it --rm --name mysqldrv_test -p 3306:3306 -e MYSQL_ROOT_PASSWORD=test -e MYSQL_DATABASE=test "mysql:${tag}"

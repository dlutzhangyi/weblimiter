# How to run redis by docker

## Start a redis container
```
docker run -itd --name redis-test -p 6379:6379 redis
```

## Use redis-cli to set and get value
```
$ docker exec -it redis-test /bin/bash
root@a7f3d7af9039:/data# redis-cli
127.0.0.1:6379> get name
"1"
127.0.0.1:6379> set name 2
OK
127.0.0.1:6379> get name
"2"
```

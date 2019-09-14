## Usage

```shell
docker-compose up
```

Then send some data in:
```shell
echo -e "26:\x82\xa6\x64\x6f\x6d\x61\x69\x6e\xa9\x6c\x6f\x63\x61\x6c\x68\x6f\x73\x74\xa2\x69\x70\xce\x7f\x00\x00\x01,24:\x82\xa6\x64\x6f\x6d\x61\x69\x6e\xab\x65\x78\x61\x6d\x70\x6c\x65\x2e\x63\x6f\x6d\xa2\x69\x70\x00," | nc localhost 8080
```

or

```shell
cd helper
go run .
```

Strings are netstring formatted i.e. 
```
11:hello world,
```

## Задание

Тестовое задание по Go:
Написать сервис, который будет обрабатывать входящие запросы по TCP в формате MessagePack, сохранять в памяти значения полей "domain" и "ip" в формате key:value с TTL 10 секунд, выводя при этом каждую секунду полный список хранимых данных.

Форматы: domain - string, ip — uint32

По-умолчанию TCP-сервер слушает на 0.0.0.0:6000. Но эти параметры можно изменить с помощью задания следующих опций при старте:
--tcpport=8080 - порт TCP-сервера;
--tcpaddr=127.0.0.1 - адрес TCP-сервера;
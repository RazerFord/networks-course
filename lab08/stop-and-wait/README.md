### Инструкция к запуску

Сначала необходимо запустить сервера, затем уже клиент

#### Запуск сервера

Для запуска сервера можно использовать следующую команду

```shell
go run ./cmd/server/server.go -file <file> [-address localhost] [-port 8080] [-timeout 100]

-address string
      server address (default "localhost")
-file string
      file name
-port int
      server port (default 8888)
-timeout int
      time-out, ms (default 100)
```

#### Запуск клиента

Для запуска клиента можно использовать следующую команду

```shell
go run ./cmd/client/client.go  -file <file> [-address localhost] [-port 8080] [-timeout 100]

-address string
      server address (default "localhost")
-file string
      file name
-port int
      server port (default 8888)
-timeout int
      time-out (default 100)
```

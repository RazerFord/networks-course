# Запуск консольной утилиты

Для запуска можно использовать следующую команду

```shell
go run ./cmd/freeport/main.go -ip 0.0.0.0 -start 1000 -end 10000

-end int
      end of port range (excluding) (default 65536)
-ip string
      ip address (default "0.0.0.0")
-start int
      start of port ranges
```
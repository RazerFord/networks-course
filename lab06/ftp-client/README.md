# Инструкция к запуску

Для того, чтобы запустить клиент достаточно ввести следующую команду

```shell
go run ./cmd/ftpclient/main.go -user <user> -pass <pass> -addr <addr> -port <port>

-addr string
      ftp server address (default "localhost")
-pass string
      user password (default "anonymous@domain.com")
-port int
      ftp server port (default 21)
-user string
      username (default "anonymous")
```

Пример команды

```shell
go run ./cmd/ftpclient/main.go -user dlpuser -pass rNrKYTX9g7z3RgJRmxWuGHbeu -addr ftp.dlptest.com
```

В `FTP` клиенте доступы следующие команды:

```shell
list - list of files and directories
retr - download file from source to target
stor - load file from source to target
help - withdraw help
quit - go out
```

Пример использования команды `list`
```shell
list
Select path:
.
drwxr-xr-x    3 1001     1001           16 Mar 25 11:10 2024
```

Пример использования команды `retr`
```shell
retr
Select source:
boo.txt
Select target:
foo.txt
File downloaded
```

Пример использования команды `stor`
```shell
stor
Select source:
boo.txt
Select target:
foo.txt
File uploaded
```

### Инструкция к запуску

Чтобы отправить сообщение, используйте следующую команду:

```shell
go run cmd/smtpclient/main.go -to recipient -from sender -password password -smtp "smtp protocol" -port port -message "path to message"

-from string
    mail sender address
-message string
    path to sent message
-password string
    sender of email password
-port int
    smtp server port
gi-smtp string
    smtp server address
-to string
    mail recipient address

```

Пример команды
```shell
go run cmd/smtpclient/main.go -to to@mail.com -from from@mail.com -password pass -smtp smtp.mail.com -port 567 -message ./test.html
```

### Получение пароля

Обычный пароль для отправки сообщения не подойдет. Нужно использовать пароль для приложений. Его можно получить на следующей странице [тык](https://id.yandex.ru/security/app-passwords) для `Яндекс` почты, для -- `Google` почты необходимо перейти на страницу с подробной инструкцией [тык](https://support.google.com/accounts/answer/185833?hl=ru).

> Для каждой почты необходимо указывать ее почтовый сервер и порт

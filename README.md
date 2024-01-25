# bore-server

Promote ADMIN role to user

```
go run . promote asdasd@asdsad.asd ADMIN
```

```
go generate ./...
```

```
go run . serve
```

## TODO

- token invalidation

### Генерация private/public ключа

Ключи должны быть в PEM формате.
Для генерации требуется openssl v3+

на macos можно поставить через `brew install openssl`
бинарник можно найти тут `/usr/local/opt/openssl/bin/openssl`

```bash
openssl version
# OpenSSL 3.0.5 5 Jul 2022 (Library: OpenSSL 3.0.5 5 Jul 2022)

# Создаем приватный ключ
openssl genpkey -algorithm Ed25519 -out private.pem

# Из private.pem генерим публичный ключ
openssl pkey -in private.pem -out public.pem -pubout
```

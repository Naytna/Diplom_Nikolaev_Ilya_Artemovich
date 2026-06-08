# RSL Learning Generator

Автономный модуль генерации учебных заданий по РЖЯ.

## Запуск

```bash
docker compose up --build
```

## Проверка

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/translated-words
curl -X POST http://localhost:8080/api/themes/1/generate
curl http://localhost:8080/api/themes/1/exercises
```

## Тестовые пользователи

admin@example.ru / password
expert@example.ru / password
student@example.ru / password
guest@example.ru / password

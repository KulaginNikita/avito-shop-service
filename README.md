# Avito Shop Service
  
## Сервис позволяет сотрудникам:
- При первой авторизации автоматически создавается аккаунт с 1000 монетами
- Покупать товары за монеты
- Переводить монеты другим сотрудникам
- Просматривать купленные товары и историю транзакций

## Стек технологий
- Go
- PostgreSQL
- Docker
- JWT

## Как запустить сервис:

1. Клонируйте репозиторий: 
```
git clone https://github.com/KulaginNikita/avito-shop-service
cd avito-shop-service
```
2. Запустите:
```
docker-compose up --build
```
## Тестирование:

- Юнит-тесты для бизнес-логики находятся в `internal/service/service_test.go`.
- E2E-тесты для API — в `tests/e2e_test.go`.


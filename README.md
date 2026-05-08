# Сервис "Гофермарт" 
 
Накопительная система лояльности

- [Структура базы данных](docs/Database.mmd)
- [Диаграмма последовательности](docs/Sequence.mmd)
- [Диаграмма последовательности асинхронной обработки заказов](docs/AsyncSequence.mmd)
- [Блок-схема goroutine и каналов асинхронной обработки заказов](docs/AsyncFlowchart.mmd)

## Сборка

```
make build
```

## Запуск docker-compose

Зависимости:
- база данных `PostgreSQL`

```bash
# -d - для запуска в фоне
docker compose up -d
```

Остановка `docker-compose`
```bash
docker compose down
# -v - для удаления volumes
docker compose down -v
```

Просмотр логов
```bash
# docker compose logs -f <имя_контейнера>
docker compose logs -f db
```

### PostgreSQL

Подключение к базе данных через консоль:
```bash
# docker exec -it <имя_контейнера> psql -U <пользователь> -d <база_данных>
docker compose exec db psql -U username -d dbname
docker compose exec db psql -U username -d gophermart_db
docker compose exec db psql -U username -d accrual_db
```

Консольные команды:
- `\dt`- список таблиц
- `\d  <имя_таблицы>` - структура таблицы
- `\q` - выход из консоли

## Запуск сервиса

### Запуск сервиса accrual - системы расчёта начислений
```
# параметры запуска
./cmd/accrual/accrual_linux_amd64 -h

# пример запуска
./cmd/accrual/accrual_linux_amd64 \
    -a ":8081" \
    -d "host=localhost user=username password=password dbname=accrual_db sslmode=disable" 
```

### Запуск сервитса gopgermart - накопительной системы лояльности
```
# параметры запуска
./cmd/gophermart/gophermart -h

# пример запуска
./cmd/gophermart/gophermart \
    -a ":8080" \
    -d "host=localhost user=username password=password dbname=gophermart_db sslmode=disable" \
    -r ":8081" \
    -ll "debug" \
    -lf "text"
```

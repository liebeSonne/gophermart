# Сервис "Гофермарт" 
 
Накопительная система лояльности

- [Структура базы данных](docs/Database.mmd)
- [Диаграмма последовательности](docs/Sequence.mmd)

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
```

Консольные команды:
- `\dt`- список таблиц
- `\d  <имя_таблицы>` - структура таблицы
- `\q` - выход из консоли

## Запуск сервиса

```
# параметры запуска
./cmd/gophermart/gophermart -h

# пример запуска
./cmd/gophermart/gophermart \
    -a ":8080" \
    -d "host=localhost user=username password=password dbname=dbname sslmode=disable" \
    -ll "debug" \
    -lf "text"

```

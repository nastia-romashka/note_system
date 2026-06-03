# Note System

Микросервисная система для ведения заметок с категориями, тегами, файловыми вложениями, графом связей между заметками и полнотекстовым поиском.

## Возможности

- регистрация, вход и обновление JWT-токена через `api_service`
- создание, редактирование, удаление и дублирование заметок
- работа с категориями и древовидной структурой категорий
- работа с тегами
- загрузка, скачивание и удаление файлов, привязанных к заметкам
- отображение связей между заметками в виде графа
- поиск по заметкам через `search_service`
- веб-интерфейс на `React + Vite`

## Архитектура

Проект состоит из API-шлюза, набора микросервисов и инфраструктурных контейнеров.

- `api_service` - единая точка входа, JWT-аутентификация, маршрутизация и агрегация запросов
- `note_service` - заметки, теги и календарные события, хранение в `MongoDB`
- `user_service` - пользователи, профиль и история действий, хранение в `PostgreSQL`
- `category_service` - категории и граф связей заметок, реализован на `FastAPI`, хранение в `Neo4j`
- `file_service` - файловые вложения, хранение в `MinIO`
- `search_service` - индексация и поиск заметок через `Typesense`
- `web_client` - клиентское приложение

Инфраструктурные сервисы:

- `mongo_notes`
- `postgres_users`
- `neo4j`
- `minio`
- `typesense`

## Технологии

- `Go` - `api_service`, `note_service`, `user_service`, `file_service`, `search_service`
- `Python + FastAPI` - `category_service`
- `React + Vite` - `web_client`
- `MongoDB`, `PostgreSQL`, `Neo4j`, `MinIO`, `Typesense`
- `Docker Compose` для локального запуска всей системы

## Структура репозитория

- `api_service/` - gateway API
- `note_service/` - сервис заметок
- `user_service/` - сервис пользователей
- `category_service/` - сервис категорий и графа
- `file_service/` - сервис файлов
- `search_service/` - сервис поиска
- `web_client/` - фронтенд
- `deploy/` - файлы для развёртывания
- `docker-compose.yml` - запуск всех контейнеров локально

## Локальный запуск

### 1. Подготовка

Убедитесь, что у вас установлены:

- `Docker`
- `Docker Compose`

Для `file_service` сначала нужно подготовить файл окружения:

- создать `file_service/.env` на основе `file_service/.env.example`

### 2. Запуск всей системы

```bash
docker compose up --build
```

### 3. Основные адреса

- API gateway: `http://localhost:8080`
- Swagger UI для `api_service`: `http://localhost:8080/swagger`
- OpenAPI JSON для `api_service`: `http://localhost:8080/openapi.json`
- FastAPI docs для `category_service`: `http://localhost:8081/docs`
- Frontend: `http://localhost:5173`
- MinIO API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001`
- Neo4j Browser: `http://localhost:7474`
- Typesense: `http://localhost:8108`

## Порты сервисов

- `8080` - `api_service`
- `8081` - `category_service`
- `8082` - `note_service`
- `8083` - `user_service`
- `8085` - `file_service`
- `8086` - `search_service`
- `5173` - `web_client`
- `27017` - `MongoDB`
- `5433` - `PostgreSQL`
- `7474` и `7687` - `Neo4j`
- `9000` и `9001` - `MinIO`
- `8108` - `Typesense`

## API и тестовые запросы

- основной набор запросов к gateway: [api_service/api.http](api_service/api.http)
- дополнительный набор интеграционных запросов: [api_service/test.http](api_service/test.http)
- локальные `http`-файлы также есть у отдельных сервисов

Основные публичные маршруты идут через `api_service`:

- `/api/signup`
- `/api/auth`
- `/api/categories`
- `/api/notes`
- `/api/tags`
- `/api/notes/{uuid}/files`
- `/api/graph`
- `/api/me`
- `/api/search/notes`

## Что важно знать

- `api_service` - основная точка входа для клиента
- `category_service` уже отдаёт встроенную документацию FastAPI
- `api_service` отдаёт Swagger UI и OpenAPI-схему
- healthcheck у сервисов доступен по `/health`

## Полезные файлы

- [docker-compose.yml](docker-compose.yml)
- [api_service/config.yml](api_service/config.yml)
- [note_service/config.yml](note_service/config.yml)
- [user_service/config.yml](user_service/config.yml)
- [category_service/config.yaml](category_service/config.yaml)
- [file_service/config.yml](file_service/config.yml)
- [search_service/config.yml](search_service/config.yml)

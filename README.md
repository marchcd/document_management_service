# Document management service

Веб-сервис для автоматизации и управления документооборотом.

**Статус** Проект в продакшене, 30+ пользователей.

## Технологии
* **Frontend:** HTML/CSS, JS
* **Backend:** Golang, Postgresql
* **DevOps:** Docker

## Локальный запуск
1. Клонировать репозиторий: `git clone https://github.com/marchcd/document_management_service`
2. Создать `.env` на основе `.env.example`.
3. Нужно создать копию data/sessions.json.example. `cp data/sessions.json.example data/sessions.json`
4. Запустить через Docker: `docker-compose up --build`

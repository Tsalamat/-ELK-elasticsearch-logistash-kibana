# Todo ELK

Мини-проект с todo API, простым frontend и ELK стеком:

- `backend` - Go/Fiber API, пишет JSON-логи в `/app/logs/todo-api.log`.
- `frontend` - статическая todo-страница на Nginx.
- `elasticsearch` - хранит логи.
- `logstash` - читает JSON-лог backend из общего volume и пишет в Elasticsearch.
- `kibana` - UI для просмотра логов.

## Запуск

```bash
docker compose up --build
```

После старта:

- Todo UI: https://elk.404tears.kz
- Todo API: https://elk.404tears.kz/api
- Kibana: https://elk.404tears.kz/kibana

## API

```bash
curl http://localhost:8080/health
curl http://localhost:8080/todos
curl -X POST http://localhost:8080/todos \
  -H 'Content-Type: application/json' \
  -d '{"title":"Check logs in Kibana"}'
```

## Логи в Kibana

1. Открой Kibana: https://elk.404tears.kz/kibana
2. Перейди в **Stack Management -> Data Views**.
3. Создай data view с паттерном `todo-api-logs-*`.
4. В **Discover** выбери этот data view и смотри события API.

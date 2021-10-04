# Прокси сервер и простой сканер уязвимости на его основе

## 1. Функционал
- Http прокси
- Https прокси
- Повторная отправка проксированных запросов
- Сканер уязвимости

## 2. Сканер уязвимости
`Param-miner` – добавляет к запросу по очереди каждый GET параметр из словаря 
https://github.com/PortSwigger/param-miner/blob/master/resources/params со случайным значением (?param=shefuisehfuishe).
После это выполняется проверка указанного случайного значения в ответе, если нашлось, выводится название скрытого параметра.

## 3. Запуск проекта
```shell
sudo docker build -t proxy .                 
sudo docker run -p 8080:8080 -p 8000:8000 -t proxy
```

## 4. API
```shell
1) Получение всех запросов и ответов
  GET "http://127.0.0.1:8000/requests"

2) Получение конкретного запроса с ответом
  GET "http://127.0.0.1:8000/requests/{id:[0-9]+}"
  
3) Сканирование запроса на уязвимость
  GET "http://127.0.0.1:8000/scan/{id:[0-9]+}"
  
4) Повторная отправка запроса
  GET "http://127.0.0.1:8000/repeat/{id:[0-9]+}"
  
5) Проксирование запросов
  "http://127.0.0.1:8080/"
```

## 5. Структура БД (PostgreSql)
```sql
CREATE TABLE IF NOT EXISTS requests (
    id        serial NOT NULL PRIMARY KEY,
    method    text NOT NULL,
    scheme    text NOT NULL,
    host      text NOT NULL,
    path      text NOT NULL,
    headers   jsonb NOT NULL,
    params   jsonb NOT NULL,
    body      text NOT NULL
);

CREATE TABLE IF NOT EXISTS responses (
    id  serial NOT NULL PRIMARY KEY,
    request_id  integer NOT NULL,
    code integer NOT NULL,
    message text NOT NULL,
    headers   jsonb NOT NULL,
    body      text NOT NULL,

    FOREIGN KEY (request_id) REFERENCES requests(id)
);
```
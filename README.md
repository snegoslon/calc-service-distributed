# Распределённый вычислитель арифметических выражений

Состоит из 2 элементов:

* Сервер, который принимает арифметическое выражение, переводит его в набор последовательных задач и обеспечивает порядок их выполнения. Будем называть его оркестратором.
* Вычислитель, который может получить от оркестратора задачу, выполнить его и вернуть серверу результат. Будем называть его агентом.

## Установка и запуск

У вас должен быть установлен Golang.

1. Клонируйте репозиторий в окне терминала (консоли)
    ```bash
    git clone https://github.com/snegoslon/calc-service-distributed.git
    cd calc-service-distributed
    ```

## Запуск тестов

```bash
go test .\tests\
```

### Запуск - самый простой способ (Docker)

NOTE: Это способ также предпочтителен для запуска под Linux. Предварительно установите `docker` через менеджер пакетов дистрибутива.

```bash
docker compose up --build
```

### Запуск без Docker под Windows
 
В окне терминала (консоли) набрать и нажать Enter

```bash
run.cmd
```

Этот скрипт запустит оркестратор и агента, каждого в отдельном терминале.

### Дополнительные настройки окружения

В условии задачи указаны некоторые параметры (длительность выполнения различных операций, "вычислительная мощность"), значения которые можно изменить.

Если для запуска используется `docker`, то значения параметров меняются путём редактирования файла `docker-compose.yml`. 

При запуске под Windows без `docker` изменить параметры можно отредактировав значения в файле `run.cmd`.

#### Оркестратор

- `PORT` - порт сервера (по умолчанию 8080)
- `TIME_ADDITION_MS` - время сложения (мс)
- `TIME_SUBTRACTION_MS` - время вычитания (мс)
- `TIME_MULTIPLICATIONS_MS` - время умножения (мс)
- `TIME_DIVISIONS_MS` - время деления (мс)

#### Агент

- `ORCHESTRATOR_URL` - URL оркестратора
- `COMPUTING_POWER` - количество параллельных задач

## Примеры сценариев 

Для выполнения сценариев в Windows необходимо использовать терминал PowerShell  - в этом случае не нужно экранировать кавычки.

#### Сценарий 1 - Успех

##### Команда 1 - Отправка выражения

```bash
curl --location 'http://localhost:8080/api/v1/calculate' --data '{"expression": "(2+13)*4-10/2"}'
```

##### Команда 2 - Проверка статуса - Ответ после завершения вычислений с учётом длительности операций

```bash

curl http://localhost:8080/api/v1/expressions/1

```

```json
{
    "expression": {
        "id": "1",
        "status": "completed",
        "result": 6
    }
}
```

#### Сценарий 2 - Ошибка

##### Команда 1 - Отправка выражения

```bash
curl --location 'http://localhost:8080/api/v1/calculate' --data '{"expression": "10/(5-5)"}'
```

##### Команда 2 - Проверка статуса - Ответ после завершения вычислений с учётом длительности операций

```bash

curl http://localhost:8080/api/v1/expressions/2
```

```json
{
    "expression": {
        "id": "2",
        "status": "error",
        "result": 0
    }
}
```

#### Сценарий 3 - Вывод несуществующего выражения

```bash

curl http://localhost:8080/api/v1/expressions/9897978
```

```json
{"error":"Expression 9897978 not found"}
```

#### Сценарий 4 - Вывод всех выражений

```bash
curl http://localhost:8080/api/v1/expressions
```

```json
{
  "expressions": [
    {
      "id": "5",
      "expression": "(23333+13)*4-10/20",
      "status": "completed",
      "result": 93383.5
    },
    {
      "id": "6",
      "expression": "(23333+13)*4-10/20",
      "status": "completed",
      "result": 93383.5
    },
    {
      "id": "1",
      "expression": "10/(5-5)",
      "status": "error",
      "result": 0
    },
    {
      "id": "2",
      "expression": "10/(5-5)",
      "status": "error",
      "result": 0
    },
    {
      "id": "3",
      "expression": "(2+13)*4-10/2",
      "status": "completed",
      "result": 55
    },
    {
      "id": "4",
      "expression": "(2+13)*4-10/20",
      "status": "completed",
      "result": 59.5
    }
  ]
}
```


## API Endpoints

Эта секция содержит справочные сведения из задания. Ниже приведены сценарии использования, которые можно использовать для проверки функционирования проекта.

<details>

<summary>Посмотреть</summary>>

Далее приведён синтаксис для Linux и Windows (PowerShell)

### 1. Добавление выражения

```bash
POST /api/v1/calculate
```

Пример запроса:

```bash
curl --location 'http://localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{ "expression": "(200000000+3.3)*14-100/2" }'
```

Успешный ответ (201):

```json
{
    "id": "1"
}
```

### 2. Получение списка выражений

```bash
GET /api/v1/expressions
```

Пример ответа (200):

```json
{
    "expressions": [
        {
            "id": "1",
            "expression": "(0+0)*0-0/2",
            "status": "completed",
            "result": 0
        }
    ]
}
```

### 3. Получение выражения по id

```bash
GET /api/v1/expressions/{id}
```

Пример запроса:

```bash
curl http://localhost:8080/api/v1/expressions/1
```

Ответ (200):

```json
{
    "expression": {
        "id": "1",
        "status": "completed",
        "result": 0
    }
}
```

</details>

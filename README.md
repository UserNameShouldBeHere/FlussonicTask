# Flussonic task
[![Lang](https://img.shields.io/github/languages/top/UserNameShouldBeHere/FlussonicTask)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/UserNameShouldBeHere/FlussonicTask)](https://goreportcard.com/report/github.com/UserNameShouldBeHere/FlussonicTask)
[![License](https://img.shields.io/github/license/UserNameShouldBeHere/FlussonicTask)](https://opensource.org/license/mit)
[![Build Status](https://img.shields.io/github/actions/workflow/status/UserNameShouldBeHere/FlussonicTask/go.yml)](https://img.shields.io/github/actions/workflow/status/UserNameShouldBeHere/FlussonicTask/go.yml)

## Задание
[Ссылка](https://gist.github.com/hjbaa/51ce3a28f0b3dc78bb5fa130524d1726)

## Проделанная работа

### Использованные технологии

- Go 1.23
- Kafka
- Docker
- CI с линтером и тестами

### На что способен сервис

Есть 4 эндпоинта:

- GET /job/all
  
  Получение списка всех задач с текущим статусом
  
- GET /job/{id}/status

  Получение статуса выбранной задачи

- POST /job/{id}/cancel

  Отмена задачи (если задача не выполняется в настоящее время, то она будет отменена в будущем)

- POST /job

  Запуск задачи

Задачи поступают на обработку через кафку в следующем формате:
```json
{
  "timeout": 3000000000
  "data": "hello"
}
```

У задачи есть 5 возможных состояний:

- Done (выполнена)
- Failed (ошибка во время выполнения спустя все перезапуски)
- Canceled (отмененная задача)
- Executing (выполняемая в текущий момент)
- Waiting (ждущая очередь на перезапуск)

Сама задача представляет собой следующую структуру:

```go
type Task struct {
	Id     uint16
	Ttl    uint8
	Task   TaskFunc
	Ctx    context.Context
	Cancel context.CancelFunc
}
```

Ttl отвечает за возможное количество перезапусков

Task - сама функция для выполнения

У каждой задачи есть свой контекст, но можно задать и единый таймаут для всех задач с помощью флага -t

## Запуск

Для запуска достаточно выполнить команду `docker-compose up -d`

Также, можно запустить из консоли. Для этого нужно иметь запущенную кафку на пору 9092, затем выполнить команду `go run cmd/app/main.go`
При запуске можно указать следующие флаги:
- -p (порт, на котором запускается сервис, по умолчанию = 8080)
- -t (таймаут, применяемый ко всем задачам, по умолчанию = 3000 = 3с)
- -kh (хост кафки, для запуска в докере используется 'kafka')
- -twrk (количество горутин для обработки задач, по умолчанию = 10)
- -rwrk (количество горутин для обработки перезапускаемых задач, по умолчанию = 10)

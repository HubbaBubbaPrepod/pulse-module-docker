# Pulse Module Docker

**Русский** · [English](README.en.md)

> Нативный сборщик Docker Engine для Pulse Monitor и Pulse Agent.

[![CI](https://github.com/HubbaBubbaPrepod/pulse-module-docker/actions/workflows/ci.yml/badge.svg)](https://github.com/HubbaBubbaPrepod/pulse-module-docker/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/HubbaBubbaPrepod/pulse-module-docker)](https://github.com/HubbaBubbaPrepod/pulse-module-docker/releases)
[![MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[Установка](#установка) · [Параметры](#параметры) · [Показатели](#показатели-и-ограничения) · [Безопасность](SECURITY.ru.md) · [Поддержка](SUPPORT.ru.md)

Модуль собирает версию движка, список контейнеров, CPU, память, сетевые счётчики и последние журналы выбранного контейнера. Панель строит карточки, таблицу, графики и журналы по манифесту. Модуль выпускается отдельно как загружаемый JSON-пакет; он не является отдельной службой или Docker-образом.

```text
Pulse Monitor → назначение модуля серверу → Pulse Agent → сборщик → Docker Engine
Pulse Monitor ← карточки, таблица, графики и журналы ← JSON-результат
```

## Установка

1. Установите [Pulse Monitor](https://github.com/HubbaBubbaPrepod/pulse-monitor) и [Pulse Agent](https://github.com/HubbaBubbaPrepod/pulse-agent), подключите реальный сервер. Первый публичный набор использует панель и агент **v0.4.0**.
2. Скачайте [docker-module.json v1.0.2](https://github.com/HubbaBubbaPrepod/pulse-module-docker/releases/download/v1.0.2/docker-module.json) и [checksums.txt](https://github.com/HubbaBubbaPrepod/pulse-module-docker/releases/download/v1.0.2/checksums.txt). Сверьте SHA-256 файла с записью в `checksums.txt`:

   ```bash
   sha256sum docker-module.json
   ```

   В PowerShell:

   ```powershell
   Get-FileHash ./docker-module.json -Algorithm SHA256
   ```

3. В разделе **Модули** панели загрузите JSON. В настройках нужного сервера назначьте модуль **Docker**, заполните поля и включите его.
4. Агент загрузит сборщик для своей платформы и выполнит его. Откройте раздел модуля в панели. Для другого сервера создайте отдельное назначение.

Пакет содержит сборки **Linux amd64/arm64** и **Windows amd64**. Go и Node.js на сервере для модуля не нужны. Нужны работающий Docker Engine и доступ службы агента к нему. SHA-256 проверяет целостность файла; проверяйте также происхождение пакета.

## Параметры

| Поле | По умолчанию | Назначение |
|---|---|---|
| `socket` | Пустая строка | Unix socket или Windows named pipe; пустое поле выбирает путь платформы |
| `logContainer` | Пустая строка | Необязательный ID контейнера для журналов, минимум 12 символов |

Linux: `/var/run/docker.sock`. Windows: `\\.\pipe\docker_engine`. Для rootless Docker укажите socket своего пользователя и предоставьте службе агента доступ. На Windows проверьте доступ учётной записи службы к named pipe выбранного Docker Engine: наличие Docker Desktop само по себе его не гарантирует.

## Показатели и ограничения

| Показатель | Как читать |
|---|---|
| Контейнеры | В списке не более 100; флаг `truncated` показывает, что список ограничен, число контейнеров относится к этому списку |
| CPU | Два пакетных прохода текущей статистики с интервалом в одну секунду; 100% — одно полностью занятое ядро |
| Память | Используемая память контейнера без файлового кэша Linux; Memory total — сумма работающих контейнеров, а не RAM сервера |
| Сеть | Накопительные входящие и исходящие счётчики, а не скорость |
| Журналы | Последние 50 строк выбранного контейнера; модуль не сохраняет полный архив |
| Графики | Последние измерения в текущем просмотре панели |

Статистика собирается максимум для 100 контейнеров. Если метрики работающего контейнера недоступны, соответствующая сумма остаётся неизвестной (`—`), а не показывает неполный результат. Неизвестное значение не означает ноль.

## Права и безопасность

Сборщик делает запросы чтения Docker API, но доступ к socket или pipe обычно даёт возможности уровня root на хосте. Это привилегированный доступ к Docker daemon. Не открывайте Docker API в сеть ради модуля.

Нативный процесс выполняется с правами агента без песочницы. Поле `permissions` описывает доступ и не ограничивает ОС. Устанавливайте только доверенные пакеты; журналы контейнеров могут содержать секреты. Подробности и приватное сообщение об уязвимости: [SECURITY.ru.md](SECURITY.ru.md).

## Обновление и удаление

Пакет v1.0.2 использует `id: docker`, `schemaVersion: 1`, `protocol: 1`. Версия модуля независима от панели и агента. Перед обновлением сохраните пакет и настройки, загрузите новую версию с тем же ID и проверьте назначения и результаты. Не переписывайте выпущенный пакет или тег: конфликт содержимого уже существующей версии отклоняется.

Чтобы остановить сбор, выключите назначение на нужном сервере. Чтобы удалить модуль, уберите его в настройках панели; удалённая конфигурация больше не выдаётся агентам. Кэш бинарников может остаться в состоянии агента. Удаляйте его только при остановленном агенте и после проверки путей.

## Сборка и протокол

Нужны Go 1.23+ и PowerShell 7.4+ для упаковки. Go должен быть в PATH; для другого пути используйте параметр `-GoExe`.

```powershell
go test ./...
go vet ./...
./scripts/build.ps1 -Version 1.0.2
```

Скрипт использует toolchain Go 1.27.2. Результаты в `artifacts/releases/`: `docker-module.json`, `pulse-module-docker_v1.0.2_source.zip`, `checksums.txt`.

Процесс получает один JSON `{"protocol":1,"config":{}}` через stdin и возвращает один JSON-объект через stdout с кодом завершения 0. Диагностика идёт в stderr; агент не передаёт его в панель. Лимит stdout — **512 KiB**, скачивание и выполнение разделяют бюджет **5 секунд**. См. [протокол модулей](https://github.com/HubbaBubbaPrepod/pulse-monitor/blob/main/docs/MODULES.md).

CI запускает `go vet` и тесты на Linux и Windows, а race detector — на Linux. Это не подтверждает доступ к реальному Docker Engine на каждой платформе.

## Диагностика и участие

`unsupported` означает отсутствие сборки для платформы. При `error` проверьте Docker Engine, socket/pipe, права службы, ID контейнера и таймаут. Недоступные журналы могут быть ограничением logging driver. Не запускайте два агента с одним ID/токеном.

[Участие](CONTRIBUTING.ru.md) · [История изменений](CHANGELOG.ru.md) · [Поддержка](SUPPORT.ru.md) · [Issues](https://github.com/HubbaBubbaPrepod/pulse-module-docker/issues)

## Экосистема Pulse

| Репозиторий | Назначение |
|---|---|
| [Pulse Monitor](https://github.com/HubbaBubbaPrepod/pulse-monitor) | Панель, серверы, пакеты и представления модулей |
| [Pulse Agent](https://github.com/HubbaBubbaPrepod/pulse-agent) | Сбор метрик и выполнение назначенных модулей на сервере |
| [Pulse Module Docker](https://github.com/HubbaBubbaPrepod/pulse-module-docker) | Этот сборщик Docker Engine |

## Лицензия

[MIT](LICENSE). Вклад участников распространяется под той же лицензией.

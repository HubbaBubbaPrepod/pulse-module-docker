# Changelog

## [Unreleased]

## [1.0.1]

### Added

- Самостоятельный репозиторий, MIT-лицензия, инструкции сборки и установки.
- Отдельная сборка JSON-пакета Linux amd64/arm64 и Windows amd64, исходный архив и SHA-256 checksums.

### Fixed

- Windows Docker named-pipe connections use asynchronous handles and retry a busy pipe within the request deadline.

### Notes

- Модуль сохраняет `id: docker`, `schemaVersion: 1`, `protocol: 1` и использует независимую версию пакета.
- Совместимый первый публичный набор: Pulse Monitor v0.4.0, Pulse Agent v0.4.0, Docker-модуль v1.0.1.
- v1.0.1 предназначена для нового содержимого пакета; уже установленная v1.0.0 остаётся неизменной.

## [1.0.0]

- Исходный Docker-сборщик: версия движка, список контейнеров, ресурсные показатели, сетевые счётчики и необязательные последние журналы контейнера.
- Нативный процесс с JSON stdin/stdout для протокола модулей v1. Пакет предшествовал выделению самостоятельного репозитория.

### Fixed

- Windows Docker named-pipe connections use asynchronous handles and retry a busy pipe within the request deadline.

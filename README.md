# ShadowRocket Ruleset Generator

[RU] Этот репозиторий автоматически генерирует оптимизированные списки доменов для Shadowrocket, используя актуальные данные из репозитория https://github.com/v2fly/domain-list-community

[EN] This repository automatically generates optimized domain lists for Shadowrocket using up-to-date data from the https://github.com/v2fly/domain-list-community repository.

---

> **RU:** Обратите внимание, что обработка регулярных выражений (`regexp:`) на данный момент может работать некорректно из-за различий в синтаксисе Shadowrocket.
>
> **EN:** Please note that regular expression processing (`regexp:`) may currently work incorrectly due to syntax differences in Shadowrocket.

## Описание (Russian)

Скрипт обрабатывает исходные файлы правил и преобразует их в формат `.list`, максимально очищенный от мусора и дубликатов.

### Основные функции:

- **Фильтрация:** Скрипт анализирует иерархию доменов. Если в списке есть основной домен (например, `DOMAIN-SUFFIX,ru`), все его поддомены (например, `yandex.ru`, `vk.ru`) будут удалены как избыточные. Это значительно сокращает размер файла.
- **Автоматизация:** Благодаря GitHub Actions, список обновляется ежедневно. Вам не нужно запускать скрипт вручную.
- **Поддержка всех типов:**
  - `full:` -> `DOMAIN`
  - `keyword:` -> `DOMAIN-KEYWORD`
  - `domain:` или без префикса -> `DOMAIN-SUFFIX`
- **Рекурсивная обработка:** Скрипт корректно проходит по всем директивам `include:`.

## Description (English)

The **Go** script processes source rule files and converts them into an optimized `.list` format, removing all redundant entries.

### Key Features:

- **Smart Filtering:** The script analyzes domain hierarchy. If a base domain exists (e.g., `DOMAIN-SUFFIX,ru`), all its subdomains (e.g., `google.ru`, `mail.ru`) are removed. This keeps the configuration lean and fast.
- **Automation:** Powered by GitHub Actions, the list is updated daily, fetching the latest changes from the source.
- **Full Syntax Support:**
  - `full:` -> `DOMAIN`
  - `keyword:` -> `DOMAIN-KEYWORD`
  - `domain:` or no prefix -> `DOMAIN-SUFFIX`
- **Recursive Inclusion:** Correctly handles all `include:` directives within files.

## Техническая часть / Technical Details

Скрипт использует только стандартную библиотеку Go. / The script uses only the Go standard library.

```bash
go run main.go
```

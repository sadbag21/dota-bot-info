# Запуск через Docker

Команды выполняются из папки проекта при запущенном Docker Desktop
в режиме Linux containers.

## Сборка

```powershell
docker build -t dota-bot-info:local .
```

Сборка запускает unit-тесты и go vet, затем создаёт Linux-бинарник.
В итоговом образе находятся бинарник, корневые сертификаты HTTPS и пустой каталог `/data`.
Бот работает без root. Токен, .env, Git и результаты Allure не входят в образ.
Внешние тесты OpenDota не запускаются.

## Первый запуск

Останови ранее запущенного бота в GoLand, чтобы два процесса не получали
обновления с одним токеном одновременно. Файл .env должен содержать
TELEGRAM_BOT_TOKEN и, при необходимости, LOG_LEVEL.

```powershell
docker run --rm --name dota-bot-info --env-file .env -e DATA_DIR=/data --mount type=volume,source=dota-bot-data,target=/data dota-bot-info:local
```

Логи видны прямо в терминале. Ctrl+C останавливает бота.
Проверка в Telegram: /player с твоим ID, кнопки профиля, /match и Impact.

## Фоновый запуск

```powershell
docker run -d --name dota-bot-info --restart unless-stopped --env-file .env -e DATA_DIR=/data --mount type=volume,source=dota-bot-data,target=/data dota-bot-info:local
docker logs --tail 50 -f dota-bot-info
```

Ctrl+C при просмотре логов завершает только просмотр. Остановка контейнера:

```powershell
docker stop --timeout 15 dota-bot-info
```

Для запуска остановленного контейнера: `docker start dota-bot-info`.
Перед созданием нового контейнера с таким же именем удали остановленный:
`docker rm dota-bot-info`. Кэш бота находится в памяти и при остановке теряется.
Выбранные игроки хранятся отдельно в томе `dota-bot-data` и сохраняются.
Том создаётся автоматически при первом запуске. Его каталог доступен
пользователю 65532, под которым работает бот.

При каждом пересоздании используй тот же `--mount` и `DATA_DIR=/data`.
Удаление контейнера не удаляет именованный том. Не удаляй сам том,
если хочешь сохранить выбранных игроков. Подробнее: [STORAGE.md](STORAGE.md).

После изменения кода повтори сборку и пересоздай контейнер.
После изменения .env также пересоздай контейнер: `docker restart` не перечитывает файл.
Порты открывать не требуется: бот использует исходящие HTTPS-запросы.

Docker Desktop на твоём ПК не обеспечивает работу 24/7 при выключенном
компьютере. Для этого следующим этапом потребуется сервер.

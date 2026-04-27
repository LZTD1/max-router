# Max Router

![max-router-funny-pic](https://i.imgur.com/IVlsf8E.png)

Гибкий и мощный роутер для [Max Bot API Client@v1.6.14](https://github.com/max-messenger/max-bot-api-client-go), вдохновленный принципами `go-chi`.

Роутер предоставляет удобный интерфейс для обработки сообщений, команд и callback-запросов, поддерживая middleware и группировку маршрутов.

## Установка

```bash
go get github.com/LZTD1/max-router@v1.0.1
```

## Возможности (Features)

Max Router делает разработку бота простой, чистой и эффективной:

1. **Интеллектуальная маршрутизация:** Роутер автоматически определяет, как обработать сообщение, стараясь сделать это за меньшее время.
2. **Гибкие Middleware:** Вы можете легко создавать цепочки обработки для всей системы или отдельных групп маршрутов.
3. **MaxContext:** Вместо работы с «сырыми» данными SDK, ваш хендлер получает готовый инструмент для работы с пользователем - `Context`.
4. **Асинхронность из коробки:** Роутер поддерживает параллельную обработку сообщений.


## Простой пример использования

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
    defer stop()
    
    api, _ := maxbot.New(
        os.Getenv("MAX_TOKEN"),
    )
    
    // Инициализация роутера
    r := maxrouter.NewRouter(api)
    r.Use(middleware.Recover()) // Защита от паники
    r.Use(middleware.DefaultLogger())
    
    // --- Регистрация хендлера на команду /start ---
    r.HandleCommand("/start", func(ctx maxrouter.Context) error {
        return ctx.Send("Привет!")
    })
    
    // --- Обработка неизвестных команд (NotFound) ---
    r.NotFound(func(c maxrouter.Context) error {
        return c.Send("Извините, такую команду я еще не умею обрабатывать")
    })
    
    // --- Start polling ---
    for update := range api.GetUpdates(ctx) {
        r.Handle(update, ctx)
    }
}
```
## Примеры

Все примеры предоставлены в директории `_examples`:

- [Обработка сообщений и нажатий на кнопки](./_examples/basic-text-and-callbacks.go)
- [Работа с регулярными выражениями](./_examples/regular-exp.go)
- [Использование Middleware: Применение глобальных и Scoped Middleware](./_examples/middleware-usage.go)

## License
Licensed under [MIT License](./LICENSE)
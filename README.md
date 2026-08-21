# Max Router

![max-router-funny-pic](https://i.imgur.com/IVlsf8E.png)

Гибкий и мощный роутер сообщений для [Max Bot API Client v2.2.6](https://github.com/max-messenger/max-bot-api-client-go/releases/tag/v2.2.6), вдохновленный принципами `go-chi`.

Роутер предоставляет удобный интерфейс для обработки сообщений, команд и callback-запросов, поддерживая middleware и группировку маршрутов.

## Установка

```bash
go get github.com/LZTD1/max-router/v2@v2.1.0
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
    
    token := os.Getenv("MAX_TOKEN")
    if token == "" {
        log.Fatal("MAX_TOKEN не задан")
    }
    api, err := maxbot.NewApi(token)
    if err != nil {
        log.Fatal(err)
    }
    
    // Инициализация роутера
    r := maxrouter.NewRouter(api)
    r.Use(middleware.Recover()) // Защита от паники
    r.Use(middleware.DefaultLogger())
    
    // --- Регистрация хендлера на команду /start ---
    r.HandleCommand("/start", func(ctx maxrouter.Context) error {
        return ctx.Send(fmt.Sprintf("Привет %s!", ctx.FullName()))
    })
    r.HandleText("Кто ты?", func(ctx maxrouter.Context) error {
        return ctx.Reply("Я бот!")
    })
	
    // --- Обработка неизвестных команд (NotFound) ---
    r.NotFound(func(c maxrouter.Context) error {
        return c.Send("Извините, такую команду я еще не умею обрабатывать")
    })
    
    // --- Start polling ---
    if err := r.RunPolling(ctx, maxrouter.WithPollingErrorHandler(func(err error) {
        log.Printf("polling error: %v", err)
    })); err != nil && !errors.Is(err, context.Canceled) {
        log.Fatal(err)
    }
}
``` 
Для корректной работы внутри роутера рекомендуется использование встроенных в контекст методов `Send`, `Reply` и т.д. Использование std API внутри роутера может привести к некорректной работе.

## Примеры

Все примеры предоставлены в директории `_examples`:

- [Обработка сообщений и нажатий на кнопки](./_examples/basic-text-and-callbacks.go)
- [Работа с регулярными выражениями](./_examples/regular-exp.go)
- [Использование Middleware: Применение глобальных и Scoped Middleware](./_examples/middleware-usage.go)

## История версий
- **v2.1.0**: Добавление обертки polling событий ( `RunPolling`, `WithPollingErrorHandler` )
- **v2.0.0**: Миграция на новую мажорную версию max sdk v2.2.6
- **v1.2.0**: Добавлены методы для работы с пользователем `User()`, `Username()`, `FullName()` и т.д.
- **v1.1.1**: Добавлены примеры и миграция на новую версию maxbot, расширены тесты middleware
- **v1.1.0**: Интеграция передачи контекста в обработчики запросов
- **v1.0.1**: Оптимизация контекста (удален mutex)
- **v1.0.0**: Релиз

## License
Licensed under [MIT License](./LICENSE)

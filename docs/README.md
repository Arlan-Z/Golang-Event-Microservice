# Шаблон сервиса

## Подтягивание приватных библиотек
Go Mod под капотом использует git, для подтягивания приватных репозиториев необходимо обращаться к гиту не через http а через ssh.

ВАЖНО!

Данная команда дожна работать без пароля (для Windows в CMD)
```shell
ssh -T git@gitlab.frhc.one
```


1) Самому go указываем список хостов для приватных репозиториев, устанавливаем env
```
GOPRIVATE=gitlab.frhc.one/*
```

2) Чтобы git ходил по дефолту через ssh, в .gitconfig, который в директории home, нужно добавить следующее
```git
[url "ssh://git@gitlab.frhc.one/"]
	insteadOf = https://gitlab.frhc.one/
```

## Открытые вопросы
- Библиотека для валидации
- Запуск Liveness/Readyness проб, вместе с метриками
- Трейсинг

Шаблон сервиса состоит из следующей структуры:

- **cmd**
- - **app** - запуск сервиса
- - **migrate** - утилита для миграций
- **internal** - код проекта, закрытый для переиспользования
- - **app** - все что касается приложения и его запуска
- - - **config** - конфигурация
- - - **connections** - внешние коннекты бд, клиенты для апи и прочее
- - - **start** - хелперы для старта листнеров
- - - **store** - сущность имеющая связь со всеми репозиториями проекта
- - **data** - слой данных, структуры для общения между слоями 
- - **deliveries** - слой доставки, точки входа делятся по доменной сущности и типу
- - - **{domain entity}**
- - - - **http**
- - - - **kafka**
- - - - **grpc**
- - **pkg** - внутренние библиотеки, закрытые для переиспользования другими сервисами
- - **repositories** - слой получения внешних данных, делятся по доменной сущности и типу
- - - **{domain entity}**
- - - - pg
- - **services** - слой сервиса (контроллер), все что умеет делать приложение, делятся по доменной сущности
- - - **{domain entity}**
- - **usecases** - слой бизнес сценариев, реализует бизнес логику, полностью независим от других слоев
- - - **{domain entity}**
- **pkg** - пакеты которые, открытые для переиспользования другими сервисами



## HTTP Client для обращения в другие сервисы

`*http.BookHTTPClient` становится частью `service` для вызова в usecases и service.

```go
// service реализует интерфейс Service
type service struct {
	st *store.RepositoryStore
	hc *http.BookHTTPClient
	mw *service_middleware.ServiceMiddleware
}
```

Методы описываются в интерфейсе и реализуются в `internal/deliveries/http/client.go`

```go
type BookClient interface {
	GetBookByID(ctx context.Context, bookID string) (domain.Book, error)
}

func (c *BookHTTPClient) GetBookByID(ctx context.Context, checkoutID string) (domain.Book, error) {
	resp, err := c.r.R().SetPathParam("id", checkoutID).
		SetHeader("Session-Id", ctx.Value("session_id").(string)).
		Get(getBookById)
	if err != nil {
		return domain.Book{}, apperror.New(apperror.CommonErrInternal)
	}

	var book domain.Book

	err = json.Unmarshal(resp.Body(), &book)
	if err != nil {
		return domain.Book{}, apperror.New(apperror.CommonErrInternal)
	}

	return book, nil
}

```
Оболочка клиента для удобства использования: [resty](https://github.com/go-resty/resty)
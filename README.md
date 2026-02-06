# Posts / Comments (GraphQL, Go)

### Посты
- создать пост
- получить список постов (пагинация limit/offset)
- получить конкретный пост
- автор поста может запретить комментарии к посту

### Комментарии
- создать комментарий к посту
- вложенные комментарии без ограничения глубины (через parentID)
- ограничение длины комментария: **до 2000 символов** (при 2001 выдает ошибку)
- пагинация комментариев

### Хранилище
Два варианта хранения:
- in-memory
- PostgreSQL
Выбор через переменную окружения STORAGE_TYPE

### Стек
- Go
- GraphQL (gqlgen)
- PostgreSQL
- Docker / docker-compose

Для запуска: `docker compose up --build`

### Примеры GraphQl запросов:
Создать пост
```
mutation {
  createPost(input: {
    title: "Hello"
    content: "Пост 1"
    commentsAllowed: true
  }) {
    id
    title
    commentsAllowed
    createdAt
  }
}
```
Получить список постов
```
query {
  posts(limit: 10, offset: 0) {
    id
    title
    commentsAllowed
    createdAt
  }
}
```
Написать комментарий
```
mutation {
  createComment(input: {
    postID: "POST_ID"
    content: "комментарий"
  }) {
    id
    postID
    parentID
    content
    createdAt
  }
}
```
Ответить на комментарий
```
mutation {
  createComment(input: {
    postID: "POST_ID"
    parentID: "PARENT_COMMENT_ID"
    content: "Ответ на комментарий"
  }) {
    id
    parentID
    content
  }
}
```
Запретить комментировать пост
```
mutation {
  setCommentsAllowed(postID: "POST_ID", allowed: false) {
    id
    commentsAllowed
  }
}
```

### Unit-Тесты
Покрытие: 75.8%

### Subscriptions
Заготовка есть, но полноценная реализация не доведена до конца из-за времени

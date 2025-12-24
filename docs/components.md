# Описание
## 1. CLI-уровень (Presentation Layer)
Командная строка - единственный интерфейс пользователя

Функции: парсинг аргументов, валидация формата, вызов соответствующего Use Case

Пример: vcs commit -m "Initial commit" → вызывает CommitUseCase.execute()

## 2. Application / Use Case Layer
### Отдельные Use Cases для каждой команды:
* CommitUseCase - создание коммита
* CheckoutUseCase - переключение версий
* BranchCreateUseCase - создание ветки
* BranchDeleteUseCase - удаление ветки
* MergeUseCase - слияние веток
* CloneUseCase - клонирование репозитория
* FetchUseCase - получение изменений
* PullUseCase - получение + слияние
* PushUseCase - отправка изменений
* LogUseCase - просмотр истории

Каждый Use Case: принимает DTO от CLI, использует Domain-модель, не знает о консоли

## 3. Domain Layer (ядро системы)
### Основные сущности:
* Repository - репозиторий со списком веток, HEAD, object storage, config, индекс, рабочая директория
* Commit - коммит с hash, author, date, message, parents[], treeHash
* Tree - дерево файлов (path → blobHash | subtreeHash)
* Blob - содержимое файла (байты, идентифицируется хешем)
* Branch - ветка (name, headCommitHash)
* Index - staging area
* Remote - удалённый репозиторий (name, url, refs)
* Ядро не зависит от CLI и сети, может использоваться как библиотека

## 4. Infrastructure Layer
* FileSystemService - чтение/запись файлов, кроссплатформенность
* Сериализация - преобразование объектов Git в байты
* Сетевой транспорт - HTTP + JSON для удалённых операций

## 5. Server Layer
* VCS Server - серверная часть (API, Repository Service, Object Storage, Auth)
* Работает только с объектами и refs, не знает о рабочей директории

## 6. Внешние зависимости
* Файловая система ОС - для работы с файлами
* Сетевой стек - для сетевых операций
* Пользовательский ввод - команды от пользователя
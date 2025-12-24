```
@startuml
left to right direction
skinparam shadowing false
skinparam wrapWidth 220
skinparam dpi 150

actor "Контролёр качества\n(QC Inspector)" as QC
actor "Диспетчер\n(Dispatcher)" as Dispatcher
actor "Ремонтник\n(Repair Technician)" as Technician
actor "Бригадир\n(Foreman)" as Foreman
actor "Начальник цеха\n(Shop Head)" as ShopHead
actor "Бухгалтерия\n(Accounting)" as Accounting
actor "Администратор\n(System Admin)" as Admin
actor "Внешние системы\n(ERP/MES)" as ExternalSystems

rectangle "Система учёта дефектов сборки" as System {
  usecase "Зарегистрировать дефект" as UC1
  usecase "Выполнить диагностику" as UC2
  usecase "Назначить дефект на ремонт" as UC3
  usecase "Начать ремонт" as UC4
  usecase "Завершить ремонт" as UC5
  usecase "Вернуть автомобиль\nна конвейер" as UC6
  usecase "Просмотреть статус\nремонтных зон" as UC7
  usecase "Направить автомобиль\nв ремонтную зону" as UC8
  usecase "Назначить ремонт\nконкретному рабочему" as UC9
  usecase "Зарегистрировать начало смены" as UC10
  usecase "Сформировать отчёт контроля\nкачества за смену" as UC11
  usecase "Сформировать отчёт для\nначальника цеха" as UC12
  usecase "Сформировать отчёт по бригаде" as UC13
  usecase "Сформировать месячный отчёт\nпо отработанным сменам" as UC14
  usecase "Управлять справочниками" as UC15
  usecase "Поиск и фильтрация\nдефектов/ремонтов" as UC16
  usecase "Загрузить фото/документы" as UC17
  usecase "Просмотреть уведомления" as UC18
  usecase "Аутентификация и авторизация" as UC19
}

QC --> UC1
QC --> UC6
QC --> UC11
QC --> UC16
QC --> UC19

Dispatcher --> UC3
Dispatcher --> UC7
Dispatcher --> UC8
Dispatcher --> UC16
Dispatcher --> UC18
Dispatcher --> UC19

Technician --> UC2
Technician --> UC4
Technician --> UC5
Technician --> UC10
Technician --> UC17
Technician --> UC18
Technician --> UC19

Foreman --> UC7
Foreman --> UC9
Foreman --> UC13
Foreman --> UC16
Foreman --> UC18
Foreman --> UC19

ShopHead --> UC12
ShopHead --> UC16
ShopHead --> UC18
ShopHead --> UC19

Accounting --> UC14
Accounting --> UC16
Accounting --> UC19

Admin --> UC15
Admin --> UC16
Admin --> UC19

ExternalSystems --> UC1 : "Данные об автомобилях\nи заказах"
ExternalSystems --> UC3 : "Информация о\nпроизводственном плане"

UC3 .> UC7 : <<extend>>
UC4 .> UC9 : <<extend>>
UC5 .> UC17 : <<include>>
UC6 .> UC5 : <<include>>
UC11 .> UC1 : <<include>>
UC11 .> UC5 : <<include>>
UC12 .> UC5 : <<include>>
UC13 .> UC4 : <<include>>
UC13 .> UC5 : <<include>>
UC14 .> UC10 : <<include>>
UC15 .> UC19 : <<extend>>

@enduml
```
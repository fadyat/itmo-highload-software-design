```
@startuml
title Конечный автомат микроволновой печи

[*] --> Idle
Idle : ВЫКЛЮЧЕНА

Idle --> DoorOpen : открыть дверцу
DoorOpen : ДВЕРЦА ОТКРЫТА

DoorOpen --> Idle : закрыть дверцу

Idle --> TimeSet : нажать цифру
TimeSet : УСТАНОВКА ВРЕМЕНИ

TimeSet --> TimeSet : нажать цифру

TimeSet --> Idle : СТОП

TimeSet --> Cooking : СТАРТ
Cooking : НАГРЕВ

Cooking --> Paused : открыть дверцу
Paused : ПАУЗА

Paused --> Cooking : закрыть дверцу + СТАРТ

Paused --> TimeSet : СТОП + закрыть дверцу

Cooking --> TimeSet : СТОП

Cooking --> Finished : таймер = 0
Finished : ЗАВЕРШЕНО

Finished --> Idle : любое действие

Finished --> DoorOpen : открыть дверцу

Idle --> Error : ошибка
TimeSet --> Error : ошибка
Cooking --> Error : ошибка
Paused --> Error : ошибка
Error : ОШИБКА

Error --> [*] : отключение

@enduml
```
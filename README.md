---
statusmate login --server=<statusmate.ru | statusmate.top>
---
stautsmate notice Log in on https://statusmate.ru/
Username:
Password:

---
statusmate config use-statuspage <domain> - использовать страницу по default
statusmate config use-releasepage <domain> - использовать
statusmate config show - вывести данные
---

---
statusmate version - вернуть версию
---


```bash
statusmate add
                -i | --incident - Создать инцидент
                -m | --maintenance - создать обслуживание
                    [--t] - использовать шаблон для создания
                    [--s] - статус
                    [--p] - страница статуса или использовать по умолчания
                    [--n] - название инцидента или обслуживания
                    [--d] - описание инцидента или обслуживания
                    // подумать над способом указания компоненты которые были затронуты
                    [--с] - компоненты op cloud/lkk
                    [--dry] - проверка данных открываем редактор из $EDITOR
```

Текст редактора в ini формате

[name]
Name

[desc]
Описание

[status]
# Available statuses:
# - incident_investigating
# - incident_identified
# - incident_monitoring
# - incident_resolved
incident_investigating


[affected]
op cloud/lkk
um cdn

# Impacts:
# o, op - operational
# u, um - under maintenance
# d, dp - degraded performance
# p, po - partial outage
# m, mo - major outage


statusmate add i  -n "Проблема с Edgeconf" -d "Мы наблюдаем проблемы с Edgeconf" -
запустить редактор в котором можно будет указать какие компонеты у нас были затронуты

statusmate add i  - запустить редактор в котором можно будет указать какие компонеты у нас были затронуты и название и описание


statusmate add i -t - показываем выбор шаблона который нужно использовать для создания инцидента




statusmate ls -l - списком
              -i | -s | -m | -t | - c - показывает список сущностей

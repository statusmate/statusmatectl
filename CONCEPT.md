---
st4 login --server=<statusmate.ru | statusmate.top>
---
stautsmate notice Log in on https://statusmate.ru/
Username:
Password:

---
st4 config use-statuspage <domain> - использовать страницу по default
st4 config use-releasepage <domain> - использовать
st4 config show - вывести данные
---


---
st4 ls - показать список страниц

st4 ls
            c | components - показать список компонентов
            i | incidents - показать список инцидентов
            m | maintenance - показать список обслуживании


---

---
st4 version - вернуть версию
---


```bash
st4 add
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





```bash

st4 list-page

st4 config use-status-page <uuid>

st4 create-incident --name "Обнаружены проблемы с работой серверов" --page "status.statumate.ru"
Incident "Database Outage" created successfully!

Summary:
    uuid=12345
    name=Database Outage
    description=All DB servers are down
    components=DB, API
    status=Investigating
    created_at=2024-10-22 12:34


st4 update-incident --uuid <uuid-incident>

```

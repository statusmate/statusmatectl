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
statusmate ls - показать список страниц

statusmate ls 
            c | components - показать список компонентов
            i | incidents - показать список инцидентов
            m | maintenance - показать список обслуживании
        

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





```bash

statusmate list-page 

statusmate config use-status-page <uuid> 

statusmate create-incident --name "У анас проблемы" --page "status.statumate.ru"  
Incident "Database Outage" created successfully!

Summary:
    uuid=12345
    name=Database Outage
    description=All DB servers are down
    components=DB, API
    status=Investigating
    created_at=2024-10-22 12:34


statusmate update-incident --uuid <uuid-incident>

```

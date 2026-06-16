# honeypot
Honeypot в рамках дипломной работы 
Состав
------
  main.go        точка входа, режим --child, управление жизненным циклом
  config.go      типы конфигурации и загрузка из JSON
  fsmimic.go     создание файловых артефактов средств защиты
  procs.go       имитация процессов через подмену argv[0]
  decoys.go      сетевые ловушки TCP/UDP и обработчики поведения
  telemetry.go   запись событий в формате JSONL и отправка во внешний сборщик
  config.example.json   пример конфигурации
  deploy/ansible        роль и playbook для массового развёртывания


Запуск
------
  ./honeyagent --config ./config.example.json

Привязка к портам ниже 1024 требует прав администратора либо capability
CAP_NET_BIND_SERVICE (выдаётся в systemd-юните).

Конфигурация
------------
Все параметры задаются в JSON-файле и читаются один раз при старте.
  agent_id        идентификатор узла, попадает в каждое событие
  log_dir         каталог для локального журнала telemetry.jsonl
  collector_url   адрес внешнего сборщика (пусто = только локальная запись)
  fake_dirs       создаваемые каталоги-артефакты
  fake_procs      имитируемые процессы (name, argv0, sleep_every_sec)
  decoys          ловушки (port, proto: tcp|udp, behavior, params)
  heartbeat       период служебных событий, секунды

Поддерживаемые типы поведения ловушек (behavior):
  telnet_banner        выдаёт приглашение и фиксирует логин и пароль
  http_admin           отвечает страницей административной панели
  mongo_decoy          имитирует отклик MongoDB
  kaspersky_like_tls   имитирует управляющий интерфейс антивируса
  default              баннер сетевой службы (по умолчанию SSH)

Телеметрия
----------
Каждое событие записывается одной строкой в формате JSONL. Поля:
time, agent_id, category, type, fields. Категории: system (жизненный цикл),
edr (имитация защитного агента), decoy (обращения к ловушкам).
Файл telemetry.jsonl напрямую читается коннекторами SIEM (Wazuh, KUMA, MaxPatrol)
через стандартные декодеры JSON.

Пример события:
  {"time":"2026-04-15T10:12:33Z","agent_id":"honeypot01","category":"decoy",
   "type":"telnet_auth_attempt","fields":{"src_ip":"203.0.113.5","dst_port":2323,
   "username":"admin","password":"password123"}}

Развёртывание


  cd deploy/ansible
  ansible-playbook -i inventory.example.ini site.yml

Роль создаёт системного пользователя с минимальными правами, разворачивает
конфигурацию из шаблона Jinja2 (индивидуально для каждого узла через group_vars
и host_vars) и регистрирует усиленный systemd-юнит.



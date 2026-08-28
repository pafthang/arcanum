# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез (2026-08-28) — runtime Docker start/stop

- `RUNTIME_DOCKER_HOST` (unix:///var/run/docker.sock или tcp://)
- create → pull + start `sleep infinity`, dockerId в строке
- stop → Engine stop
- пустой host → только metadata

Ранее — runtime kernel + agents LLM + media S3.

## Не делаем сейчас

- Telegram / Discord / Slack адаптеры (`integ`)
- exec / machine-gateway
- cycles / projects / views / inbox
- Svelte UI
- второй HTTP-listen

## Следом

Один живой integ или exec в контейнер.

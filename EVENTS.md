# Eventos

Cada evento é um JSON com o campo `type`. O mesmo corpo sai no SSE (`GET /api/events`) e, quando a sessão tem `webhook_url`, num `POST` para esse endereço.

O webhook é definido em `POST /api/sessions` com `webhook_url` e `webhook_secret`. Se houver segredo, a entrega inclui o header `X-Wacalls-Signature` com o valor `sha256=` mais o HMAC-SHA256 do corpo bruto.

`session-list` e `call-list` misturam todas as sessões e só saem no SSE.

## Webhook e SSE

### `auth-state`

Mudança de conexão da sessão.

| Campo | Tipo | Notas |
|---|---|---|
| `sessionId` | string | |
| `state` | string | `connecting`, `qr`, `open`, `logged_out` |
| `paired` | boolean | |
| `qr` | string | Presente quando `state` é `qr` |

### `session-qr`

Novo código para parear a sessão.

| Campo | Tipo |
|---|---|
| `sessionId` | string |
| `qr` | string |

### `incoming`

Chamada recebida.

| Campo | Tipo | Notas |
|---|---|---|
| `sessionId` | string | |
| `id` | string | ID da chamada |
| `peer` | string | |
| `offeredAt` | number | Unix em milissegundos |

### `incoming-claimed`

Um operador aceitou a chamada recebida.

| Campo | Tipo |
|---|---|
| `sessionId` | string |
| `id` | string |
| `owner` | string |

### `call-status`

Atualização de status da chamada.

| Campo | Tipo | Notas |
|---|---|---|
| `sessionId` | string | |
| `id` | string | |
| `owner` | string ou null | |
| `status` | string | `starting`, `ringing`, `connected`, `ended` |
| `peer` | string | |
| `startedAt` | number | Unix em milissegundos |

### `call-ended`

Chamada encerrada.

| Campo | Tipo | Notas |
|---|---|---|
| `sessionId` | string | |
| `id` | string | |
| `owner` | string ou null | |
| `reason` | string | |
| `endedAt` | number | Unix em milissegundos |

## Só SSE

### `session-list`

| Campo | Tipo |
|---|---|
| `sessions` | array de sessões (`id`, `name`, `jid`, `state`, `paired`) |

### `call-list`

| Campo | Tipo |
|---|---|
| `calls` | array de chamadas ativas |

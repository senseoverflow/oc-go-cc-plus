# oc-go-cc-plus — Guida in italiano

Proxy locale che collega **Claude Code** al piano **OpenCode Go**, con preset, sync modelli e routing endpoint corretto.

Fork potenziato di [oc-go-cc](https://github.com/samueltuyizere/oc-go-cc) (AGPL-3.0).

## Requisiti

- Piano [OpenCode Go](https://opencode.ai/docs/go/) attivo
- API key da [opencode.ai/auth](https://opencode.ai/auth)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) installato
- Go 1.24+ (solo per compilare da sorgente)

## Installazione

### Homebrew (consigliato)

```bash
brew tap senseoverflow/tap
brew install oc-go-cc-plus
```

Oppure installa la formula locale:

```bash
brew install --formula ./Formula/oc-go-cc-plus.rb
```

### Da sorgente

```bash
git clone https://github.com/senseoverflow/oc-go-cc-plus.git
cd oc-go-cc-plus
make install
```

## Setup rapido

```bash
# 1. API key OpenCode Go
export OC_GO_CC_PLUS_API_KEY="sk-..."

# 2. Crea config con preset (deepseek | budget | balanced | quality)
oc-go-cc-plus init --preset deepseek

# 3. Sincronizza tutti i modelli dall'API
oc-go-cc-plus sync-models

# 4. Verifica setup
oc-go-cc-plus doctor

# 5. Avvia proxy
oc-go-cc-plus serve
```

In un altro terminale, configura Claude Code come descritto in [Configurazione Claude Code](#configurazione-claude-code) e avvia:

```bash
claude
```

## Comandi

| Comando | Descrizione |
|---|---|
| `serve` | Avvia il proxy |
| `stop` | Ferma il proxy |
| `status` | Stato del proxy |
| `init --preset NAME` | Crea config di default |
| `validate` | Validazione config estesa |
| `doctor` | Diagnostica config + API + proxy |
| `models` | Elenco modelli con endpoint |
| `models --remote` | Interroga l'API OpenCode Go |
| `sync-models` | Aggiunge modelli mancanti al config |
| `sync-models --dry-run` | Anteprima senza scrivere |
| `preset list` | Elenco preset |
| `preset apply NAME` | Applica preset (preserva api_key) |

## Preset disponibili

| Preset | Quando usarlo |
|---|---|
| `deepseek` | DeepSeek V4 Pro/Flash con max thinking — ottimo per agent coding |
| `budget` | Massimo risparmio — MiMo V2.5, Qwen3.6 Plus |
| `balanced` | Bilanciato qualità/costo — Kimi K2.6, GLM-5.1 |
| `quality` | Massima qualità — GLM-5.1, Qwen3.7 Max |

```bash
oc-go-cc-plus preset apply deepseek
```

## Routing endpoint

OpenCode Go usa **due endpoint diversi**. oc-go-cc-plus li seleziona automaticamente:

| Modelli | Endpoint |
|---|---|
| MiniMax M2.5, M2.7 | `https://opencode.ai/zen/go/v1/messages` (Anthropic) |
| Qwen3.6 Plus, Qwen3.7 Max | `https://opencode.ai/zen/go/v1/messages` (Anthropic) |
| DeepSeek, GLM, Kimi, MiMo, ecc. | `https://opencode.ai/zen/go/v1/chat/completions` (OpenAI) |

Verifica con:

```bash
oc-go-cc-plus models
oc-go-cc-plus validate
```

## Sync automatico modelli

Quando OpenCode Go aggiunge nuovi modelli:

```bash
oc-go-cc-plus sync-models
```

Il comando:
1. Interroga `GET /zen/go/v1/models`
2. Aggiunge entry nominate per ogni modello mancante
3. **Non sovrascrive** gli scenario di routing (`default`, `think`, ecc.)

## Configurazione

Percorso: `~/.config/oc-go-cc-plus/config.json`

Variabili d'ambiente (con fallback a `OC_GO_CC_*` per compatibilità):

| Variabile | Descrizione |
|---|---|
| `OC_GO_CC_PLUS_API_KEY` | API key OpenCode Go |
| `OC_GO_CC_PLUS_CONFIG` | Percorso config personalizzato |
| `OC_GO_CC_PLUS_PORT` | Porta proxy |
| `OC_GO_CC_PLUS_HOST` | Host proxy |

## Configurazione Claude Code

Dopo aver avviato il proxy (`oc-go-cc-plus serve`), Claude Code deve puntare al proxy locale invece che all'API Anthropic. Senza queste modifiche vedrai la schermata di login Anthropic o modelli tipo "Sonnet 4.5 · API Usage Billing".

### 1. Variabili d'ambiente (shell)

Aggiungi al tuo `~/.zshrc` (o `~/.bashrc`):

```bash
# Proxy oc-go-cc-plus
export ANTHROPIC_BASE_URL=http://127.0.0.1:3456
export ANTHROPIC_AUTH_TOKEN=unused

# Modello predefinito
export ANTHROPIC_MODEL=deepseek-v4-pro

# Alias tier → modelli OpenCode Go (per /model sonnet|opus|haiku)
export ANTHROPIC_DEFAULT_SONNET_MODEL=deepseek-v4-pro
export ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.1
export ANTHROPIC_DEFAULT_HAIKU_MODEL=deepseek-v4-flash
export ANTHROPIC_DEFAULT_SONNET_MODEL_NAME="DeepSeek V4 Pro"
export ANTHROPIC_DEFAULT_OPUS_MODEL_NAME="GLM-5.1"
export ANTHROPIC_DEFAULT_HAIKU_MODEL_NAME="DeepSeek V4 Flash"

# Subagenti / task paralleli
export CLAUDE_CODE_SUBAGENT_MODEL=deepseek-v4-flash

# Popola /model con i modelli del proxy (Claude Code ≥ 2.1.129)
export CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1

# Voce extra opzionale nel picker
export ANTHROPIC_CUSTOM_MODEL_OPTION=kimi-k2.6
export ANTHROPIC_CUSTOM_MODEL_OPTION_NAME="Kimi K2.6"
```

Poi ricarica la shell: `source ~/.zshrc`.

### 2. Settings di Claude Code (consigliato)

Crea o aggiorna `~/.claude/settings.json` così le variabili valgono anche fuori dalla shell (es. app avviata dal launcher):

```json
{
  "model": "deepseek-v4-pro",
  "env": {
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:3456",
    "ANTHROPIC_AUTH_TOKEN": "unused",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "deepseek-v4-pro",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-5.1",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "deepseek-v4-flash",
    "ANTHROPIC_DEFAULT_SONNET_MODEL_NAME": "DeepSeek V4 Pro",
    "ANTHROPIC_DEFAULT_OPUS_MODEL_NAME": "GLM-5.1",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL_NAME": "DeepSeek V4 Flash",
    "CLAUDE_CODE_SUBAGENT_MODEL": "deepseek-v4-flash",
    "CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY": "1"
  }
}
```

Riavvia Claude Code dopo ogni modifica a `settings.json`.

### 3. Config del proxy (oc-go-cc-plus)

In `~/.config/oc-go-cc-plus/config.json` assicurati che sia attivo:

```json
{
  "respect_requested_model": true,
  "hot_reload": true
}
```

- **`respect_requested_model`** — Claude Code può forzare un modello con `/model` o `--model`
- **`hot_reload`** — ricarica il config senza riavviare il proxy

### 4. Selezionare un modello

| Metodo | Esempio | Note |
|---|---|---|
| Picker `/model` | `/model` → scegli **From gateway** | Richiede `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1` |
| Alias tier | `/model sonnet`, `/model opus`, `/model haiku` | Mappati dalle variabili `ANTHROPIC_DEFAULT_*_MODEL` |
| ID diretto | `/model deepseek-v4-pro` | Con `respect_requested_model: true` |
| ID gateway | `/model anthropic-opencode-deepseek-v4-pro` | ID esposto da `GET /v1/models`; il proxy traduce automaticamente |
| Avvio CLI | `claude --model qwen3.7-max` | Solo per quella sessione |

Nel picker: **`s`** = solo sessione corrente, **`Enter`** = default permanente.

### 5. Discovery automatica modelli (`/v1/models`)

Il proxy espone `GET http://127.0.0.1:3456/v1/models` in formato Anthropic. Con `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1`, all'avvio Claude Code interroga questo endpoint e aggiunge al picker i modelli presenti nel tuo `config.json`, etichettati **From gateway**.

Gli ID usano il prefisso `anthropic-opencode-` (richiesto dal filtro interno di Claude Code). Esempio:

```
anthropic-opencode-deepseek-v4-pro  →  deepseek-v4-pro (upstream)
```

Verifica che l'endpoint risponda:

```bash
curl -s http://127.0.0.1:3456/v1/models | head
```

### 6. Comandi slash opzionali (locale)

Puoi aggiungere comandi personalizzati in `~/.claude/commands/og/` (non inclusi nel repo). Esempio `~/.claude/commands/og/deepseek.md`:

```markdown
---
description: OpenCode Go — DeepSeek V4 Pro
model: deepseek-v4-pro
---

$ARGUMENTS
```

Diventa `/og:deepseek` e usa quel modello solo per quella invocazione.

### Checklist rapida

1. `oc-go-cc-plus serve` in esecuzione
2. `ANTHROPIC_BASE_URL=http://127.0.0.1:3456` nella shell o in `settings.json`
3. `respect_requested_model: true` nel config del proxy
4. (Opzionale) `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1` per il picker completo
5. Riavvia Claude Code

## Validazione e diagnostica

```bash
# Validazione statica
oc-go-cc-plus validate

# Diagnostica completa (config + API + proxy)
oc-go-cc-plus doctor

# Salta test API
oc-go-cc-plus doctor --skip-api
```

## Modelli supportati (maggio 2026)

- `deepseek-v4-pro`, `deepseek-v4-flash`
- `glm-5`, `glm-5.1`
- `kimi-k2.5`, `kimi-k2.6`
- `mimo-v2.5`, `mimo-v2.5-pro`
- `minimax-m2.5`, `minimax-m2.7`
- `qwen3.6-plus`, `qwen3.7-max`

## Troubleshooting

### Proxy già in esecuzione

```bash
oc-go-cc-plus stop
oc-go-cc-plus serve
```

### API key non valida

```bash
export OC_GO_CC_PLUS_API_KEY="sk-..."
oc-go-cc-plus doctor
```

### Modello sbagliato / errori endpoint

```bash
oc-go-cc-plus validate   # mostra endpoint per ogni modello
oc-go-cc-plus models     # tabella modelli
```

### Aggiornare modelli dopo release OpenCode

```bash
oc-go-cc-plus sync-models
kill -HUP $(cat ~/.config/oc-go-cc-plus/oc-go-cc-plus.pid)  # se hot_reload attivo
```

### Claude Code mostra login Anthropic o modelli sbagliati

- Verifica `echo $ANTHROPIC_BASE_URL` → deve essere `http://127.0.0.1:3456`
- Controlla che il proxy sia attivo: `oc-go-cc-plus status`
- Se usi il launcher macOS, imposta le variabili in `~/.claude/settings.json` (non basta solo `.zshrc`)

### `/model` non mostra modelli OpenCode Go

- Abilita `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1`
- Verifica `curl http://127.0.0.1:3456/v1/models`
- Riavvia Claude Code (la cache è in `~/.claude/cache/gateway-models.json`)

## Licenza

AGPL-3.0 — vedi [LICENSE](LICENSE). Basato su [oc-go-cc](https://github.com/samueltuyizere/oc-go-cc) di Samuel Tuyizere.

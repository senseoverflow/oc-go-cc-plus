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
brew tap gabrielecipolloni/tap
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

In un altro terminale:

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:3456
export ANTHROPIC_AUTH_TOKEN=unused
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

## Forzare un modello in Claude Code

Con `respect_requested_model: true` (default nei preset):

```bash
claude --model deepseek-v4-pro
claude --model qwen3.7-max
claude --model minimax-m2.7
```

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

## Licenza

AGPL-3.0 — vedi [LICENSE](LICENSE). Basato su [oc-go-cc](https://github.com/samueltuyizere/oc-go-cc) di Samuel Tuyizere.

# oc-go-cc-plus

Enhanced fork of [oc-go-cc](https://github.com/samueltuyizere/oc-go-cc) — use **Claude Code** with your **OpenCode Go** subscription.

**🇮🇹 [Documentazione in italiano](docs/it/README.md)**

## Features

- **Auto sync models** — `sync-models` fetches from OpenCode Go API
- **Presets** — `deepseek`, `budget`, `balanced`, `quality`
- **Correct endpoint routing** — Qwen/MiniMax → Anthropic, others → OpenAI
- **Extended validation** — `validate` + `doctor`
- **Homebrew formula** — see `Formula/oc-go-cc-plus.rb`

## Quick start

```bash
export OC_GO_CC_PLUS_API_KEY="your-key"
oc-go-cc-plus init --preset deepseek
oc-go-cc-plus sync-models
oc-go-cc-plus doctor
oc-go-cc-plus serve
```

Then configure Claude Code (see below) and run `claude`.

**🇮🇹 Full Claude Code setup guide (Italian):** [docs/it/README.md#configurazione-claude-code](docs/it/README.md#configurazione-claude-code)

## Claude Code setup

Claude Code must use the local proxy instead of Anthropic's API.

### Shell environment

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:3456
export ANTHROPIC_AUTH_TOKEN=unused
export ANTHROPIC_MODEL=deepseek-v4-pro

# Map /model sonnet|opus|haiku to OpenCode Go models
export ANTHROPIC_DEFAULT_SONNET_MODEL=deepseek-v4-pro
export ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.1
export ANTHROPIC_DEFAULT_HAIKU_MODEL=deepseek-v4-flash

export CLAUDE_CODE_SUBAGENT_MODEL=deepseek-v4-flash
export CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1
```

### Claude Code settings (recommended)

`~/.claude/settings.json`:

```json
{
  "model": "deepseek-v4-pro",
  "env": {
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:3456",
    "ANTHROPIC_AUTH_TOKEN": "unused",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "deepseek-v4-pro",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-5.1",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "deepseek-v4-flash",
    "CLAUDE_CODE_SUBAGENT_MODEL": "deepseek-v4-flash",
    "CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY": "1"
  }
}
```

### Proxy config

In `~/.config/oc-go-cc-plus/config.json`:

```json
{
  "respect_requested_model": true,
  "hot_reload": true
}
```

### Selecting models

| Method | Example |
|---|---|
| `/model` picker | Gateway models labeled **From gateway** |
| Tier aliases | `/model sonnet`, `/model opus`, `/model haiku` |
| Direct ID | `/model deepseek-v4-pro` or `claude --model qwen3.7-max` |
| Gateway ID | `/model anthropic-opencode-deepseek-v4-pro` |

The proxy exposes `GET /v1/models` (Anthropic format) for gateway discovery. Model IDs use the `anthropic-opencode-` prefix so Claude Code includes them in the picker.

## Commands

```
oc-go-cc-plus serve|stop|status|init|validate|doctor|models|sync-models|preset
```

## License

AGPL-3.0 — based on oc-go-cc by Samuel Tuyizere.

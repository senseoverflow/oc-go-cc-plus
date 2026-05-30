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

export ANTHROPIC_BASE_URL=http://127.0.0.1:3456
export ANTHROPIC_AUTH_TOKEN=unused
claude
```

## Commands

```
oc-go-cc-plus serve|stop|status|init|validate|doctor|models|sync-models|preset
```

## License

AGPL-3.0 — based on oc-go-cc by Samuel Tuyizere.

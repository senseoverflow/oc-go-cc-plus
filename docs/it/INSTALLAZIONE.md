# Installazione oc-go-cc-plus

## Homebrew

```bash
# Dalla formula nel repo
brew install --formula ./Formula/oc-go-cc-plus.rb

# Dopo pubblicazione su GitHub (sostituisci USER con il tuo username)
brew tap USER/tap
brew install oc-go-cc-plus
```

> **Nota:** aggiorna gli SHA256 nella formula dopo la prima release GitHub (`make dist`).

## Compilazione manuale

```bash
make build
sudo cp bin/oc-go-cc-plus /usr/local/bin/
```

Richiede Go 1.24+.

## Aggiornamento

```bash
brew upgrade oc-go-cc-plus
# oppure
cd oc-go-cc-plus && git pull && make install
oc-go-cc-plus sync-models
```

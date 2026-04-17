# PriceNexus

CLI en Go para buscar y trackear precios de productos en tiendas online argentinas usando una arquitectura multi-agente.

## Características

- Descubrimiento dinámico de URLs con `openrouter:web_search`
- Filtrado de dominios argentinos (`.com.ar`, `.ar`)
- Pipeline de extracción con varias estrategias y fallback
- Validación determinística antes de persistir resultados
- Persistencia histórica en SQLite
- CLI basada en Cobra

## Configuración rápida

### Requisitos

- La versión de Go declarada en `go.mod` (actualmente `1.26.2`)
- Una API key de OpenRouter para la configuración por defecto
- Ollama solo si querés usar extracción local

### Instalación

```bash
cp .env.example .env
make build
```

### Variables importantes

```bash
OPENROUTER_API_KEY=tu_clave
PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash
PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free
PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:xiaomi/mimo-v2-flash
```

### Ollama opcional

Solo si querés correr el extractor localmente:

```bash
ollama serve
ollama pull gemma4:e4b
```

## Uso

```bash
./pricenexus search "Game Stick Lite"
./pricenexus history "Game Stick Lite"
./pricenexus add shop "Mi Tienda" "https://mitienda.com.ar"
```

Flags globales:

```bash
./pricenexus search "Game Stick Lite" -v
./pricenexus search "Game Stick Lite" --db-path ./prices.db
```

## Arquitectura resumida

1. **Orchestrator** coordina el flujo.
2. **Web Searcher** usa OpenRouter `web_search` para descubrir URLs argentinas.
3. **Page Loader** descarga el HTML.
4. **Data Extractor** intenta extracción estructurada, content finder y gentle cleaning.
5. **Validator** normaliza y filtra resultados válidos.
6. **Storage** persiste en SQLite.

### Defaults actuales

| Componente | Default actual |
|---|---|
| Orchestrator | `openrouter:xiaomi/mimo-v2-flash` |
| Web Searcher | `openrouter:nvidia/nemotron-3-super-120b-a12b:free` |
| Data Extractor | `openrouter:xiaomi/mimo-v2-flash` |
| Validator | validación determinística sin uso activo de LLM |

## Notas operativas

- `cmd/cli/main.go` carga `.env` automáticamente si existe.
- Si `OPENROUTER_API_KEY` está presente y `OPENAI_API_KEY` no, el CLI copia el valor para compatibilidad.
- El web searcher debe usar OpenRouter en la práctica para tener `openrouter:web_search` real.
- Ollama no es obligatorio salvo que configures explícitamente un modelo `ollama:*` para extracción.

## Documentación

- [AGENTS.md](AGENTS.md)
- [docs/SETUP.md](docs/SETUP.md)
- [docs/LLM_CONFIG.md](docs/LLM_CONFIG.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- [docs/COMMANDS.md](docs/COMMANDS.md)

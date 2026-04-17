# Guía de instalación y configuración

Esta guía refleja el estado actual del proyecto y su configuración por defecto.

## 1. Prerrequisitos

### Go

Usá la versión declarada en `go.mod`:

```bash
go version
```

### OpenRouter

La configuración por defecto necesita OpenRouter para:

- Orchestrator
- Web Searcher
- Data Extractor por defecto

Configurá al menos:

```bash
OPENROUTER_API_KEY=tu_api_key_aqui
```

### Ollama (opcional)

Solo es necesario si cambiás el extractor a un modelo local `ollama:*`.

```bash
ollama serve
ollama pull gemma4:e4b
```

## 2. Instalación

```bash
cp .env.example .env
make build
```

También podés compilar sin Makefile:

```bash
go build -o pricenexus ./cmd/cli
```

## 3. Configuración básica

Variables más importantes:

```bash
# Orchestrator
PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash

# Web Searcher
PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free

# Data Extractor
PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:xiaomi/mimo-v2-flash

# OpenRouter
OPENROUTER_API_KEY=tu_api_key_aqui
```

Variables adicionales soportadas por el código actual:

```bash
PRICE_NEXUS_WEBSEARCH_ALLOWED_DOMAINS=.com.ar,.ar
PRICE_NEXUS_WEBSEARCH_MAX_RESULTS=10
PRICE_NEXUS_DEFAULT_CURRENCY=ARS
```

## 4. Defaults actuales

| Componente | Default actual |
|---|---|
| Orchestrator | `openrouter:xiaomi/mimo-v2-flash` |
| Web Searcher | `openrouter:nvidia/nemotron-3-super-120b-a12b:free` |
| Data Extractor | `openrouter:xiaomi/mimo-v2-flash` |

Notas:

- El web searcher debe usar OpenRouter para disponer de `openrouter:web_search`.
- El extractor puede usar OpenRouter u Ollama.
- El validator actual no depende de un LLM para validar resultados.

## 5. Comportamiento del CLI

`cmd/cli/main.go` hace dos cosas importantes al arrancar:

1. Carga `.env` automáticamente si existe.
2. Si `OPENROUTER_API_KEY` está definido y `OPENAI_API_KEY` no, copia el valor a `OPENAI_API_KEY` para compatibilidad.

## 6. Verificación rápida

### Probar OpenRouter

```bash
curl -X GET https://openrouter.ai/api/v1/models \
  -H "Authorization: Bearer $OPENROUTER_API_KEY"
```

### Probar el binario

```bash
./pricenexus search "Game Stick Lite"
```

### Si usás Ollama

```bash
ollama list
```

## 7. Troubleshooting

### `OPENROUTER_API_KEY` faltante

- Verificá que esté definido en `.env` o en tu shell.
- La configuración por defecto no funciona sin esa key.

### Web search sin resultados útiles

- Confirmá que `PRICE_NEXUS_WEBSEARCHER_LLM` siga en OpenRouter.
- Evitá `openrouter:xiaomi/mimo-v2-flash` como modelo de búsqueda web en este proyecto.

### Error de Ollama

Solo relevante si configuraste el extractor con `ollama:*`:

```bash
ollama serve
ollama pull gemma4:e4b
```

## 8. Próximos pasos

- Referencia de modelos: [LLM_CONFIG.md](LLM_CONFIG.md)
- Arquitectura: [ARCHITECTURE.md](ARCHITECTURE.md)
- Comandos: [COMMANDS.md](COMMANDS.md)

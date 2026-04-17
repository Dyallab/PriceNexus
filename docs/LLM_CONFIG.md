# Configuración de LLMs

Este documento describe la configuración real que hoy usa PriceNexus.

## Defaults actuales

Los defaults salen de `internal/agent/config.go`:

| Componente | Default actual | Observaciones |
|---|---|---|
| Orchestrator | `openrouter:xiaomi/mimo-v2-flash` | Coordinación general |
| Web Searcher | `openrouter:nvidia/nemotron-3-super-120b-a12b:free` | Necesita `openrouter:web_search` |
| Data Extractor | `openrouter:xiaomi/mimo-v2-flash` | Puede cambiarse a Ollama |

## Variables de entorno

```bash
export PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash
export PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free
export PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:xiaomi/mimo-v2-flash
export OPENROUTER_API_KEY=tu_api_key_de_openrouter_aqui
```

Variables de búsqueda soportadas:

```bash
export PRICE_NEXUS_WEBSEARCH_ALLOWED_DOMAINS=.com.ar,.ar
export PRICE_NEXUS_WEBSEARCH_MAX_RESULTS=10
export PRICE_NEXUS_DEFAULT_CURRENCY=ARS
```

## OpenRouter y compatibilidad con OpenAI

`cmd/cli/main.go` carga `.env` automáticamente y además hace este puente:

- si `OPENROUTER_API_KEY` existe
- y `OPENAI_API_KEY` no existe
- entonces copia el valor de OpenRouter a `OPENAI_API_KEY`

Eso permite compatibilidad con clientes que esperan una key estilo OpenAI.

## Web Searcher

El web searcher está pensado para OpenRouter con el server tool `openrouter:web_search`.

Configuración efectiva del tool:

```json
{
  "type": "openrouter:web_search",
  "parameters": {
    "allowed_domains": [".com.ar", ".ar"],
    "max_results": 10
  }
}
```

Notas importantes:

- En el código existe una rama que permite construir un web searcher con Ollama, pero eso no aporta `openrouter:web_search` real.
- Para el comportamiento soportado del proyecto, el web searcher debe correr con OpenRouter.
- `openrouter:xiaomi/mimo-v2-flash` no es la opción recomendada para el web searcher en PriceNexus.

## Data Extractor

El extractor soporta dos modos:

### Default: OpenRouter

```bash
export PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:xiaomi/mimo-v2-flash
```

### Opcional: Ollama local

```bash
export PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:gemma4:e4b
```

Si usás Ollama:

```bash
ollama serve
ollama pull gemma4:e4b
```

## Validator

El validator actual no usa activamente un LLM para validar resultados.

- Su constructor acepta un `llms.Model`
- La lógica actual normaliza strings, verifica precios, currency y URLs
- Si no queda ningún resultado válido, devuelve error

Por eso no hay un “default de validator” configurable en `internal/agent/config.go` hoy.

## Recomendaciones de uso

### Configuración por defecto del proyecto

```bash
PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash
PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free
PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:xiaomi/mimo-v2-flash
OPENROUTER_API_KEY=tu_api_key_aqui
```

### Configuración híbrida con extractor local

```bash
PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash
PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free
PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:gemma4:e4b
OPENROUTER_API_KEY=tu_api_key_aqui
```

## Resumen rápido

- OpenRouter es obligatorio para la configuración por defecto.
- El web searcher depende de OpenRouter en la práctica.
- Ollama es opcional y hoy solo tiene sentido para el data extractor.
- El validator actual es determinístico, no LLM-driven.

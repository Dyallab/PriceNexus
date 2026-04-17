# Arquitectura de PriceNexus

## Resumen

PriceNexus es un CLI en Go que busca productos en tiendas argentinas, descarga las páginas encontradas, extrae datos de producto y persiste resultados históricos en SQLite.

La arquitectura sigue un flujo multi-componente:

1. **Orchestrator** coordina
2. **Web Searcher** descubre URLs con `openrouter:web_search`
3. **Page Loader** descarga HTML
4. **Data Extractor** intenta extraer productos con varias estrategias
5. **Validator** normaliza y filtra resultados válidos
6. **Storage** persiste en SQLite

## Componentes

### Orchestrator

- Archivo principal: `internal/agent/orchestrator/orchestrator.go`
- Crea y coordina el resto de los componentes
- Usa la configuración cargada desde `internal/agent/config.go`

### Web Searcher

- Archivo principal: `internal/agent/websearcher/agent.go`
- Usa OpenRouter `openrouter:web_search`
- Aplica filtros de dominios permitidos y resultados máximos
- Default actual: `openrouter:nvidia/nemotron-3-super-120b-a12b:free`

### Page Loader

- Ubicación: `internal/agent/pageloader/`
- Descarga contenido HTML
- No usa LLM

### Data Extractor

- Archivo principal: `internal/agent/dataextractor/agent.go`
- Default actual: `openrouter:xiaomi/mimo-v2-flash`
- Soporta extractor local con Ollama si se configura `PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:...`
- Tiene un pipeline por capas:
  - extracción de structured data
  - content finder con LLM
  - fallback de gentle cleaning

### Validator

- Archivo principal: `internal/agent/validator/agent.go`
- La implementación actual es determinística
- Normaliza strings, currency, URLs y descarta resultados inválidos
- Aunque el constructor acepta un `llms.Model`, hoy no usa activamente un LLM en la validación

### Storage

- Persistencia en SQLite
- Repositorio bajo `internal/db/`

## Defaults actuales

Definidos en `internal/agent/config.go`:

| Componente | Default |
|---|---|
| Orchestrator | `openrouter:xiaomi/mimo-v2-flash` |
| Web Searcher | `openrouter:nvidia/nemotron-3-super-120b-a12b:free` |
| Data Extractor | `openrouter:xiaomi/mimo-v2-flash` |

## Configuración de búsqueda

Defaults del search config:

- `allowed_domains`: `.com.ar`, `.ar`
- `excluded_domains`: `mercadolibre.com.ar`
- `max_results`: `10`
- `default_currency`: `ARS`

Variables relacionadas:

```bash
PRICE_NEXUS_WEBSEARCH_ALLOWED_DOMAINS=.com.ar,.ar
PRICE_NEXUS_WEBSEARCH_MAX_RESULTS=10
PRICE_NEXUS_DEFAULT_CURRENCY=ARS
```

## OpenRouter web_search

El web searcher configura este tool:

```json
{
  "type": "openrouter:web_search",
  "parameters": {
    "allowed_domains": [".com.ar", ".ar"],
    "max_results": 10
  }
}
```

Notas:

- El proyecto depende de OpenRouter para este paso.
- Aunque el factory tiene una rama para Ollama en el web searcher, eso no provee `web_search` real y no representa el flujo soportado del proyecto.

## Flujo de búsqueda

```text
Usuario
  ↓
CLI (`cmd/`)
  ↓
Orchestrator
  ↓
Web Searcher → URLs de tiendas argentinas
  ↓
Page Loader → HTML
  ↓
Data Extractor → productos candidatos
  ↓
Validator → resultados válidos
  ↓
Storage → SQLite
```

## Layout relevante

```text
cmd/cli/main.go               # entrypoint, carga .env y ejecuta Cobra
cmd/                         # comandos search, history, add
internal/agent/config.go     # defaults de LLM y search config
internal/agent/openrouter.go # cliente OpenRouter y tool wiring
internal/agent/orchestrator/ # coordinación del flujo
internal/agent/websearcher/  # búsqueda web
internal/agent/pageloader/   # carga de páginas
internal/agent/dataextractor/# extracción de datos
internal/agent/validator/    # validación determinística
internal/db/                 # persistencia SQLite
```

## Decisiones importantes del estado actual

- El descubrimiento de URLs es dinámico; no depende de una lista hardcodeada de tiendas.
- El web searcher debe usar OpenRouter en la práctica.
- Ollama es opcional, no obligatorio, y hoy aplica al extractor.
- El validator actual no es un agente LLM-driven.

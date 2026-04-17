# Arquitectura de `internal/agent`

## Componentes actuales

```text
internal/agent/
├── orchestrator/      # Coordinación del flujo de búsqueda
├── websearcher/       # Integración con OpenRouter web_search
├── pageloader/        # Descarga de HTML
├── dataextractor/     # Extracción de productos desde HTML
├── validator/         # Validación determinística de resultados
├── storage/           # Persistencia auxiliar para agentes
├── shared/            # Tipos compartidos
├── openrouter.go      # Cliente OpenRouter y tool wiring
└── config.go          # Defaults de LLM y search config
```

## Defaults actuales

Definidos en `config.go`:

- Orchestrator: `openrouter:xiaomi/mimo-v2-flash`
- Web Searcher: `openrouter:nvidia/nemotron-3-super-120b-a12b:free`
- Data Extractor: `openrouter:xiaomi/mimo-v2-flash`

## Flujo actual

```text
Orchestrator
  ↓
Web Searcher
  ↓
Page Loader
  ↓
Data Extractor
  ↓
Validator
  ↓
Storage / DB
```

## Notas por componente

### Web Searcher

- Usa OpenRouter `openrouter:web_search`
- Filtra resultados a dominios argentinos configurados
- En la práctica necesita OpenRouter para funcionar como está documentado el proyecto

### Data Extractor

- Usa un pipeline con varias estrategias
- Puede correr con OpenRouter o con Ollama según `PRICE_NEXUS_DATAEXTRACTOR_LLM`
- Default actual: OpenRouter

### Validator

- La implementación actual no usa activamente un LLM
- Normaliza campos y descarta resultados inválidos

## Configuración soportada

Variables relevantes:

```bash
PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash
PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free
PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:xiaomi/mimo-v2-flash
PRICE_NEXUS_WEBSEARCH_ALLOWED_DOMAINS=.com.ar,.ar
PRICE_NEXUS_WEBSEARCH_MAX_RESULTS=10
PRICE_NEXUS_DEFAULT_CURRENCY=ARS
OPENROUTER_API_KEY=tu_api_key
```

Si querés extracción local:

```bash
PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:gemma4:e4b
ollama serve
ollama pull gemma4:e4b
```

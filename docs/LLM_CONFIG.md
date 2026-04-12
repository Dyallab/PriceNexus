# Configuración de LLMs por Agente

PriceNexus utiliza una arquitectura multi-agente donde cada agente puede usar un LLM diferente según sus necesidades.

## Configuración por Defecto

| Agente | LLM | Modelo | Proveedor | Propósito |
|--------|-----|--------|-----------|-----------|
| **Orchestrator** | OpenRouter | xiaomi/mimo-v2-flash | OpenRouter | Coordinación y planificación |
| **Web Searcher** | Ollama | phi3:mini | Local | Búsqueda web básica |
| **Data Extractor** | Ollama | phi3:mini | Local | Extracción de datos HTML |

## Variables de Entorno

Puedes configurar los LLMs usando variables de entorno:

```bash
# Orchestrator (OpenRouter con modelo específico)
export PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash

# Web Searcher (Ollama local)
export PRICE_NEXUS_WEBSEARCHER_LLM=ollama:phi3:mini

# Data Extractor (Ollama local)
export PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:phi3:mini
```

## Configuración de OpenRouter

Para usar OpenRouter con modelos como `xiaomi/mimo-v2-flash`, necesitas configurar la API key:

```bash
export OPENROUTER_API_KEY=tu_api_key_de_openrouter_aqui
```

O en un archivo `.env`:

```
OPENROUTER_API_KEY=tu_api_key_de_openrouter_aqui
```

**Nota**: OpenRouter usa la misma API que OpenAI, así que puedes usar tu API key de OpenRouter directamente.

## Configuración de Ollama (Local)

Para usar modelos locales con Ollama:

1. Instalar Ollama: https://ollama.ai
2. Descargar el modelo phi3:mini:

```bash
ollama pull phi3:mini
```

3. Iniciar el servidor Ollama:

```bash
ollama serve
```

## Cambiar Modelos

Puedes cambiar los modelos usando variables de entorno o modificando `internal/agent/config.go`:

### Usando Variables de Entorno

```bash
# Usar Claude 3.5 Sonnet via OpenRouter para orchestrator
export PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:anthropic/claude-3.5-sonnet

# Usar Llama 3 70B local para web searcher
export PRICE_NEXUS_WEBSEARCHER_LLM=ollama:llama3:70b

# Usar GPT-4o mini para data extractor
export PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:openai/gpt-4o-mini
```

### Editando el Código

Puedes cambiar los modelos modificando `internal/agent/config.go`:

```go
func DefaultLLMConfig() LLMConfig {
    return LLMConfig{
        Orchestrator:  "openrouter:xiaomi/mimo-v2-flash",  // Cambiar aquí
        WebSearcher:   "ollama:phi3:mini",
        DataExtractor: "ollama:phi3:mini",
    }
}
```

## Recomendaciones

### Para Producción
- **Orchestrator**: xiaomi/mimo-v2-flash via OpenRouter (costo-efectivo, buen rendimiento)
- **Web Searcher**: Ollama local (phi3:mini) - sin costos de API
- **Data Extractor**: Ollama local (phi3:mini) - sin costos de API

### Alternativas de Modelos OpenRouter

| Modelo | Proveedor | Costo | Rendimiento |
|--------|-----------|-------|-------------|
| xiaomi/mimo-v2-flash | Xiaomi | Bajo | Bueno |
| anthropic/claude-3.5-sonnet | Anthropic | Medio | Excelente |
| openai/gpt-4o-mini | OpenAI | Muy bajo | Bueno |
| meta/llama-3.1-70b | Meta | Medio | Excelente |
| google/gemini-pro-1.5 | Google | Medio | Bueno |

### Recomendación con Tus Modelos

| Agente | LLM | Modelo | Costo | Razón |
|--------|-----|--------|-------|-------|
| **Orchestrator** | OpenRouter | `xiaomi/mimo-v2-flash` | ~$0.05/1M tokens | Buen precio, buen razonamiento |
| **Web Searcher** | Ollama local | `gemma4:e4b` | $0 (gratis) | Procesamiento rápido, sin límites |
| **Data Extractor** | Ollama local | `gemma4:e4b` | $0 (gratis) | Extracción determinística |

**Costo total por búsqueda**: ~$0.001 (solo el orchestrator)

### Para Desarrollo/Local
- **Orchestrator**: Ollama (gemma4:e4b o phi3:mini)
- **Web Searcher**: Ollama (gemma4:e4b)
- **Data Extractor**: Ollama (gemma4:e4b)

### Para Costos
- **Low cost**: Tu configuración actual (MiMo-V2-Flash + Gemma4 local)
- **Balanced**: OpenRouter para orchestrator, Ollama para otros
- **High performance**: OpenRouter (Claude 3.5/GPT-4) para todo

# Configuración de LLMs por Agente

PriceNexus utiliza una arquitectura multi-agente donde cada agente usa un LLM específico según sus necesidades.

## Configuración por Defecto

| Agente | LLM | Modelo | Proveedor | Herramientas |
|--------|-----|--------|-----------|--------------|
| **Orchestrator** | OpenRouter | xiaomi/mimo-v2-flash | OpenRouter | - |
| **Web Searcher** | OpenRouter | xiaomi/mimo-v2-flash | OpenRouter | `openrouter:web_search` (dominios .com.ar, .ar) |
| **Data Extractor** | Ollama | gemma4:e4b | Local | - |
| **Validator** | Ollama | gemma4:e4b | Local | - |

## Variables de Entorno

Puedes configurar los LLMs usando variables de entorno:

```bash
# Orchestrator (OpenRouter con modelo específico)
export PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash

# Web Searcher (OpenRouter - necesita OPENROUTER_API_KEY)
export PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:xiaomi/mimo-v2-flash

# Data Extractor (Ollama local)
export PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:gemma4:e4b

# API Key de OpenRouter (requerida para Orchestrator y Web Searcher)
export OPENROUTER_API_KEY=tu_api_key_de_openrouter_aqui
```

## Configuración de OpenRouter

Para usar OpenRouter necesitas configurar la API key:

```bash
export OPENROUTER_API_KEY=tu_api_key_de_openrouter_aqui
```

O en un archivo `.env`:

```
OPENROUTER_API_KEY=tu_api_key_de_openrouter_aqui
```

### Web Search Tool

El **Web Searcher** ahora usa el server tool `openrouter:web_search` para buscar productos automáticamente.

**Características:**
- ✅ Búsqueda en tiempo real
- ✅ Filtrado automático: solo dominios `.com.ar` y `.ar`
- ✅ Legal y oficial (no requiere scraping)
- ✅ El modelo decide cuándo buscar
- ✅ Costo: $4 por 1,000 resultados (~$0.04 por búsqueda típica)

**Configuración del tool:**
```go
{
  "type": "openrouter:web_search",
  "parameters": {
    "allowed_domains": [".com.ar", ".ar"],
    "max_results": 10
  }
}
```

## Configuración de Ollama (Local)

Para usar modelos locales con Ollama:

1. Instalar Ollama: https://ollama.ai
2. Descargar el modelo gemma4:e4b (recomendado):

```bash
ollama pull gemma4:e4b
```

O alternativas:
```bash
ollama pull phi3:mini
ollama pull llama3:8b
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

# Usar GPT-4o mini para data extractor
export PRICE_NEXUS_DATAEXTRACTOR_LLM=openrouter:openai/gpt-4o-mini

# Usar Llama 3 70B local
export PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:llama3:70b
```

### Editando el Código

Puedes cambiar los modelos modificando `internal/agent/config.go`:

```go
func DefaultLLMConfig() LLMConfig {
    return LLMConfig{
        Orchestrator:  "openrouter:xiaomi/mimo-v2-flash",  // Cambiar aquí
        WebSearcher:   "openrouter:xiaomi/mimo-v2-flash",  // Web Searcher requiere OpenRouter
        DataExtractor: "ollama:gemma4:e4b",
    }
}
```

## Recomendaciones

### Para Producción (Costo-Efectiva) - RECOMENDADA
- **Orchestrator**: `xiaomi/mimo-v2-flash` via OpenRouter (~$0.05/1M tokens)
- **Web Searcher**: `xiaomi/mimo-v2-flash` via OpenRouter con `web_search` tool (~$0.04 por búsqueda)
- **Data Extractor**: `gemma4:e4b` local via Ollama ($0, sin costos)
- **Validator**: `gemma4:e4b` local via Ollama ($0, sin costos)

**Costo total por búsqueda**: ~$0.05 (solo OpenRouter para búsqueda + orquestación)

### Para Máximo Rendimiento
- **Orchestrator**: `anthropic/claude-3.5-sonnet` via OpenRouter
- **Web Searcher**: `anthropic/claude-3.5-sonnet` via OpenRouter con `web_search`
- **Data Extractor**: `openai/gpt-4o-mini` via OpenRouter
- **Validator**: `openai/gpt-4o-mini` via OpenRouter

### Para Desarrollo/Local (Sin Costos)
- **Orchestrator**: `ollama:gemma4:e4b` ($0)
- **Web Searcher**: NO recomendado en local - requiere búsqueda web real. Usar OpenRouter.
- **Data Extractor**: `ollama:gemma4:e4b` ($0)
- **Validator**: `ollama:gemma4:e4b` ($0)

**Nota**: El Web Searcher requiere OpenRouter porque necesita la herramienta `openrouter:web_search` para buscar productos. Usar Ollama local no proporcionaría búsqueda web real.

## Alternativas de Modelos OpenRouter

| Modelo | Proveedor | Costo | Rendimiento | Recomendado |
|--------|-----------|-------|-------------|-------------|
| xiaomi/mimo-v2-flash | Xiaomi | Muy bajo | Bueno | ✅ Sí |
| openai/gpt-4o-mini | OpenAI | Muy bajo | Bueno | ✅ Sí |
| anthropic/claude-3.5-sonnet | Anthropic | Medio | Excelente | Para máximo rendimiento |
| meta/llama-3.1-70b | Meta | Medio | Excelente | Para máximo rendimiento |
| google/gemini-2.0-flash | Google | Bajo | Bueno | ✅ Sí |

## Tus Modelos Locales (Ollama)

| Modelo | Costo | Velocidad | Recomendación |
|--------|-------|-----------|---------------|
| gemma4:e4b | $0 | Rápido | ✅ Recomendado para extracción |
| phi3:mini | $0 | Muy rápido | Bueno para recursos limitados |
| llama3:8b | $0 | Rápido | Bueno para precisión |
| llama3:70b | $0 | Lento | Mejor rendimiento, más recursos |

## Comparación: OpenRouter vs Local

| Aspecto | OpenRouter | Ollama Local |
|--------|-----------|--------------|
| **Búsqueda web** | ✅ Sí (web_search tool) | ❌ No |
| **Costo** | Pagado | Gratis |
| **Latencia** | Medio (red) | Bajo (local) |
| **Privacidad** | Datos en OpenRouter | Local |
| **Disponibilidad** | Requiere internet | Offline posible |
| **Escalabilidad** | Ilimitada | Hardware limitado |

## Flujo Típico de Búsqueda

```
Usuario: ./pricenexus search "Game Stick Lite"
         │
         ├─→ Orchestrator (OpenRouter)
         │   └─→ Web Searcher (OpenRouter + web_search)
         │       └─→ Busca: "Game Stick Lite Argentina"
         │           └─→ Filtra: solo .com.ar, .ar
         │               └─→ Retorna URLs argentinas
         │
         ├─→ Page Loader
         │   └─→ Descarga HTML de cada URL
         │
         ├─→ Data Extractor (Ollama local)
         │   └─→ Extrae precios, stock, shipping
         │
         ├─→ Validator (Ollama local)
         │   └─→ Valida datos extraídos
         │
         └─→ Storage
             └─→ Guarda en SQLite

Costo total: ~$0.05 (solo OpenRouter)
Tiempo total: ~10-15 segundos
```

## Troubleshooting

### Error: "OPENROUTER_API_KEY not set"
```bash
export OPENROUTER_API_KEY=tu_clave_aqui
# O agrega a .env:
# OPENROUTER_API_KEY=tu_clave_aqui
```

### Error: "web_search tool not available"
- Asegúrate que `PRICE_NEXUS_WEBSEARCHER_LLM` está configurado con OpenRouter
- No puedes usar Ollama local para búsqueda web - requiere OpenRouter

### Error: "Ollama connection refused"
```bash
# Inicia Ollama en otra terminal
ollama serve

# O verifica que esté corriendo:
curl http://localhost:11434/api/tags
```

### Slow data extraction
- Usa gemma4:e4b en lugar de phi3:mini (más rápido)
- Considera gpt-4o-mini vía OpenRouter si necesitas máxima precisión

## Referencias

- [OpenRouter API Docs](https://openrouter.ai/docs)
- [OpenRouter Web Search Tool](https://openrouter.ai/docs/guides/features/server-tools/web-search)
- [Ollama](https://ollama.ai)
- [LangChain Go](https://github.com/tmc/langchaingo)
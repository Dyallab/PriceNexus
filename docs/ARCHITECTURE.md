# PriceNexus - Arquitectura Multi-Agente

## Resumen Ejecutivo

PriceNexus es un sistema multi-agente para búsqueda y tracking de precios en tiendas argentinas. Usa LangChain Go y OpenRouter's web_search server tool para encontrar dinámicamente productos en línea.

## Arquitectura Actual

### Agentes Especializados

```
┌─────────────────────────────────────────────────────────────┐
│                   ORCHESTRATOR AGENT                        │
│              (xiaomi/mimo-v2-flash via OpenRouter)          │
│  - Coordina flujos de trabajo                              │
│  - Delega tareas a agentes especializados                  │
│  - Mantiene contexto de conversación                       │
└─────────────────────────────────────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┬──────────┐
         │                 │                 │          │
         ▼                 ▼                 ▼          ▼
┌──────────────────┐  ┌──────────────┐  ┌──────────┐  ┌──────────┐
│  WEB SEARCHER    │  │ PAGE LOADER  │  │ DATA     │  │ VALIDATE │
│ (OpenRouter +    │  │  (HTTP)      │  │ EXTRACT  │  │ (Ollama) │
│  web_search)     │  │              │  │ (Ollama) │  │          │
│                  │  │ Descarga     │  │ Extrae   │  │ Valida   │
│ Busca URLs       │  │ HTML         │  │ datos    │  │ datos    │
│ en Argentina     │  │              │  │ del HTML │  │          │
└──────────────────┘  └──────────────┘  └──────────┘  └──────────┘
                                    │
                                    ▼
                              ┌──────────┐
                              │ STORAGE  │
                              │  (DB)    │
                              │          │
                              └──────────┘
```

### Configuración de LLMs

| Agente | LLM | Modelo | Propósito | Tool |
|--------|-----|--------|-----------|------|
| Orchestrator | OpenRouter | xiaomi/mimo-v2-flash | Coordinación y planificación | - |
| Web Searcher | OpenRouter | xiaomi/mimo-v2-flash | Buscar URLs en Argentina | openrouter:web_search |
| Page Loader | N/A | N/A | Descargar HTML | - |
| Data Extractor | Ollama | gemma4:e4b local | Extraer datos de HTML | - |
| Validator | Ollama | gemma4:e4b local | Validar resultados | - |
| Storage Agent | N/A | N/A | Persistencia en SQLite | - |

Ver [LLM_CONFIG.md](LLM_CONFIG.md) para detalles completos.

### Estructura de Directorios

```
internal/agent/
├── orchestrator/          # Agente orquestador
│   ├── agent.go
│   └── orchestrator.go
├── websearcher/           # Agente buscador web (usa OpenRouter web_search)
│   └── agent.go
├── pageloader/            # Carga HTML
│   └── pageloader.go
├── dataextractor/         # Agente extractor de datos
│   └── agent.go
├── validator/             # Agente validador
│   └── agent.go
├── storage/               # Agente de almacenamiento
│   ├── agent.go
│   └── agent_test.go
├── shared/                # Componentes compartidos
│   └── models.go
├── openrouter.go          # Cliente OpenRouter con soporte a tools
└── config.go              # Configuración de LLMs
```

## Búsqueda Web: OpenRouter web_search

### Cómo Funciona

1. **Web Searcher Agent** genera un prompt instructivo en español
2. **OpenRouter** recibe la solicitud con el tool `openrouter:web_search`
3. **Server Tool** ejecuta la búsqueda con filtros de dominio (.com.ar, .ar)
4. **Resultados** se retornan con URLs de tiendas argentinas
5. **Agent** parsea las URLs y las convierte a `SearchResult`

### Configuración

El Web Searcher está configurado con:

```go
{
  "type": "openrouter:web_search",
  "parameters": {
    "allowed_domains": [".com.ar", ".ar"],
    "max_results": 10
  }
}
```

**Beneficios:**
- ✅ Búsqueda automática y dinámica
- ✅ Filtrado de dominios argentinos integrado
- ✅ Legal y oficial
- ✅ Sin mantenimiento de scrapers
- ✅ Modelo decide cuándo buscar
- ✅ Costo: $4 por 1,000 resultados (~$0.04 por búsqueda típica)

### Cambios Recientes

- ❌ **Eliminado**: URLFinder (hardcodeaba tiendas)
- ❌ **Eliminado**: SearchTool (scraping manual DuckDuckGo/Bing)
- ✅ **Añadido**: Soporte a server tools en OpenRouter model
- ✅ **Integrado**: openrouter:web_search con filtro de dominios

## Implementación

### Dependencias

- `github.com/tmc/langchaingo` - Framework de agentes LangChain para Go
- `github.com/PuerkitoBio/goquery` - Parsing HTML (para Data Extractor)
- `github.com/jmoiron/sqlx` - ORM SQLite
- `github.com/sirupsen/logrus` - Logging

### Características Clave

1. **Búsqueda dinámica**: Solo dominios argentinos (.com.ar, .ar)
2. **Sin hardcodeo**: Cualquier tienda argentina es descubierta automáticamente
3. **Flexible**: Fácil cambiar modelos vía variables de entorno
4. **Escalable**: Arquitectura preparada para nuevos agentes
5. **Testeable**: Cada agente puede testearse independientemente

## Flujo de Búsqueda Completo

```
1. Usuario ejecuta: ./pricenexus search "Game Stick Lite"
                            │
                            ▼
2. Orchestrator.Search() inicia
                            │
                            ▼
3. WebSearcher busca usando OpenRouter web_search
   "Search for online stores selling 'Game Stick Lite' in Argentina"
   [Tool: openrouter:web_search con allowed_domains: [.com.ar, .ar]]
                            │
                            ▼
4. OpenRouter ejecuta búsqueda y retorna URLs argentinas
   - compagamer.com.ar/game-stick-lite
   - mexx.com.ar/productos?q=game-stick
   - etc.
                            │
                            ▼
5. PageLoader descarga HTML de cada URL
                            │
                            ▼
6. DataExtractor extrae precios usando Ollama local
                            │
                            ▼
7. Validator valida los datos extraídos
                            │
                            ▼
8. Storage persiste en SQLite
                            │
                            ▼
9. Resultados retornados al usuario
```

## Uso

### Búsqueda de Productos

```bash
./pricenexus search "Game Stick Lite"
```

Busca automáticamente en tiendas argentinas únicamente.

### Historial de Precios

```bash
./pricenexus history "Game Stick Lite"
```

### Programático

```go
import (
    "context"
    "github.com/dyallo/pricenexus/internal/agent/orchestrator"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    orch, err := orchestrator.NewOrchestrator("prices.db", logger)
    if err != nil {
        panic(err)
    }
    defer orch.Close()

    results, err := orch.Search(context.Background(), "Game Stick Lite")
    if err != nil {
        panic(err)
    }
    
    for _, r := range results {
        println(r.ProductName, r.Price, r.URL)
    }
}
```

## Beneficios de la Arquitectura

1. **Separación de responsabilidades**: Cada agente hace una cosa bien
2. **Escalabilidad**: Nuevos agentes sin afectar al core
3. **Flexibilidad**: Cambiar modelos vía environment variables
4. **Mantenibilidad**: Sin scrapers frágiles, búsqueda oficial
5. **Dinamismo**: Descubre cualquier tienda argentina automáticamente
6. **Costo-efectivo**: OpenRouter + Ollama local = máximo valor

## Consideraciones de Seguridad

- Rate limiting implícito (OpenRouter)
- Filtrado de dominios integrado
- Validación de datos antes de persistencia
- User-Agent configurable (via OpenRouter)
- Logs detallados para auditoría

## Próximos Pasos Posibles

1. **Caché de búsquedas**: Evitar búsquedas duplicadas
2. **Notificaciones**: Alertar cuando baja el precio
3. **Más tiendas**: Argentina tiene muchas tiendas menores
4. **Multihilo**: Procesar múltiples búsquedas en paralelo
5. **APIs de tiendas**: Integrar APIs oficiales cuando sea posible

## Referencias

- [OpenRouter Web Search Docs](https://openrouter.ai/docs/guides/features/server-tools/web-search)
- [LangChain Go Documentation](https://tmc.github.io/langchaingo/docs/)
- [LangChain Go GitHub](https://github.com/tmc/langchaingo)
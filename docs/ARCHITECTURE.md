# PriceNexus - Arquitectura Multi-Agente

## Resumen Ejecutivo

PriceNexus ha sido transformado de un sistema monolítico de scraping a una arquitectura multi-agente moderna usando LangChain Go. Esta arquitectura permite mayor flexibilidad, escalabilidad y mantenibilidad.

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
┌──────────────┐  ┌──────────────┐  ┌──────────┐  ┌──────────┐
│ URL FINDER   │  │ PAGE LOADER  │  │ DATA     │  │ VALIDATE │
│  (Gemma4)    │  │  (HTTP)      │  │ EXTRACT  │  │ (Gemma4) │
│              │  │              │  │ (Gemma4) │  │          │
│ Busca URLs   │  │ Descarga     │  │ Extrae   │  │ Valida   │
│ en la web    │  │ HTML         │  │ datos    │  │ datos    │
└──────────────┘  └──────────────┘  └──────────┘  └──────────┘
                                    │
                                    ▼
                              ┌──────────┐
                              │ STORAGE  │
                              │  (DB)    │
                              │          │
                              └──────────┘
```

### Configuración de LLMs

| Agente | LLM | Modelo | Propósito | Carga de Contexto |
|--------|-----|--------|-----------|-------------------|
| Orchestrator | OpenRouter | xiaomi/mimo-v2-flash | Coordinación y planificación | Media |
| URL Finder | Ollama | gemma4:e4b local | Buscar URLs en la web | Baja |
| Page Loader | N/A | N/A | Descargar HTML | N/A |
| Data Extractor | Ollama | gemma4:e4b local | Extraer datos de HTML | Baja |
| Validator | Ollama | gemma4:e4b local | Validar resultados | Baja |
| Storage Agent | N/A | N/A | Persistencia en SQLite | N/A |

Ver [LLM_CONFIG.md](LLM_CONFIG.md) para detalles completos.

### Estructura de Directorios

```
internal/agent/
├── orchestrator/          # Agente orquestador
│   ├── agent.go           # Implementación base
│   └── orchestrator.go    # Orquestación principal
├── websearcher/           # Agente buscador web
│   └── agent.go
├── dataextractor/         # Agente extractor de datos
│   └── agent.go
├── storage/               # Agente de almacenamiento
│   ├── agent.go
│   └── agent_test.go
└── shared/                # Componentes compartidos
    └── models.go          # Tipos de datos
```

## Implementación

### Dependencias

- `github.com/tmc/langchaingo` - Framework de agentes LangChain para Go
- `github.com/gocolly/colly` - Web scraping
- `github.com/jmoiron/sqlx` - ORM SQLite
- `github.com/sirupsen/logrus` - Logging

### Configuración

1. **go.mod**: Añadida dependencia de LangChain Go
2. **Modelos de datos**: Estructuras compartidas en `internal/agent/shared/models.go`
3. **Tests**: Tests básicos implementados para el agente de almacenamiento

## Beneficios de la Arquitectura Multi-Agente

1. **Separación de responsabilidades**: Cada agente hace una cosa bien
2. **Escalabilidad**: Puedes añadir más agentes sin modificar el core
3. **Flexibilidad**: Cambiar modelos sin afectar otros agentes
4. **Mantenibilidad**: Código más limpio y organizado
5. **Testing**: Puedes testear agentes individualmente
6. **Performance**: Ollama local para tareas simples, OpenRouter para complejas

## Uso

### Búsqueda de productos

```bash
./pricenexus search "Game Stick Lite"
```

### Historial de precios

```bash
./pricenexus history "Game Stick Lite"
```

### Programación

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
}
```

## Próximos Pasos Recomendados

1. **Implementar herramientas reales** para los agentes:
   - Web search tool (navegación y scraping)
   - Data extraction tool (parseo HTML)
   - Storage tool (persistencia en DB)

2. **Conectar con LLMs reales**:
   - OpenRouter para orchestrator (GPT-4, Claude)
   - Ollama local para websearcher y dataextractor

3. **Añadir tests más completos**:
   - Tests de integración para cada agente
   - Tests de mocks para herramientas
   - Tests de rendimiento

4. **Documentación y monitorización**:
   - Logs estructurados
   - Métricas de rendimiento
   - Health checks

## Consideraciones de Seguridad

- Rate limiting para evitar bloqueos
- User-Agent configurable
- Cache de resultados para evitar scraping innecesario
- Validación de datos extraídos

## Referencias

- [LangChain Go Documentation](https://tmc.github.io/langchaingo/docs/)
- [LangChain Go GitHub](https://github.com/tmc/langchaingo)
- [MRKL Agent Example](https://github.com/tmc/langchaingo/tree/main/examples/mrkl-agent-example)

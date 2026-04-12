# Arquitectura Multi-Agente de PriceNexus

## Estructura de Agentes

```
internal/agent/
├── orchestrator/          # Agente orquestador principal (MiMo-V2-Flash)
├── urlfinder/             # Buscador de URLs (Gemma4)
├── pageloader/            # Cargador de páginas (no LLM)
├── dataextractor/         # Extractor de datos (Gemma4)
├── validator/             # Validador de datos (Gemma4)
├── storage/               # Almacenamiento (no LLM)
└── shared/                # Componentes compartidos
```

## Flujo de Trabajo

```
┌─────────────────────────────────────────────────────────────┐
│                   ORCHESTRATOR AGENT                        │
│              (xiaomi/mimo-v2-flash via OpenRouter)          │
│  - Coordinación de flujo                                    │
│  - Delegación de tareas                                     │
└─────────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ URL FINDER   │  │ PAGE LOADER  │  │ DATA EXTRACT │
│  (Gemma4)    │  │  (HTTP)      │  │  (Gemma4)    │
│              │  │              │  │              │
│ Busca URLs   │  │ Descarga     │  │ Extrae       │
│ en la web    │  │ HTML         │  │ datos        │
└──────────────┘  └──────────────┘  └──────────────┘
                          │
                          ▼
                   ┌──────────┐
                   │ VALIDATE │
                   │  (Gemma4)│
                   │          │
                   │ Valida   │
                   │ datos    │
                   └──────────┘
                          │
                          ▼
                   ┌──────────┐
                   │ STORAGE  │
                   │  (DB)    │
                   │          │
                   └──────────┘
```

## Beneficios de la Nueva Arquitectura

### 1. Contexto Reducido por Tarea

| Agente | Tarea | Contexto | Carga de LLM |
|--------|-------|----------|--------------|
| URL Finder | Buscar URLs | Query del producto | 🟢 Baja |
| Page Loader | Descargar HTML | URL (no usa LLM) | ⚪ N/A |
| Data Extractor | Extraer datos | HTML de una sección | 🟢 Baja |
| Validator | Validar datos | Resultados extraídos | 🟢 Baja |
| Orchestrator | Coordinar todo | Flujo completo | 🟡 Media |

### 2. Mejor Control de Errores

Cada paso es independiente y se puede debuggear por separado.

### 3. Tareas Reutilizables

- **Page Loader**: Reusable para cualquier scraper
- **Data Extractor**: Reusable para cualquier tienda
- **Validator**: Reusable para cualquier validación

## Uso

### Búsqueda de productos

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
        fmt.Printf("%s: $%.2f %s\n", r.ProductName, r.Price, r.Currency)
    }
}
```

### Historial de precios

```go
history, err := orch.GetHistory(context.Background(), "Game Stick Lite")
```

## Tests

Ejecutar tests:

```bash
go test ./internal/agent/...
```

## Próximos Pasos

1. Implementar parsing real de JSON en Data Extractor
2. Añadir herramientas específicas para cada tienda
3. Implementar caching de resultados
4. Añadir rate limiting

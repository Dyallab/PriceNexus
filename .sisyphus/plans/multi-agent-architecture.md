# Arquitectura Multi-Agente para PriceNexus

## TL;DR

> **Transformación completa**: Transformar el sistema monolítico actual a una arquitectura de agentes independientes, cada uno especializado en una tarea y con su propio modelo de LLM.

> **Proveedores**: OpenRouter (potencia) + Ollama (local/velocidad)

---

## Arquitectura de Agentes

### Agentes Especializados

```
┌─────────────────────────────────────────────────────────────┐
│                   ORCHESTRATOR AGENT                        │
│              (GPT-4 via OpenRouter)                         │
│  - Coordina flujos de trabajo                              │
│  - Delega tareas a agentes especializados                  │
│  - Mantiene contexto de conversación                       │
└─────────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ WEB SEARCHER │  │ DATA EXTRACT │  │ STORAGE AGENT│
│   (Ollama)   │  │   (Ollama)   │  │   (Local)    │
│              │  │              │  │              │
│ - Buscar en  │  │ - Parsear    │  │ - Guardar    │
│   la web     │  │   HTML       │  │   en SQLite  │
│ - Encontrar  │  │ - Extraer    │  │ - Recuperar  │
│   URLs       │  │   datos      │  │   datos      │
└──────────────┘  └──────────────┘  └──────────────┘
```

### Flujo de Trabajo

```
1. Usuario: "buscar Game Stick Lite"
   ↓
2. Orchestrator: Analiza solicitud
   ↓
3. Orchestrator → Web Searcher: Buscar URLs
   ↓
4. Web Searcher → Orchestrator: URLs encontradas
   ↓
5. Orchestrator → Data Extractor: Extraer datos de URLs
   ↓
6. Data Extractor → Orchestrator: Datos extraídos
   ↓
7. Orchestrator → Storage Agent: Guardar datos
   ↓
8. Orchestrator: Formatear y mostrar resultados
```

---

## Estructura de Directorios

```
internal/agent/
├── orchestrator/          # Agente orquestador
│   ├── agent.go
│   ├── tools.go
│   └── prompts.go
├── websearcher/           # Agente buscador web
│   ├── agent.go
│   ├── search_tool.go
│   └── browser_tool.go
├── dataextractor/         # Agente extractor de datos
│   ├── agent.go
│   ├── parse_tool.go
│   └── validate_tool.go
├── storage/               # Agente de almacenamiento
│   ├── agent.go
│   └── db_tool.go
├── shared/                # Componentes compartidos
│   ├── models.go          # Tipos de datos
│   ├── memory.go          # Memoria de conversación
│   └── tools.go           # Herramientas comunes
└── executor.go            # Ejecutor de agentes
```

---

## Especificaciones por Agente

### 1. Orchestrator Agent

**Modelo**: OpenRouter (GPT-4 o Claude 3.5)
**Responsabilidades**:
- Recibir solicitud del usuario
- Analizar y planificar el flujo de trabajo
- Delegar tareas a agentes especializados
- Mantener contexto de conversación
- Formatear y presentar resultados

**Herramientas**:
- `search_product`: Delegar a Web Searcher
- `extract_data`: Delegar a Data Extractor
- `save_prices`: Delegar a Storage Agent
- `format_results`: Formatear salida para usuario

### 2. Web Searcher Agent

**Modelo**: Ollama (phi3:mini o similar ligero)
**Responsabilidades**:
- Buscar productos en tiendas online
- Encontrar URLs relevantes
- Verificar disponibilidad de sitios

**Herramientas**:
- `search_mercadolibre`: Buscar en MercadoLibre
- `search_garbarino`: Buscar en Garbarino
- `search_tecnoshops`: Buscar en Tecnoshops
- `browse_page`: Navegar a URL específica

### 3. Data Extractor Agent

**Modelo**: Ollama (phi3:mini o similar ligero)
**Responsabilidades**:
- Extraer datos de HTML
- Validar información extraída
- Parsear precios, stock, envío

**Herramientas**:
- `extract_price`: Extraer precio
- `extract_stock`: Verificar stock
- `extract_shipping`: Verificar envío
- `extract_product_name`: Extraer nombre

### 4. Storage Agent

**Modelo**: Local (no requiere LLM)
**Responsabilidades**:
- Persistir datos en SQLite
- Recuperar datos históricos
- Gestionar transacciones

**Herramientas**:
- `save_product`: Guardar producto
- `save_price`: Guardar precio
- `get_history`: Obtener histórico
- `get_prices`: Obtener precios actuales

---

## Implementación

### Paso 1: Actualizar go.mod

Añadir dependencias de LangChain Go y LLM providers:

```go
require (
    github.com/tmc/langchaingo v0.1.0
    github.com/sashabaranov/go-openai v1.17.0  // OpenRouter compatible
)
```

### Paso 2: Implementar Agentes Base

Crear estructuras comunes para todos los agentes:

- `Agent` interface
- `Tool` interface
- `Memory` interface
- `Executor` para ejecutar flujos

### Paso 3: Implementar Cada Agente

Cada agente implementa:
- Sistema de prompts específico
- Herramientas propias
- Memoria (si es necesario)
- Conexión al modelo LLM

### Paso 4: Integrar con CLI existente

Modificar `cmd/search.go` para usar el orchestrator:
- Mantener interfaz CLI igual
- Internamente usar agentes
- Mostrar resultados formateados

---

## Beneficios de la Arquitectura Multi-Agente

1. **Separación de responsabilidades**: Cada agente hace una cosa bien
2. **Escalabilidad**: Puedes añadir más agentes sin modificar el core
3. **Flexibilidad**: Cambiar modelos sin afectar otros agentes
4. **Mantenibilidad**: Código más limpio y organizado
5. **Testing**: Puedes testear agentes individualmente
6. **Performance**: Ollama local para tareas simples, OpenRouter para complejas

---

## Próximos Pasos

- [ ] 1. Configurar dependencias de LangChain Go
- [ ] 2. Crear estructura base de agentes
- [ ] 3. Implementar Orchestrator Agent
- [ ] 4. Implementar Web Searcher Agent
- [ ] 5. Implementar Data Extractor Agent
- [ ] 6. Implementar Storage Agent
- [ ] 7. Integrar con CLI
- [ ] 8. Tests y documentación

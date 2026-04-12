# PriceNexus

Sistema CLI en Go para búsqueda y trackeo de precios de productos en tiendas online de Argentina con arquitectura multi-agente.

## Características

- **Arquitectura Multi-Agente**: Agentes especializados para cada tarea
- **Búsqueda en tiendas pequeñas**: Foco en tiendas especializadas y nichos
- **Almacenamiento histórico**: SQLite con tracking de precios
- **CLI intuitiva**: Fácil de usar desde terminal
- **Tiendas configurables**: Agregá tus propias tiendas fácilmente

## Documentación

Para información detallada sobre instalación, configuración y uso, consulta la documentación en `/docs`:

- **[Instalación y Configuración](docs/SETUP.md)** - Guía paso a paso para configurar PriceNexus
- **[Arquitectura Multi-Agente](docs/ARCHITECTURE.md)** - Detalles técnicos de la arquitectura
- **[Configuración de LLMs](docs/LLM_CONFIG.md)** - Opciones de modelos y variables de entorno
- **[Referencia de Comandos](docs/COMMANDS.md)** - Lista completa de comandos CLI

## Uso Rápido

### Buscar precios
```bash
./pricenexus search "Game Stick Lite"
```

### Ver historial de precios
```bash
./pricenexus history "Game Stick Lite"
```

### Agregar tienda
```bash
./pricenexus add shop "Tienda Ejemplo" "https://ejemplo.com"
```

## Estructura del Proyecto

```
PriceNexus/
├── cmd/                    # Comandos CLI
├── internal/               # Lógica interna (agentes, scrapers, db)
├── docs/                   # Documentación
├── migrations/             # Migraciones SQL
└── ...
```

## Agentes

- **Orchestrator**: Coordina todo el flujo de trabajo (LLM: xiaomi/mimo-v2-flash via OpenRouter)
- **Web Searcher**: Busca productos en la web (LLM: Ollama phi3:mini local)
- **Data Extractor**: Extrae datos de páginas HTML (LLM: Ollama phi3:mini local)
- **Storage Agent**: Gestiona el almacenamiento en SQLite (no usa LLM)

## Licencia

MIT

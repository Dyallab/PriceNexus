# PriceNexus

Sistema CLI en Go para búsqueda y trackeo de precios de productos en **tiendas online argentinas** usando arquitectura multi-agente y búsqueda web dinámica.

## ✨ Características

- **Búsqueda Dinámica**: Usa OpenRouter's `web_search` tool para encontrar productos automáticamente en tiendas argentinas
- **Filtrado Automático**: Solo busca en dominios `.com.ar` y `.ar` (Argentina)
- **Arquitectura Multi-Agente**: Agentes especializados (Orchestrator, Web Searcher, Data Extractor, Validator)
- **Sin Hardcodeo**: Descubre cualquier tienda argentina automáticamente, no requiere mantenimiento de lista de tiendas
- **Almacenamiento Histórico**: SQLite con tracking de precios en el tiempo
- **CLI Intuitiva**: Fácil de usar desde terminal
- **Costo-Efectiva**: OpenRouter para búsqueda + Ollama local para extracción = máximo valor

## 🚀 Inicio Rápido

### Instalación

1. **Requisitos previos:**
   - Go 1.21+
   - [Ollama](https://ollama.ai) (para extracción local)
   - [OpenRouter API Key](https://openrouter.ai) (para búsqueda web)

2. **Descargar e instalar:**
```bash
git clone https://github.com/dyallo/pricenexus.git
cd PriceNexus
make build
```

3. **Configurar variables de entorno:**
```bash
cp .env.example .env
```

4. **Iniciar Ollama (en otra terminal):**
```bash
ollama serve
ollama pull gemma4:e4b
```

### Uso

```bash
# Buscar un producto
./pricenexus search "Game Stick Lite"

# Ver historial de precios
./pricenexus history "Game Stick Lite"

# Agregar una tienda manualmente
./pricenexus add shop "Mi Tienda" "https://mitienda.com.ar"
```

## 📊 Arquitectura

### Flujo de Búsqueda

```
1. Usuario busca: "Game Stick Lite"
                    ↓
2. Orchestrator coordina
                    ↓
3. Web Searcher usa OpenRouter web_search
   → "Search for 'Game Stick Lite' in Argentina"
   → Filtra: solo .com.ar, .ar
   → Retorna URLs argentinas
                    ↓
4. Page Loader descarga HTML
                    ↓
5. Data Extractor (Ollama local) extrae precios
                    ↓
6. Validator (Ollama local) valida datos
                    ↓
7. Storage guarda en SQLite
                    ↓
8. Resultados mostrados al usuario
```

### Agentes

| Agente | LLM | Herramientas | Función |
|--------|-----|-------------|---------|
| **Orchestrator** | OpenRouter | - | Coordina flujos de trabajo |
| **Web Searcher** | OpenRouter | `openrouter:web_search` | Busca productos en Argentina (.com.ar, .ar) |
| **Page Loader** | N/A | HTTP | Descarga contenido HTML |
| **Data Extractor** | Ollama local | - | Extrae precios y datos |
| **Validator** | Ollama local | - | Valida información extraída |
| **Storage** | N/A | SQLite | Persiste en base de datos |

### Búsqueda con OpenRouter Web Search

El Web Searcher usa el **server tool** `openrouter:web_search` de OpenRouter:

- ✅ **Dinámico**: Encuentra tiendas automáticamente
- ✅ **Inteligente**: El modelo decide qué buscar
- ✅ **Preciso**: Filtra solo dominios argentinos
- ✅ **Legal**: Sin scraping manual, uso oficial de OpenRouter
- ✅ **Económico**: ~$0.04 por búsqueda típica

```go
{
  "type": "openrouter:web_search",
  "parameters": {
    "allowed_domains": [".com.ar", ".ar"],
    "max_results": 10
  }
}
```

## ⚙️ Configuración

### Variables de Entorno Obligatorias

```bash
export OPENROUTER_API_KEY=tu_clave_aqui
```

### Variables de Entorno Opcionales

```bash
export PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free
export PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:nvidia/nemotron-3-super-120b-a12b:free
export PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:gemma4:e4b
```

Ver [LLM_CONFIG.md](docs/LLM_CONFIG.md) para opciones completas.

### Ollama Local

Descargar modelos para extracción local:

```bash
ollama pull gemma4:e4b
```

## 📚 Documentación Completa

- **[SETUP.md](docs/SETUP.md)** - Instalación y configuración detallada
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Detalles técnicos de la arquitectura
- **[LLM_CONFIG.md](docs/LLM_CONFIG.md)** - Opciones de modelos y configuración
- **[COMMANDS.md](docs/COMMANDS.md)** - Referencia completa de comandos
- **[AGENTS.md](AGENTS.md)** - Guía de agents para el proyecto

## 💰 Costos

| Componente | Costo | Notas |
|-----------|-------|-------|
| OpenRouter Orchestrator | ~$0.01/búsqueda | xiaomi/mimo-v2-flash |
| OpenRouter Web Search | ~$0.04/búsqueda | $4 por 1,000 resultados |
| Ollama local | $0 | Sin costo de API |
| **Total por búsqueda** | **~$0.05** | Muy económico |

## 🔧 Características Técnicas

- **LangChain Go**: Framework de agentes para orquestación
- **OpenRouter API**: Acceso a múltiples LLMs (incluye web_search)
- **Ollama**: Modelos locales sin costo de API
- **SQLite**: Base de datos embebida sin dependencias
- **Logging**: Logrus con rotación de logs
- **CLI**: Cobra para interfaz de línea de comandos

## 📝 Ejemplos de Uso

### Buscar un producto

```bash
$ ./pricenexus search "PlayStation 5"

Buscando: PlayStation 5

✓ Encontradas 8 URLs de tiendas argentinas
✓ Extrayendo datos de 8 páginas...
  - Game Store: $599,990 ARS (Con stock, Con envío)
  - ElectrónicaMax: $589,999 ARS (Con stock, Con envío)
  ...

✓ Búsqueda completada
```

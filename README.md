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
# Edita .env y agrega tu OPENROUTER_API_KEY
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
# API Key de OpenRouter (requerida)
export OPENROUTER_API_KEY=tu_clave_aqui
```

### Variables de Entorno Opcionales

```bash
# Modelos a usar (tienen defaults sensatos)
export PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash
export PRICE_NEXUS_WEBSEARCHER_LLM=openrouter:xiaomi/mimo-v2-flash
export PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:gemma4:e4b
```

Ver [LLM_CONFIG.md](docs/LLM_CONFIG.md) para opciones completas.

### Ollama Local

Descargar modelos para extracción local:

```bash
ollama pull gemma4:e4b    # Recomendado
# O alternativas:
ollama pull phi3:mini
ollama pull llama3:8b
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

## 🛠️ Desarrollo

### Compilar

```bash
make build
```

### Ejecutar tests

```bash
make test
```

### Limpiar

```bash
make clean
```

### Estructura del Proyecto

```
PriceNexus/
├── cmd/
│   └── cli/                    # CLI principal (main.go)
├── internal/
│   ├── agent/
│   │   ├── orchestrator/       # Agente orquestador
│   │   ├── websearcher/        # Búsqueda web (OpenRouter web_search)
│   │   ├── pageloader/         # Carga HTML
│   │   ├── dataextractor/      # Extracción de datos
│   │   ├── validator/          # Validación
│   │   ├── storage/            # Persistencia SQLite
│   │   ├── shared/             # Tipos compartidos
│   │   ├── openrouter.go       # Cliente OpenRouter con tools
│   │   └── config.go           # Configuración de LLMs
│   ├── scraper/                # Scrapers específicos de tiendas
│   └── db/                     # Repositorio de base de datos
├── docs/                       # Documentación completa
├── migrations/                 # Migraciones SQL
├── Makefile                    # Comandos útiles
└── prices.db                   # Base de datos SQLite (se crea al ejecutar)
```

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

### Ver historial de precios

```bash
$ ./pricenexus history "PlayStation 5"

Historial de precios para: PlayStation 5

2024-12-15: $599,990 (Game Store)
2024-12-14: $589,999 (ElectrónicaMax)
2024-12-13: $599,990 (Game Store)
```

## 🌍 Dominios Soportados

PriceNexus busca **únicamente en tiendas argentinas**:

- ✅ `.com.ar` - Dominios comerciales argentinos
- ✅ `.ar` - Dominios de Argentina
- ❌ `.com` - No se incluyen sitios internacionales

Esto asegura que los precios estén en **pesos argentinos (ARS)** y sean relevantes localmente.

## 🚨 Solución de Problemas

### Error: "OPENROUTER_API_KEY not set"
```bash
export OPENROUTER_API_KEY=tu_clave_aqui
```

### Error: "Ollama connection refused"
```bash
# Inicia Ollama en otra terminal
ollama serve
```

### Resultados vacíos en búsqueda
- Verifica que tienes conexión a internet
- Intenta con un producto más común
- Revisa los logs en `logs/search.log`

## 📄 Licencia

MIT - Ver [LICENSE](LICENSE) para detalles

## 🤝 Contribuciones

Las contribuciones son bienvenidas. Por favor:

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/mi-feature`)
3. Commit tus cambios (`git commit -am 'Add mi-feature'`)
4. Push a la rama (`git push origin feature/mi-feature`)
5. Abre un Pull Request

## 📞 Soporte

- 📖 [Documentación completa](docs/)
- 🐛 [Reportar problemas](https://github.com/dyallo/pricenexus/issues)
- 💬 [Discussiones](https://github.com/dyallo/pricenexus/discussions)

---

**Hecho con ❤️ para encontrar mejores precios en Argentina**
# Guía de Configuración e Instalación

Esta guía te ayudará a instalar y configurar PriceNexus en tu sistema.

## 1. Prerrequisitos

### Go
Asegúrate de tener Go instalado (versión 1.18 o superior):
```bash
go version
```

### Ollama (Opcional, para modelos locales)
Si deseas usar modelos locales:
1. Instala Ollama: https://ollama.ai
2. Inicia el servidor:
   ```bash
   ollama serve
   ```
3. Descarga el modelo recomendado:
   ```bash
   ollama pull gemma4:e4b
   ```

## 2. Instalación

### Compilar el proyecto
Clona o navega al directorio del proyecto y compila:
```bash
go build ./cmd/cli
```

Esto generará el binario `pricenexus` en el directorio raíz.

### Configurar variables de entorno
Copia el archivo de ejemplo y edítalo con tus credenciales:
```bash
cp .env.example .env
# Edita .env con tus credenciales
```

## 3. Configuración de LLMs

PriceNexus permite configurar diferentes LLMs para cada agente mediante variables de entorno.

### Variables de Entorno Básicas
Agrega estas líneas a tu archivo `.env`:

```bash
# Orchestrator (Recomendado: OpenRouter para razonamiento complejo)
PRICE_NEXUS_ORCHESTRATOR_LLM=openrouter:xiaomi/mimo-v2-flash

# Web Searcher (Recomendado: Ollama local para bajo costo)
PRICE_NEXUS_WEBSEARCHER_LLM=ollama:gemma4:e4b

# Data Extractor (Recomendado: Ollama local)
PRICE_NEXUS_DATAEXTRACTOR_LLM=ollama:gemma4:e4b

# API Key de OpenRouter (obtenla en https://openrouter.ai/keys)
OPENROUTER_API_KEY=tu_api_key_aqui
```

### Opciones de Modelos

| Agente | Proveedor | Modelo Recomendado | Costo |
|--------|-----------|--------------------|-------|
| **Orchestrator** | OpenRouter | `xiaomi/mimo-v2-flash` | Bajo |
| **Web Searcher** | Ollama Local | `gemma4:e4b` | Gratis |
| **Data Extractor** | Ollama Local | `gemma4:e4b` | Gratis |

*Ver [LLM_CONFIG.md](LLM_CONFIG.md) para más opciones y detalles avanzados.*

## 4. Verificación

### Verificar Ollama
```bash
ollama list
# Deberías ver gemma4:e4b o phi3:mini en la lista
```

### Verificar API de OpenRouter
```bash
curl -X GET https://openrouter.ai/api/v1/models \
  -H "Authorization: Bearer $OPENROUTER_API_KEY"
```

### Probar PriceNexus
```bash
./pricenexus search "Game Stick Lite"
```

## 5. Solución de Problemas Comunes

### "Ollama no está corriendo"
```bash
ollama serve
# O en otro terminal: ollama run gemma4:e4b
```

### "Modelo no encontrado"
```bash
ollama pull gemma4:e4b
```

### "API Key inválida"
1. Verifica que la API key esté correcta en `.env`
2. Regenera la API key en https://openrouter.ai/keys
3. Reinicia la aplicación

## 6. Costos Estimados

| Tarea | Costo |
|-------|-------|
| Búsqueda de producto | ~$0.001 (solo orchestrator) |
| Historial de precios | $0 (local) |
| **Costo mensual estimado** | $0.50 - $5.00 |

## Próximos Pasos

Una vez configurado, puedes:
- Buscar precios: `./pricenexus search "producto"`
- Ver historial: `./pricenexus history "producto"`
- Agregar tiendas: `./pricenexus add shop "Nombre" "URL"`

Para detalles de arquitectura, consulta [ARCHITECTURE.md](ARCHITECTURE.md).

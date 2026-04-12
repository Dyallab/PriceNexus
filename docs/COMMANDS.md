# Comandos de PriceNexus

Esta página referencia todos los comandos disponibles en PriceNexus.

## Comandos CLI

### Búsqueda de Precios
Busca precios de un producto en tiendas configuradas.

```bash
./pricenexus search "Game Stick Lite"
```

### Historial de Precios
Muestra el historial de precios registrado para un producto.

```bash
./pricenexus history "Game Stick Lite"
```

### Agregar Tienda
Agrega una nueva tienda a la base de datos para ser rastreada.

```bash
./pricenexus add shop "Tienda Ejemplo" "https://ejemplo.com"
```

## Comandos de Desarrollo

### Compilar el Proyecto
Compila el binario de la CLI.

```bash
go build ./cmd/cli
```

### Ejecutar Tests
Ejecuta el suite de tests del proyecto.

```bash
go test ./...
```

### Ver Logs (Modo Verboso)
Ejecuta un comando con logs detallados para depuración.

```bash
./pricenexus search "test" -v
```

### Usar Base de Datos Específica
Especifica una ruta diferente para la base de datos.

```bash
./pricenexus search "test" --db-path /ruta/a/tu/db.db
```

## Comandos Ollama (Modelos Locales)

### Listar Modelos Disponibles
Verifica qué modelos tienes descargados localmente.

```bash
ollama list
```

### Descargar un Modelo
Descarga un nuevo modelo para usarlo localmente.

```bash
ollama pull gemma4:e4b
# o
ollama pull phi3:mini
```

### Iniciar Servidor Ollama
Inicia el servidor Ollama si no está corriendo.

```bash
ollama serve
```

### Probar un Modelo
Ejecuta una interacción directa con un modelo.

```bash
ollama run gemma4:e4b "Hola, ¿cómo estás?"
```

## Referencia Rápida

| Comando | Descripción |
|---------|-------------|
| `./pricenexus search "producto"` | Busca precios de un producto |
| `./pricenexus history "producto"` | Muestra historial de precios |
| `./pricenexus add shop "nombre" "url"` | Agrega una nueva tienda |
| `go build ./cmd/cli` | Compila el proyecto |
| `go test ./...` | Ejecuta tests |
| `ollama list` | Ver modelos locales |
| `ollama pull <modelo>` | Descargar modelo |

Para configuración de entorno y LLMs, consulta [LLM_CONFIG.md](LLM_CONFIG.md).
Para instalación detallada, consulta [SETUP.md](SETUP.md).

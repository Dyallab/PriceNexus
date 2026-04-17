# Comandos de PriceNexus

## CLI

### Buscar precios

```bash
./pricenexus search "Game Stick Lite"
```

### Ver historial

```bash
./pricenexus history "Game Stick Lite"
```

### Agregar una tienda manualmente

```bash
./pricenexus add shop "Tienda Ejemplo" "https://ejemplo.com.ar"
```

## Flags globales

Estos flags viven en el root command y aplican a todos los subcomandos:

```bash
./pricenexus search "Game Stick Lite" -v
./pricenexus search "Game Stick Lite" --db-path ./prices.db
```

- `-v, --verbose`: activa salida verbal adicional del root command
- `--db-path`: cambia la ruta de la base SQLite (default `prices.db`)

## Desarrollo

### Build

```bash
make build
```

Equivalente directo:

```bash
go build -o pricenexus ./cmd/cli
```

### Run

```bash
make run
```

### Test

```bash
make test
```

Equivalente directo:

```bash
go test -v ./...
```

### Instalar binario con Go

```bash
make install
```

### Limpiar binario

```bash
make clean
```

## Ollama opcional

Solo relevante si configuraste el extractor para usar `ollama:*`.

```bash
ollama serve
ollama pull gemma4:e4b
ollama list
```

## Referencia rápida

| Comando | Descripción |
|---|---|
| `./pricenexus search "producto"` | Busca precios de un producto |
| `./pricenexus history "producto"` | Muestra historial de precios |
| `./pricenexus add shop "nombre" "url"` | Agrega una tienda a la base |
| `make build` | Compila el binario |
| `make run` | Ejecuta el CLI con `go run` |
| `make test` | Ejecuta tests |
| `make install` | Instala el binario con Go |
| `make clean` | Elimina el binario generado |

Para setup y variables de entorno, ver [SETUP.md](SETUP.md) y [LLM_CONFIG.md](LLM_CONFIG.md).

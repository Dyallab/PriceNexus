# PriceNexus TODO

## Recap

PriceNexus es un CLI en Go para buscar y seguir precios en tiendas online argentinas.

Flujo actual:

`CLI -> Orchestrator -> Web Searcher -> Page Loader -> Data Extractor -> Validator -> Storage(SQLite)`

Defaults actuales:

- Orchestrator: `openrouter:xiaomi/mimo-v2-flash`
- Web Searcher: `openrouter:nvidia/nemotron-3-super-120b-a12b:free`
- Data Extractor: `openrouter:xiaomi/mimo-v2-flash`
- Validator: determinístico, sin LLM activo

## Ya está bastante cubierto

- Descubrimiento dinámico de URLs con `openrouter:web_search`
- Filtro de dominios argentinos
- Extracción con múltiples estrategias y fallback
- Validación previa a persistir resultados
- Persistencia histórica en SQLite
- CLI con Cobra y docs principales alineadas
- Cache persistente en SQLite para búsquedas, páginas y extracción

## Falta / por revisar

### Validación base

- [ ] Probar el flujo real end-to-end con un producto de ejemplo (`search`, `history`, `add`)
- [ ] Correr la suite completa de tests y anotar fallas reales vs. fallas preexistentes
- [ ] Ver si hace falta un smoke test mínimo para el comando `search`
- [ ] Confirmar que docs y `.env.example` sigan sincronizados con el código

### Mejoras útiles manteniendo SQLite

- [ ] Mejorar `history` con filtros por fecha, tienda y orden por precio/fecha
- [ ] Mostrar variación de precio vs. última observación en `history`
- [ ] Sumar exportación de historial a CSV y/o JSON
- [ ] Guardar búsquedas tipo watchlist para seguimiento de productos
- [ ] Agregar alertas simples cuando un precio baja o sube
- [ ] Ajustar TTLs e invalidación del cache por etapa
- [ ] Medir ahorro real de OpenRouter con cache activado

### Robustez del flujo de búsqueda

- [ ] Mejorar mensajes de error y UX de la CLI en casos de URLs vacías, HTML roto o sin resultados
- [ ] Revisar cobertura de casos borde del extractor (JSON truncado, HTML comprimido, páginas raras)
- [ ] Agregar retry/recuperación más clara en etapas de web search y extracción
- [ ] Revisar trazabilidad por etapa para entender dónde falla cada búsqueda

### Limpieza y deuda técnica

- [ ] Revisar `internal/scraper/` y decidir si sigue en uso o ya es legado
- [ ] Definir un plan corto de CI/lint si todavía no está cerrado
- [ ] Agregar tests de repositorio SQLite y fixtures reales de HTML
- [ ] Separar backlog futuro en “bugfix”, “tests”, “feature” y “cleanup”

## Próximo orden sugerido

1. Validar runtime real con un `search`.
2. Correr tests y diagnosticar fallas reales.
3. Priorizar mejoras útiles sobre SQLite (`history`, export, watchlist, alertas).
4. Limpiar legado y cerrar deuda técnica.

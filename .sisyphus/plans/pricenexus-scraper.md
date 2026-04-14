# PriceNexus - Sistema de Agentes de Precios

## TL;DR

> **Quick Summary**: Sistema CLI en Go para buscar productos en tiendas online de Argentina, verificar stock y envío, comparar precios de 5+ variantes, con almacenamiento histórico en SQLite.
> 
> **Deliverables**:
> - CLI en Go (`pricenexus`) para búsqueda manual
> - Motor de scraping extensible con arquitectura de plugins
> - Base de datos SQLite con schema para precios históricos
> - Agendador para ejecución programada
> - Documentación de arquitectura
> 
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 3+ waves
> **Critical Path**: Project setup → DB schema → Core scraper → CLI → Scheduler

---

## Context

### Original Request
Usuario necesita un sistema de agentes que:
1. Dado un nombre de producto (ej "Game Stick Lite") buscar en tiendas de Argentina
2. Verificar disponibilidad/stock
3. Verificar si tiene envío
4. Armar listado de precios con al menos 5 variantes
5. Guardar histórico en SQLite
6. Disponer CLI en Golang para consultar precios

### Interview Summary

**Key Discussions**:
- Scraping de cualquier página (no APIs oficiales)
- Uso personal del sistema
- Actualización programada (no bajo demanda)
- Producto ejemplo: "Game Stick Lite"

**Research Findings**:
- Necesidad de arquitectura extensible para múltiples scrapers
- Go es ideal para CLI por performance y binario estático
- SQLite es apropiado para uso personal (sin configuración)
- Considerar rate limiting y user-agent para evitar bloquos

### Metis Review

**Identified Gaps** (addressed):
- Sitios objetivo no especificados → Arquitectura extensible
- Scheduling no definido → Arquitectura modular

---

## Work Objectives

### Core Objective
Construir sistema CLI para búsqueda y trackeo de precios de productos en tiendas online de Argentina con almacenamiento histórico.

### Concrete Deliverables
- [ ] CLI ejecutable: `pricenexus search "Game Stick Lite"`
- [ ] Scrapers configurables para múltiples tiendas
- [ ] Schema SQLite con tablas: products, prices, shops, price_history
- [ ] Sistema de agendado (cron/programable)
- [ ] README con usage instructions

### Definition of Done
- [ ] `pricenexus search "Game Stick Lite"` retorna precios de al menos 5 tiendas
- [ ] Cada resultado tiene: tienda, precio, stock, envío
- [ ] `pricenexus history "Game Stick Lite"` muestra histórico
- [ ] Los datos se guardan en SQLite automáticamente

### Must Have
- [ ] CLI funcional en Go
- [ ] Scraper extensible (añadir tiendas sin modificar core)
- [ ] SQLite con datos persistentes
- [ ] Búsqueda por nombre de producto
- [ ] Verificación de stock y envío

### Must NOT Have (Guardrails)
- [ ] NO usar APIs de terceros sin autorización
- [ ] NO integrar servicios comerciales de affiliates
- [ ] NO almacenar credenciales de usuarios
- [ ] NO hacer scraping sin rate limiting (min 2s entre requests)

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification agent-executed. No exceptions.
> Acceptance criteria: "user manually tests/confirms" is FORBIDDEN.

### Test Decision
- **Infrastructure exists**: NO - nuevo proyecto
- **Automated tests**: YES - Tests-after
- **Framework**: go test / httptest
- **Test coverage**: Unit tests para modelos, integration tests para CLI

### QA Policy
Every task MUST have agent-executed QA scenarios (TODO template below).

- **CLI**: Use Bash - invocar comando y validar output
- **DB**: Use Bash - sqlite3 para queries de verificación
- **Scraper**: Use Bash/curl - testing de endpoints

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately - foundation + scaffolding):
├── Task 1: Project setup - go.mod, Makefile [quick]
├── Task 2: DB schema - SQLite models [quick]
├── Task 3: Models - Product, Price, Shop [quick]
├── Task 4: DB repository layer [quick]
└── Task 5: CLI framework - Cobra/Viper [quick]

Wave 2 (After Wave 1 - core scraping, MAX PARALLEL):
├── Task 6: Scraper interface + base types [quick]
├── Task 7: Generic web scraper (Colly) [quick]
├── Task 8: MercadoLibre scraper [deep]
├── Task 9: Garbarino scraper [deep]
├── Task 10: Tecnoshops scraper [deep]
├── Task 11: Parser HTML genérico [quick]
└── Task 12: Storage repository impl [deep]

Wave 3 (After Wave 2 - CLI commands + integration):
├── Task 13: CLI search command [deep]
├── Task 14: CLI history command [deep]
├── Task 15: CLI add command [quick]
├── Task 16: Scheduler - basic cron [unspecified-high]

└── Task 18: Error handling + logging [unspecified-high]

Wave FINAL (After ALL tasks — 4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay
```

---

## TODOs

- [ ] 1. Project setup - go.mod, Makefile

  **What to do**:
  - Inicializar proyecto Go con go.mod
  - Crear estructura de directorios (cmd/, internal/, pkg/)
  - Crear Makefile con targets: build, test, run
  - Agregar dependencias: cobra, colly, sqlite3, viper

  **Must NOT do**:
  - Ninguna dependencia sin verificar

  **Recommended Agent Profile**:
  > Category: quick
  - Reason: Tareas de setup simple
  - Skills: []
  - Skills Evaluated but Omitted: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 1
  - Blocks: Tasks 6-12
  - Blocked By: None

  **References**:
  - Go modules docs para estructura standard

  **Acceptance Criteria**:
  - [ ] go mod init y go.mod creado
  - [ ] make build compila sin errores

  **QA Scenarios**:
  ```
  Scenario: Project setup compiles
    Tool: Bash
    Preconditions: Ninguno
    Steps:
      1. go build -o pricenexus ./cmd/cli
    Expected Result: Binario creado sin errores
    Failure Indicators: Errores de compilación
    Evidence: pricenexus binary exists
  ```

  **Commit**: YES
  - Message: `feat(core): project setup`
  - Files: go.mod, Makefile, cmd/, internal/`

- [ ] 2. DB schema - SQLite models

  **What to do**:
  - Definir schema SQLite:
    - shops (id, name, url, active, created_at)
    - products (id, name, search_term, created_at)
    - prices (id, product_id, shop_id, price, currency, has_stock, has_shipping, url, scraped_at)
    - price_history (mismo schema que prices para auditoría)
  - Crear archivo migrations/001_init.sql
  - Crear función de inicialización de DB

  **Must NOT do**:
  - No agregar índices prematuramente

  **Recommended Agent Profile**:
  - Category: quick
  - Reason: Schema definition straightforward
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 1
  - Blocks: Tasks 4, 12
  - Blocked By: None

  **References**:
  - SQLite best practices

  **Acceptance Criteria**:
  - [ ] Schema define las 4 tablas
  - [ ] migrations/001_init.sql existe

  **QA Scenarios**:
  ```
  Scenario: Schema file exists
    Tool: Bash
    Preconditions: Ninguno
    Steps:
      1. ls migrations/001_init.sql
    Expected Result: File exists
    Failure Indicators: File not found
    Evidence: ls output
  ```

- [ ] 3. Models - Product, Price, Shop

  **What to do**:
  - Crear internal/models/types.go
  - Definir struct Shop{name, URL, Active}
  - Definir struct Product{name,SearchTerm}
  - Definir struct Price{ProductID,ShopID,Price,Currency,HasStock,HasShipping,URL,ScrapedAt}
  - Agregar tags JSON y validaciones básicas

  **Must NOT do**:
  - No lógica de negocio en modelos

  **Recommended Agent Profile**:
  - Category: quick

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 1
  - Blocks: Tasks 4, 6, 7
  - Blocked By: None

  **Acceptance Criteria**:
  - [ ] internal/models/types.go existe
  - [ ] 3 structs definidos

  **QA Scenarios**:
  ```
  Scenario: Models compile
    Tool: Bash
    Preconditions: Ninguno
    Steps:
      1. go build ./internal/models
    Expected Result: Sin errores
    Evidence: go build output
  ```

- [ ] 4. DB repository layer

  **What to do**:
  - Crear internal/db/repository.go
  - Implementar interfaz Repository:
    - Init() error
    - GetAllShops() ([]Shop, error)
    - AddShop(Shop) error
    - GetProduct(string) (Product, error)
    - AddProduct(Product) error
    - AddPrice(Price) error
    - GetPricesByProduct(int) ([]Price, error)
  - Usar library: github.com/jmoiron/sqlx

  **Must NOT do**:
  - No exponer sqlx directamente

  **Recommended Agent Profile**:
  - Category: quick
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 1
  - Blocks: Tasks 12, 14
  - Blocked By: 2, 3

  **Acceptance Criteria**:
  - [ ] Repository interface definida
  - [ ] Implementación SQLite funciona
  - [ ] go build ./internal/db sin errores

  **QA Scenarios**:
  ```
  Scenario: DB compiles
    Tool: Bash
    Preconditions: Ninguno
    Steps:
      1. go build ./internal/db
    Expected Result: Sin errores
    Evidence: go build output
  ```

- [ ] 5. CLI framework - Cobra/Viper

  **What to do**:
  - Inicializar Cobra para CLI
  - Configurar Viper para config
  - Crear cmd/root.go con estructura basic
  - Agregar flags: --config, --verbose

  **Must NOT do**:
  - No crear subcommands aún

  **Recommended Agent Profile**:
  - Category: quick
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 1
  - Blocks: Tasks 13-15
  - Blocked By: 1

  **Acceptance Criteria**:
  - [ ] go run ./cmd/cli compila
  - [ ] --help muestra ayuda

  **QA Scenarios**:
  ```
  Scenario: CLI displays help
    Tool: Bash
    Preconditions: Ninguno
    Steps:
      1. go run ./cmd/cli --help
    Expected Result: Muestra ayuda
    Failure Indicators: Error
    Evidence: help output
  ```

- [ ] 6. Scraper interface + base types

  **What to do**:
  - Crear internal/scraper/scraper.go
  - Definir interfaz Scraper:
    - Name() string
    - BaseURL() string
    - Search(query string) ([]Result, error)
  - Definir struct Result{Price,ProductName,URL,HasStock,HasShipping}
  - Interfaz flexible para implementar nuevos scrapers

  **Must NOT do**:
  - No lógica específica de tiendas

  **Recommended Agent Profile**:
  - Category: quick
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 2
  - Blocks: Tasks 8-10
  - Blocked By: 3

  **Acceptance Criteria**:
  - [ ] Interfaz definida
  - [ ] Go compila

- [ ] 7. Generic web scraper (Colly)

  **What to do**:
  - Crear base scraper usando github.com/gocolly/colly
  - Implementar rate limiting (2s entre requests)
  - User-agent configurable
  - Retry logic (3 intentos)
  - Timeout configurado

  **Must NOT do**:
  - No hardcodear URLs

  **Recommended Agent Profile**:
  - Category: quick

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 2
  - Blocks: Tasks 8-10
  - Blocked By: 6

  **Acceptance Criteria**:
  - [ ] Base scaper compilable
  - [ ] Rate limiting funciona

- [ ] 8. MercadoLibre scraper

  **What to do**:
  - Crear internal/scraper/mercadolibre.go
  - Implementar Search para MercadoLibre Argentina
  - Parsear resultados: precio, título, URL
  - Detectar stock (disponible/consultar)
  - Detectar envío (llega/código postal)

  **Must NOT do**:
  - No hacer más de 1 request cada 5 segundos

  **Recommended Agent Profile**:
  - Category: deep
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 2
  - Blocks: Tasks 12, 14
  - Blocked By: 6, 7

  **Acceptance Criteria**:
  - [ ] Búsqueda retorna resultados
  - [ ] Compila sin errores

- [ ] 9. Garbarino scraper

  **What to do**:
  - Crear internal/scraper/garbarino.go
  - Implementar Search para Garbarino
  - Mismo patrón que Task 8

  **Must NOT do**:
  - Rate limiting mínimo 2s

  **Recommended Agent Profile**:
  - Category: deep
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 2
  - Blocks: 12, 14
  - Blocked By: 6, 7

- [ ] 10. Tecnoshops scraper

  **What to do**:
  - Crear internal/scraper/tecnoshops.go
  - Implementar Search para Tecnoshops
  - Mismo patrón que Task 8

  **Recommended Agent Profile**:
  - Category: deep
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 2
  - Blocks: 12, 14
  - Blocked By: 6, 7

- [ ] 11. Parser HTML genérico

  **What to do**:
  - Crear internal/scraper/parser.go
  - Utilidades para parsear:
    - Price (manejar "$", ".", ",")
    - Stock status
    - Shipping info
  - Reutilizable entre scrapers

  **Recommended Agent Profile**:
  - Category: quick

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 2
  - Blocks: 8-10
  - Blocked By: 6

- [ ] 12. Storage repository impl

  **What to do**:
  - Implementar Repository completo
  - Conectar scrapers con DB
  - Implementar AddPrice con timestamp auto

  **Recommended Agent Profile**:
  - Category: deep
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 2
  - Blocks: 14, 15
  - Blocked By: 4

- [ ] 13. CLI search command

  **What to do**:
  - Crear cmd/search.go
  - Implementar: pricenexus search "producto"
  - Invocar todos los scrapers configurados
  - Mostrar resultados tabulados
  - Guardar en DB automáticamente

  **Must NOT do**:
  - No mostrar errores de parsing como failures

  **Recommended Agent Profile**:
  - Category: deep
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 3
  - Blocks: 16, 17
  - Blocked By: 5, 8-11

  **Acceptance Criteria**:
  - [ ] ./pricenexus search "Game Stick Lite" funciona
  - [ ] Muestra al menos 5 precios

  **QA Scenarios**:
  ```
  Scenario: Search command works
    Tool: Bash
    Preconditions: DB inicializada
    Steps:
      1. ./pricenexus search "Game Stick Lite"
    Expected Result: Lista precios
    Failure Indicators: Error en comando
    Evidence: stdout output
  ```

- [ ] 14. CLI history command

  **What to do**:
  - Crear cmd/history.go
  - Implementar: pricenexus history "producto"
  - Mostrar histórico de precios
  - Soporte para filtrar por tienda

  **Recommended Agent Profile**:
  - Category: deep
  - Skills: []

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 3
  - Blocks: None
  - Blocked By: 12

  **Acceptance Criteria**:
  - [ ] ./pricenexus history "Game Stick Lite" muestra precios

- [ ] 15. CLI add command

  **What to do**:
  - Crear cmd/add.go
  - Implementar: pricenexus add "tienda" "URL"
  - Agregar nuevas tiendas al sistema

  **Recommended Agent Profile**:
  - Category: quick

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 3
  - Blocks: 16
  - Blocked By: 5

- [ ] 16. Scheduler - basic cron

  **What to do**:
  - Crear internal/scheduler/scheduler.go
  - Support para ejecución programada
  - Configurar intervals: cada 1h, 6h, 24h
  - Opcional: daemon mode

  **Recommended Agent Profile**:
  - Category: unspecified-high

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 3
  - Blocks: None
  - Blocked By: 13, 14



  **Recommended Agent Profile**:
  - Category: quick

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 3
  - Blocks: None
  - Blocked By: 5

- [ ] 18. Error handling + logging

  **What to do**:
  - Agregar logging con logrus o zerolog
  - Manejo de errores consistente
  - Verbose flag para debug

  **Recommended Agent Profile**:
  - Category: unspecified-high

  **Parallelization**:
  - Can Run In Parallel: YES
  - Parallel Group: Wave 3
  - Blocks: None
  - Blocked By: 5

---

---

## Final Verification Wave (MANDATORY - after ALL tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go build` + `go vet` + linter. Review all changed files for: `any`, empty catches, commented-out code, unused imports. Check AI slop patterns: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- Wave 1: `feat(core): project setup`
- Wave 2: `feat(scraper): addMercadoLibre scraper`
- Wave 3: `feat(cli): add search history commands`

---

## Success Criteria

### Verification Commands
```bash
go build -o pricenexus ./cmd/cli
./pricenexus search "Game Stick Lite"  # Expected: output con precios
./pricenexus history "Game Stick Lite"  # Expected: output con histórico
sqlite3 prices.db ".schema"  # Expected: tables config
```

### Final Checklist
- [ ] CLI funcional
- [ ] Scrapers extensible
- [ ] SQLite persistence
- [ ] Rate limiting implementado
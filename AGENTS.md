# AGENTS.md

## Project Overview

PriceNexus is a Go CLI for searching and tracking product prices in Argentine online stores. It uses a multi-agent workflow built with LangChain Go and relies on OpenRouter's `openrouter:web_search` server tool to discover product URLs dynamically on `.com.ar` and `.ar` domains.

## Setup & Prerequisites

1. **Go**: Use the Go version declared in `go.mod` (currently `1.26.2`).
2. **Environment**: Copy `.env.example` to `.env` and fill in the required values.
3. **OpenRouter**: Required for the default setup because the orchestrator, web searcher, and default data extractor use OpenRouter models.
   - Recommended orchestrator model: `openrouter:xiaomi/mimo-v2-flash`
   - Recommended web searcher model: `openrouter:nvidia/nemotron-3-super-120b-a12b:free`
   - Default data extractor model: `openrouter:xiaomi/mimo-v2-flash`
   - Sign up at https://openrouter.ai
   - Generate an API key and set `OPENROUTER_API_KEY=your_key_here`
4. **Ollama (optional)**: Only required if you explicitly switch the data extractor to a local Ollama model such as `ollama:gemma4:e4b`.

### Optional Ollama Setup

If you want to run extraction locally with Ollama:

```bash
ollama serve
ollama pull gemma4:e4b
```

## Essential Commands

Use the `Makefile` for the standard workflow:

- **Build**: `make build` — creates the `pricenexus` binary in the repository root
- **Run**: `make run`
- **Test**: `make test`
- **Install**: `make install`
- **Clean**: `make clean`

Direct Go commands:

- **Build**: `go build -o pricenexus ./cmd/cli`
- **Run**: `go run ./cmd/cli`
- **Test**: `go test -v ./...`

## Project Structure

```text
PriceNexus/
├── cmd/cli/                  # CLI entrypoint (main.go)
├── cmd/                      # Cobra commands (search, history, add)
├── internal/
│   ├── agent/
│   │   ├── orchestrator/     # Workflow coordinator
│   │   ├── websearcher/      # OpenRouter web_search integration
│   │   ├── pageloader/       # HTML fetching and page loading
│   │   ├── dataextractor/    # Product extraction pipeline
│   │   ├── validator/        # Deterministic post-processing and validation
│   │   ├── storage/          # Persistence helpers
│   │   ├── shared/           # Shared agent models
│   │   ├── openrouter.go     # OpenRouter client and tool wiring
│   │   └── config.go         # LLM and search configuration
│   ├── db/                   # SQLite repository layer
│   └── models/               # Domain models
├── docs/                     # Project documentation
├── migrations/               # SQL migrations
└── prices.db                 # SQLite database (created on run)
```

## Configuration & Environment

Key variables from `.env.example`:

- `OPENROUTER_API_KEY`: Required for the default setup.
- `PRICE_NEXUS_ORCHESTRATOR_LLM`: Default `openrouter:xiaomi/mimo-v2-flash`
- `PRICE_NEXUS_WEBSEARCHER_LLM`: Default `openrouter:nvidia/nemotron-3-super-120b-a12b:free`
- `PRICE_NEXUS_DATAEXTRACTOR_LLM`: Default `openrouter:xiaomi/mimo-v2-flash`
- `PRICE_NEXUS_WEBSEARCH_ALLOWED_DOMAINS`: Default `.com.ar,.ar`
- `PRICE_NEXUS_WEBSEARCH_MAX_RESULTS`: Default `10`
- `PRICE_NEXUS_DEFAULT_CURRENCY`: Default `ARS`

### CLI Environment Loading

- `cmd/cli/main.go` loads `.env` automatically when present.
- If `OPENROUTER_API_KEY` is set and `OPENAI_API_KEY` is not, the CLI copies the OpenRouter key into `OPENAI_API_KEY` for compatibility with OpenAI-style clients.

## Architecture & Agents

The current workflow is:

1. **Orchestrator**
   - Coordinates the search flow
   - Creates agent instances from `internal/agent/config.go`
   - Default LLM: OpenRouter `xiaomi/mimo-v2-flash`

2. **Web Searcher**
   - Uses OpenRouter `openrouter:web_search`
   - Searches only on allowed Argentine domains
   - Default model: `openrouter:nvidia/nemotron-3-super-120b-a12b:free`
   - In practice it must use OpenRouter to get real `web_search` support

3. **Page Loader**
   - Fetches HTML for discovered URLs
   - No LLM required

4. **Data Extractor**
   - Extracts product data from HTML
   - Current default: OpenRouter `xiaomi/mimo-v2-flash`
   - Optional local mode: Ollama, for example `ollama:gemma4:e4b`
   - Uses a layered extraction pipeline with structured data, LLM content-finder extraction, and gentle-cleaning fallback

5. **Validator**
   - Performs deterministic normalization and filtering on extracted results
   - Current implementation does not actively use an LLM, even though the constructor accepts one

6. **Storage**
   - Persists validated results to SQLite
   - No LLM required

### Web Search with OpenRouter

The web searcher configures the OpenRouter server tool like this:

```json
{
  "type": "openrouter:web_search",
  "parameters": {
    "allowed_domains": [".com.ar", ".ar"],
    "max_results": 10
  }
}
```

Benefits:

- Dynamic store discovery
- Domain filtering for Argentina-focused results
- No hardcoded store catalog
- Official OpenRouter tool integration

## Removed Components

The project no longer relies on older URL-discovery approaches:

- **URLFinder Agent** — removed in favor of dynamic search
- **pricetools/SearchTool** — removed in favor of OpenRouter `web_search`

## LLM Dependencies

- **Orchestrator**: OpenRouter in the default configuration
- **Web Searcher**: OpenRouter in the supported configuration
- **Data Extractor**: OpenRouter by default, Ollama optional
- **Validator**: No active LLM usage in the current implementation

## Testing

- Run all tests: `make test` or `go test -v ./...`
- Run agent-focused tests: `go test ./internal/agent/...`
- Only start Ollama if you intentionally configured the extractor to use Ollama

## Common Pitfalls / Gotchas

1. **Missing API Key**: `OPENROUTER_API_KEY` is required for the default setup.
2. **Web Searcher Must Use OpenRouter**: Local Ollama models do not provide the OpenRouter `web_search` tool.
3. **MiMo as Web Searcher**: `xiaomi/mimo-v2-flash` is fine for orchestration and extraction, but not reliable as the web-search model in this project.
4. **Ollama Is Optional**: Only required if you explicitly set an Ollama model for extraction.
5. **Database Locked**: If `prices.db` is locked, make sure no other process is using it.
6. **Wrong Domain Filters**: The default search is intentionally limited to `.com.ar` and `.ar`.

## References

- [README](README.md)
- [Setup Guide](docs/SETUP.md)
- [LLM Configuration](docs/LLM_CONFIG.md)
- [Architecture Details](docs/ARCHITECTURE.md)
- [Command Reference](docs/COMMANDS.md)
- [OpenRouter Web Search Docs](https://openrouter.ai/docs/guides/features/server-tools/web-search)

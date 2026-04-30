# Delta Spec: Translate to Spanish

## Intent
Localize all user-facing strings in the OpenRouter Discord Monitor to Spanish to improve accessibility for Spanish-speaking users.

## Scope
- Discord embed titles, descriptions, and labels.
- Service operational logs and health check responses.
- Documentation (`README.md`).

## Localization Requirements
- **MUST** translate all user-facing strings to Spanish.
- **MUST NOT** translate environment variable names or command line tools.

### Discord Notifications (internal/discord)
- Titles:
  - "New Models Discovered" -> "Nuevos Modelos Detectados"
  - "Model Price Updates" -> "Actualizaciones de Precios"
  - "Models No Longer Available" -> "Modelos No Disponibles"
  - "Daily Digest" -> "Resumen Diario"
- Descriptions & Labels:
  - "new model(s) detected" -> "nuevo(s) modelo(s) detectado(s)"
  - "price change detected" -> "cambio de precio detectado"
  - "model removed" -> "modelo eliminado"
  - "Best by Cost" -> "Mejor por Costo"
  - "Best by Context/Cost" -> "Mejor relación Contexto/Costo"

### Service & Health (internal/service)
- Health response: "healthy" -> "saludable", "unhealthy" -> "no saludable"
- Logs: Translate main operational logs (e.g., "Starting service" -> "Iniciando servicio").

### Documentation (README.md)
- Translate the entire README.md to Spanish.

## Affected Scenarios (Updates to openrouter-discord-monitor-spec.md)
- SC-1: Initial Poll message "New models discovered" -> "Nuevos modelos detectados".
- SC-2: Price change notification.
- SC-3: New model added notification.
- Discord Embed Format: Update examples with Spanish strings.

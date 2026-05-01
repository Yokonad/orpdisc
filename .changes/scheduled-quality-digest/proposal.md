# Change Proposal: scheduled-quality-digest

## Summary

Implement time-windowed quality digest notifications for the OpenRouter Discord Monitor, sending enhanced model rankings every 30 minutes during business hours (9am-7pm Peru time), while maintaining 24/7 silent polling for change detection.

## Motivation

Currently:
- The service polls every 30 minutes but the digest only shows Top 1 by cost + Top 1 by context/cost ratio
- Digest notifications fire regardless of time (can send at night)
- Users want richer digest content showing newest and most capable models, not just cheapest

Desired:
- **Time-windowed notifications**: Only send digest between 9am-7pm (America/Lima, UTC-5)
- **Enhanced content**: Show 4 categories - cheapest, best value, most capable (context length), newest
- **Maintain 24/7 polling**: Continue detecting changes silently outside business hours

## Scope

### In Scope

1. **Configuration (`internal/config/config.go`)**
   - Add `ActiveStartHour int` with env `ACTIVE_START_HOUR` (default: `9`)
   - Add `ActiveEndHour int` with env `ACTIVE_END_HOUR` (default: `19`)
   - Add `DigestInterval time.Duration` with env `DIGEST_INTERVAL_MINUTES` (default: `30`)

2. **Processor (`internal/processor/processor.go`)**
   - Add `TopByContextLength(models, n)` - sort by context length descending
   - Add `TopByNewest(models, n)` - sort by FirstSeen descending (newest first)

3. **Service (`internal/service/service.go`)**
   - Add `isWithinActiveHours()` method using `time.Now()` in `America/Lima` timezone
   - Modify `Start()` to check active hours before calling `SendDigest()`
   - Enhance `SendDigest()` to include all 4 ranking categories
   - Change detection notifications (`poll()`) remain time-agnostic (fire immediately)

4. **Discord Webhook (`internal/discord/webhook.go`)**
   - Enhance digest embed to show 4 categories in Spanish:
     - "Mejor por Costo" (cheapest)
     - "Mejor relación Contexto/Costo" (best value)
     - "Más Capaz" (largest context)
     - "Más Nuevo" (newest model)
   - Update embed structure to accommodate multiple categories

### Out of Scope

- Database schema changes (all required fields exist: `ContextLength`, `FirstSeen`)
- API changes (no new endpoints)
- New data sources (use existing OpenRouter API)
- Change detection notification timing (still fire immediately when detected)
- Timezone configuration (hardcoded to America/Lima as per requirements)

## Approach

### 1. Configuration Changes

**File**: `internal/config/config.go`

Add operating hours configuration:

```go
type Config struct {
    // ... existing fields ...
    ActiveStartHour int `env:"ACTIVE_START_HOUR" envDefault:"9"`
    ActiveEndHour   int `env:"ACTIVE_END_HOUR" envDefault:"19"`
    DigestInterval  time.Duration `env:"DIGEST_INTERVAL_MINUTES" envDefault:"30m"`
}
```

**Validation**: Ensure `ActiveStartHour < ActiveEndHour` and both are in range [0, 23].

### 2. Processor Ranking Functions

**File**: `internal/processor/processor.go`

Add two new ranking functions following the existing pattern:

```go
// TopByContextLength returns the top N models sorted by largest context length
func TopByContextLength(modelList []models.Model, n int) []models.Model {
    sorted := make([]models.Model, len(modelList))
    copy(sorted, modelList)
    
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].ContextLength > sorted[j].ContextLength
    })
    
    if n > len(sorted) {
        n = len(sorted)
    }
    return sorted[:n]
}

// TopByNewest returns the top N models sorted by FirstSeen (newest first)
func TopByNewest(modelList []models.Model, n int) []models.Model {
    sorted := make([]models.Model, len(modelList))
    copy(sorted, modelList)
    
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].FirstSeen.After(sorted[j].FirstSeen)
    })
    
    if n > len(sorted) {
        n = len(sorted)
    }
    return sorted[:n]
}
```

### 3. Service Time Window Logic

**File**: `internal/service/service.go`

Add timezone-aware time checking:

```go
import "time"

const peruTimezone = "America/Lima"

func (s *Service) isWithinActiveHours() bool {
    now := time.Now()
    peruLocation, err := time.LoadLocation(peruTimezone)
    if err != nil {
        s.logger.Warn("Failed to load Peru timezone, defaulting to UTC: %v", err)
        peruLocation = time.UTC
    }
    
    peruTime := now.In(peruLocation)
    hour := peruTime.Hour()
    
    return hour >= s.cfg.ActiveStartHour && hour < s.cfg.ActiveEndHour
}
```

Modify `Start()` to check before sending digest:

```go
case <-ticker.C:
    s.poll()
    // Only send digest during active hours
    if s.isWithinActiveHours() {
        if err := s.SendDigest(s.ctx); err != nil {
            s.logger.Error("Error al enviar resumen: %v", err)
        }
    }
```

**Note**: Initial poll on startup should also respect the time window for digest.

### 4. Enhanced Digest Content

**File**: `internal/service/service.go` - `SendDigest()`

Update to fetch all 4 categories:

```go
func (s *Service) SendDigest(ctx context.Context) error {
    allModels, err := s.db.GetAllModels()
    if err != nil {
        return fmt.Errorf("failed to get models for digest: %w", err)
    }
    
    if len(allModels) == 0 {
        return nil
    }
    
    // Get top model in each category
    topByCost := processor.TopByCostPer1K(allModels, 1)
    topByRatio := processor.TopByContextCostRatio(allModels, 1)
    topByContext := processor.TopByContextLength(allModels, 1)
    topByNewest := processor.TopByNewest(allModels, 1)
    
    changeset := &models.Changeset{
        NewModels:     topByCost,      // Reuse for cheapest
        UpdatedModels: topByRatio,     // Reuse for best value
        // Will need to extend Changeset or pass additional data
    }
    
    return s.webhook.SendDigest(ctx, allModels) // Pass all models for ranking
}
```

### 5. Discord Embed Enhancement

**File**: `internal/discord/webhook.go`

Update `BuildEmbedsForChangeset` to handle enhanced digest format:

```go
// For digest, build a rich embed with 4 categories
if changeset.IsDigest {
    var fields []DiscordField
    
    // Category 1: Mejor por Costo
    if len(changeset.NewModels) > 0 {
        m := changeset.NewModels[0]
        fields = append(fields, DiscordField{
            Name: "💰 Mejor por Costo",
            Value: fmt.Sprintf("[%s](%s%s)\n$%.6f/1K tokens | %d context", 
                m.Name, OpenRouterBaseURL, m.ID, 
                m.CostPer1KTokens(), m.ContextLength),
            Inline: true,
        })
    }
    
    // Category 2: Mejor relación Contexto/Costo
    if len(changeset.UpdatedModels) > 0 {
        m := changeset.UpdatedModels[0]
        fields = append(fields, DiscordField{
            Name: "📊 Mejor relación Contexto/Costo",
            Value: fmt.Sprintf("[%s](%s%s)\n$%.6f/1K tokens | %d context",
                m.Name, OpenRouterBaseURL, m.ID,
                m.CostPer1KTokens(), m.ContextLength),
            Inline: true,
        })
    }
    
    // Category 3: Más Capaz (context length)
    // Category 4: Más Nuevo (first seen)
    // ... similar pattern
    
    embed := DiscordEmbed{
        Title: "🤖 Resumen de Modelos - OpenRouter",
        Description: "Ranking actualizado cada 30 minutos (9am-7pm hora Perú)",
        Color: ColorBlue,
        Timestamp: timestamp,
        Fields: fields,
        Footer: &DiscordFooter{Text: "Monitor de OpenRouter"},
    }
    embeds = append(embeds, embed)
    return embeds
}
```

**Consideration**: May need to refactor `SendDigest` to pass all 4 categories explicitly rather than overloading `Changeset` fields.

## Technical Considerations

### Timezone Handling
- Use `time.LoadLocation("America/Lima")` for accurate Peru time
- Handle load errors gracefully (fallback to UTC with warning)
- Peru is UTC-5 year-round (no DST)

### Edge Cases
1. **Empty model list**: Return early, no notification
2. **Fewer than 4 unique models**: Same model may appear in multiple categories (acceptable)
3. **FirstSeen is zero**: Models without `first_seen` timestamp sort to end
4. **Midnight boundary**: `hour < ActiveEndHour` handles 7pm cutoff correctly

### Backward Compatibility
- Existing environment variables remain unchanged
- New configs have sensible defaults
- Change detection notifications unchanged

### Testing Strategy
1. Unit tests for ranking functions (verify sort order)
2. Unit tests for `isWithinActiveHours()` (test boundary conditions)
3. Integration test: mock time, verify digest only sent in window

## Files to Modify

| File | Changes |
|------|---------|
| `internal/config/config.go` | Add 3 new config fields + validation |
| `internal/processor/processor.go` | Add 2 ranking functions |
| `internal/service/service.go` | Add time check, modify digest logic |
| `internal/discord/webhook.go` | Enhance digest embed format |
| `internal/models/types.go` | (possibly) extend Changeset for 4 categories |

## Dependencies

- No new external dependencies
- Uses Go standard library `time` package for timezone handling

## Success Criteria

1. ✅ Digest notifications only sent between 9am-7pm Peru time
2. ✅ Digest shows 4 ranking categories in Spanish
3. ✅ Change detection notifications still fire immediately (24/7)
4. ✅ Service continues polling every 30 minutes regardless of time window
5. ✅ All user-facing strings in Spanish
6. ✅ Default configuration works out-of-the-box

## Risks

| Risk | Mitigation |
|------|------------|
| Timezone load failure | Fallback to UTC with warning log |
| Model with zero FirstSeen | Sorts to end naturally; handle gracefully |
| Discord embed character limits | Test with 4 categories; may need to split into multiple embeds |
| Performance with large model lists | Ranking is O(n log n); acceptable for ~100-200 models |

## Future Enhancements (Out of Scope)

- Configurable timezone via environment variable
- Weekend/holiday schedule exceptions
- Per-channel digest customization
- Historical digest summary (daily/weekly recap)

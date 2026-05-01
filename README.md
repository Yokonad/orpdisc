# orpdisc

Servicio en Go que consulta la API de OpenRouter cada 30 min, detecta cambios en modelos de IA y envia notificaciones a Discord. Opera unicamente entre 9:00 y 19:00 (hora Peru) y muestra ranking de los mejores modelos por costo, relacion contexto/costo, capacidad y novedad.

## Instalacion

```bash
git clone https://github.com/Yokonad/orpdisc.git
cd orpdisc
go build -o monitor ./cmd/monitor
```

```bash
sudo tee /etc/openrouter-monitor.env << 'EOF'
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/tu-webhook
POLL_INTERVAL_MINUTES=30m
DB_PATH=/var/lib/orpdisc/data.db
HEALTH_CHECK_PORT=:9090
ACTIVE_START_HOUR=9
ACTIVE_END_HOUR=19
EOF

./monitor
```

Para produccion con systemd:

```bash
sudo cp openrouter-monitor.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now openrouter-monitor
```

## Configuracion

| Variable | Default | Descripcion |
|:--|:--:|:--|
| `DISCORD_WEBHOOK_URL` | requerido | URL del webhook de Discord |
| `POLL_INTERVAL_MINUTES` | `30m` | Intervalo entre consultas |
| `DB_PATH` | `./data.db` | Ruta de la base de datos |
| `LOG_LEVEL` | `info` | debug, info, warn, error |
| `HTTP_TIMEOUT_SECONDS` | `30s` | Timeout HTTP |
| `OPENROUTER_BASE_URL` | `https://openrouter.ai/api/v1` | URL base de la API |
| `MAX_RETRIES` | `5` | Reintentos maximos |
| `CIRCUIT_BREAKER_THRESHOLD` | `5` | Fallos para abrir circuito |
| `CIRCUIT_BREAKER_TIMEOUT_MINUTES` | `60m` | Espera del circuit breaker |
| `HEALTH_CHECK_PORT` | `:9090` | Puerto health check |
| `ACTIVE_START_HOUR` | `9` | Hora inicio notificaciones |
| `ACTIVE_END_HOUR` | `19` | Hora fin notificaciones |

## Uso

```bash
sudo systemctl start|stop|restart openrouter-monitor
sudo journalctl -u openrouter-monitor -f
curl http://localhost:9090/health   # 200 = saludable
```

## Licencia

MIT. Ver [LICENSE](https://github.com/Yokonad/orpdisc/blob/main/LICENSE).

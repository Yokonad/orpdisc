<div align="center">

<a href="https://github.com/Yokonad/orpdisc">
  <img src="https://img.shields.io/badge/ORPDISC-OpenRouter%20Discord%20Monitor-00E5FF?style=for-the-badge&logo=discord&logoColor=white&labelColor=0D1117&color=00E5FF" alt="ORPDISC" />
</a>

<br /><br />

<img src="https://media4.giphy.com/media/L1R1tvI9kwzkP8TgIr/giphy.gif" alt="Tech" width="640" style="border-radius: 8px;" />

<br /><br />

<img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=0D1117&color=00ADD8" />
<img src="https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite&logoColor=white&labelColor=0D1117&color=003B57" />
<img src="https://img.shields.io/badge/Discord-5865F2?style=for-the-badge&logo=discord&logoColor=white&labelColor=0D1117&color=5865F2" />
<img src="https://img.shields.io/badge/OpenRouter-FF6B35?style=for-the-badge&logo=openai&logoColor=white&labelColor=0D1117&color=FF6B35" />
<img src="https://img.shields.io/badge/systemd-FFDD00?style=for-the-badge&logo=systemd&logoColor=black&labelColor=0D1117&color=FFDD00" />
<img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white&labelColor=0D1117&color=2496ED" />

</div>

<br />

<div align="left">
  <img src="https://img.shields.io/badge/ACERCA%20DEL%20PROYECTO-00E5FF?style=for-the-badge&labelColor=0D1117&color=00E5FF" />
</div>

Servicio en Go que consulta la API de OpenRouter cada 30 min, detecta cambios en modelos de IA y envia notificaciones a Discord. Opera unicamente entre 9:00 y 19:00 (hora Peru) y muestra ranking de los mejores modelos por costo, relacion contexto/costo, capacidad y novedad.

<br />

<div align="left">
  <img src="https://img.shields.io/badge/INSTALACION-00E676?style=for-the-badge&labelColor=0D1117&color=00E676" />
</div>

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

<br />

<div align="left">
  <img src="https://img.shields.io/badge/CONFIGURACION-FFD700?style=for-the-badge&labelColor=0D1117&color=FFD700" />
</div>

| Variable | Default | Descripcion |
|:--|:--:|:--|
| `DISCORD_WEBHOOK_URL` | requerido | URL del webhook de Discord |
| `POLL_INTERVAL_MINUTES` | `30m` | Intervalo entre consultas |
| `DB_PATH` | `./data.db` | Ruta de la base de datos |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `HTTP_TIMEOUT_SECONDS` | `30s` | Timeout HTTP |
| `OPENROUTER_BASE_URL` | `https://openrouter.ai/api/v1` | URL base de la API |
| `MAX_RETRIES` | `5` | Reintentos maximos |
| `CIRCUIT_BREAKER_THRESHOLD` | `5` | Fallos para abrir circuito |
| `CIRCUIT_BREAKER_TIMEOUT_MINUTES` | `60m` | Espera del circuit breaker |
| `HEALTH_CHECK_PORT` | `:9090` | Puerto health check |
| `ACTIVE_START_HOUR` | `9` | Hora inicio notificaciones |
| `ACTIVE_END_HOUR` | `19` | Hora fin notificaciones |

<br />

<div align="left">
  <img src="https://img.shields.io/badge/USO-FF00FF?style=for-the-badge&labelColor=0D1117&color=FF00FF" />
</div>

```bash
sudo systemctl start|stop|restart openrouter-monitor
sudo journalctl -u openrouter-monitor -f
curl http://localhost:9090/health   # 200 = saludable
```

<br />

<div align="left">
  <img src="https://img.shields.io/badge/LICENCIA-00E676?style=for-the-badge&labelColor=0D1117&color=00E676" />
</div>

MIT. Ver [LICENSE](https://github.com/Yokonad/orpdisc/blob/main/LICENSE).

<br />

<div align="center">
  <sub>_Built with Go, SQLite & Discord Webhooks_</sub>
</div>

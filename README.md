<div align="center">

<!-- ===== MAIN TITLE BANNER ===== -->
<a href="https://github.com/Yokonad/orpdisc">
  <img src="https://img.shields.io/badge/ORPDISC-OpenRouter%20Discord%20Monitor-00E5FF?style=for-the-badge&logo=discord&logoColor=white&labelColor=0D1117&color=00E5FF" alt="ORPDISC" />
</a>

<br />

<!-- ===== SUBTITLE ===== -->
<img src="https://img.shields.io/badge/Automated%20AI%20Model%20Monitor-FF00FF?style=for-the-badge&logo=openai&logoColor=white&labelColor=0D1117&color=FF00FF" alt="Automated AI Model Monitor" />

<br />

<!-- ===== TECH STACK BADGES ===== -->
<a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=0D1117&color=00ADD8" /></a>
<a href="https://www.sqlite.org/"><img src="https://img.shields.io/badge/SQLite-3-003B57?style=for-the-badge&logo=sqlite&logoColor=white&labelColor=0D1117&color=003B57" /></a>
<a href="https://discord.com/developers/docs/resources/webhook"><img src="https://img.shields.io/badge/Discord%20Webhook-5865F2?style=for-the-badge&logo=discord&logoColor=white&labelColor=0D1117&color=5865F2" /></a>
<a href="https://openrouter.ai/"><img src="https://img.shields.io/badge/OpenRouter%20API-FF6B35?style=for-the-badge&logo=openai&logoColor=white&labelColor=0D1117&color=FF6B35" /></a>
<a href="https://systemd.io/"><img src="https://img.shields.io/badge/systemd-FFDD00?style=for-the-badge&logo=systemd&logoColor=black&labelColor=0D1117&color=FFDD00" /></a>
<a href="https://www.docker.com/"><img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white&labelColor=0D1117&color=2496ED" /></a>

<br />

<!-- ===== ANIMATED GIF ===== -->
<img src="https://media4.giphy.com/media/L1R1tvI9kwzkP8TgIr/giphy.gif" alt="Anime Coding" width="720" style="border-radius: 12px; box-shadow: 0 0 30px rgba(0, 229, 255, 0.3);" />

<br />
<br />

<!-- ===== QUICK STATUS BADGES ===== -->
<a href="https://github.com/Yokonad/orpdisc/actions"><img src="https://img.shields.io/badge/build-passing-00E676?style=for-the-badge&logo=githubactions&logoColor=white&labelColor=0D1117&color=00E676" /></a>
<a href="https://github.com/Yokonad/orpdisc/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-FFD700?style=for-the-badge&logo=openaccess&logoColor=white&labelColor=0D1117&color=FFD700" /></a>
<a href="https://go.dev"><img src="https://img.shields.io/badge/go%20report-A+-00E676?style=for-the-badge&logo=go&logoColor=white&labelColor=0D1117&color=00E676" /></a>

</div>

<br />

<!-- ================================================================================== -->
<!-- SECTION: DESCRIPTION                                                               -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%F0%9F%93%96%20DESCRIPCION-FF00FF?style=for-the-badge&labelColor=0D1117&color=FF00FF" />
</div>

<p align="center">
  <b>orpdisc</b> es un servicio automatizado en <b>Go</b> que consulta la API de <b>OpenRouter</b>, detecta cambios en modelos de IA (nuevos, actualizados, eliminados) y envía notificaciones enriquecidas a <b>Discord</b> con precios exactos, ranking de calidad y horario inteligente.
</p>

<br />

<!-- ================================================================================== -->
<!-- SECTION: FEATURES                                                                   -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%E2%9A%A1%20CARACTERISTICAS-00E5FF?style=for-the-badge&labelColor=0D1117&color=00E5FF" />
</div>

<br />

| Característica | Descripción |
|:--|:--|
| 🤖 **Monitor Automatizado** | Consulta la API de OpenRouter cada `30 min` y detecta cambios |
| 🧠 **Ranking Inteligente** | Top modelos por: costo, relación contexto/costo, capacidad y novedad |
| ⏰ **Horario Activo** | Notificaciones solo entre **9:00 - 19:00** (hora Peru) |
| 🛡️ **Circuit Breaker** | Previene saturación de API con tolerancia a fallos configurable |
| 💚 **Health Check** | Endpoint HTTP `:9090/health` para monitoreo |
| 🗄️ **SQLite** | Almacenamiento local sin dependencias externas |
| ♻️ **Auto-Restart** | Systemd reinicia el servicio automáticamente si falla |
| 🐳 **Docker** | Despliegue en contenedor listo para producción |

<br />

<!-- ================================================================================== -->
<!-- SECTION: NOTIFICATIONS PREVIEW                                                      -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%F0%9F%93%AC%20VISTA%20PREVIA%20DE%20NOTIFICACIONES-FF6B35?style=for-the-badge&labelColor=0D1117&color=FF6B35" />
</div>

<br />

Cada 30 min (9am-7pm Peru) recibirás un resumen como este en Discord:

<div align="center">

```
╔══════════════════════════════════════════════════════════╗
║              RESÚMEN DE MODELOS                          ║
║           Mejores modelos del momento                    ║
╠══════════════════╦═══════════════════════╦═══════════════╣
║ Mejor por Costo  ║ Mejor Relacion        ║ Mas Capaz     ║
║ [Gemini 2.5 Pro] ║ Contexto/Costo        ║ (Mayor Ctx)   ║
║ $0.005/1K tokens ║ [Claude Opus]         ║ [Gemini 2.5]  ║
║ 1M context       ║ $0.015/1K tokens      ║ $0.005/1K     ║
║ Ratio: 200       ║ 200K context          ║ 1M context    ║
╠══════════════════╩═══════════════════════╩═══════════════╣
║  🆕 Modelo Mas Nuevo: [GPT-5] — visto 01/05/2026        ║
╚══════════════════════════════════════════════════════════╝
```

</div>

<br />

<!-- ================================================================================== -->
<!-- SECTION: INSTALLATION                                                               -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%F0%9F%9B%A0%20INSTALACION-00E676?style=for-the-badge&labelColor=0D1117&color=00E676" />
</div>

<br />

### 📋 Requisitos Previos

| Requisito | Versión Mínima |
|:--|:--:|
| [Go](https://go.dev/dl/) | `1.21+` |
| [SQLite](https://www.sqlite.org/download.html) | `3.x` |
| [Discord Webhook](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) | URL válida |

<br />

### 🐚 Desde Código Fuente

```bash
# Clonar repositorio
git clone https://github.com/Yokonad/orpdisc.git
cd orpdisc

# Compilar binario
go build -o monitor ./cmd/monitor

# Crear archivo de entorno
sudo tee /etc/openrouter-monitor.env << 'EOF'
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/tu-webhook-aqui
POLL_INTERVAL_MINUTES=30m
DB_PATH=/home/usuario/orpdisc/data.db
LOG_LEVEL=info
HEALTH_CHECK_PORT=:9090
EOF

# Ejecutar
./monitor
```

<br />

### 🐳 Desde Docker

```bash
# Construir imagen
docker build -t orpdisc .

# Ejecutar contenedor
docker run -d \
  --name orpdisc \
  --restart unless-stopped \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -e POLL_INTERVAL_MINUTES=30m \
  -e ACTIVE_START_HOUR=9 \
  -e ACTIVE_END_HOUR=19 \
  -v /path/to/data:/data \
  orpdisc
```

<br />

### ⚙️ Desde systemd (Recomendado para VPS)

```bash
# 1. Copiar archivo de servicio
sudo cp openrouter-monitor.service /etc/systemd/system/

# 2. Crear archivo de entorno
sudo nano /etc/openrouter-monitor.env

# 3. Activar e iniciar servicio
sudo systemctl daemon-reload
sudo systemctl enable openrouter-monitor
sudo systemctl start openrouter-monitor

# 4. Verificar estado
sudo systemctl status openrouter-monitor
```

<br />

<!-- ================================================================================== -->
<!-- SECTION: CONFIGURATION                                                              -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%E2%9A%99%20CONFIGURACION-FFD700?style=for-the-badge&labelColor=0D1117&color=FFD700" />
</div>

<br />

Todas las opciones se configuran mediante **variables de entorno** en `/etc/openrouter-monitor.env`:

| Variable | Default | Descripción |
|:--|:--:|:--|
| `DISCORD_WEBHOOK_URL` | **requerido** | URL del webhook de Discord |
| `POLL_INTERVAL_MINUTES` | `30m` | Intervalo entre consultas a OpenRouter |
| `DB_PATH` | `./data.db` | Ruta de la base de datos SQLite |
| `LOG_LEVEL` | `info` | Nivel de log: `debug`, `info`, `warn`, `error` |
| `HTTP_TIMEOUT_SECONDS` | `30s` | Timeout para requests HTTP |
| `OPENROUTER_BASE_URL` | `https://openrouter.ai/api/v1` | URL base de la API |
| `MAX_RETRIES` | `5` | Intentos máximos antes de fallar |
| `CIRCUIT_BREAKER_THRESHOLD` | `5` | Fallos consecutivos para abrir circuito |
| `CIRCUIT_BREAKER_TIMEOUT_MINUTES` | `60m` | Tiempo de espera del circuit breaker |
| `HEALTH_CHECK_PORT` | `:9090` | Puerto del health check HTTP |
| `ACTIVE_START_HOUR` | `9` | Hora de inicio de notificaciones (Peru) |
| `ACTIVE_END_HOUR` | `19` | Hora de fin de notificaciones (Peru) |

<br />

**Ejemplo completo de archivo de entorno:**

```env
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/abc123/def456
POLL_INTERVAL_MINUTES=30m
DB_PATH=/var/lib/orpdisc/data.db
LOG_LEVEL=info
HTTP_TIMEOUT_SECONDS=30s
MAX_RETRIES=5
CIRCUIT_BREAKER_THRESHOLD=5
CIRCUIT_BREAKER_TIMEOUT_MINUTES=60m
HEALTH_CHECK_PORT=:9090
ACTIVE_START_HOUR=9
ACTIVE_END_HOUR=19
```

<br />

<!-- ================================================================================== -->
<!-- SECTION: USAGE                                                                      -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%F0%9F%9A%80%20USO-FF00FF?style=for-the-badge&labelColor=0D1117&color=FF00FF" />
</div>

<br />

### ▶️ Comandos Básicos

```bash
# Iniciar servicio
sudo systemctl start openrouter-monitor

# Ver estado
sudo systemctl status openrouter-monitor

# Ver logs en tiempo real
sudo journalctl -u openrouter-monitor -f

# Detener
sudo systemctl stop openrouter-monitor

# Reiniciar
sudo systemctl restart openrouter-monitor
```

<br />

### 🩺 Health Check

```bash
# Verificar que el servicio responde
curl http://localhost:9090/health

# Respuesta esperada:
#   200 → "saludable"      (todo bien)
#   503 → "no saludable"   (base de datos caída)
```

<br />

<!-- ================================================================================== -->
<!-- SECTION: TROUBLESHOOTING                                                            -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%F0%9F%94%A7%20SOLUCION%20DE%20PROBLEMAS-FF4444?style=for-the-badge&labelColor=0D1117&color=FF4444" />
</div>

<br />

| Problema | Solución |
|:--|:--|
| ❌ El servicio no inicia | Verificar `DISCORD_WEBHOOK_URL`, permisos de DB y logs con `journalctl -u openrouter-monitor -n 50` |
| 🔇 No llegan notificaciones | Verificar webhook URL, canal de Discord, y horario activo (9:00-19:00 Peru) |
| 🔒 Circuit breaker abierto | Revisar estado de OpenRouter API, aumentar `CIRCUIT_BREAKER_TIMEOUT_MINUTES` |
| 🗄️ Base de datos bloqueada | Solo una instancia a la vez, verificar espacio en disco |
| 🚫 Health check 503 | `sqlite3 /path/data.db "SELECT 1;"` — verificar permisos y disco |
| ⏰ Notificaciones fuera de horario | Ajustar `ACTIVE_START_HOUR` y `ACTIVE_END_HOUR` en el archivo de entorno |

<br />

<!-- ================================================================================== -->
<!-- SECTION: ARCHITECTURE                                                               -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%F0%9F%8F%97%20ARQUITECTURA-00E5FF?style=for-the-badge&labelColor=0D1117&color=00E5FF" />
</div>

<br />

```
┌─────────────────────────────────────────────────────────┐
│                    cmd/monitor/main.go                    │
│                     Punto de entrada                       │
└────────────┬────────────────────────────────┬────────────┘
             │                                │
    ┌────────▼────────┐            ┌─────────▼─────────┐
    │  internal/service │            │  internal/config   │
    │   Orquestrador    │            │  Variables de      │
    │  ┌─────────────┐ │            │  entorno (env)     │
    │  │ Start()     │ │            └───────────────────┘
    │  │  ├─ poll()  │ │
    │  │  └─ maybe   │ │
    │  │     Send    │ │
    │  │     Digest()│ │
    │  └─────────────┘ │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐   ┌──────────────────┐   ┌──────────────────┐
    │ internal/openrouter│   │ internal/processor │   │ internal/discord   │
    │  FetchModels()    │──▶│  ProcessModels()  │──▶│  SendNotification()│
    │  Circuit Breaker  │   │  TopByCost()      │   │  BuildEmbeds()     │
    │  Exponential      │   │  TopByContext()   │   │  Rate Limit Retry  │
    │  Backoff Retry    │   │  TopByNewest()    │   │  Color Coding      │
    └───────────────────┘   └────────┬──────────┘   └──────────────────┘
                                     │
                            ┌────────▼──────────┐
                            │ internal/database   │
                            │  SQLite (modernc)   │
                            │  CGO-free           │
                            │  Models, Price      │
                            │  History, Notifs    │
                            └───────────────────┘
```

<br />

<!-- ================================================================================== -->
<!-- SECTION: LICENSE                                                                    -->
<!-- ================================================================================== -->

<div align="center">
  <img src="https://img.shields.io/badge/%F0%9F%93%84%20LICENCIA-00E676?style=for-the-badge&labelColor=0D1117&color=00E676" />
</div>

<br />

<p align="center">
  Distribuido bajo licencia <b>MIT</b>. Consulta el archivo <a href="https://github.com/Yokonad/orpdisc/blob/main/LICENSE"><code>LICENSE</code></a> para más detalles.
</p>

<br />

<!-- ================================================================================== -->
<!-- FOOTER                                                                              -->
<!-- ================================================================================== -->

<div align="center">
  <hr style="border: 1px solid #00E5FF; width: 80%; opacity: 0.3;" />

  <br />

  <img src="https://img.shields.io/badge/Made%20with%20%E2%9D%A4%EF%B8%8F%20in%20Peru-0D1117?style=for-the-badge&logo=go&logoColor=00E5FF&labelColor=0D1117&color=0D1117" />

  <br />

  <sub>_Built with Go, SQLite & Discord Webhooks_</sub>
</div>

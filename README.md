# orpdisc

Servicio automatizado que consulta la API de OpenRouter para obtener precios/capacidades de modelos y envía notificaciones de webhook de Discord para el mejor modelo.

## Características

- **Consulta Automatizada**: Consulta la API de OpenRouter cada 4 horas (configurable)
- **Detección de Cambios**: Detección basada en hash para nuevos modelos, cambios de precios y modelos eliminados
- **Notificaciones de Discord**: Notificaciones enrichidas con colores dinámicos
- **Circuit Breaker**: Patrón de resiliencia para prevenir agotamiento de la API
- **Verificaciones de Salud**: Endpoint HTTP opcional para verificaciones de salud

## Instalación

### Requisitos Previos

- Go 1.21+
- SQLite3
- URL de webhook de Discord

### Desde el Código Fuente

```bash
git clone https://github.com/Yokonad/orpdisc.git
cd orpdisc
go build -o monitor ./cmd/monitor
```

### Desde Docker

```bash
docker build -t openrouter-monitor .
docker run -d \
  --name openrouter-monitor \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -v /path/to/data:/data \
  openrouter-monitor
```

### Desde systemd (Recomendado para VPS)

```bash
# Copiar binario
sudo cp monitor /opt/monitor/

# Copiar archivo de servicio
sudo cp openrouter-monitor.service /etc/systemd/system/

# Crear archivo de entorno
sudo nano /etc/openrouter-monitor.env
# Agregar: DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...

# Crear usuario
sudo useradd -r -m -d /var/lib/monitor -s /bin/false monitor

# Establecer permisos
sudo chown -R monitor:monitor /var/lib/monitor

# Recargar systemd e iniciar
sudo systemctl daemon-reload
sudo systemctl enable openrouter-monitor
sudo systemctl start openrouter-monitor
```

## Configuración

El servicio se configura mediante variables de entorno:

| Variable | Valor por defecto | Descripción |
|----------|---------|-------------|
| `DISCORD_WEBHOOK_URL` | **requerido** | URL del webhook de Discord |
| `POLL_INTERVAL_MINUTES` | `240` | Minutos entre consultas |
| `DB_PATH` | `./data.db` | Ruta del archivo de base de datos SQLite |
| `LOG_LEVEL` | `info` | Nivel de log: debug, info, warn, error |
| `HTTP_TIMEOUT_SECONDS` | `30` | Tiempo de espera para solicitudes HTTP |
| `OPENROUTER_BASE_URL` | `https://openrouter.ai/api/v1` | URL base de la API de OpenRouter |
| `MAX_RETRIES` | `5` | Intentos máximos para solicitudes fallidas |
| `CIRCUIT_BREAKER_THRESHOLD` | `5` | Número de fallos consecutivos antes de abrir el circuito |
| `CIRCUIT_BREAKER_TIMEOUT_MINUTES` | `60` | Minutos de espera antes de reintentar después de abrir el circuito |
| `HEALTH_CHECK_PORT` | `:8080` | Puerto para el servidor HTTP de verificación de salud (vacío para desactivar) |

### Ejemplo de Archivo de Entorno

```
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/1498708885681209364/R2dWL1LoGb3jINU0OuHWm-bgM6d_P4s39w0upvoUY3kOhy0elTv2ZcwNe4uHKqNJj8nd
POLL_INTERVAL_MINUTES=30
DB_PATH=/var/lib/monitor/data.db
LOG_LEVEL=info
CIRCUIT_BREAKER_THRESHOLD=5
CIRCUIT_BREAKER_TIMEOUT_MINUTES=60
HEALTH_CHECK_PORT=:8080
```

## Uso

### Ejecutar desde Línea de Comando

```bash
# Con variables de entorno
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... ./monitor

# Con archivo de entorno
./monitor  # lee desde archivo .env si existe
```

### Docker

```bash
docker run -d \
  --name openrouter-monitor \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -e POLL_INTERVAL_MINUTES=30 \
  -v /path/to/data:/data \
  openrouter-monitor
```

### systemd

```bash
# Iniciar el servicio
sudo systemctl start openrouter-monitor

# Verificar estado
sudo systemctl status openrouter-monitor

# Ver logs
sudo journalctl -u openrouter-monitor -f

# Detener el servicio
sudo systemctl stop openrouter-monitor

# Reiniciar el servicio
sudo systemctl restart openrouter-monitor
```

## Verificaciones de Salud

El servicio expone opcionalmente un endpoint HTTP para verificaciones de salud:

```bash
# Establecer HEALTH_CHECK_PORT para activar
HEALTH_CHECK_PORT=:8080 ./monitor

# Verificar salud
curl http://localhost:8080/health

# Respuesta:
# - 200 OK: "saludable" cuando la base de datos es accesible
# - 503 Service Unavailable: "no saludable: <error>" cuando la base de datos está caída
```

## Notificaciones de Discord

El monitor envía embeds enriquecidos a Discord:

- **🆕 Nuevos Modelos** (verde): Cuando aparecen nuevos modelos en OpenRouter
- **📝 Actualizaciones de Modelos** (amarillo): Cuando cambian precios o contexto
- **🗑️ Modelos Ya No Disponibles** (rojo): Cuando se eliminan modelos

## Solución de Problemas

### El servicio no inicia

1. Verificar que `DISCORD_WEBHOOK_URL` esté configurado correctamente
2. Verificar que la ruta de la base de datos tenga permisos de escritura
3. Revisar logs: `journalctl -u openrouter-monitor -n 50`

### No se envían notificaciones

1. Verificar que la URL del webhook sea válida
2. Verificar que el canal de Discord no haya sido eliminado
3. Revisar logs en busca de errores de API
4. Verificar que los modelos realmente hayan cambiado (la primera ejecución llena la DB sin notificaciones)

### El circuit breaker sigue abriéndose

1. Verificar el estado de la API de OpenRouter
2. Aumentar `CIRCUIT_BREAKER_TIMEOUT_MINUTES`
3. Aumentar `MAX_RETRIES` para problemas temporales

### Errores de base de datos bloqueada

1. Asegurarse de que solo una instancia esté ejecutándose
2. Verificar espacio en disco
3. Para alta concurrencia, considerar cambiar a PostgreSQL

### La verificación de salud retorna 503

1. Verificar conectividad de la base de datos: `sqlite3 /path/to/data.db "SELECT 1;"`
2. Verificar que el disco no esté lleno
3. Verificar permisos de archivos

## Licencia

MIT

# Guía exhaustiva de desarrollo de drivers (para desarrolladores)

Este documento está enfocado a **desarrolladores** que van a construir drivers de Netsocs en Go usando `github.com/Netsocs-Team/driver.sdk_go`.

## Modelo mental: qué hace un driver

Un driver es un “bridge” entre:

- **Plataforma Netsocs** (DriverHub) — acciones/configuración/eventos/estado
- **Dispositivo / Sistema** — API HTTP, RTSP, SDK, Cloud, colas (SQS), Webhooks, etc.

En términos del SDK, tu driver implementa:

- **Configuration handlers**: endpoints lógicos invocados por la plataforma (WebSocket).
- **Objects**: representación de entidades en Netsocs (sensores, cámaras, puertas, paneles…).
- **State updates**: cambios de estado y atributos de estado, persistidos en la plataforma.
- **Events**: notificación de eventos relevantes.

## Arquitectura recomendada

Separar en capas evita drivers “monolíticos” difíciles de mantener:

- **`devices/`**: integración con fabricante (HTTP/SDK/colas).
- **`config/`**: handlers (validación, orquestación, respuesta a plataforma).
- **`objects/`**: construcción/acciones/loops por objeto.
- **`main.go`**: wiring + registro de handlers + inicio del loop.

## `driver.netsocs.json`: contrato de runtime

El SDK inicializa el cliente leyendo `driver.netsocs.json`.

Buenas prácticas:

- Mantén `driver.netsocs.json.example` con placeholders y **commitéalo**.
- `driver.netsocs.json` real **no** se commitea.
- Declara `settings_available` de forma coherente con los handlers que implementas.

## Ciclo de vida típico

### 1) Startup

- `client.New()` → autentica y prepara canales de comunicación.
- `AddConfigHandler(...)` → registra handlers.
- `ListenConfig()` → loop bloqueante.

### 2) Configuración de un “device”

La plataforma dispara handlers típicos:

- `ACTION_PING_DEVICE`: validar reachability/credenciales.
- `GET_EXTRA_DEVICE_FIELDS`: si necesitas API keys u otros parámetros.
- `REQUEST_CREATE_OBJECTS`: crear y registrar objetos para ese device.

### 3) Operación

En background:

- polling periódico (si el dispositivo no “pushea”)
- consumo de eventos (colas/cloud/webhooks)
- actualización de estados y dispatch de eventos

## Handlers: reglas de oro

### 1) Validación

Antes de pegarle al dispositivo:

- valida `DeviceData` (IP, Port, Username, Password, IsSSL)
- valida `ExtraFields` (si aplica)
- valida `msg.Value` si trae JSON

### 2) Timeouts

Todo acceso a dispositivo debe tener timeout (TCP/HTTP/SDK/colas).

### 3) Errores accionables

El error debe guiar al usuario:

- “credenciales inválidas”
- “no responde en \(ip:port\)”
- “campo extra faltante: X”

### 4) DeviceState

En drivers reales suele ser clave actualizar estados:

- `AuthenticationFailure` cuando detectas 401/invalid credentials.
- `ConfigurationFailure` cuando faltan parámetros.
- `Online` cuando ya está operativo.
- `DuplicatedDevice` si detectas un “origen” duplicado.

> El objetivo es que el usuario entienda rápidamente por qué el driver no opera.

## `REQUEST_CREATE_OBJECTS`: el handler más importante

Es el punto donde el driver “materializa” la integración:

- consulta inventario del dispositivo (canales/puertas/sensores)
- registra objetos con `sdkClient.RegisterObject(obj)`
- registra event types con `sdkClient.AddEventTypes(...)` (si corresponde)
- arranca goroutines de operación (escucha de eventos, polling, health loops)

Checklist:

- **IDs estables**: `ObjectID` debe ser determinístico para evitar duplicados entre reinicios.
- **Orden de creación**: primero objetos “padre” (si los hay), luego “hijos”.
- **Logs**: escribe logs por etapa (cuántos objetos/eventos se registraron).
- **Concurrencia**: no lances 100 goroutines por canal sin control; usa pools si aplica.

## ExtraFields: UX de configuración

Si tu driver necesita campos adicionales (API keys, región, cola, etc.):

- implementa `GET_EXTRA_DEVICE_FIELDS` devolviendo definiciones con:
  - `Name` (string exacto; luego lo leerás en `DeviceData.Extrafields`)
  - `Description` (texto claro para usuario)
  - `Type` (string/number/bool según SDK)

Y en `REQUEST_CREATE_OBJECTS`:

- valida que existan
- si faltan, marca `ConfigurationFailure` y devuelve error

## Objects: cómo decidir qué crear

Regla práctica: cada “cosa” que un usuario quiere ver/controlar en la UI debe ser un **objeto**.

Ejemplos:

- NVR: un objeto por canal (`VideoChannelObject`)
- Acceso: un `ReaderObject` y uno o más `Door/Lock` según el modelo
- Alarmas: un `AlarmPanelObject` y `SensorObject` por zona

### Metadata consistente

- `ObjectID`: único, estable, sin espacios.
- `Name`: amigable.
- `Domain`: consistente (camera, alarm, access_control…).
- `DeviceID`: asocia el objeto al device.
- `ParentID`: si es jerárquico.
- `Tags`: filtros útiles.

## Estados vs atributos de estado

La plataforma diferencia:

- **State**: estado principal (enum conocido por el objeto).
- **StateAttributes**: mapa flexible key/value para datos de contexto.

Ejemplo:

- State: `sensor.state.measurement`
- Attributes: `value=23.5`, `unit=°C`, `battery=85`

## Events: diseño y contrato

Los event types se registran una vez (por driver) y luego se “disparan” con data.

Buenas prácticas:

- `EventType`: estable y versionable (no lo cambies sin migración).
- `DisplayName/Description`: orientado a usuario.
- `EventLevel`: info/warn/error según semántica.
- `Properties`: llaves consistentes (ej. `user_id`, `door_id`, `confidence`, etc.).

## Integraciones por categoría (patrones probados)

### Video (Cámaras/NVR)

- `GET_CHANNELS` para inventario.
- Objetos `VideoChannelObject` por canal.
- Implementa snapshot/playback/PTZ según soporte del fabricante.
- Mantén un “health loop” para marcar offline/online por canal o por device.

### Alarmas

- inventario (particiones/zonas)
- eventos (alarma activada, bypass, tamper…)
- acciones (arm/disarm/bypass)

### Acceso (biométrico / lectores / controladores)

- objetos para lectores/puertas
- eventos de acceso (granted/denied) con propiedades (credencial/usuario/razón)
- enrolamiento (cara/tarjeta/qr) via handlers dedicados

### Cloud / colas (ej. SQS)

Si la fuente de eventos es externa al dispositivo:

- la “key” de deduplicación suele ser el nombre/ARN de la cola o endpoint.
- valida credenciales cloud como extra fields.
- asegúrate de poder “replay” / reintentar sin duplicar eventos.

## Testing

Mínimos recomendados:

- tests de parsing de payload JSON de handlers
- tests de validación de extra fields
- tests de formatter de eventos (si conviertes payloads crudos)
- tests unitarios del “client wrapper” si es mockeable

Comandos:

```bash
go test ./...
```

## Build/Release (recomendación)

Aunque cada equipo puede variar, un release reproducible debería incluir:

- `-ldflags` para versión (o setearla en runtime con `SetDriverVersion`)
- binarios por OS/arch si distribuyes fuera de Docker
- artefactos con nombres consistentes con `driver_binary_filename`
- documentación asociada (`documentation_url`) y changelog

## Seguridad (no negociable)

- No commitear `driver.netsocs.json`
- Sanitizar logs al compartirlos (no imprimir tokens/keys)
- No exponer credenciales en errores devueltos a plataforma


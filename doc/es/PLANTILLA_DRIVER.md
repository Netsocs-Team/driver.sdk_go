# Plantilla de driver (template) — estructura recomendada

Este documento es una **plantilla práctica** para que desarrolladores creen drivers consistentes dentro del ecosistema Netsocs.

> Nota: el SDK ya incluye un template en `SDK/doc/template/`. Este documento explica **cómo usarlo** y qué piezas debes completar para un driver real.

## Estructura mínima del proyecto

```
my-driver/
├── main.go
├── go.mod
├── go.sum
├── driver.netsocs.json              # NO se commitea (credenciales)
├── driver.netsocs.json.example      # sí se commitea (placeholders)
├── .gitignore
├── config/
│   └── handlers.go                  # config handlers (acciones de la plataforma)
├── devices/
│   └── device_manager.go            # pool/conexiones + wrapper de API del fabricante
└── objects/
    ├── <objetos>.go                 # wrappers para objetos Netsocs
    └── ...
```

## `driver.netsocs.json.example` (obligatorio)

Incluye placeholders y declara `settings_available`.

- **Regla de oro**: `driver.netsocs.json` real se genera copiando el `.example`.

## `main.go` (responsabilidades)

En `main.go` se hace el “boot” del driver:

- Crear logger
- `sdkClient, err := client.New()`
- (Opcional) `sdkClient.SetDriverVersion(...)` y `sdkClient.SetDriverDocumentation(...)`
- Registrar handlers: `sdkClient.AddConfigHandler(...)`
- Entrar al loop: `sdkClient.ListenConfig()` (bloqueante)

### Pattern recomendado: handlers como lista/mapa

Define una tabla de handlers, y registra en un loop (más mantenible que llamadas sueltas).

## `config/handlers.go` (qué debe vivir acá)

Handlers típicos por integración:

- **Base**
  - `ACTION_PING_DEVICE`
  - `REQUEST_CREATE_OBJECTS`
  - `GET_EXTRA_DEVICE_FIELDS` (si necesitas extra fields)
- **Video**
  - `GET_CHANNELS`
  - `GET_RECORDING_RANGES`, playback, snapshots, PTZ (según features)
- **Acceso**
  - handlers para personas/credenciales (card/face/qr), logs, puertas
- **Alarmas**
  - particiones, zonas, arm/disarm, bypass

Recomendación: cada handler debe:

- Validar input (incluye `DeviceData` + `msg.Value` si trae payload JSON)
- Operar con timeouts
- Reportar errores con mensajes accionables
- Actualizar `DeviceState` cuando corresponda (auth/config/online)

## `devices/device_manager.go` (cómo diseñarlo)

Objetivo: que tu código de negocio no dependa del transporte.

- **DeviceManager**: cachea conexiones por \(ip:port\) o por un identificador de “origen” (ej. una cola).
- **Device**: wrapper con métodos semánticos (Ping, GetChannels, OpenDoor, GetEvents…).

Checklist:

- Mutex para concurrencia
- reconexión / invalidación (Remove)
- Cleanup al shutdown (si implementas señales)

## `objects/` (qué poner aquí)

Aunque puedes registrar objetos directamente desde handlers, suele ser mejor encapsular:

- Builders de objetos (por tipo)
- `SetupFn` para inicialización + loops
- Acciones (`RunAction`) cuando el objeto controla algo (switch/lock/ptz)

## Guía rápida: crear un driver nuevo copiando el template

1) Copia `SDK/doc/template` a una carpeta nueva (por ejemplo `driver.mi_driver`).

2) Cambia:

- `go.mod` (`module ...`)
- imports `your-module-name/...`
- `driver.netsocs.json.example` (nombre/versión/handlers declarados)

3) Implementa:

- Cliente del fabricante en `devices/`
- Lógica en handlers de `config/`
- Creación/registro de objetos en `REQUEST_CREATE_OBJECTS`

4) Prueba:

```bash
go test ./...
go run .
```

## Convenciones recomendadas (para consistencia)

- **ObjectID**: estable y único; evita números “sueltos”. Ej.: `nvr_<deviceId>_ch_<n>`, `door_<id>`, `reader_<id>`.
- **Domain**: consistente con el tipo funcional (camera, access_control, alarm, temperature…).
- **Eventos**: `EventType` estable; `DisplayName` humano; `Origin` “driver”.
- **Versioning**: SemVer (`1.2.3`) y propaga a `sdkClient.SetDriverVersion`.


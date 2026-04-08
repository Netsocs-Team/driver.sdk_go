# Uso del SDK y de los Drivers (Guía para usuarios)

Este documento está pensado para **usuarios/integradores** que van a **instalar y operar** drivers de Netsocs (por ejemplo: `driver.dahua_nvr`, `driver.ajax`, `driver.controlid`).

## Qué es un “driver” en Netsocs

Un driver es un binario (Go) que:

- Se autentica contra **DriverHub** con credenciales del sitio.
- Recibe **solicitudes** desde la plataforma (WebSocket) mediante **configuration handlers**.
- Registra **objetos** (cámaras, sensores, puertas, paneles, etc.) y actualiza sus **estados**.
- Publica **eventos** (por ejemplo, acceso concedido/denegado, motion, alarmas) con propiedades y, opcionalmente, media.

## Requisitos

- **Go** (solo si compilarás desde código; si ya tienes el binario, no hace falta).
- Acceso a la plataforma Netsocs para obtener:
  - **Driver Key**
  - **DriverHub Host**
  - **Token/Site Token** (si aplica en tu despliegue)
  - **Driver ID** / **Site ID** (si aplica en tu despliegue)

## Archivo de configuración: `driver.netsocs.json`

El driver lee un archivo `driver.netsocs.json` desde el **directorio de trabajo** (normalmente, la carpeta del binario).

### Ejemplo (plantilla)

Usa placeholders. **Nunca** pegues credenciales reales en documentación ni las subas a Git.

```json
{
  "driver_key": "YOUR_DRIVER_KEY_HERE",
  "driver_hub_host": "https://<host>/api/netsocs/dh",
  "token": "YOUR_AUTH_TOKEN",
  "driver_id": "YOUR_DRIVER_ID",
  "site_id": "YOUR_SITE_ID",
  "name": "Nombre visible del driver",
  "version": "1.0.0",
  "driver_binary_filename": "driver.mi_driver",
  "documentation_url": "https://<repo o pdf>",
  "settings_available": [
    "actionPingDevice",
    "getExtraDeviceFields",
    "requestCreateObjects"
  ],
  "log_level": "info",
  "device_models_supported_all": true,
  "device_firmwares_supported_all": true
}
```

### Campos clave (cómo leerlos como usuario)

- **`driver_key`**: clave de autenticación del driver.
- **`driver_hub_host`**: endpoint HTTP del DriverHub.
- **`settings_available`**: lista de capacidades/acciones que el driver declara (la UI suele usar esto para mostrar opciones).
- **`documentation_url`**: donde la UI o el soporte debería encontrar la guía del driver.
- **`log_level`**: nivel de log (si el driver lo respeta).

## Ejecutar un driver

### Desde binario

- Copia el binario y `driver.netsocs.json` en la misma carpeta.
- Ejecuta el binario desde esa carpeta (para que encuentre el JSON).

En Windows (PowerShell):

```powershell
cd D:\ruta\al\driver\
.\driver.mi_driver.exe
```

En Linux:

```bash
cd /opt/driver/
chmod +x ./driver.mi_driver
./driver.mi_driver
```

### Desde código (compilando)

```bash
go mod download
go build -o driver.mi_driver
./driver.mi_driver
```

## Flujo típico en la plataforma (qué esperar)

### 1) Probar conexión (“Ping device”)

La plataforma dispara el handler **`ACTION_PING_DEVICE`**. Si falla:

- Revisa IP/puerto/SSL del dispositivo
- Usuario/clave
- firewall/rutas/ACL
- logs del driver

### 2) Crear objetos (“Create objects”)

La plataforma dispara **`REQUEST_CREATE_OBJECTS`**. El driver:

- Consulta el dispositivo (canales, puertas, sensores, etc.)
- Registra objetos (aparecen en UI)
- Configura event types (si aplica)
- Entra en “operación” (polling, escucha de eventos, etc.)

### 3) Operación

Según el tipo de integración:

- **Video (NVR/Cámaras)**: se crean `VideoChannelObject` (canales), snapshots, playback, PTZ.
- **Alarmas**: paneles, zonas, particiones; eventos de alarmas.
- **Acceso (ControlID)**: lectores/puertas; eventos de acceso; enrolamiento facial/credenciales.

## Campos extra del dispositivo (ExtraFields)

Algunos drivers requieren **campos extra** además de IP/puerto/usuario/clave.

Ejemplo típico (patrón real visto en drivers existentes):

- El driver implementa el handler **`GET_EXTRA_DEVICE_FIELDS`** para que la UI sepa **qué pedirle** al usuario.
- En `REQUEST_CREATE_OBJECTS`, el driver valida que esos campos existan y, si faltan, marca el device como “ConfigurationFailure”.

Como usuario:

- Si en UI aparece un formulario con “campos extra”, **rellénalos** (API keys, credenciales cloud, nombres de cola, etc.)
- Si el driver reporta “missing field …”, revisa exactamente el nombre del campo; suelen ser sensibles a mayúsculas/espacios.

## Estados del dispositivo (DeviceState) y resolución de fallas

Los drivers suelen informar estados como:

- **Online**: operativo.
- **ConfigurationFailure**: configuración incompleta o inválida (faltan extra fields, parámetros incorrectos, etc.).
- **AuthenticationFailure**: credenciales del dispositivo/API inválidas.
- **DuplicatedDevice**: el driver detecta instancia duplicada con claves equivalentes (por ejemplo, misma cola/event source).

Acciones recomendadas:

- **ConfigurationFailure**: completar/ajustar campos extra; reintentar “Create objects”.
- **AuthenticationFailure**: corregir credenciales y reintentar.
- **DuplicatedDevice**: evita registrar el mismo “origen” dos veces (y/o ajusta la llave de deduplicación).

## Logs: qué mirar

Buenas señales:

- “Driver started” (con versión)
- “Client initialized”
- “objects created …”
- “DeviceStateOnline”

Señales de problema:

- “failed to create client” (no lee `driver.netsocs.json`, JSON inválido, credenciales platform)
- “failed to register object” (IDs duplicados, error plataforma)
- timeouts a dispositivo/API

## Seguridad

- **Nunca subas** `driver.netsocs.json` con credenciales reales a Git.
- Para soporte, comparte **logs** y **config sanitizada** (sin claves/tokens).


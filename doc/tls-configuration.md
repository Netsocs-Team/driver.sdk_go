# HTTPS / SSL en `driver_hub_host`

Cuando `driver_hub_host` apunta a un endpoint HTTPS cuyo certificado está emitido por
una CA privada o corporativa (lo habitual en instalaciones on-premise), Go solo
confía en el almacén de raíces del sistema operativo, y las peticiones del SDK
fallaban con:

```
Put "https://nutresa.netsocs.com/api/netsocs/dh/devices/states/8": tls: failed to verify certificate: x509: certificate signed by unknown authority
Post "https://nutresa.netsocs.com/api/netsocs/dh/devices/audit-logs": tls: failed to verify certificate: x509: certificate signed by unknown authority
```

Los websockets del SDK ya aceptaban esos certificados; el resto de peticiones no.

## Solución

Todo el tráfico saliente del SDK pasa ahora por un único transport construido en
[`pkg/httpx`](../pkg/httpx), con la misma configuración TLS que ya usaban los
websockets. **No hay nada que configurar: funciona por defecto.**

| Área | Ficheros |
| --- | --- |
| Estados de dispositivo / audit logs | `pkg/client/device.go`, `pkg/client/device_logger.go` |
| Objetos, eventos, grupos, usuarios, licencia, streams | `pkg/client/*.go`, `pkg/objects/object_controller.go` |
| Websocket de configuración (`ws/v1/config_communication`) | `pkg/config/main.go` |
| Websocket de objetos, audio de micrófono/altavoz | `pkg/objects/*.go` |
| Subida de ficheros y snapshots | `pkg/tools/upload_file.go`, `pkg/tools/upload_snapshot.go` |
| Data plane de control de acceso | `pkg/accessctl/dataplane.go` |
| Dispatchers de eventos legacy | `pkg/event/*.go` |

## Peticiones propias con la misma configuración

Si un driver necesita hablar con el DriverHub (o con un dispositivo por HTTPS)
por su cuenta:

```go
import "github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"

res, err := httpx.Client().Get(url)          // *http.Client compartido
res, err := httpx.Resty().R().Get(url)       // cliente resty
c, _, err := httpx.WebsocketDialer().Dial(url, header)
cli := httpx.NewClient(10 * time.Second)     // cliente con timeout
```

`httpx.CloneTransport()` devuelve una copia privada del transport para casos que
necesiten personalizarlo sin afectar al resto del SDK.

## Verificar la central (`ca_mode`, desde v0.8.0-rc.2)

Lo anterior no verifica el certificado. Cuando SmartConnect registra el driver
con una central que publica su CA, el bloque `auth` de `driver.netsocs.json`
dice cómo verificarla:

```json
"auth": { "client_id": "drv-…", "token_url": "https://…", "key_file": "driver.key",
          "ca_file": "central-ca.pem", "ca_mode": "pinned" }
```

| `ca_mode` | Conexiones al DriverHub y a `token_url` |
| --- | --- |
| `pinned` | Solo se confía en la CA raíz de `ca_file` (PEM, relativo a `driver.netsocs.json`). Se verifican la cadena y el nombre del host (DNS o IP). Si la CA es la de Caddy de la propia central ("Caddy Local Authority"), solo la cadena: esa CA es única de la instalación y sus llaves no salen de Caddy, y en instalaciones locales usadas por IP Caddy pone en el certificado la IP del contenedor |
| `public` | Raíces del sistema, verificación normal |
| ausente | Sin verificar, igual que v0.8.0-rc.1 (centrales antiguas); `ca_file` se ignora |

Con `ca_mode`, `driver_hub_host` y `token_url` deben ser `https://` y el SDK
rechaza cualquier `http://`/`ws://` a esos hosts. Si `ca_mode` es desconocido,
falta `ca_file` en `pinned` o no contiene un certificado PEM, el driver sale
con el código 78 antes de hacer ninguna petición: nunca vuelve al modo sin
verificar. Un certificado que no valida no es fatal: el SDK reintenta con
espera creciente y lo registra en el log.

Solo cambian las conexiones a los hosts de la central. El resto (cámaras y
otros dispositivos por HTTPS con `httpx.Client()`) sigue igual. `httpx.Client()`,
`httpx.NewClient()`, `httpx.Resty()`, `httpx.RoundTripper()` y
`httpx.DialWebsocket()` aplican la política completa. Desde v0.8.0-rc.3,
`httpx.Transport()`, `httpx.CloneTransport()` y `httpx.WebsocketDialer()`
también verifican el TLS de los hosts de la central (lo que cambie el driver
en su copia solo vale para los demás hosts), pero no rechazan `http://`/`ws://`
ni ponen credenciales, y detrás de un proxy HTTP usan su propia configuración.
Para hablar con la central sin el token del driver (por ejemplo WHIP, que
lleva su propio token de publicación) use `httpx.RoundTripper()`.

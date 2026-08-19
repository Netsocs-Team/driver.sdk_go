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

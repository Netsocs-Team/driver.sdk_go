# Actions Payload Reference

Payloads JSON esperados por cada action de cada objeto del SDK.  
Los campos marcados con `*` son opcionales (`omitempty` en la struct).

> **Acciones custom:** además de las actions predefinidas listadas aquí, un driver puede
> registrar acciones propias en cualquier objeto con `obj.RegisterCustomAction(name, handler)`.
> El payload de una acción custom es JSON arbitrario definido por el driver — el handler recibe
> los bytes crudos (`CustomActionContext.Payload`) y los deserializa a su propia struct. Ver
> `pkg/objects/custom_actions.go` y la sección "Custom Actions" en `doc/quick-start/03-understanding-objects.md`.

---

## Switch

Archivo: `pkg/objects/switch.go`

### `switch.action.turn_on`
No requiere payload (el campo es ignorado).
```json
{}
```

### `switch.action.turn_off`
No requiere payload (el campo es ignorado).
```json
{}
```

---

## Door

Archivo: `pkg/objects/door.go`

### `door.action.open`
No requiere payload (el campo es ignorado).
```json
{}
```

### `door.action.close`
No requiere payload (el campo es ignorado).
```json
{}
```

---

## Lock

Archivo: `pkg/objects/lock.go`

### `lock`
No requiere payload (el campo es ignorado).
```json
{}
```

### `unlock`
No requiere payload (el campo es ignorado).
```json
{}
```

### `reboot`
No requiere payload (el campo es ignorado).
```json
{}
```

---

## Alarm Panel

Archivo: `pkg/objects/alarm_panel.go`

Todas las actions de este objeto comparten la misma struct `actionPayload`. Cada action usa solo los campos que necesita.

### `alarm_panel.action.arm`
Usa `arm_mode` y `code`.
```json
{
  "arm_mode": "away",
  "code": "1234"
}
```
Valores válidos de `arm_mode`: `away`, `stay`, `stay_no_entry_delay`, `away_no_entry_delay`, `night`, `interior`.

### `alarm_panel.action.disarm`
Usa `code`.
```json
{
  "code": "1234"
}
```

### `alarm_panel.action.fire`
Usa `code`.
```json
{
  "code": "1234"
}
```

### `alarm_panel.action.panic`
Usa `code`.
```json
{
  "code": "1234"
}
```

### `alarm_panel.action.auxiliary`
Usa `code`.
```json
{
  "code": "1234"
}
```

### `alarm.action.bypass`
Usa `code` y `zone`.
```json
{
  "code": "1234",
  "zone": "zone-001"
}
```

### `alarm.action.bypass_rest`
Usa `code` y `zone`.
```json
{
  "code": "1234",
  "zone": "zone-001"
}
```

### `alarm.action.restore_alarm`
Usa `code`.
```json
{
  "code": "1234"
}
```

---

## Sensor

Archivo: `pkg/objects/sensor.go`

### `bypass`
No requiere payload (struct vacía, el campo no es deserializado).
```json
{}
```

### `unbypass`
No requiere payload (struct vacía, el campo no es deserializado).
```json
{}
```

### `custom`
No requiere payload (struct vacía en la llamada actual, aunque la struct tiene campos).
```json
{}
```

---

## Octopus

Archivo: `pkg/objects/octopus.go`

### `octopus.action.turn_on`
```json
{
  "relay_id": "relay-1"
}
```

### `octopus.action.turn_off`
```json
{
  "relay_id": "relay-1"
}
```

---

## Microphone

Archivo: `pkg/objects/microphone.go`

### `microphone.action.start_stream`
```json
{
  "session_id": "sess-abc123"
}
```

### `microphone.action.stop_stream`
```json
{
  "session_id": "sess-abc123"
}
```

---

## Speaker

Archivo: `pkg/objects/speaker.go`

### `speaker.action.start_talkback`
```json
{
  "session_id": "sess-abc123"
}
```

### `speaker.action.stop_talkback`
```json
{
  "session_id": "sess-abc123"
}
```

### `speaker.action.play_audio_clip`
`timeout` en segundos; `0` = sin límite de tiempo.
```json
{
  "url": "https://example.com/clip.wav",
  "timeout": 30
}
```

---

## Reader

Archivo: `pkg/objects/reader.go`

### `read`
`type` debe ser uno de los tipos soportados por el dispositivo.  
`timeout` en segundos.
```json
{
  "type": "face",
  "timeout": 15
}
```
Valores válidos de `type`: `face`, `normal_card`, `fingerprint_iso_19794_2`, `fingerprint_ansi_378_2004`.

### `reader.action.stop`
No requiere payload (no hay handler implementado en RunAction).
```json
{}
```

### `reader.action.reset`
No requiere payload (no hay handler implementado en RunAction).
```json
{}
```

### `reader.action.restart`
No requiere payload (no hay handler implementado en RunAction).
```json
{}
```

### `reader.action.store_qrs`
```json
{
  "person_id": "person-001",
  "name": "Juan Pérez",
  "values": ["QR_VALUE_1", "QR_VALUE_2"]
}
```

### `reader.action.delete_qrs`
```json
{
  "person_id": "person-001",
  "name": "Juan Pérez",
  "values": ["QR_VALUE_1"]
}
```

### `reader.action.delete_person`
```json
{
  "person_id": "person-001"
}
```

### `get_people`
No requiere payload (no hay datos de entrada, retorna la lista de personas).
```json
{}
```

### `set_people`
Reemplaza la base de personas completa en el dispositivo.  
`support_schedule`: si el dispositivo soporta horarios por persona.
```json
{
  "people": [
    {
      "person_id": "person-001",
      "name": "Juan Pérez",
      "credentials": [
        {
          "id": "cred-001",
          "type": "normal_card",
          "data": "AABBCCDD",
          "metadata": {
            "facility_code": "10"
          },
          "last_updated": "2024-01-15T10:30:00Z"
        }
      ]
    }
  ],
  "support_schedule": true,
  "schedule": [
    {
      "id": "sched-001",
      "last_updated": "2024-01-15T10:30:00Z",
      "monday":    { "start": "2024-01-15T08:00:00Z", "end": "2024-01-15T18:00:00Z", "enabled": true },
      "tuesday":   { "start": "2024-01-15T08:00:00Z", "end": "2024-01-15T18:00:00Z", "enabled": true },
      "wednesday": { "start": "2024-01-15T08:00:00Z", "end": "2024-01-15T18:00:00Z", "enabled": true },
      "thursday":  { "start": "2024-01-15T08:00:00Z", "end": "2024-01-15T18:00:00Z", "enabled": true },
      "friday":    { "start": "2024-01-15T08:00:00Z", "end": "2024-01-15T18:00:00Z", "enabled": true },
      "saturday":  { "start": "2024-01-15T08:00:00Z", "end": "2024-01-15T18:00:00Z", "enabled": false },
      "sunday":    { "start": "2024-01-15T08:00:00Z", "end": "2024-01-15T18:00:00Z", "enabled": false },
      "holidays": [
        { "date": "2024-12-25", "enabled": false }
      ]
    }
  ]
}
```

### `sync_access_database`
> **Importante:** DriversHub envuelve el payload real en un campo `data` como string JSON.  
> El wrapper que llega al driver es:
> ```json
> { "data": "<SyncAccessDatabasePayload serializado como string>" }
> ```
>
> El payload interno (`SyncAccessDatabasePayload`) tiene la siguiente estructura:

**Modo `full`** — reemplaza toda la base:
```json
{
  "mode": "full",
  "persons": [
    {
      "person_id": "person-001",
      "name": "Juan Pérez",
      "photo": "https://example.com/photo.jpg",
      "credentials": [
        {
          "type": "normal_card",
          "value": "AABBCCDD",
          "data": null
        },
        {
          "type": "face",
          "value": "",
          "data": "<base64-decoded biometric template>"
        }
      ],
      "bands": [
        {
          "weekdays": ["monday", "tuesday", "wednesday", "thursday", "friday"],
          "start_time": "08:00",
          "end_time": "18:00"
        }
      ],
      "holidays": ["2024-12-25", "2025-01-01"],
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2025-12-31T23:59:59Z",
      "enabled": true,
      "apb_exempt": false,
      "extended_unlock": false,
      "escort_required": false
    }
  ],
  "two_person_rule": false,
  "apb_area": {
    "area_id": "area-001",
    "name": "Acceso Principal",
    "mode": "soft",
    "direction": "entry"
  }
}
```

**Modo `incremental`** — solo cambios desde `since`:
```json
{
  "mode": "incremental",
  "since": "2024-06-01T00:00:00Z",
  "persons": [
    {
      "person_id": "person-002",
      "name": "María García",
      "credentials": [
        {
          "type": "normal_card",
          "value": "11223344"
        }
      ],
      "bands": [
        {
          "weekdays": ["monday", "friday"],
          "start_time": "09:00",
          "end_time": "17:00"
        }
      ],
      "holidays": [],
      "enabled": true
    }
  ],
  "deleted_ids": ["person-003", "person-004"]
}
```

Valores válidos:
- `mode`: `full` | `incremental`
- `apb_area.mode`: `none` | `soft` | `hard`
- `apb_area.direction`: `entry` | `exit` | `both`
- `bands.weekdays` items: `monday` | `tuesday` | `wednesday` | `thursday` | `friday` | `saturday` | `sunday`

---

## Video Channel

Archivo: `pkg/objects/video_channel.go`

### `video_channel.action.snapshot`
`snapshot_timestamp` vacío = captura inmediata. `filename` vacío = usa el timestamp como nombre.
```json
{
  "snapshot_timestamp": "2024-06-15T14:30:00Z",
  "resolution": "1920x1080",
  "filename": "cam1_snapshot"
}
```

### `video_channel.action.videoclip`
`timeout` en segundos; `0` = sin límite.
```json
{
  "start_timestamp": "2024-06-15T14:00:00Z",
  "end_timestamp": "2024-06-15T14:05:00Z",
  "resolution": "1920x1080",
  "timeout": 60
}
```

### `video_channel.action.ptz_control`
`value`: velocidad de 1 a 10.
```json
{
  "command": "up",
  "value": 5,
  "relative": false
}
```
Valores válidos de `command`: `up`, `down`, `left`, `right`, `up_left`, `up_right`, `down_left`, `down_right`, `zoom_in`, `zoom_out`, `stop`, `focus_near`, `focus_far`, `iris_open`, `iris_close`, `home`.

### `video_channel.action.ptz_goto_preset`
```json
{
  "token": "preset-token-001",
  "name": "Entrada Principal",
  "position": {
    "pan_tilt": {
      "x": 0.5,
      "y": 0.3,
      "space": "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace"
    },
    "zoom": {
      "x": 0.1,
      "space": "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace"
    }
  }
}
```

### `video_channel.action.ptz_get_status`
No requiere payload.
```json
{}
```

### `video_channel.action.seek`
```json
{
  "playback_id": "playback-abc123",
  "seek_to": "2024-06-15T14:02:30Z",
  "speed": 1.0,
  "reverse": false,
  "destroy": false,
  "video_engine_hostname": "video-engine.local",
  "video_engine_rtsp_port": "8554"
}
```

### `video_channel.action.get_recording_segments`
```json
{
  "start_time": "2024-06-15T00:00:00Z",
  "end_time": "2024-06-15T23:59:59Z"
}
```

### `video_channel.action.request_dolynk_stream_url`
```json
{
  "deviceId": "SN123456789",
  "devCode": "password123",
  "channelId": 1,
  "businessType": "Real",
  "encryptMode": 0,
  "protoType": "rtsp",
  "streamType": 0,
  "deviceType": "Channel",
  "assistStream": 0,
  "beginTime": "",
  "endTime": "",
  "recordPlayType": null,
  "recordFileName": ""
}
```
Valores válidos:
- `businessType`: `Real` | `talk` | `localRecord` | `cloudRecord`
- `encryptMode`: `0` (sin cifrado) | `1` (cifrado)
- `protoType`: `rtsp` | `rtsv`
- `streamType`: `0` (stream principal) | `1` (sub stream)
- `deviceType`: `Channel` | `device`
- `recordPlayType`: `0` (por tiempo) | `1` (por nombre de archivo)

### `video_channel.action.request_dahua_playback_media_files`
```json
{
  "start_time": "2024-06-15T00:00:00Z",
  "end_time": "2024-06-15T23:59:59Z"
}
```

### `publish_stream_start`
`quality` vacío o `"main"` inicia stream principal (101); `"sub"` inicia sub stream (102).
```json
{
  "quality": "main"
}
```

### `publish_stream_stop`
`quality` vacío o `"main"` detiene stream principal; `"sub"` detiene sub stream.
```json
{
  "quality": "main"
}
```

### `video_channel.action.download_video_clip`
`timeout` en segundos; `0` = sin timeout.  
El driver debe hacer POST de los bytes del clip a `{driverhub_host}/video-downloads/receive/{job_id}`.
```json
{
  "object_id": "obj-cam-001",
  "channel_idx": 0,
  "start_time": "2024-06-15T14:00:00Z",
  "end_time": "2024-06-15T14:05:00Z",
  "job_id": "job-xyz789",
  "timeout": 120
}
```

---

## Notifier

Archivo: `pkg/objects/notify.go`

### `create`
`title`, `notification_id`, `target` y todos los campos dentro de `data` son opcionales.
```json
{
  "message": "Movimiento detectado en la entrada principal",
  "title": "Alerta de seguridad",
  "notification_id": "notif-001",
  "target": "user@example.com",
  "data": {
    "image_urls": ["https://example.com/snapshot1.jpg"],
    "audio_urls": [],
    "video_urls": ["https://example.com/clip1.mp4"]
  }
}
```

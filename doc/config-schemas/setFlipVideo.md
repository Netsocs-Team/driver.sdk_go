# Set Flip Video

## Descripcion

Aplicar un efecto de volteo (vertical) a la imagen del video

## Request Message

```json
{
    ...
    "data": {
        "deviceData": {...},
        "configKey": "setFlipVideo",
        "value":{
            "enabled": true,
            "channelNumber": "111-aaa"

        }
    }
}
```

| Campo         | Tipo    | Descripcion                                  |
| ------------- | ------- | -------------------------------------------- |
| enabled       | boolean | Habilitar o deshabilitar el efecto de espejo |
| channelNumber | string  | Id del canal de video                        |

## Response Message

```json
{
    ...,
    "data": {
        "error" : false,
        "msg" : ""
    }
}
```

| Campo | Tipo    | Descripcion                    |
| ----- | ------- | ------------------------------ |
| error | boolean | Error al hacer backup          |
| msg   | string  | Mensaje de error o log interno |

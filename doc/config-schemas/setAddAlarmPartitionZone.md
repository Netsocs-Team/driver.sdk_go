# Adds a zone to a partition

## Descripcion

Añade una zona a una particion

## Request Message

```json
{
    ...
    "data": {
        "deviceData": {...},
        "configKey": "setAddAlarmPartitionZone",
        "value":{
            "number":1,
            "systemId": "xxx-xxxx",
            "zoneNumber":14
        }
    }
}

```

| Campo      | Tipo   | Descripcion                   |
| ---------- | ------ | ----------------------------- |
| number     | int    | numero de particion           |
| systemId   | string | Id de sistema de la particion |
| zoneNumber | int    | numero de la zona             |

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
| error | boolean | Error                          |
| msg   | string  | Mensaje de error o log interno |
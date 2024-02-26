# Netsocs Driver SDK

This is the official SDK for Netsocs Drivers

## Installation

```bash
go get github.com/Netsocs-Team/driver.sdk_go
```

## Usage

<img src="/doc/driverhub-ws-driver.png" />

```go
import "github.com/Netsocs-Team/driver.sdk_go/pkg/config"

type Config struct {
	// ConfigKey represents the key for the Netsocs configuration.
	ConfigKey      config.NetsocsConfigKey
	// HandleFunction is the function that handles the value message and returns a result.
	HandleFunction func(valueMessage config.HandlerValue) (interface{}, error)
}

// eventsHandlers is a map that stores event handlers associated with their respective IDs. This have cancel functions to stop the event listener.
var eventsHandlers = map[int]context.CancelFunc{}

var configs = []Config{
	{
		ConfigKey: config.ACTION_PING_DEVICE,
		HandleFunction: func(valueMessage config.HandlerValue) (interface{}, error) {
			return nil, conn.Ping()
		},
	},
	{
		ConfigKey: config.SET_ADD_PERSON_TO_AC,
		HandleFunction: func(valueMessage config.HandlerValue) (interface{}, error) {
			request := config.SetAddPersonToACRequest{}
			json.Unmarshal([]byte(valueMessage.Value), &request)
			conn, _ := buildConnection(valueMessage)
			return nil, conn.AddPersonToAC(request)
		},
	},
	{
		ConfigKey: config.SET_DEL_PERSON_TO_AC,
		HandleFunction: func(valueMessage config.HandlerValue) (interface{}, error) {
			request := config.SetDelToPersonToACRequest{}
			json.Unmarshal([]byte(valueMessage.Value), &request)
			conn, _ := buildConnection(valueMessage)
			return nil, conn.DelPersonToAC(request)
		},
	},
	{
		ConfigKey: config.SET_CARD_TO_PERSON_AC,
		HandleFunction: func(valueMessage config.HandlerValue) (interface{}, error) {
			request := config.SetCardToPersonACRequest{}
			json.Unmarshal([]byte(valueMessage.Value), &request)
			conn, _ := buildConnection(valueMessage)
			return nil, conn.SetCardToPersonAC(request)
		
		},
	},
	{
		ConfigKey: config.ACTION_LISTEN_EVENTS,
		HandleFunction: func(valueMessage config.HandlerValue) (interface{}, error) {
			conn, err := buildConnection(valueMessage)
			ctx, cancel := context.WithCancel(context.Background())
			go conn.ListenEvents(ctx, valueMessage.DeviceData.ID)
			eventsHandlers[valueMessage.DeviceData.ID] = cancel
			return nil, err
		},
	},
	{
		ConfigKey: config.ACTION_STOP_LISTEN_EVENT,
		HandleFunction: func(valueMessage config.HandlerValue) (interface{}, error) {
			cancel, ok := eventsHandlers[valueMessage.DeviceData.ID]
			if ok {
				cancel()
				delete(eventsHandlers, valueMessage.DeviceData.ID)
			} else {
				return nil, fmt.Errorf("No event listener found for device %d", valueMessage.DeviceData.ID)
			}
			return nil, nil
		},
	},
}

func main() {
   for _, _config := range configs {
		config.AddConfigHandler(_config.ConfigKey, _config.HandleFunction)
	}
	config.ListenConfig("netsocs.local:3196", "...")
}
```

En el ejemplo de arriba muestra lo que seria la base recomenda para el uso del Netsocs Driver SDK en Go. Se recomienda crear una estructura que contenga la Key de la configuracion que se quiere manejar y un manejador que se encargue de las acciones en el dispositivo correspodiente. 

## Configuraciones

Todas las 'Keys' de configuraciones estan relacionadas a un mensaje de tipo 'Request' y 'Response' que se envia al dispositivo. Estos mensajes estan definidos en el SDK y son accesibles desde el paquete `github.com/Netsocs-Team/driver.sdk_go/pkg/config`.

### Configuraciones soportadas

Los marcados con * son obligatorios.

| ConfigKey | Request struct | Response struct | Description | Details |
| --- | --- | --- | --- | --- |
| ACTION_LISTEN_EVENTS (*) | ListenEventsRequest | ListenEventsResponse | Listen the events of the device | [Open](./doc/config-schemas/actionListenEvent.md) |
| ACTION_OUTPUT_PULSE | OutputPulseRequest | OutputPulseResponse | Send a pulse to the device | [Open](./doc/config-schemas/actionOutputPulse.md) |
| ACTION_PING_DEVICE (*) | PingDeviceRequest | PingDeviceResponse | Ping the device | [Open](./doc/config-schemas/actionPingDevice.md) |
| ACTION_RESTART_DEVICE | RestartDeviceRequest | RestartDeviceResponse | Restart the device | [Open](./doc/config-schemas/actionRestart.md) |
| ACTION_STOP_LISTEN_EVENT (*) | StopListenEventRequest | StopListenEventResponse | Stop listening the events of the device | [Open](./doc/config-schemas/actionStopListenEvent.md) |
| ACTION_ZOOM | ZoomRequest | ZoomResponse | Zoom the device | MISSING |
| DELETE_USER | DeleteUserRequest | DeleteUserResponse | Delete a user from the device | [Open](./doc/config-schemas/deleteUser.md) |
| GET_ADMIN_USER | GetAdminUserRequest | GetAdminUserResponse | Get the admin user of the device | [Open](./doc/config-schemas/getAdminUser.md) |
| GET_CHANNELS | GetChannelsRequest | GetChannelsResponse | Get the channels of the device | [Open](./doc/config-schemas/getChannels.md) |
| GET_RECORDING_SOURCE | GetRecordingSourceRequest | GetRecordingSourceResponse | Get the recording source of the device | MISSING |
| GET_USERS | GetUsersRequest | GetUsersResponse | Get the users of the device | [Open](./doc/config-schemas/getUsers.md) |
| SET_ADD_PERSON_TO_AC | SetAddPersonToACRequest | SetAddPersonToACResponse | Add a person to the access control | [Open](./doc/config-schemas/setAddPersonToAC.md) |
| SET_CARD_TO_PERSON_AC | SetCardToPersonACRequest | SetCardToPersonACResponse | Set a card to a person in the access control | [Open](./doc/config-schemas/setCardToPersonAC.md) |
| SET_DEL_PERSON_TO_AC | SetDelPersonToACRequest | SetDelPersonToACResponse | Delete a person from the access control | [Open](./doc/config-schemas/setDelPersonToAC.md) |
| SET_DELETE_STORAGE | DeleteStorageRequest | DeleteStorageResponse | Delete the storage of the device | [Open](./doc/config-schemas/setDeleteStorage.md) |
| SET_DELETE_USER | DeleteUserRequest | DeleteUserResponse | Delete a user from the device | MISSING |
| SET_FACE_TO_PERSON_AC | SetFaceToPersonACRequest | SetFaceToPersonACResponse | Set a face to a person in the access control | [Open](./doc/config-schemas/setFaceToPersonAC.md) |
| SET_USERS | SetUsersRequest | SetUsersResponse | Set the users of the device | [Open](./doc/config-schemas/setUsers.md) |


# Netsocs Driver SDK

This is the official SDK for Netsocs Drivers

## Installation

```bash
go get github.com/Netsocs-Team/driver.sdk_go
```

## Usage

<img src="/doc/driverhub-ws-driver.png" />

El SDK de Driver de Netsocs establecerá comunicación con el servicio de drivers de Netsocs llamado DriverHub. 

Para que la conexión entre el driver y netsocs sea exitosa, el sdk provee el metodo `ListenConfig` el cual necesita dos parametros, el primero es la direccion del driver hub y el segundo es un texto llamado DriverKey el cual le permite el DriverHub conocer los detalles del driver que intenta establecer conexion, el DriverKey debe ser unico por driver y es clave en la comunicacion entre ambas partes, un DriverKey no emitido por Netsocs y completado la cuidadosa metodologia de testing podria no funcionar en una instancia de netsocs en produccion ya que el DriverKey contiene unos metadatos que contienen la informacion de las pruebas realizadas con el driver y si estas pruebas fueron exitosas, de alli se toma la decision de funcionar en una instancia de produccion. Una vez llamado el metodo `ListenConfig` el SDK intentara establecer conexion con el DriverHub y de ser exitosa se quedara escuchando a la espera de alguna configuracion que se haya habilitado en el driver. Para habilitar configuraciones en un driver se usa el metodo `AddConfigHandler` el cual necesita dos parametros, el primero es la `key` de la configuracion, el segundo es la funciona que se encarga de la ejecucion de esta configuración.

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
| ACTION_PLAY_AUDIO_CLIP | PlayAudioClipRequest | PlayAudioClipResponse | Play an audio clip in the device | [Open](./doc/config-schemas/actionPlayAudioClip.md) |
| ACTION_RESTART_DEVICE | RestartDeviceRequest | RestartDeviceResponse | Restart the device | [Open](./doc/config-schemas/actionRestart.md) |
| ACTION_STOP_LISTEN_EVENT (*) | StopListenEventRequest | StopListenEventResponse | Stop listening the events of the device | [Open](./doc/config-schemas/actionStopListenEvent.md) |
| ACTION_ZOOM | ZoomRequest | ZoomResponse | Zoom the device | MISSING |
| DELETE_USER | DeleteUserRequest | DeleteUserResponse | Delete a user from the device | [Open](./doc/config-schemas/deleteUser.md) |
| GET_ADMIN_USER | GetAdminUserRequest | GetAdminUserResponse | Get the admin user of the device | [Open](./doc/config-schemas/getAdminUser.md) |
| GET_ADMIN_USERS | GetAdminUsersRequest | GetAdminUsersResponse | Get the admin users of the device | [Open](./doc/config-schemas/getAdminUsers.md) |
| GET_AVAILABLE_OUTPUTS (*) | GetAvailableOutputsRequest | GetAvailableOutputsResponse | Get the available outputs of the device | [Open](./doc/config-schemas/getAvailableOutputs.md) |
| GET_AVAILABLE_SPEAKERS (*) | GetAvailableSpeakersRequest | GetAvailableSpeakersResponse | Get the available speakers of the device | [Open](./doc/config-schemas/getAvailableSpeakers.md) |
| GET_AVAILABLE_VIDEO_RESOLUTIONS (*) | GetAvailableVideoResolutionsRequest | GetAvailableVideoResolutionsResponse | Get the available video resolutions of the device | [Open](./doc/config-schemas/getAvailableVideoResolutions.md) |
| GET_CHANNELS | GetChannelsRequest | GetChannelsResponse | Get the channels of the device | [Open](./doc/config-schemas/getChannels.md) |
| GET_FLIP_VIDEO_STATUS | GetFlipVideoStatusRequest | GetFlipVideoStatusResponse | Get the flip video status of the device | [Open](./doc/config-schemas/getFlipVideoStatusResponse.md) |
| GET_HEATMAP_IMAGE | GetHeatmapImageRequest | GetHeatmapImageResponse | Get the heatmap image of the device | [Open](./doc/config-schemas/getHeatmapImage.md) |
| GET_MIRROR_VIDEO_STATUS | GetMirrorVideoStatusRequest | GetMirrorVideoStatusResponse | Get the mirror video status of the device | [Open](./doc/config-schemas/getMirrorVideoStatusResponse.md) |
| GET_RECORDING_SOURCE | GetRecordingSourceRequest | GetRecordingSourceResponse | Get the recording source of the device | MISSING |
| GET_STORAGES | GetStoragesRequest | GetStoragesResponse | Get the storages of the device | [Open](./doc/config-schemas/getStorages.md) |
| GET_USERS | GetUsersRequest | GetUsersResponse | Get the users of the device | [Open](./doc/config-schemas/getUsers.md) |
| GET_VIDEO_IN_BLACK_AND_WHITE_STATUS | GetVideoInBlackAndWhiteStatusRequest | GetVideoInBlackAndWhiteStatusResponse | Get the video in black and white status of the device | [Open](./doc/config-schemas/getVideoInBlackAndWhiteStatus.md) |
| SET_ADD_PERSON_TO_AC | SetAddPersonToACRequest | SetAddPersonToACResponse | Add a person to the access control | [Open](./doc/config-schemas/setAddPersonToAC.md) |
| SET_ADMIN_USER | SetAdminUserRequest | SetAdminUserResponse | Set the admin user of the device | [Open](./doc/config-schemas/setAdminUser.md) |
| SET_CARD_TO_PERSON_AC | SetCardToPersonACRequest | SetCardToPersonACResponse | Set a card to a person in the access control | [Open](./doc/config-schemas/setCardToPersonAC.md) |
| SET_DEL_PERSON_TO_AC | SetDelPersonToACRequest | SetDelPersonToACResponse | Delete a person from the access control | [Open](./doc/config-schemas/setDelPersonToAC.md) |
| SET_DELETE_STORAGE | DeleteStorageRequest | DeleteStorageResponse | Delete the storage of the device | [Open](./doc/config-schemas/setDeleteStorage.md) |
| SET_DELETE_USER | DeleteUserRequest | DeleteUserResponse | Delete a user from the device | MISSING |
| SET_FACE_TO_PERSON_AC | SetFaceToPersonACRequest | SetFaceToPersonACResponse | Set a face to a person in the access control | [Open](./doc/config-schemas/setFaceToPersonAC.md) |
| SET_FLIP_VIDEO | SetFlipVideoRequest | SetFlipVideoResponse | Set the flip video of the device | [Open](./doc/config-schemas/setFlipVideo.md) |
| SET_MIRROR_VIDEO | SetMirrorVideoRequest | SetMirrorVideoResponse | Set the mirror video of the device | [Open](./doc/config-schemas/setMirrorVideo.md) |
| SET_USERS | SetUsersRequest | SetUsersResponse | Set the users of the device | [Open](./doc/config-schemas/setUsers.md) |
| SET_VIDEO_IN_BLACK_AND_WHITE | SetVideoInBlackAndWhiteRequest | SetVideoInBlackAndWhiteResponse | Set the video in black and white of the device | [Open](./doc/config-schemas/setVideoInBlackAndWhite.md) |
| SET_VIDEO_RESOLUTION | SetVideoResolutionRequest | SetVideoResolutionResponse | Set the video resolution of the device | [Open](./doc/config-schemas/setVideoResolution.md) |
| SET_BLOCK_PERSON_TO_AC | SetBlockPersonToACRequest | SetBlockPersonToACResponse | Block a person in the access control | [Open](./doc/config-schemas/setBlockPersonToAC.md) |
| DELETE_ALL_PEOPLE_AC | DeleteAllPeopleACRequest | DeleteAllPeopleACResponse | Eliminar todas las personas de la lista de acceso del dispositivo | [Open](./doc/config-schemas/deleteAllPeopleAC.md) |
|GET_INPUTS | GetInputsRequest | GetInputsResponse | Get a list of inputs in the device | [Open](./doc/config-schemas/getInputs.md) |
| GET_MICROPHONES | GetMicrophonesRequest | GetMicrophonesResponse | Get the microphones of the device |[Open](./doc/config-schemas/getMicrophones.md) |
| GET_CURRENT_VIDEO_RESOLUTION_BY_CHANNEL | GetCurrentVideoResolutionByChannelRequest | GetCurrentVideoResolutionByChannelResponse | Get the current video resolution of a channel |[Open](./doc/config-schemas/getCurrentVideoResolutionByChannel.md) |


package objects

import (
	"errors"
	"strings"
	"sync"

	"github.com/Netsocs-Team/driver.sdk_go/internal/eventbus"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type objectRunner struct {
	controller ObjectController
	// objectsMap map[string][]RegistrableObject
	objectsMap sync.Map
	// registerMu serialises the read-modify-write of a domain's object list.
	registerMu sync.Mutex
}

// GetController implements ObjectRunner.
func (o *objectRunner) GetController() ObjectController {
	return o.controller
}

type requestActionExecutionEventData struct {
	Payload           map[string]interface{} `json:"payload"`
	Domain            string                 `json:"domain"`
	Action            string                 `json:"action"`
	ObjectID          []string               `json:"object_id"`
	ActionExecutionID string                 `json:"id"`
}

// SubscribeToActionsRequest implements objects.ObjectRunner.
func (o *objectRunner) listenActions() {
	eventbus.Pubsub.Subscribe("REQUEST_ACTION_EXECUTION", func(data interface{}) {
		logger, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		req := requestActionExecutionEventData{}
		jsoncontent, _ := json.Marshal(data)
		err = json.Unmarshal(jsoncontent, &req)
		if err != nil {
			sugar.Error("failed to unmarshal request action execution data", zap.Error(err))
			return
		}
		payloadBytes, err := json.Marshal(req.Payload)
		if err != nil {
			sugar.Error("failed to marshal payload", zap.Error(err))
			return
		}

		objectsRaw, ok := o.objectsMap.Load(req.Domain)
		if !ok {
			sugar.Info("no objects found for domain", zap.String("domain", req.Domain))
			return
		}

		objects, ok := objectsRaw.([]RegistrableObject)
		if !ok {
			sugar.Info("invalid objects type", zap.String("domain", req.Domain))
			return
		}

		if len(objects) == 0 {
			sugar.Info("no objects found for domain", zap.String("domain", req.Domain))
			return
		}

		sugar.Info("running action", zap.String("action", req.Action), zap.String("domain", req.Domain), zap.Strings("object_id", req.ObjectID), zap.String("payload", string(payloadBytes)))

		RunActionRoutine := func(obj RegistrableObject) {
			resp, err := obj.RunAction(req.ActionExecutionID, req.Action, payloadBytes)

			if err != nil {
				sugar.Info("action execution error", zap.Error(err))
				o.GetController().UpdateResultAttributes(req.ActionExecutionID, map[string]string{"error": err.Error()})
			} else {
				sugar.Info("action executed", zap.Any("response", resp))
				o.GetController().UpdateResultAttributes(req.ActionExecutionID, resp)
			}
		}

		if len(req.ObjectID) > 0 {
			for _, obj := range objects {
				for _, objID := range req.ObjectID {
					if obj.GetMetadata().ObjectID == objID {
						go RunActionRoutine(obj)
					}
				}
			}
		} else {
			for _, obj := range objects {
				go RunActionRoutine(obj)
			}
		}
	})

}

// RegisterObject implements objects.ObjectRunner.
func (o *objectRunner) RegisterObject(object RegistrableObject) error {

	if err := o.controller.CreateObject(object); err != nil {
		if !strings.Contains(err.Error(), "ERR_ITEM_ALREADY_EXIST") {
			return err
		}
	}

	for _, action := range object.GetAvailableActions() {
		if err := o.controller.NewAction(action); err != nil {
			if !strings.Contains(err.Error(), "ERR_ITEM_ALREADY_EXIST") {
				return err
			}
		}
	}

	if err := o.storeObject(object); err != nil {
		return err
	}
	eventbus.Pubsub.Publish("SUBSCRIBE_OBJECTS_COMMANDS_LISTENING", struct{ Domain string }{Domain: object.GetMetadata().Domain})

	return object.Setup(o.controller)
}

func NewObjectRunner(controller ObjectController) ObjectRunner {
	runner := &objectRunner{
		controller: controller,
		// objectsMap: make(map[string][]RegistrableObject),
	}

	runner.listenActions()
	return runner
}

// storeObject adds object to its domain's list, replacing an object already
// registered with the same ObjectID.
//
// Drivers register their objects again every time the hub asks them to
// provision a device (a hub restart, a reconnect, an edited device). Appending
// unconditionally kept every previous instance: each action then ran once per
// registration — after one hub restart a single publish_stream_start opened two
// RTSP sessions and two WHIP publishers for the same stream, the second one
// replaced the first in the hub, and the viewer sat on a spinner until the
// player reconnected. The newest registration wins because it carries the
// device's current configuration.
func (o *objectRunner) storeObject(object RegistrableObject) error {
	o.registerMu.Lock()
	defer o.registerMu.Unlock()

	domain := object.GetMetadata().Domain
	id := object.GetMetadata().ObjectID
	existingRaw, ok := o.objectsMap.Load(domain)
	if !ok {
		o.objectsMap.Store(domain, []RegistrableObject{object})
		return nil
	}
	existing, validType := existingRaw.([]RegistrableObject)
	if !validType {
		return errors.New("invalid objects type")
	}
	// Copy on write: listenActions may be iterating the current slice.
	next := make([]RegistrableObject, 0, len(existing)+1)
	for _, obj := range existing {
		if obj.GetMetadata().ObjectID != id {
			next = append(next, obj)
		}
	}
	o.objectsMap.Store(domain, append(next, object))
	return nil
}

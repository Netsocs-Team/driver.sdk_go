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

	existingObjectsRaw, ok := o.objectsMap.Load(object.GetMetadata().Domain)
	if !ok {
		o.objectsMap.Store(object.GetMetadata().Domain, []RegistrableObject{object})
	} else {
		existingObjects, validType := existingObjectsRaw.([]RegistrableObject)
		if !validType {
			// sugar.Info("invalid objects type", zap.String("domain", object.GetMetadata().Domain))
			return errors.New("invalid objects type")
		}
		o.objectsMap.Store(object.GetMetadata().Domain, append(existingObjects, object))
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

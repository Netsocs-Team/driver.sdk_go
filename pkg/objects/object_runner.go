package objects

import (
	"strings"

	"github.com/Netsocs-Team/driver.sdk_go/internal/eventbus"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type objectRunner struct {
	controller ObjectController
	objectsMap map[string][]RegistrableObject
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

		objects := o.objectsMap[req.Domain]
		if objects == nil || len(objects) == 0 {
			sugar.Info("no objects found for domain", zap.String("domain", req.Domain))
			return
		}

		sugar.Info("running action", zap.String("action", req.Action), zap.String("domain", req.Domain), zap.Strings("object_id", req.ObjectID), zap.String("payload", string(payloadBytes)))

		if len(req.ObjectID) > 0 {
			for _, obj := range objects {
				for _, objID := range req.ObjectID {
					if obj.GetMetadata().ObjectID == objID {
						go obj.RunAction(req.ActionExecutionID, req.Action, payloadBytes)
					}
				}
			}
		} else {
			for _, obj := range objects {
				go obj.RunAction(req.ActionExecutionID, req.Action, payloadBytes)
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
	var registerNew = true
	existingObjects := o.objectsMap[object.GetMetadata().Domain]
	for _, existingObject := range existingObjects {
		if existingObject.GetMetadata().ObjectID == object.GetMetadata().ObjectID {
			registerNew = false
			break
		}
	}
	if registerNew {
		o.objectsMap[object.GetMetadata().Domain] = append(o.objectsMap[object.GetMetadata().Domain], object)
	}
	eventbus.Pubsub.Publish("SUBSCRIBE_OBJECTS_COMMANDS_LISTENING", struct{ Domain string }{Domain: object.GetMetadata().Domain})

	return object.Setup(o.controller)
}

func NewObjectRunner(controller ObjectController) ObjectRunner {
	runner := &objectRunner{
		controller: controller,
		objectsMap: make(map[string][]RegistrableObject),
	}

	runner.listenActions()
	return runner
}

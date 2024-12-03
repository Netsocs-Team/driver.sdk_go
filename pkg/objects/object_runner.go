package objects

import (
	"fmt"
	"strings"

	"github.com/Netsocs-Team/driver.sdk_go/internal/eventbus"
	"github.com/goccy/go-json"
)

type objectRunner struct {
	controller ObjectController
	objectsMap map[string][]RegistrableObject
}

type requestActionExecutionEventData struct {
	Payload  map[string]interface{} `json:"payload"`
	Domain   string                 `json:"domain"`
	Action   string                 `json:"action"`
	ObjectID string                 `json:"object_id"`
}

// SubscribeToActionsRequest implements objects.ObjectRunner.
func (o *objectRunner) listenActions() {
	eventbus.Pubsub.Subscribe("REQUEST_ACTION_EXECUTION", func(data interface{}) {
		fmt.Println("Received action request")
		req := requestActionExecutionEventData{}
		jsoncontent, _ := json.Marshal(data)
		err := json.Unmarshal(jsoncontent, &req)
		if err != nil {
			return
		}
		payloadBytes, _ := json.Marshal(req.Payload)

		objects := o.objectsMap[req.Domain]
		if objects == nil || len(objects) == 0 {
			return
		}
		if req.ObjectID != "" {
			for _, obj := range objects {
				if obj.GetMetadata().ObjectID == req.ObjectID {
					obj.RunAction(req.Action, payloadBytes)
				}
			}
		} else {
			for _, obj := range objects {
				obj.RunAction(req.Action, payloadBytes)
			}
		}

	})

}

// RegisterObject implements objects.ObjectRunner.
func (o *objectRunner) RegisterObject(object RegistrableObject) error {
	var mustDisable bool = true
	if err := o.controller.CreateObject(object); err != nil {
		if strings.Contains(err.Error(), "ERR_ITEM_ALREADY_EXIST") {
			mustDisable = false
		} else {
			return err
		}
	}

	if mustDisable {
		defer o.controller.DisabledObject(object.GetMetadata().ObjectID)
	}

	err := o.controller.EnabledObject(object.GetMetadata().ObjectID)
	if err != nil {
		return err
	}

	for _, action := range object.GetAvailableActions() {
		if err := o.controller.NewAction(action); err != nil {
			if !strings.Contains(err.Error(), "ERR_ITEM_ALREADY_EXIST") {
				return err
			}
		}
	}

	o.objectsMap[object.GetMetadata().Domain] = append(o.objectsMap[object.GetMetadata().Domain], object)

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

package objects

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/valyala/fastjson"
)

type createObjectRequest struct {
	ObjectID string `json:"object_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Icon     string `json:"icon"`
	DeviceID int    `json:"device_id"`
	State    struct {
		StateID string `json:"state_id"`
		Name    struct {
			Es string `json:"es"`
		} `json:"name"`
		Icon string `json:"icon"`
	} `json:"state"`
}

func RegisterObject(objectRunner ObjectRunner, objectHandler ObjectHandler, driverKey string, driverhubHost string) error {
	if err := objectHandler.AppendObject(objectRunner); err != nil {
		return err
	}

	client := resty.New().R()
	req := createObjectRequest{}
	req.ObjectID = objectRunner.GetID()
	req.Type = objectRunner.GetType()
	req.Name = objectRunner.GetName()
	req.Icon = objectRunner.GetIcon()
	req.DeviceID = objectRunner.GetDeviceID()

	client.Header.Set("Content-Type", "application/json")
	client.Header.Set("Authorization", driverKey)

	rawresp, err := client.SetBody(req).Post(fmt.Sprintf("http://%s/api/v1/objects", driverhubHost))
	if err != nil {
		return err
	}
	response := rawresp.String()

	if fastjson.MustParse(response).Get("error") != nil {
		errMessage := fastjson.MustParse(response).Get("error").String()
		errCode := fastjson.MustParse(response).Get("code").String()
		msg := fmt.Sprintf("Error code: %s, Error message: %s", errCode, errMessage)
		switch errCode {
		case "\"ERR_OBJECT_ALREADY_EXIST\"":
			break
		default:
			log.Printf("Error: %s", msg)
			return errors.New(msg)
		}

	}

	go func() {
		for {
			select {
			case cmd := <-objectRunner.GetCommandsChannel():
				fmt.Println("Command received: ", cmd.Command)

				switch cmd.Command {
				case _CHANGE_ATTRIBUTE_COMMAND:
					req := interface{}(nil)
					json.Unmarshal([]byte(cmd.Params), &req)
					resp, err := client.SetBody(req).Post(fmt.Sprintf("http://%s/api/v1/object/%s/set_attribute", driverhubHost, objectRunner.GetID()))
					if err != nil {
						log.Printf("Error: %s", err)
						continue
					}
					response := resp.String()
					fmt.Println(response)
				}
			}
		}
	}()

	log.Printf("Object %s registered successfully", objectRunner.GetID())

	return nil
}

package objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndGetState(t *testing.T) {
	o := Object{}
	o.SetState("on")
	assert.Equal(t, "on", o.GetState())
	assert.Equal(t, "on", o.State)

	o.SetState("off")
	assert.Equal(t, "off", o.GetState())
	assert.Equal(t, "off", o.State)
}

func TestSetAndGetStateAttributes(t *testing.T) {
	o := Object{}
	o.SetStateProperties("key", "value")
	assert.Equal(t, "value", o.GetStateProperties("key"))
	assert.Equal(t, "value", o.StateProperties["key"])

	o.SetStateProperties("key2", "value2")
	assert.Equal(t, "value2", o.GetStateProperties("key2"))
	assert.Equal(t, "value2", o.StateProperties["key2"])
}

func TestSetAndGetIcon(t *testing.T) {
	o := Object{}
	o.SetIcon("icon.png")
	assert.Equal(t, "icon.png", o.GetIcon())
	assert.Equal(t, "icon.png", o.Icon)
}

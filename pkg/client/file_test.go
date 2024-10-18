package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFileData(t *testing.T) {
	data, err := getDriverNetsocsDotJsonContent("driver.netsocs.json")

	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "key123", data.DriverKey)
	assert.Equal(t, "192.168.6.43:3196", data.DriverHubHost)
}

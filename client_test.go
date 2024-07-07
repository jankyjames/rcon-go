package rcon

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c := New("", "")
	do, err := c.Do("DoExit")
	assert.NoError(t, err)
	fmt.Println(do)
	for {
		time.Sleep(time.Minute)
	}
}

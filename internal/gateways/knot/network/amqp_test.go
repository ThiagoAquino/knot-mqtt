package network

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAMQPConnection(t *testing.T) {
	url := os.Getenv("AMQP_URL")
	amqp := NewAMQP(url)
	err := amqp.Start()
	assert.Nil(t, err)
}

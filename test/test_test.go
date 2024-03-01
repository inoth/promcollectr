package test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/push"
)

func TestPush(t *testing.T) {
	ph := push.New("localhost", "job").Gatherer(nil)

	ph.Push()
}

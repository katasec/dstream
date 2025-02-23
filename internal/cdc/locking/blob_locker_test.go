package locking

import (
	"log"
	"testing"

	"github.com/katasec/dstream/internal/config"
)

func TestHello(t *testing.T) {
	config := config.NewConfig("../../dstream.hcl")
	log.Println(config.Ingester.Tables[0].Name)
}

package stream

import (
	"github.com/satori/go.uuid"
	"log"
	"os"
	"testing"
)

var testKafkaAddr string

func TestMain(m *testing.M) {
	testKafkaAddr = os.Getenv("KAFKA_ADDR")
	if testKafkaAddr == "" {
		testKafkaAddr = "127.0.0.1:9092"
	}

	log.Printf("using addr: %s", testKafkaAddr)

	os.Exit(m.Run())
}

func TestSLCreate(t *testing.T) {
	_, err := NewSL(&SLOptions{Addrs: []string{testKafkaAddr}, ClientID: uuid.NewV4().String()})
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
}

func TestSLLock(t *testing.T) {
	sl, err := NewSL(&SLOptions{Addrs: []string{testKafkaAddr}, ClientID: uuid.NewV4().String(), SLConfig: &Config{Topic: uuid.NewV4().String()}})

	err = sl.Lock()
	if err != nil {
		if err == ErrSLInUse {
			t.Error("SL was locked prematurely")
		} else {
			t.Error(err)
		}
		t.FailNow()
		return
	}

	err = sl.Lock()
	if err != ErrSLInUse {
		t.Errorf("SL was not locked when it should be locked, output was %s", err.Error())
		t.FailNow()
		return
	}

}

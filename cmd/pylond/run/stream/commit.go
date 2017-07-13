package stream

import (
	"gopkg.in/Shopify/sarama.v1"
	"hash/fnv"
)

type Committer struct {
	p  sc.
	pc chan<- *sarama.ProducerMessage
	e  <-chan *sarama.ProducerError
}

func NewCommitter(sl *SL) *Committer {
	sl.Lock()
	p := sarama.NewAsyncProducerFromClient(sl.Kafka)
	return &Committer{
		sl: sl,
		p:  p,
		pc: p.Input(),
		e:  p.Errors(),
	}
}

func (c *Committer) ErrorListener(fn func(<-chan *sarama.ProducerError)) {
	fn(c.e)
}

func (c *Committer) Write(cmd []byte, metadata interface{}) error {
	h := fnv.New64()
	h.Write(cmd)

	m := &sarama.ProducerMessage{
		Topic:    c.sl.Config.Topic,
		Key:      h.Sum(),
		Value:    cmd,
		Metadata: metadata,
	}

	c.pc <- m
}

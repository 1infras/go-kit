package consume

import (
	"github.com/Shopify/sarama"
)

type consumerGroup struct {
	cg sarama.ConsumerGroup
}

func (_this *consumerGroup) close() error {
	return _this.cg.Close()
}
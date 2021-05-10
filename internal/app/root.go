package app

import (
	"github.com/core-go/firestore"
	"github.com/core-go/health/server"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/pubsub"
)

type Root struct {
	Server            server.ServerConf       `mapstructure:"server"`
	Log               log.Config              `mapstructure:"log"`
	Firestore         firestore.Config        `mapstructure:"firestore"`
	BatchWorkerConfig mq.BatchWorkerConfig    `mapstructure:"batch_worker"`
	Sub               pubsub.SubscriberConfig `mapstructure:"sub"`
	Pub               *pubsub.PublisherConfig `mapstructure:"pub"`
}

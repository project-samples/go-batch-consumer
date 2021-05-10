package app

import (
	"context"
	"reflect"

	"github.com/core-go/firestore"
	"github.com/core-go/health"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/pubsub"
	"github.com/core-go/mq/validator"
	val "github.com/go-playground/validator/v10"
)

type ApplicationContext struct {
	HealthHandler *health.HealthHandler
	BatchWorker   mq.BatchWorker
	Receive       func(ctx context.Context, handle func(context.Context, *mq.Message, error) error)
	Subscription  *mq.Subscription
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	client, er1 := firestore.Connect(ctx, []byte(root.Firestore.Credentials))
	if er1 != nil {
		return nil, er1
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	receiver, er2 := pubsub.NewSubscriberByConfig(ctx, root.Sub, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}

	userType := reflect.TypeOf(User{})
	batchWriter := firestore.NewBatchInserter(client, "user", userType)
	batchHandler := mq.NewBatchHandler(userType, batchWriter.Write, logError, logInfo)

	firestoreChecker := firestore.NewHealthChecker(ctx, []byte(root.Firestore.Credentials))
	receiverChecker := pubsub.NewSubHealthChecker("pubsub_subscriber", receiver.Client, root.Sub.SubscriptionId)
	var healthHandler *health.HealthHandler
	var batchWorker mq.BatchWorker

	if root.Pub != nil {
		sender, er3 := pubsub.NewPublisherByConfig(ctx, *root.Pub)
		if er3 != nil {
			log.Error(ctx, "Cannot new a new sender. Error: "+er3.Error())
			return nil, er3
		}
		retryService := mq.NewRetryService(sender.Publish, logError, logInfo)
		batchWorker = mq.NewDefaultBatchWorker(root.BatchWorkerConfig, batchHandler.Handle, retryService.Retry, logError, logInfo)
		senderChecker := pubsub.NewPubHealthChecker("pubsub_publisher", sender.Client, root.Pub.TopicId)
		healthHandler = health.NewHealthHandler(firestoreChecker, receiverChecker, senderChecker)
	} else {
		batchWorker = mq.NewDefaultBatchWorker(root.BatchWorkerConfig, batchHandler.Handle, nil, logError, logInfo)
		healthHandler = health.NewHealthHandler(firestoreChecker, receiverChecker)
	}
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	validator := mq.NewValidator(userType, checker.Check)
	subscription := mq.NewSubscription(batchWorker.Handle, validator.Validate, logError, logInfo)

	return &ApplicationContext{
		HealthHandler: healthHandler,
		BatchWorker:   batchWorker,
		Receive:       receiver.Subscribe,
		Subscription:  subscription,
	}, nil
}

func NewUserValidator() validator.Validator {
	val := validator.NewDefaultValidator()
	val.CustomValidateList = append(val.CustomValidateList, validator.CustomValidate{Fn: CheckActive, Tag: "active"})
	return val
}
func CheckActive(fl val.FieldLevel) bool {
	return fl.Field().Bool()
}

package worker

import (
	"crypto/tls"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/handysuherman/clean-arch-payment-service/internal/config" //nolint
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	messageClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka" //nolint
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"go.uber.org/mock/gomock"
)

var (
	conf     *config.App
	tlog     logger.Logger
	val      *validator.Validate
	producer *messageClient.ProducerImpl
)

func TestMain(m *testing.M) {
	logger := logger.NewLogger()
	cm := config.NewManager(logger, 15*time.Second)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	cfgs, err := cm.Bootstrap(fmt.Sprintf("%v/%s", findModuleRoot(cwd), "etcd-config.yaml"))
	if err != nil {
		logger.Debug(err)
		return
	}

	tlog = logger
	val = validator.New()
	conf = cfgs

	producer = producerWorker()

	os.Exit(m.Run())
}

func producerWorker() *messageClient.ProducerImpl {
	kafkaTls := &tls.Config{}

	if conf.Brokers.Kafka.Config.EnableTLS {
		tlsCfg, err := helper.Base64EncodedTLS(conf.TLS.Kafka.Ca, conf.TLS.Kafka.Cert, conf.TLS.Kafka.Key)
		if err != nil {
			tlog.Warnf("connectKafkaBrokers.helper.base64encodedtls.err: %v", err)
			// return err
		}

		kafkaTls = tlsCfg
	}

	return messageClient.NewProducerImpl(tlog, conf.Brokers.Kafka.Config.Brokers, conf.Brokers.Kafka.Config.EnableTLS, kafkaTls)
}

type eqGetByIDParamsMatcher struct {
	arg string
}

func EqGetByIDParamsMatcher(arg string) gomock.Matcher {
	return &eqGetByIDParamsMatcher{arg: arg}
}

func (ex *eqGetByIDParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(string)
	if !ok {
		return false
	}

	ex.arg = arg

	if ex.arg == "" {
		return false
	}

	if !reflect.DeepEqual(ex.arg, arg) {
		return false
	}

	return true
}

func (ex *eqGetByIDParamsMatcher) String() string {
	if ex.arg == "" {
		return "id should not be empty"
	}

	return fmt.Sprintf("matches arg: %v", ex.arg)
}

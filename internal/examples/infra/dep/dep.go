package dep

import (
	"context"
	"fmt"

	"github.com/ekuu/dgo/internal/examples/domain/order"

	dr "github.com/ekuu/dgo/repository"

	"github.com/ekuu/dgo/internal/examples/domain/product"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/ekuu/dgo"
	ebrocket "github.com/ekuu/dgo/bus/rocketmq"
	"github.com/ekuu/dgo/internal/examples/domain/account"
	"github.com/ekuu/dgo/internal/examples/infra/repo/mongo"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MustBus() dgo.Bus {
	producerOptions := []producer.Option{
		producer.WithNameServer([]string{"192.168.31.210:9876"}),
		producer.WithGroupName(fmt.Sprintf("%s-event-producer", "ddd-test")),
		producer.WithRetry(2),
	}
	consumerOptions := []consumer.Option{
		consumer.WithNameServer([]string{"192.168.31.210:9876"}),
		consumer.WithGroupName(fmt.Sprintf("%s-push-consumer", "ddd-test")),
	}
	eb, err := ebrocket.NewTransactionRocketMQ(producerOptions, consumerOptions, func(ext *primitive.MessageExt) primitive.LocalTransactionState {
		fmt.Println(ext)
		return primitive.RollbackMessageState
	})
	if err != nil {
		panic(err)
	}
	return eb
}

func MustDB() *mg.Database {
	client, err := mg.Connect(context.Background(), options.Client().ApplyURI("mongodb://root:123456@192.168.31.210:27017"))
	if err != nil {
		panic(err)
	}
	return client.Database("ddd")
}

func AccountSvc() *dgo.Service[*account.Account] {
	return dgo.NewService[*account.Account](
		mongo.NewAccountRepo(MustDB()),
		func() *account.Account {
			return account.New(dgo.NewAggBase())
		},
		dgo.WithServiceBus[*account.Account](MustBus()),
		//dgo.WithServiceSnapshotSaveStrategy[*account.Account](dgo.AlwaysSaveSnapshot[*account.Account]),
		//dgo.WithServiceGenID[*account.Account](func(ctx context.Context) (dgo.ID, error) {
		//	return repo.NewObjectID().Reverse(), nil
		//}),
	)
}

func ProductSvc() *dgo.Service[*product.Product] {
	return dgo.NewService[*product.Product](
		mongo.NewProductRepo(MustDB()),
		func() *product.Product {
			return product.New(dgo.NewAggBase())
		},
		dgo.WithServiceIdGenFunc[*product.Product](func(ctx context.Context) (dgo.ID, error) {
			return dr.NewObjectID().Reverse(), nil
		}),
	)
}

func OrderSvc() *dgo.Service[*order.Order] {
	return dgo.NewService[*order.Order](
		mongo.NewOrderRepo(MustDB()),
		func() *order.Order {
			return order.New(dgo.NewAggBase())
		},
	)
}

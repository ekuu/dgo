package app

import (
	"context"

	"github.com/davecgh/go-spew/spew"

	"github.com/ekuu/dgo/internal/examples/domain/account"
	"github.com/ekuu/dgo/internal/examples/infra/dep"
)

func Translate(ctx context.Context, cmd *account.TransferCmd) error {
	as, err := dep.AccountSvc().Batch(ctx, cmd.BatchEntries())
	if err != nil {
		return err
	}
	spew.Dump(as)
	return nil

	//err = svc.Delete(context.Background(), new(account.DeleteCmd), dgo.ID("6458b4cb0ad733309a4134e0"))
	//err = svc.Delete(context.Background(), new(account.DeleteCmd), dgo.ID("30648790f4da40f3a609c5e355921f4a"))
	//a, err := svc.Update(context.Background(), &account.UpdateNameCmd{Name: "lisi"}, dgo.ID("wangwu"))
	//spew.Dump(a)

	//as, err := svc.BatchCreate(
	//	context.Background(),
	//	dgo.NewBatchHandler(func(ctx context.Context) (entries []dgo.BatchEntry[*account.Account], err error) {
	//		entries = append(
	//			entries,
	//			dgo.NewBatchEntry[*account.Account](&account.CreateCmd{"zhangsan", 100}, dgo.ID("zhangsan")),
	//			dgo.NewBatchEntry[*account.Account](&account.CreateCmd{"wangwu", 0}, dgo.ID("lisi")),
	//		)
	//		return
	//	}),
	//)
}

func CreateAccount(ctx context.Context, cmd *account.CreateCmd) (*account.Account, error) {
	return dep.AccountSvc().Create(ctx, cmd)
	//return dep.AccountSvc().Save(context.Background(), cmd, dgo.ID("wangwu"))
}

func UpdateAccountName(ctx context.Context, cmd *account.UpdateNameCmd) (*account.Account, error) {
	return dep.AccountSvc().Update(ctx, cmd, cmd.ID)
}

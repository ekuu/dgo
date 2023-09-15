package app

import (
	"context"

	"github.com/ekuu/dgo/internal/examples/domain/account"
	"github.com/ekuu/dgo/internal/examples/infra/dep"
)

func Translate(ctx context.Context, cmd *account.TransferCmd) error {

	return dep.AccountSvc().BatchUpdate(ctx, cmd)

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
	//dep.AccountSvc().Execute(context.Background(), func(ctx context.Context, r dgo.Repo[*account.Account]) error {
	//	a, err := r.Get(ctx, dgo.ID("zhangsan"))
	//	if err != nil {
	//		return err
	//	}
	//
	//})
	//dep.AccountSvc().CreateBy(ctx, fn)
	return dep.AccountSvc().Create(ctx, cmd)
	//return dep.AccountSvc().Save(context.Background(), cmd, dgo.ID("wangwu"))
}

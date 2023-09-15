package app

import (
	"context"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal/examples/domain/account"
)

func TestTranslate(t *testing.T) {
	err := Translate(context.Background(), account.NewTransferCmd(dgo.ID("zhangsan"), dgo.ID("wangwu"), 1))
	if err != nil {
		t.Fatalf("%+v", err)
	}
	time.Sleep(time.Millisecond * 300)
}

func TestCreateAccount(t *testing.T) {
	a, err := CreateAccount(context.Background(), &account.CreateCmd{Name: "zhangsan2", Balance: 100})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	spew.Dump(a)
	time.Sleep(time.Millisecond * 300)
}

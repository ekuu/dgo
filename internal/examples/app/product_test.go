package app

import (
	"context"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/ekuu/dgo/internal/examples/domain/product"
)

func TestCreateProduct(t *testing.T) {
	a, err := CreateProduct(context.Background(), &product.CreateCmd{Name: "牛奶", Price: 100})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	spew.Dump(a)
	time.Sleep(time.Millisecond * 300)
}

func TestCreateProducts(t *testing.T) {
	cmds := []*product.CreateCmd{
		{Name: "西红柿", Price: 200},
		{Name: "黄瓜", Price: 150},
	}
	a, err := CreateProducts(context.Background(), cmds)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	spew.Dump(a)
	time.Sleep(time.Millisecond * 300)
}

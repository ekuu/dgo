package app

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/ekuu/dgo"
	"github.com/samber/lo"
)

func TestCreateOrder(t *testing.T) {
	products := []lo.Entry[dgo.ID, uint32]{
		{Key: "65017182a113f3b19a27ec69", Value: 2},
		{Key: "6501755af62cddbfcbae62cf", Value: 1},
	}
	a, err := CreateOrder(context.Background(), products, 20)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	spew.Dump(a)
}

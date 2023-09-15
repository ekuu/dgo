package account

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ekuu/dgo"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true,
		Level: slog.LevelDebug,
	})))
}

//go:generate gogen option -n Account -p _ -r AggBase
type Account struct {
	dgo.AggBase
	name    string
	balance uint64
}

func (a *Account) Name() string {
	return a.name
}

func (a *Account) Balance() uint64 {
	return a.balance
}

func (a *Account) Validate() error {
	fmt.Println("this is validate")
	return nil
}

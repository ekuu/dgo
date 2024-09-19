package initial

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func init() {
	Cmd.Flags().BoolP("lowercase", "", false, "make constructor names lowercase.")
	//if err := Cmd.MarkFlagRequired("name"); err != nil {
	//	panic(err)
	//}
}

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "初始化领域/项目",
	Example: `  # Simple
  gogen option -n structName

  # Specifies the filename or directory where the structure is located
  gogen option -n structName [file.go|directory]
`,
	Args: cobra.MinimumNArgs(1),
	Run:  Init,
}

func Init(cmd *cobra.Command, args []string) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	name := args[0]
	root := filepath.Join(pwd, name)
	if err := os.MkdirAll(root, 0755); err != nil {
		panic(err)
	}
	fmt.Println(filepath.Join(pwd, name))
	fmt.Println(os.Args)
	fmt.Println(os.Getwd())
	fmt.Println(args)
}

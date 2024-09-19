package create

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.Flags().StringP("name", "n", "", "struct name.")
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
	//Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
	},
}

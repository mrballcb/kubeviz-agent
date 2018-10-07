package cmd

import (
  "fmt"
  "os"
  "github.com/spf13/cobra"
  "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
  Use:   "kubeviz-agent",
  Short: "Kubernetes visualization agent",
  Long: `Client (Kubernetes) side of kubeviz implementation. Documentation is available at https://github.com/bartlettc22/kubeviz-agent`,
  Run: func(cmd *cobra.Command, args []string) {
    // Do Stuff Here
  },
}

func init() {
  viper.SetEnvPrefix("KUBEVIZ")
  viper.AutomaticEnv()
}

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

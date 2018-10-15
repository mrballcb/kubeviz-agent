package cmd

import (
  "github.com/spf13/cobra"
  "github.com/spf13/viper"
  "github.com/bartlettc22/kubeviz-agent/agent"
)

var serverAddress string
var token string
var awsAccessKeyId string
var awsSecretAccessKey string
var clusterName string

func init() {
  startCmd.Flags().StringVarP(&serverAddress, "server-address", "s", "", "Kubeviz server address")
  viper.BindPFlag("server_address", startCmd.Flags().Lookup("server-address"))
  startCmd.Flags().StringVarP(&token, "token", "t", "", "server auth token")
  viper.BindPFlag("token", startCmd.Flags().Lookup("token"))
  startCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "Set cluster name")
  viper.BindPFlag("cluster_name", startCmd.Flags().Lookup("cluster-name"))
  rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
  Use:   "start",
  Short: "Start agent",
  Long:  `Starts the kubeviz agent`,
  Run: func(cmd *cobra.Command, args []string) {
    agent.Start(viper.GetString("server_address"), viper.GetString("token"), viper.GetString("cluster_name"))
  },
}

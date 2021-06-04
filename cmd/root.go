package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/t94j0/Mythic_CLI/util"
)

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./.env)")
}

func initConfig() {
	// nginx configuration
	viper.SetDefault("nginx_port", 7443)
	// mythic server configuration
	viper.SetDefault("documentation_host", "127.0.0.1")
	viper.SetDefault("documentation_port", 8090)
	viper.SetDefault("mythic_debug", false)
	viper.SetDefault("mythic_server_port", 17443)
	viper.SetDefault("mythic_server_host", "127.0.0.1")
	// postgres configuration
	viper.SetDefault("postgres_host", "127.0.0.1")
	viper.SetDefault("postgres_port", 5432)
	viper.SetDefault("postgres_db", "mythic_db")
	viper.SetDefault("postgres_user", "mythic_user")
	viper.SetDefault("postgres_password", util.GenerateRandomPassword(30))
	// rabbitmq configuration
	viper.SetDefault("rabbitmq_host", "127.0.0.1")
	viper.SetDefault("rabbitmq_port", 5672)
	viper.SetDefault("rabbitmq_user", "mythic_user")
	viper.SetDefault("rabbitmq_password", util.GenerateRandomPassword(30))
	viper.SetDefault("rabbitmq_vhost", "mythic_vhost")
	// jwt configuration
	viper.SetDefault("jwt_secret", util.GenerateRandomPassword(30))
	// hasura configuration
	viper.SetDefault("hasura_host", "127.0.0.1")
	viper.SetDefault("hasura_port", 8080)
	viper.SetDefault("hasura_secret", util.GenerateRandomPassword(30))
	// redis configuration
	viper.SetDefault("redis_port", 6379)
	// docker-compose configuration
	viper.SetDefault("COMPOSE_PROJECT_NAME", "mythic")
	// Mythic instance configuration
	viper.SetDefault("mythic_admin_user", "mythic_admin")
	viper.SetDefault("mythic_admin_password", util.GenerateRandomPassword(30))
	viper.SetDefault("default_operation_name", "Operation Chimera")
	viper.SetDefault("allowed_ip_blocks", "0.0.0.0/0")
	viper.SetDefault("server_header", "nginx 1.2")
	viper.SetDefault("web_log_size", 1024000)
	viper.SetDefault("web_keep_logs", false)
	viper.SetDefault("siem_log_name", "")
	viper.SetDefault("excluded_payload_types", "")
	viper.SetDefault("excluded_c2_profiles", "")

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(util.GetCwdFromExe())
	viper.AutomaticEnv()
	if !util.FileExists(filepath.Join(util.GetCwdFromExe(), ".env")) {
		_, err := os.Create(filepath.Join(util.GetCwdFromExe(), ".env"))
		if err != nil {
			log.Fatalf("[-] .env doesn't exist and couldn't be created")
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("[-] Error while reading in .env file: %s", err)
		} else {
			log.Fatalf("[-]Error while parsing .env file: %s", err)
		}
	}
}

var rootCmd = &cobra.Command{
	Use:     "mythic-cli",
	Short:   "",
	Version: "0.0.1",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

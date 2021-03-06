package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// SetupConfigSources sets up Viper configuration.
//
// If specific file path is provided but fails when loading it will return an error.
//
// In case of default config location it will not fail if file does not exist,
// but will in any other case such as parse error.
//
// Config precedence (each item takes precedence over the item below it):
// . Flags
// . Env variables
// . Config file
//
// Environment variables are prefixed with `IKE` and have fully qualified names, for example
// in case of `develop` command and its `port` flag corresponding environment variable is
// `IKE_DEVELOP_PORT`.
func SetupConfigSources(configFile string, notDefault bool) error {
	viper.Reset()
	viper.SetEnvPrefix("BINARY_NAME")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetTypeByDefaultValue(true)

	ext := path.Ext(configFile)
	viper.SetConfigName(strings.TrimSuffix(path.Base(configFile), ext))
	if !contains(SupportedExtensions(), strings.TrimPrefix(path.Ext(ext), ".")) {
		return fmt.Errorf("'%s' extension is not supported. Use one of [%s]", ext, strings.Join(SupportedExtensions(), ", "))
	}
	viper.SetConfigType(ext[1:])
	viper.AddConfigPath(path.Dir(configFile))

	if err := viper.ReadInConfig(); err != nil {
		if notDefault {
			return err
		}

		if _, fileDoesNotExist := err.(viper.ConfigFileNotFoundError); !fileDoesNotExist {
			return err
		}
	}
	return nil
}

// SupportedExtensions returns a slice of all supported config format (as file extensions)
func SupportedExtensions() []string {
	return viper.SupportedExts
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// SyncFlag ensures that if configuration provides a value for a given cmd.flag it will be set back to the flag itself,
// but only if the flag was not set through CLI.
//
// This way we can make flags required but still have their values provided by the configuration source
func SyncFlag(cmd *cobra.Command, flagName string) error { //nolint[:unused]
	value := viper.GetString(cmd.Name() + "." + flagName)
	if value != "" && !cmd.Flag(flagName).Changed {
		return cmd.Flags().Set(flagName, value)
	}
	return nil
}

// SyncFlags ensures that if configuration provide a value for any of defined flags it will be set
// back to the flag itself.
//
// This function iterates over all flags defined for cobra.Command and accumulates errors if they occur while
// calling SyncFlag for every flag.
func SyncFlags(cmd *cobra.Command) error { //nolint[:unused]
	var accErrors *multierror.Error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		syncFlagErr := SyncFlag(cmd, flag.Name)
		accErrors = multierror.Append(accErrors, syncFlagErr)
	})
	return accErrors.ErrorOrNil()
}

// BindFullyQualifiedFlag ensures that each flag used in commands is bound to a key using fully qualified name
// which has a following form:
//
// 		commandName.flagName
//
// This lets us  keep structure of yaml file:
//
//	commandName:
//		flagName: value
func BindFullyQualifiedFlag(cmd *cobra.Command) func(flag *pflag.Flag) { //nolint[:unused]
	return func(flag *pflag.Flag) {
		_ = viper.BindPFlag(cmd.Name()+"."+flag.Name, flag)
	}
}

package util

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

func GetEnv(names []string) {
	if len(names) == 0 {
		// we want to just get all of the environment variables that mythic uses
		c := viper.AllSettings()
		// to make it easier to read and look at, get all the keys, sort them, and display variable  s in order
		keys := make([]string, 0, len(c))
		for k := range c {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Println(strings.ToUpper(key), "=", viper.Get(key))
		}
		return
	}

	for _, name := range names {
		val := viper.Get(name)
		fmt.Println(strings.ToUpper(name), "=", val)
	}
}

func SetEnv(key, value string) {
	if strings.ToLower(value) == "true" {
		viper.Set(key, true)
	} else if strings.ToLower(value) == "false" {
		viper.Set(key, false)
	} else {
		viper.Set(key, value)
	}
	viper.Get(key)
	err := viper.WriteConfig()
	if err != nil {
		fmt.Printf("[-] Failed to write config: %v\n", err)
	} else {
		fmt.Printf("[+] Successfully updated configuration in .env\n")
	}
}

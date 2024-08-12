package utils

import (
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	AllowedOriginServiceAccount string `envconfig:"ALLOWED_ORIGIN_SERVICE_ACCOUNTS"`
	AllowedOriginImages         string `envconfig:"ALLOWED_ORIGIN_IMAGES""`
	AllowedTargetImages         string `envconfig:"ALLOWED_TARGET_IMAGES"`
	TargetServiceAccount        string `envconfig:"TARGET_SERVICE_ACCOUNT" default:"litmus-admin"`
	SelfManagedDependencies     bool   `envconfig:"SELF_MANAGED_DEPENDENCIES" default:"false"`
	ChaosNamespace              string `envconfig:"CHAOS_NAMESPACE" required:"true"`
}

type Filters struct {
	AllowedOriginServiceAccount Filter
	AllowedOriginImages         Filter
	AllowedTargetImages         Filter
	TargetServiceAccount        string
	SelfManagedDependencies     bool
	ChaosNamespace              string
}

type Filter struct {
	AllowedAll  bool
	AllowedList []string
}

func InitENV() error {
	var config Configuration
	// starting env config
	err := envconfig.Process("", &config)
	if err != nil {
		return err
	}
	WebHookFilters.AllowedOriginImages, err = parseAllowedList(config.AllowedOriginImages)
	if err != nil {
		return err
	}
	WebHookFilters.AllowedTargetImages, err = parseAllowedList(config.AllowedTargetImages)
	if err != nil {
		return err
	}
	WebHookFilters.AllowedOriginServiceAccount, err = parseAllowedList(config.AllowedOriginServiceAccount)
	if err != nil {
		return err
	}
	WebHookFilters.TargetServiceAccount = config.TargetServiceAccount
	WebHookFilters.SelfManagedDependencies = config.SelfManagedDependencies
	WebHookFilters.ChaosNamespace = config.ChaosNamespace
	return nil
}

var WebHookFilters Filters

func parseAllowedList(x string) (Filter, error) {
	if x == "" {
		return Filter{
			AllowedAll: true,
		}, nil
	}

	var (
		allowedList []string
	)

	if err := json.Unmarshal([]byte(x), &allowedList); err != nil {
		return Filter{}, fmt.Errorf("failed to unmarshal allowed list: %s", err.Error())
	}

	if SliceContains(allowedList, "*") {
		return Filter{
			AllowedAll: true,
		}, nil
	}

	return Filter{
		AllowedList: allowedList,
	}, nil
}

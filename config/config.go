package config

import (
	"github.com/spf13/viper"
)

/**
[
	 {"":"" ,"":"" , "":""},
]
**/

// struct to hold clients from config file
type clientAddr struct {
	InetAddr string  `json:"inet"`
	Rate     float64 `json:"rate"`
	Burst    int     `json:"burst"`
}

type config struct {
	ClientAddrs []clientAddr `json:"clients"`
	LogConfig   loging       `json:"logging"`
}

type loging struct {
	Level    string `json:"level"`
	Encoding string `json:"encoding"`
}

// loads client data to middleware
func LoadConfig(filepath string) (*config, error) {
	viper.AddConfigPath(filepath)
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	conf := &config{}
	err = viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil

}

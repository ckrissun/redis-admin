package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/revel/revel"
	"github.com/hoisie/redis"
)

type App struct {
	*revel.Controller
}

type Config struct {
	RedisHost string
	RedisPort int
}

type Ret struct {
	Keys []string
	CurKey string
	Value []string
}

var (
	config = &Config{}
	client redis.Client
)

func init() {
	if file, err := os.Open("./conf/custom.json"); err != nil {
		panic(err)
	} else {
		decoder := json.NewDecoder(file)
		decoder.Decode(config)
	}

	client.Addr = fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort)
}

func (c App) Index(key string) revel.Result {
	Keys, err := client.Keys("*")
	var ret Ret
	ret.Keys = Keys
	if err != nil {
		c.Flash.Error(err.Error())
	}

	if key != "" {
		ret.CurKey = key
		kind, err := client.Type(key)
		if err != nil {
			c.Flash.Error(err.Error())
		}
		switch string(kind) {
		case "string":
			val, _ := client.Get(key)
			ret.Value = []string{string(val)}
		case "list":
			vals, _ := client.Lrange(key, 0, -1)
			for _, v := range vals {
				ret.Value = append(ret.Value, string(v))
			}
		case "set":
			vals, _ := client.Smembers(key)
			for _, v := range vals {
				ret.Value = append(ret.Value, string(v))
			}
		case "zset":
			vals, _ := client.Zrange(key, 0, -1)
			for _, v := range vals {
				ret.Value = append(ret.Value, string(v))
			}
		case "hash":
			kvs := map[string][]byte{}
			client.Hgetall(key, &kvs)
			ret.Value = []string{}
			for k, v := range kvs {
				ret.Value = append(ret.Value, "key: " + string(k) + ", value: " + string(v))
			}
		}
	}

	return c.Render(ret)
}

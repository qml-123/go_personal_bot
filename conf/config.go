package conf

type Config struct {
	AppID      string `json:"app_id"`
	AppSecret  string `json:"app_secret"`
	EncryptKey string `json:"encrypt_key"`
}

type DbConfig struct {
	User string
	Pass string
	Addr string
	Port string
}

var DbConf *DbConfig
var Conf *Config

// replace conf when in use
const (
	appID      = "cli_a2df1267a038900d"
	secret     = "qRX3TA7xYunEoCjIeqS9OcYhgZos6Oje"
	encryptKey = "M80Sjcx0WNSIzxUhfeEX2fHF24NFtzZj"
)

func init() {
	Conf = &Config{
		AppID:      appID,
		AppSecret:  secret,
		EncryptKey: encryptKey,
	}
	DbConf = &DbConfig{
		User: "root",
		Pass: "136901",
		Addr: "127.0.0.1",
		Port: "3306",
	}
}

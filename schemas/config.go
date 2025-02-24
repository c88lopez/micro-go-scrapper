package schemas

type Config struct {
	Influx   InfluxConfig   `json:"influxDB"`
	RabbitMQ RabbitMQConfig `json:"rabbitMQ"`
	Postgres PostgresConfig `json:"postgres"`
	Provider ProviderConfig `json:"provider"`
	Telegram TelegramConfig `json:"telegram"`
}

type ProviderConfig struct {
	ProvidersCount    int    `json:"providersCount"`
	ScrapWorkersCount int    `json:"scrapWorkersCount"`
	Source            Source `json:"source"`
}

type Source struct {
	Domain string `json:"domain"`
	Url    string `json:"url"`
}

type PostgresConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Database string `json:"database"`
}

type RabbitMQConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

type InfluxConfig struct {
	Enabled  bool   `json:"enabled"`
	Url      string `json:"url"`
	Database string `json:"database"`
}

type TelegramConfig struct {
	Token string `json:"token"`
}

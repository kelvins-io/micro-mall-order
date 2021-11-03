package vars

type EmailConfigSettingS struct {
	Enable   bool   `json:"enable"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

type G2CacheSettingS struct {
	CacheDebug             bool
	CacheMonitor           bool
	OutCachePubSub         bool
	CacheMonitorSecond     int
	EntryLazyFactor        int
	GPoolWorkerNum         int
	GPoolJobQueueChanLen   int
	FreeCacheSize          int // 100MB
	PubSubRedisChannel     string
	RedisConfDSN           string
	RedisConfDB            int
	RedisConfPwd           string
	RedisConfMaxConn       int
	PubSubRedisConfDSN     string
	PubSubRedisConfDB      int
	PubSubRedisConfPwd     string
	PubSubRedisConfMaxConn int
}
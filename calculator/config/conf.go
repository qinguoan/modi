package config

var (
	DomainBindUrl = "http://loki.hy01.internal.wandoujia.com/ptree/api/nodes/domains"
	//TsdbHosts     = "sa-tsdb0-bgp2.hy01.wandoujia.com,sa-tsdb0-bgp1.hy01.wandoujia.com,sa-tsdb0-bgp0.hy01.wandoujia.com"
	TsdbHost    = "srefeed.hy01.internal.wandoujia.com"
	TsdbPort    = "4242"
	NatsServers = "sa-broker0-ct0.db01.wandoujia.com:4242,sa-broker0-ct1.db01.wandoujia.com:4242,sa-broker0-ct2.db01.wandoujia.com:4242,sa-broker0-cnc0.hlg01.wandoujia.com:4242,sa-broker0-cnc1.hlg01.wandoujia.com:4242,sa-broker0-cnc2.hlg01.wandoujia.com:4242"

	//Kafka Broker server
	KafkaServers = "kafka-cluster0-bgp0.hy01:9092,kafka-cluster0-bgp1.hy01:9092,kafka-cluster0-bgp2.hy01:9092,kafka-cluster0-bgp3.hy01:9092,kafka-cluster0-bgp4.hy01:9092,kafka-cluster0-bgp5.hy01:9092,kafka-cluster0-bgp6.hy01:9092,kafka-cluster0-bgp7.hy01:9092"
	KafkaTopicName = "sre_online_nginx_log"

	UrlCodeMetric     = "url.code"
	UrlTimeMetric     = "url.time" // need aggregate again.
	UrlUpstreamMetric = "url.upstreamtime"
	UrlTrafficMetric  = "url.bandwidth"
	UrlQpsMetric      = "url.qps"

	Code499Timeout  = 0.5 //second
	SubGroupName    = "sre-dasboard-subscriber"
	SubTopicName    = "*"
	TotalUrlMetric  = []string{UrlCodeMetric, UrlTimeMetric, UrlTrafficMetric, UrlUpstreamMetric, UrlQpsMetric}
	RequestTimeout  = 500
	UploadFrequency = 60                  // sencond, must be times of 60, or timestamp will be disordered.
	PathAggreNumber = 2 * UploadFrequency // total count of each path
	PathAggrelength = 50                  // total length of whole path
	PathLastlength  = 20                  // length of string after last slash.
	PathMaxDepth    = 4                   // max depth of each path.
	UrlDefaultTags  = []string{"path", "code", "upstream", "scheme", "domain", "source"}
	DomainIdleTime  = 1800 //second
)

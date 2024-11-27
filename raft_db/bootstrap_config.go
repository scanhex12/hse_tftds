package raft_db

type Config struct {
	is_leader bool
	url string
	other_node_urls []string
	ttls_milliseconds int
	period_milliseconds int
	max_master_silence_ms int
}

func NewConfig(is_leader bool,
	url string,
	other_node_urls []string,
	ttls_milliseconds int,
	period_milliseconds int,
	max_master_silence_ms int) *Config {
	return &Config{
		is_leader : is_leader,
		url : url,
		other_node_urls : other_node_urls,
		ttls_milliseconds : ttls_milliseconds,
		period_milliseconds : period_milliseconds,
		max_master_silence_ms: max_master_silence_ms,
	}
}

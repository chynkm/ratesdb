package redisdb

import (
	"strconv"
	"testing"
	"time"
)

func TestCreateKey(t *testing.T) {
	time := time.Now()
	ip := "::1"
	got := createKey(ip, time)
	want := api_user_prefix + ip + ":" + strconv.Itoa(time.Minute())

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

package gateway

import (
	"strings"
	"testing"
)

const (
	readOps  = `BITCOUNT BITPOS GET GETBIT GETRANGE MGET STRLEN SCARD SDIFF SINTER SISMEMBER SMEMBERS SRANDMEMBER SUNION SSCAN ZCARD ZCOUNT ZLEXCOUNT ZRANGE ZRANGEBYLEX ZREVRANGEBYLEX ZRANGEBYSCORE ZRANK ZREVRANGE ZREVRANGEBYSCORE ZREVRANK ZSCORE ZSCAN HEXISTS HGET HGETALL HKEYS HLEN HMGET HVALS HSTRLEN HSCAN LINDEX LLEN LRANGE GEOHASH GEOPOS GEODIST GEORADIUS GEORADIUSBYMEMBER DUMP EXISTS KEYS PTTL RANDOMKEY SORT TTL TYPE SCAN ECHO PING TIME`
	writeOps = `GEOADD HDEL HINCRBY HINCRBYFLOAT HMSET HSET HSETNX PFADD PFMERGE DEL EXPIRE EXPIREAT PERSIST PEXPIRE PEXPIREAT RENAME RENAMENX RESTORE LINSERT LPOP LPUSH LPUSHX LREM LSET LTRIM RPOP RPOPLPUSH RPUSH RPUSHX SADD SDIFFSTORE SINTERSTORE SMOVE SPOP SREM SUNIONSTORE ZADD ZINCRBY ZINTERSTORE ZREM ZREMRANGEBYLEX ZREMRANGEBYRANK ZREMRANGEBYSCORE ZUNIONSTORE APPEND BITFIELD BITOP DECR DECRBY INCR INCRBY INCRBYFLOAT MSET MSETNX PSETEX SET SETBIT SETEX SETNX SETRANGE`
	blackOps = `AUTH CLUSTER READONLY READWRITE MIGRATE MOVE OBJECT WAIT BLPOP BRPOP BRPOPLPUSH PUNSUBSCRIBE PUBSUB PUBLISH PUNSUBSCRIBE SUBSCRIBE UNSUBSCRIBE EVAL EVALSHA SCRIPT CLIENT COMMAND DBSIZE DEBUG BGWRITEAOF BGSAVE DEBUG CONFIG FLUSHALL FLUSHDB INFO LASTSAVE MONITOR ROLE SAVE SHUTDOWN SLAVEOF SLOWLOG SYNC GETSET DISCARD EXEC MULTI UNWATCH WATCH`
)

var g Gateway

func TestReadResolve(t *testing.T) {
	t.Parallel()

	for _, cmda := range strings.Split(readOps, " ") {
		cmd := strings.TrimSpace(cmda)

		if cmd == "" {
			continue
		}
		ty, _ := PylonRWResolver(cmd)
		if ty != Read {
			t.Errorf("command %s has been miscalculated as %d", cmd, ty)
			t.Fail()
			break
		}
	}

}

func TestWriteResolve(t *testing.T) {
	t.Parallel()

	for _, cmda := range strings.Split(writeOps, " ") {
		cmd := strings.TrimSpace(cmda)

		if cmd == "" {
			continue
		}
		ty, _ := PylonRWResolver(cmd)
		if ty != Write {
			t.Errorf("command %s has been miscalculated as %d", cmd, ty)
			t.Fail()
			break
		}
	}

}

func TestBlacklistResolve(t *testing.T) {
	t.Parallel()

	for _, cmda := range strings.Split(blackOps, " ") {
		cmd := strings.TrimSpace(cmda)

		if cmd == "" {
			continue
		}
		ty, _ := PylonRWResolver(cmd)
		if ty != Blacklist {
			t.Errorf("command %s has been miscalculated as %d", cmd, ty)
			t.Fail()
			break
		}
	}

}

func TestUnknownResolve(t *testing.T) {
	t.Parallel()

	ty, _ := PylonRWResolver("ZZZZZZZZZZ")
	if ty != Unknown {
		t.Fail()
	}

}

func BenchmarkReadResolve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PylonRWResolver("GET")
	}
}

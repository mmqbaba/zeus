package sequence

import (
	"bytes"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

/*
 * 产生不重复字符串
 * aaaaaaaabbbbbbbbccccdd 其中aaaaaaaa是服务 id，bbbbbbbb是当前时间戳, cccc是随机数, dd是这一秒内的计数
 * 测试方法:
 * go test -v
 * go test -bench=".*" -parallel 100000
 * cat result*.txt|sort -n |uniq -c | awk '{if($1 != 1){print $0}}' 没有输出说明没有生成重复的id
 * 在我的pc(2.7 GHz Intel Core i5/8 GB 1867 MHz DDR3)上压测结果如下：
 * 10000000	  160 ns/op   5.858s
 * 这种测试条件下，每秒可以产生171w+不重复数字,性能不俗
 */
const (
	PROJECT_ID_BITS = 8
	TIMESTAMP_BITS  = 32
	RAND_BITS       = 8
	COUNT_BITS      = 16
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}
var id uint32
var pool = make(chan uint64, 5)
var last_sec uint64 = 0
var last_count uint32 = 1

var serverId string

func init() {
	go gen()
}

func Load(srvid string) {
	serverId = srvid
	id = BKDRHash([]byte(srvid))
}

func GetServerId() string {
	return serverId
}

func gen() {
	for {
		current_sec := uint64(time.Now().Unix() & 0x00000000ffffffff)
		if current_sec != last_sec {
			last_count = 1
			last_sec = current_sec
		}
		c := uint64(last_sec << (RAND_BITS + COUNT_BITS))
		rand.Seed(time.Now().UnixNano())
		if last_count&0x0000000000ff0000 == 0 {
			c += rand.Uint64() & 0x0000000000ff0000
			c += uint64(last_count)
			last_count++
			pool <- c
		} else {
			c += uint64(last_count)
			last_count++
			pool <- c
		}
	}
}

func GetUUID() (string, error) {
	if c, ok := <-pool; !ok {
		return "", errors.ECodeUUIDErr.ParseErr("")
	} else {
		buffer := bufferPool.Get().(*bytes.Buffer)
		defer bufferPool.Put(buffer)
		buffer.Reset()
		buffer.WriteString(strconv.Itoa(int(id)))
		buffer.WriteString("-")
		buffer.WriteString(strconv.Itoa(int(c)))
		return buffer.String(), nil
	}
}

// BKDR Hash Function 64
func BKDRHash(str []byte) uint32 {
	var seed uint32 = 131 // 31 131 1313 13131 131313 etc..
	var hash uint32 = 0
	for i := 0; i < len(str); i++ {
		hash = hash*seed + uint32(str[i])
	}
	return (hash & 0x7FFFFFFF)
}

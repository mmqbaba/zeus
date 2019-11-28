package obs

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func setUpTifSignature(httpReq *http.Request) {
	paasId := zeusCfg.EBus.PaasId
	paasToken := zeusCfg.EBus.PaasToken
	now := time.Now()
	nonce := tifNonce(now)
	sign := tifSign(paasToken, now.Unix(), nonce)

	httpReq.Header.Set("x-tif-paasid", paasId)
	httpReq.Header.Set("x-tif-signature", sign)
	httpReq.Header.Set("x-tif-timestamp", fmt.Sprintf("%d", now.Unix()))
	httpReq.Header.Set("x-tif-nonce", nonce)
}

func tifNonce(now time.Time) string {
	r := rand.New(rand.NewSource(now.Unix()))
	str := fmt.Sprintf("%d_%d_%d", os.Getpid(), now.Unix(), r.Uint32())
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

func tifSign(secret string, now int64, nonce string) string {
	rawStr := fmt.Sprintf("%d%s%s%d", now, secret, nonce, now)
	return strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(rawStr))))
}

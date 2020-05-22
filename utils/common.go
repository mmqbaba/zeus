package utils

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func IsEmptyString(text string) bool {
	s := strings.TrimSpace(text)
	if len(s) > 0 {
		return false
	}

	return true
}

func GenerateRandString() string {
	strLen := 32
	codes := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	codeLen := len(codes)
	data := make([]byte, strLen)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < strLen; i++ {
		idx := rand.Intn(codeLen)
		data[i] = byte(codes[idx])
	}

	return string(data)
}

func HttpReq(ctx context.Context, method string, url string, buf []byte) []byte {
	client := http.DefaultClient
	httpreq, err := http.NewRequest(method, url, bytes.NewBuffer(buf))
	httpreq.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(httpreq)
	if err != nil {
		log.Println("forward err:", err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s", body)
	return body
}

// 正则过滤sql注入的方法
// 参数 : 要匹配的语句
func FilteredSQLInject(to_match_str string) bool {
	//过滤 ‘
	//关键字过滤 update ,delete
	// 正则的字符串, 不能用 " " 因为" "里面的内容会转义
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|sleep|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		panic(err.Error())
		return false
	}
	return re.MatchString(strings.ToLower(to_match_str))
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

//加密生成base64字符串
func AesEncrypt64(pass string) string {

	var aeskey = []byte("322423t0y8d2fwf5")

	xpass, err := AesEncrypt([]byte(pass), aeskey)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	pass64 := base64.StdEncoding.EncodeToString(xpass)
	return pass64
}

//解密生成原文
func AesDecrypt64(pass64 string) string {
	var aeskey = []byte("322423t0y8d2fwf5")
	bytesPass, err := base64.StdEncoding.DecodeString(pass64)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	tpass, err := AesDecrypt(bytesPass, aeskey)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(tpass)
}

const (
	KC_RAND_KIND_NUM   = 0 // 纯数字
	KC_RAND_KIND_LOWER = 1 // 小写字母
	KC_RAND_KIND_UPPER = 2 // 大写字母
	KC_RAND_KIND_ALL   = 3 // 数字、大小写字母
)

// 随机字符串
func Krand(size int, kind int) []byte {
	ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	is_all := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if is_all { // random ikind
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return result
}

// AsyncFuncSafe 异步执行函数
func AsyncFuncSafe(ctx context.Context, f func(args ...interface{}), args ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()
		f(args...)
	}()
}

//
func GenRandomStr() string {
	rndStr := fmt.Sprint(
		os.Getpid(), time.Now().UnixNano(), rand.Float64())
	h := md5.New()
	io.WriteString(h, rndStr)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ItoIP convert ip from uint32 to string, like 174343044 to 10.100.67.132, if fail return empty string: ""
func ItoIP(ip uint32) string {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, ip)
	if err != nil {
		return ""
	}

	b := buf.Bytes()
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

var localIP uint32

// GetLocalIP return local eth1 ip
func GetLocalIP() uint32 {
	if localIP != 0 {
		return localIP
	}
	localIP = GetIP("eth1")

	return localIP
}

// GetIP get local ip from inteface name like eth1
func GetIP(name string) uint32 {
	ifaces, err := net.Interfaces()
	if err != nil {
		return 0
	}

	for _, v := range ifaces {
		if v.Name == name {
			addrs, err := v.Addrs()
			if err != nil {
				return 0
			}

			for _, addr := range addrs {
				var ip net.IP
				switch val := addr.(type) {
				case *net.IPNet:
					ip = val.IP
				case *net.IPAddr:
					ip = val.IP
				}

				if len(ip) == 16 {
					return binary.BigEndian.Uint32(ip[12:16])
				} else if len(ip) == 4 {
					return binary.BigEndian.Uint32(ip)
				}
			}
		}
	}

	return 0
}

func DesensitizeIdNumber(idNumber string) string {
	if IsEmptyString(idNumber) {
		return ""
	}
	return idNumber[0:2] + "***************" + idNumber[17:18]
}

// IsBlank 判断空值
func IsBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

//mysql fields transfer tool

func GetFieldsForSql(model interface{}) (ret []string) {
	tmp := make([]string, 0)
	getType := reflect.TypeOf(model)
	for i := 0; i < getType.NumField(); i++ {
		f := getType.Field(i)
		val := f.Tag.Get("json")
		arr := strings.Split(val, ",")
		tmp = append(tmp, arr[0])
	}
	if len(tmp) > 0 {
		ret = tmp
	}
	return
}

func GetFieldMapForSql(model interface{}) (ret map[string]string) {
	tmp := make(map[string]string)
	getType := reflect.TypeOf(model)
	for i := 0; i < getType.NumField(); i++ {
		f := getType.Field(i)
		val := f.Tag.Get("json")
		arr := strings.Split(val, ",")
		tmp[f.Name] = strings.TrimSpace(arr[0])
	}
	if len(tmp) > 0 {
		ret = tmp
	}
	return
}

func FromModelsForSql(models []interface{}) (keys []string, vals []interface{}, holder []string) {
	if len(models) == 0 {
		return
	}

	fieldMap := GetFieldMapForSql(models[0])
	fields := []string{}
	fields4DB := []string{}
	for k, v := range fieldMap {
		fields = append(fields, k)
		fields4DB = append(fields4DB, v)
	}

	values4DB, placeHolder := getValue4DB(fields, models)

	return fields4DB, values4DB, placeHolder
}

func getValue4DB(fields []string, models []interface{}) (vals []interface{}, holder []string) {
	placeHolder := []string{}
	values4DB := []interface{}{}
	for _, item := range models {
		placeHolderTmp := []string{}
		getValue := reflect.ValueOf(item)
		for _, name := range fields {
			placeHolderTmp = append(placeHolderTmp, "?")
			val := getValue.FieldByName(name).Interface()
			values4DB = append(values4DB, val)
		}
		placeHolder = append(placeHolder, "("+strings.Join(placeHolderTmp, ",")+")")
	}
	return values4DB, placeHolder
}

func DBRowParseForSql(model interface{}, row map[string]interface{}) (err error) {
	var jsonD []byte
	if jsonD, err = json.Marshal(row); err != nil {
		return
	}
	if err = json.Unmarshal(jsonD, model); err != nil {
		return
	}
	return
}

//avoid value is null set db
func RowFieldValueToStringForSql(field string, row map[string]interface{}) string {
	if v, ok := row[field]; ok && v != nil {
		return v.(string)
	}
	return ""
}

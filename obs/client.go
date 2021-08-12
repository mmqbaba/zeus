package client

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"

	"gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	zeuserr "gitlab.dg.com/BackEnd/deliver/tif/zeus/errors"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/obs/obs"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/sequence"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/utils"
)

type lconfig struct {
	Endpoint   string
	AK         string
	SK         string
	BucketName string
	Location   string // "/a/b/c/"
}

type Client struct {
	Conf *lconfig
	C    *obs.ObsClient
	rw   sync.RWMutex
}

func (c *Client) PutObject(input *obs.PutObjectInput) (out *obs.PutObjectOutput, err error) {
	if input.Bucket == "" {
		// 默认BucketName
		input.Bucket = c.Conf.BucketName
	}
	input.Key = c.Conf.Location + input.Key
	return c.C.PutObject(input)
}

func (c *Client) GetObject(input *obs.GetObjectInput) (out *obs.GetObjectOutput, err error) {
	if input.Bucket == "" {
		// 默认BucketName
		input.Bucket = c.Conf.BucketName
	}
	input.Key = c.Conf.Location + input.Key
	return c.C.GetObject(input)
}

func (c *Client) GetObjectMeta(input *obs.GetObjectMetadataInput) (out *obs.GetObjectMetadataOutput, err error) {
	if input.Bucket == "" {
		// 默认BucketName
		input.Bucket = c.Conf.BucketName
	}
	input.Key = c.Conf.Location + input.Key
	return c.C.GetObjectMetadata(input)
}

var DefaultCli *Client

func New(appconf *config.AppConf) (c *Client, err error) {
	endpoint := appconf.Obs.Endpoint
	if utils.IsEmptyString(endpoint) {
		if len(appconf.EBus.Hosts) == 0 {
			err = errors.New("appconf.Ebus.Hosts was empty")
			log.Println(err)
			return
		}
		endpoint = appconf.EBus.Hosts[0]
	}
	if utils.IsEmptyString(endpoint) {
		err = errors.New("appconf.Ebus.Hosts and appconf.Obs.Endpoint was empty")
		log.Println(err)
		return
	}

	lconf := &lconfig{
		Endpoint:   endpoint,
		AK:         appconf.Obs.AK,
		SK:         appconf.Obs.SK,
		BucketName: appconf.Obs.BucketName,
		Location:   appconf.Obs.Location,
	}
	return newLocal(lconf)
}

func newLocal(lconf *lconfig) (c *Client, err error) {
	tmp := new(Client)
	tmp.Conf = lconf
	tmp.C, err = obs.New(lconf.AK, lconf.SK, lconf.Endpoint)
	if err != nil {
		log.Println(err)
		return
	}
	c = tmp
	return
}

var onceDefaultInit sync.Once

func InitDefault(appconf *config.AppConf) error {
	var err error
	onceDefaultInit.Do(func() {
		DefaultCli, err = New(appconf)
		if err != nil {
			log.Println(err)
			panic(err)
		}
	})
	obs.InitZeusCfg(appconf)
	log.Println("init default obs")
	return nil
}

func ReloadDefault(appconf *config.AppConf) (err error) {
	if DefaultCli == nil {
		log.Println("DefaultCli未初始化")
		return errors.New("DefaultCli未初始化")
	}
	DefaultCli.rw.Lock()
	defer DefaultCli.rw.Unlock()

	DefaultCli.C.Close()

	var tmp *Client
	tmp, err = New(appconf)
	if err != nil {
		log.Println(err)
		return
	}
	obs.InitZeusCfg(appconf)
	DefaultCli.Conf = tmp.Conf
	DefaultCli.C = tmp.C
	log.Println("reload default obs")
	return
}

func ReleaseDefault() (err error) {
	if DefaultCli == nil {
		log.Println("DefaultCli未初始化")
		err = errors.New("DefaultCli未初始化")
		return
	}
	DefaultCli.C.Close()
	DefaultCli = nil
	log.Println("release default obs")
	return
}

// Put 上传
func Put(ctx context.Context, data io.Reader, fid string) (key string, err error) {
	if DefaultCli == nil {
		log.Println("client未初始化")
		err = errors.New("client未初始化")
		return
	}

	var output *obs.PutObjectOutput
	input := &obs.PutObjectInput{}
	// input.Metadata = map[string]string{"meta": "value"}
	input.Body = data
	input.Key = fid

	if utils.IsEmptyString(input.Key) {
		input.Key, err = sequence.GetUUID()
	}
	if err != nil {
		log.Println(err)
		return
	}

	output, err = DefaultCli.PutObject(input)

	if err == nil {
		log.Printf("StatusCode:%d, RequestId:%s\n", output.StatusCode, output.RequestId)
		log.Printf("Key: %s, ETag:%s, StorageClass:%s\n", key, output.ETag, output.StorageClass)
		key = input.Key
	} else {
		if obsError, ok := err.(obs.ObsError); ok {
			log.Println(obsError.StatusCode)
			log.Println(obsError.Code)
			log.Println(obsError.Message)
		} else {
			log.Println(err)
		}
	}

	return
}

// Get 下载
func Get(ctx context.Context, key string) (data io.ReadCloser, err error) {
	if DefaultCli == nil {
		log.Println("DefaultCli未初始化")
		err = errors.New("DefaultCli未初始化")
		return
	}

	var output *obs.GetObjectOutput
	input := &obs.GetObjectInput{}
	input.Key = key
	output, err = DefaultCli.GetObject(input)
	if err == nil {
		log.Printf("StatusCode:%d, RequestId:%s\n", output.StatusCode, output.RequestId)
		log.Printf("StorageClass:%s, ETag:%s, ContentType:%s, ContentLength:%d, LastModified:%s\n", output.StorageClass, output.ETag, output.ContentType, output.ContentLength, output.LastModified)
		data = output.Body
	} else {
		if obsError, ok := err.(obs.ObsError); ok {
			log.Println(obsError.StatusCode)
			log.Println(obsError.Code)
			log.Println(obsError.Message)
			if obsError.StatusCode == 404 && obsError.Code == "NoSuchKey" {
				err = zeuserr.ECodeNoFile.ParseErr("")
				return
			}
		} else {
			log.Println(err)
		}
	}
	return
}

// GetMeta GetMeta
func GetMeta(ctx context.Context, key string) (data *obs.GetObjectMetadataOutput, err error) {
	if DefaultCli == nil {
		log.Println("client未初始化")
		err = errors.New("client未初始化")
		return
	}

	var output *obs.GetObjectMetadataOutput
	input := &obs.GetObjectMetadataInput{}
	input.Key = key
	output, err = DefaultCli.GetObjectMeta(input)
	if err == nil {
		log.Printf("StatusCode:%d, RequestId:%s\n", output.StatusCode, output.RequestId)
		log.Printf("StorageClass:%s, ETag:%s, ContentType:%s, ContentLength:%d, LastModified:%s\n", output.StorageClass, output.ETag, output.ContentType, output.ContentLength, output.LastModified)
		data = output
	} else {
		if obsError, ok := err.(obs.ObsError); ok {
			log.Println(obsError.StatusCode)
			log.Println(obsError.Code)
			log.Println(obsError.Message)
			if obsError.StatusCode == 404 {
				err = zeuserr.ECodeNoFile.ParseErr("")
				return
			}
		} else {
			log.Println(err)
		}
	}
	return
}

// // Put 上传
// func Put(input *obs.PutObjectInput) (key string, out *obs.PutObjectOutput, err error) {
// 	if DefaultCli == nil {
// 		fmt.Println("DefaultCli未初始化")
// 		err = errors.New("DefaultCli未初始化")
// 		return
// 	}
// 	if strings.TrimSpace(input.Key) == "" {
// 		var tifferr *tifferrors.TiffError
// 		input.Key, tifferr = sequence.GetUUID()
// 		if tifferr != nil {
// 			fmt.Println(tifferr)
// 			err = tifferr
// 			return
// 		}
// 	}

// 	out, err = DefaultCli.PutObject(input)
// 	if err != nil {
// 		return
// 	}
// 	key = input.Key
// 	return
// }

// // Get 下载
// func Get(input *obs.GetObjectInput) (out *obs.GetObjectOutput, err error) {
// 	if DefaultCli == nil {
// 		fmt.Println("DefaultCli未初始化")
// 		err = errors.New("DefaultCli未初始化")
// 		return
// 	}
// 	return DefaultCli.GetObject(input)
// }

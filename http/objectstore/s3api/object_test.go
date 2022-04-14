package s3api

import (
	"bytes"
	"fmt"
	"github.com/filedag-project/filedag-storage/http/objectstore/utils/testsign"
	"net/http"
	"testing"
)

const (
	DefaultTestAccessKey = "test"
	DefaultTestSecretKey = "test"
)

func TestS3ApiServer_PutObjectHandler(t *testing.T) {
	u := "/testbucket/1.txt"
	r1 := "123456"

	req := testsign.MustNewSignedV4Request(http.MethodPut, u, int64(len(r1)), bytes.NewReader([]byte(r1)), "s3", DefaultTestAccessKey, DefaultTestSecretKey, t)
	fmt.Println("put:", reqTest(req).String())

}

//func TestS3ApiServer_GetObjectHandler(t *testing.T) {
//	u := "http://127.0.0.1:9985/test/1.txt"
//	req := testsign.MustNewSignedV4Request(http.MethodGet, u, 0, nil, "s3", DefaultTestAccessKey, DefaultTestSecretKey, t)
//
//	//req.Header.Set("Content-Type", "text/plain")
//	client := &http.Client{}
//	res, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer res.Body.Close()
//	body, err := ioutil.ReadAll(res.Body)
//
//	fmt.Println(res)
//	fmt.Println(string(body))
//
//}
//func TestS3ApiServer_CopyObjectHandler(t *testing.T) {
//	u := "http://127.0.0.1:9985/test1/11.txt"
//	req := testsign.MustNewSignedV4Request(http.MethodPut, u, 0, nil, "s3", DefaultTestAccessKey, DefaultTestSecretKey, t)
//	req.Header.Set("X-Amz-Copy-Source", url.QueryEscape("/test/1.txt"))
//	//req.Header.Set("Content-Type", "text/plain")
//	client := &http.Client{}
//	res, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer res.Body.Close()
//	body, err := ioutil.ReadAll(res.Body)
//
//	fmt.Println(res)
//	fmt.Println(string(body))
//
//}
//func TestS3ApiServer_HeadObjectHandler(t *testing.T) {
//	u := "http://127.0.0.1:9985/test/1.txt"
//	req := testsign.MustNewSignedV4Request(http.MethodHead, u, 0, nil, "s3", DefaultTestAccessKey, DefaultTestSecretKey, t)
//
//	//req.Header.Set("Content-Type", "text/plain")
//	client := &http.Client{}
//	res, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer res.Body.Close()
//	body, err := ioutil.ReadAll(res.Body)
//
//	fmt.Println(res)
//	fmt.Println(string(body))
//
//}
//func TestS3ApiServer_ListObjectsV1Handler(t *testing.T) {
//	u := "http://127.0.0.1:9985/test22"
//	req := testsign.MustNewSignedV4Request(http.MethodGet, u, 0, nil, "s3", DefaultTestAccessKey, DefaultTestSecretKey, t)
//
//	//req.Header.Set("Content-Type", "text/plain")
//	client := &http.Client{}
//	res, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer res.Body.Close()
//	body, err := ioutil.ReadAll(res.Body)
//
//	fmt.Println(res)
//	fmt.Println(string(body))
//
//}

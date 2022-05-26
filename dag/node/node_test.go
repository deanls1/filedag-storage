package node

import (
	"encoding/json"
	"fmt"
	"github.com/filedag-project/filedag-storage/dag/config"
	"github.com/filedag-project/filedag-storage/http/objectstore/utils"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestDagNode(t *testing.T) {
	var nc config.NodeConfig
	file, err := ioutil.ReadFile("./node_config2.json")
	if err != nil {

	}
	err = json.Unmarshal(file, &nc)
	dagNode, err := NewDagNode(nc)
	//time.Sleep(time.Millisecond * 50)
	//os.Setenv(mutcask.Host, "127.0.0.1")
	//os.Setenv(mutcask.Port, "9011")
	//os.Setenv(mutcask.Path, utils.TmpDirPath(t))
	//go mutcask.MutServer()
	data, err := ioutil.ReadFile("./node.go")
	aa := cid.Cid{}
	b, err := blocks.NewBlockWithCid(data, aa)
	if err == blocks.ErrWrongHash {
		fmt.Println(err)
	}
	err = dagNode.Put(b)
	if err != nil {
		fmt.Println(err)
	}
	dataBlock, err := dagNode.Get(aa)
	fmt.Println(dataBlock)
	if err != nil {
		fmt.Println(err)
	}
	err = dagNode.DeleteBlock(aa)
	if err != nil {
		fmt.Println(err)
	}
	dataBlock, err = dagNode.Get(aa)
	fmt.Println(dataBlock)
}
func TestNewDagNode(t *testing.T) {
	var nc config.NodeConfig
	file, err := ioutil.ReadFile("./node_config2.json")
	if err != nil {

	}
	err = json.Unmarshal(file, &nc)
	NewDagNode(nc)
	time.Sleep(time.Millisecond * 50)
	os.Setenv(Host, "127.0.0.1")
	os.Setenv(Port, "9011")
	os.Setenv(Path, utils.TmpDirPath(t))
	go MutServer()
	time.Sleep(time.Second * 20)
}

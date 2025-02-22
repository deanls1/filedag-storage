package dagnode

import (
	"bytes"
	"context"
	"fmt"
	"github.com/filedag-project/filedag-storage/dag/config"
	"github.com/filedag-project/filedag-storage/dag/proto"
	"github.com/filedag-project/filedag-storage/http/objectstore/uleveldb"
	"github.com/filedag-project/filedag-storage/kv"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	format "github.com/ipfs/go-ipld-format"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"strings"
	"sync"
)

var _ blockstore.Blockstore = (*DagNode)(nil)

//DagNode Implemented the Blockstore interface
type DagNode struct {
	Nodes        []*DataNodeClient
	db           *uleveldb.ULevelDB
	dataBlocks   int // Number of data shards
	parityBlocks int //Number of parity shards
}

//DataNodeClient is a node that stores erasure-coded sharded data
type DataNodeClient struct {
	Client      proto.DataNodeClient
	HeartClient healthpb.HealthClient
	Ip          string
	Port        string
	Conn        *grpc.ClientConn
}

//NewDagNode creates a new DagNode
func NewDagNode(cfg config.DagNodeConfig) (*DagNode, error) {
	var s []*DataNodeClient
	for _, c := range cfg.Nodes {
		dateNode, err := NewDataNodeClient(c)
		if err != nil {
			return nil, err
		}
		s = append(s, dateNode)
	}
	db, _ := uleveldb.OpenDb(cfg.LevelDbPath)
	return &DagNode{s, db, cfg.DataBlocks, cfg.ParityBlocks}, nil
}

//NewDataNodeClient creates a grpc connection to a slice
func NewDataNodeClient(cfg config.DataNodeConfig) (datanode *DataNodeClient, err error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", cfg.Ip, cfg.Port), grpc.WithInsecure())
	if err != nil {
		log.Errorf("did not connect: %v", err)
		return nil, err
	}
	datanode = &DataNodeClient{
		Client:      proto.NewDataNodeClient(conn),
		HeartClient: healthpb.NewHealthClient(conn),
		Ip:          cfg.Ip,
		Port:        cfg.Port,
		Conn:        conn,
	}
	return datanode, nil
}

//func (d DagNode) GetIP() []string {
//	var s []string
//	for _, n := range d.Nodes {
//		s = append(s, n.Ip)
//	}
//	return s
//}

//DeleteBlock deletes a block from the DagNode
func (d DagNode) DeleteBlock(ctx context.Context, cid cid.Cid) (err error) {
	log.Warnf("delete block, cid :%v", cid)
	keyCode := cid.String()
	wg := sync.WaitGroup{}
	wg.Add(len(d.Nodes))
	for _, node := range d.Nodes {
		go func(node *DataNodeClient) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("%s:%s, keyCode:%s, delete block err :%v", node.Ip, node.Port, keyCode, err)
				}
				wg.Done()
			}()
			_, err = node.Client.Delete(ctx, &proto.DeleteRequest{Key: keyCode})
			if err != nil {
				log.Debugf("%s:%s, keyCode:%s, delete block err :%v", node.Ip, node.Port, keyCode, err)
			}
		}(node)
	}
	wg.Wait()
	return err
}

//Has returns true if the given cid is in the DagNode
func (d DagNode) Has(ctx context.Context, cid cid.Cid) (bool, error) {
	_, err := d.GetSize(ctx, cid)
	if err != nil {
		if strings.Contains(err.Error(), kv.ErrNotFound.Error()) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

//Get returns the block with the given cid
func (d DagNode) Get(ctx context.Context, cid cid.Cid) (blocks.Block, error) {
	log.Debugf("get block, cid :%v", cid)
	keyCode := cid.String()
	var err error
	var size int
	err = d.db.Get(cid.String(), &size)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, format.ErrNotFound{Cid: cid}
		}
		return nil, err
	}
	merged := make([][]byte, len(d.Nodes))
	wg := sync.WaitGroup{}
	wg.Add(len(d.Nodes))
	for i, node := range d.Nodes {
		go func(i int, node *DataNodeClient) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("%s:%s, keyCode:%s, kvdb get err :%v", node.Ip, node.Port, keyCode, err)
				}
				wg.Done()
			}()
			res, err := node.Client.Get(ctx, &proto.GetRequest{Key: keyCode})
			if err != nil {
				log.Errorf("%s:%s, keyCode:%s,kvdb get :%v", node.Ip, node.Port, keyCode, err)
				merged[i] = nil
			} else {
				merged[i] = res.DataBlock
			}
		}(i, node)
	}
	wg.Wait()
	enc, err := NewErasure(d.dataBlocks, d.parityBlocks, int64(size))
	if err != nil {
		log.Errorf("new erasure fail :%v", err)
		return nil, err
	}
	err = enc.DecodeDataBlocks(merged)
	if err != nil {
		log.Errorf("decode date blocks fail :%v", err)
		return nil, err
	}
	var data []byte
	data = bytes.Join(merged, []byte(""))
	data = data[:size]
	b, err := blocks.NewBlockWithCid(data, cid)
	if err == blocks.ErrWrongHash {
		return nil, blockstore.ErrHashMismatch
	}
	return b, err
}

//GetSize returns the size of the block with the given cid
func (d DagNode) GetSize(ctx context.Context, cid cid.Cid) (int, error) {
	keyCode := cid.String()
	var err error
	var count int64
	for _, node := range d.Nodes {
		size, err := node.Client.Size(ctx, &proto.SizeRequest{
			Key: keyCode,
		})
		if err != nil {
			return 0, err
		}
		count = count + size.Size
	}
	return int(count), err
}

//Put adds the given block to the DagNode
func (d DagNode) Put(ctx context.Context, block blocks.Block) (err error) {
	log.Debugf("put block, cid :%v", block.Cid())
	// copy data from block, because reedsolomon may modify data
	buf := bytes.NewBuffer(nil)
	buf.Write(block.RawData())
	blockData := buf.Bytes()
	blockDataSize := len(blockData)
	keyCode := block.Cid().String()
	//todo store this info in datanode
	err = d.db.Put(keyCode, blockDataSize)
	if err != nil {
		return err
	}
	enc, err := NewErasure(d.dataBlocks, d.parityBlocks, int64(blockDataSize))
	if err != nil {
		log.Errorf("newErasure fail :%v", err)
		return err
	}
	shards, err := enc.EncodeData(blockData)
	if err != nil {
		log.Errorf("encodeData fail :%v", err)
		return err
	}
	ok, err := enc.encoder().Verify(shards)
	if err != nil {
		log.Errorf("encode fail :%v", err)
		return err
	}
	if ok {
		//log.Debugf("encode ok, the data is the same format as Encode. No data is modified")
	}
	wg := sync.WaitGroup{}
	wg.Add(len(d.Nodes))
	for i, node := range d.Nodes {
		go func(i int, node *DataNodeClient) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("%s:%s,keyCode:%s,kvdb put :%v", node.Ip, node.Port, keyCode, err)
				}
				wg.Done()
			}()
			_, err = node.Client.Put(ctx, &proto.AddRequest{Key: keyCode, DataBlock: shards[i]})
			if err != nil {
				log.Errorf("%s:%s,keyCode:%s,kvdb put :%v", node.Ip, node.Port, keyCode, err)
			}
		}(i, node)
	}
	wg.Wait()
	return err
}

//PutMany adds the given blocks to the DagNode
func (d DagNode) PutMany(ctx context.Context, blocks []blocks.Block) (err error) {
	for _, block := range blocks {
		err = d.Put(ctx, block)
	}
	return err
}

//AllKeysChan returns a channel that will yield every key in the dag
func (d DagNode) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	panic("implement me")
}

//HashOnRead tells the dag node to calculate the hash of the block
func (d DagNode) HashOnRead(enabled bool) {
	panic("implement me")
}

func (d *DagNode) Close() {
	for _, nd := range d.Nodes {
		nd.Conn.Close()
	}
}

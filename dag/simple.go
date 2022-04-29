package pool

import (
	"context"
	"github.com/filedag-project/filedag-storage/dag/node"
	"github.com/filedag-project/filedag-storage/dag/pool"
	"github.com/filedag-project/filedag-storage/dag/pool/config"
	"github.com/filedrive-team/filehelper"
	"io"

	"github.com/filedrive-team/filehelper/importer"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	"github.com/ipfs/go-merkledag"
	ufsio "github.com/ipfs/go-unixfs/io"
	"golang.org/x/xerrors"
)

const importerBatchNum = 32
const UnixfsLinksPerLevel = 1 << 10
const UnixfsChunkSize uint64 = 1 << 20

var _ DAGPool = (*simplePool)(nil)

type simplePool struct {
	bs               blockstore.Blockstore
	dagserv          *pool.DagPool
	cidBuilder       cid.Builder
	importerBatchNum int
}

func NewSimplePool(cfg *config.SimplePoolConfig) (*simplePool, error) {

	if cfg.StorePath == "" {
		return nil, xerrors.New("Need path to set store up for dag pool")
	}
	if cfg.BatchNum == 0 {
		cfg.BatchNum = importerBatchNum
	}

	bs, err := node.NewDagNode(&node.Config{
		CaskNum: cfg.CaskNum,
		Batch:   cfg.BatchNum,
		Path:    cfg.StorePath,
	})
	if err != nil {
		return nil, err
	}
	cidBuilder, err := merkledag.PrefixForCidVersion(0)
	if err != nil {
		return nil, err
	}
	sp := &simplePool{
		bs:               bs,
		dagserv:          pool.NewDagPoolService(blockservice.New(bs, offline.Exchange(bs))),
		cidBuilder:       cidBuilder,
		importerBatchNum: cfg.BatchNum,
	}
	return sp, nil
}

func (p *simplePool) Add(ctx context.Context, r io.ReadCloser, user, pass string) (cidstr string, err error) {
	nd, err := filehelper.BalanceNode(r, p.dagserv, p.cidBuilder)
	if err != nil {
		return "", err
	}
	return nd.String(), nil
}

func (p *simplePool) AddWithSize(ctx context.Context, r io.ReadCloser, fsize int64, user, pass string) (cidstr string, err error) {

	ndcid, err := importer.BalanceNode(ctx, r, fsize, p.dagserv, p.cidBuilder, p.importerBatchNum)
	if err != nil {
		return "", err
	}
	return ndcid.String(), nil
}

func (p *simplePool) Get(ctx context.Context, cidstr string, user, pass string) (r io.ReadSeekCloser, err error) {
	cid, err := cid.Decode(cidstr)
	if err != nil {
		return nil, err
	}
	dagNode, err := p.dagserv.Get(ctx, cid)
	if err != nil {
		return nil, err
	}
	return ufsio.NewDagReader(ctx, dagNode, p.dagserv)
}

package referencecount

import (
	"errors"
	"github.com/filedag-project/filedag-storage/http/objectstore/uleveldb"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"strconv"
	"strings"
	"sync"
	"time"
)

var log = logging.Logger("refer-count")

//ReferSys reference sys
type ReferSys struct {
	cacheMu sync.RWMutex
	storeMu sync.RWMutex
	DB      *uleveldb.ULevelDB
}

const gcExpiredTime = time.Minute * 1
const gcTime = time.Minute
const dagPoolReferCache = "dagPoolReferCache/"
const dagPoolReferPin = "dagPoolReferPin/"

//AddReference add refer for block
func (i *ReferSys) AddReference(cid string, isPin bool) error {
	if !isPin {
		i.cacheMu.Lock()
		ti := time.Now().Unix()
		err := i.DB.Put(dagPoolReferCache+cid, ti)
		i.cacheMu.Unlock()
		if err != nil {
			return err
		}
	} else {
		var count int64
		i.storeMu.Lock()
		err := i.DB.Get(dagPoolReferPin+cid, &count)
		count++
		err = i.DB.Put(dagPoolReferPin+cid, count)
		i.storeMu.Unlock()
		if err != nil {
			return err
		}
	}
	return nil
}

//QueryReference query block refer
func (i *ReferSys) QueryReference(cid string, isPin bool) (uint64, error) {
	if !isPin {
		ti := 0
		i.cacheMu.RLock()
		err := i.DB.Get(dagPoolReferCache+cid, &ti)
		i.cacheMu.RUnlock()
		if err != nil {
			return 0, err
		}
		if ti != 0 {
			return 1, nil
		}
		return 0, errors.New("no record")

	} else {
		var count uint64
		i.storeMu.RLock()
		err := i.DB.Get(dagPoolReferPin+cid, &count)
		i.storeMu.RUnlock()
		if err != nil {
			return 0, err
		}
		return count, nil
	}
}

//HasReference has block reference
func (i *ReferSys) HasReference(cid string) bool {
	ti := 0
	i.cacheMu.RLock()
	err := i.DB.Get(dagPoolReferCache+cid, &ti)
	i.cacheMu.RUnlock()
	if err == nil && ti != 0 {
		return true
	} else {
		var count uint64
		i.storeMu.RLock()
		err := i.DB.Get(dagPoolReferPin+cid, &count)
		i.storeMu.RUnlock()
		if err == nil && count != 0 {
			return true
		}
		return false
	}
}

//RemoveReference reduce refer
func (i *ReferSys) RemoveReference(cid string, isPin bool) error {
	if isPin {
		var count int
		i.storeMu.Lock()
		err := i.DB.Get(dagPoolReferPin+cid, &count)
		if count == 0 {
			return errors.New("cid does not exist")
		} else if count >= 1 {
			count--
			err = i.DB.Put(dagPoolReferPin+cid, count)
		} else {
			return errors.New("cid does not exist")
		}
		i.storeMu.Unlock()
		if err != nil {
			return err
		}
	}
	return nil
}

//QueryAllCacheReference query all cache refer record
func (i *ReferSys) QueryAllCacheReference() ([]cid.Cid, error) {
	i.cacheMu.RLock()
	defer i.cacheMu.RUnlock()
	all, err := i.DB.ReadAll(dagPoolReferCache)
	if err != nil {
		return nil, err
	}
	var m []cid.Cid
	for k, v := range all {
		tmp, _ := strconv.ParseInt(v, 10, 64)

		if time.Now().After(time.Unix(tmp, 0).Add(gcExpiredTime)) {
			c, _ := cid.Decode(strings.Split(k, "/")[1])
			m = append(m, c)
		}

	}
	return m, nil
}

//QueryAllStoreNonRefer query all store refer which count 0
func (i *ReferSys) QueryAllStoreNonRefer() ([]cid.Cid, error) {
	i.storeMu.RLock()
	defer i.storeMu.RUnlock()
	all, err := i.DB.ReadAll(dagPoolReferPin)
	if err != nil {
		return nil, err
	}
	var m []cid.Cid
	for k, v := range all {
		if v == "0" {
			c, _ := cid.Decode(strings.Split(k, "/")[1])
			m = append(m, c)
		}
	}
	return m, nil
}

//RemoveRecord remove record in db
func (i *ReferSys) RemoveRecord(c string, pin bool) error {
	i.cacheMu.Lock()
	defer i.cacheMu.Unlock()
	var path = dagPoolReferCache
	if pin {
		path = dagPoolReferPin
	}

	err := i.DB.Delete(path + c)
	if err != nil {
		return err
	}

	return nil
}

//NewIdentityRefe new a reference sys
func NewIdentityRefe(db *uleveldb.ULevelDB) *ReferSys {
	return &ReferSys{DB: db}
}

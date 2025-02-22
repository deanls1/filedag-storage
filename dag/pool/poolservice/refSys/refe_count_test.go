package refSys

import (
	"fmt"
	"github.com/filedag-project/filedag-storage/http/objectstore/uleveldb"
	"testing"
)

func TestIdentityRefer(t *testing.T) {
	db, _ := uleveldb.OpenDb(t.TempDir())
	identityRefe := ReferSys{DB: db}
	cid := "123456789"
	testCases := []struct {
		isRemove bool
		cid      string
	}{
		// Test case - 1.
		{
			isRemove: true,
			cid:      "123456789",
		},
		// Test case - 2.
		{
			isRemove: true,
			cid:      "123456789",
		},
		// Test case - 3.
		{
			isRemove: false,
			cid:      "123456789",
		},
	}
	for _, testCase := range testCases {
		err := identityRefe.AddReference(testCase.cid)
		if err != nil {
			fmt.Println(err)
		}
	}
	count, err := identityRefe.queryReference(cid)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(count)
	for _, testCase := range testCases {
		err = identityRefe.RemoveReference(testCase.cid)
		if err != nil {
			fmt.Println(err)
		}
	}
	err = identityRefe.RemoveReference(cid)
	if err != nil {
		fmt.Println(err)
	}
}

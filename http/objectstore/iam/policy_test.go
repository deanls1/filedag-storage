package iam

import (
	"context"
	"fmt"
	"github.com/filedag-project/filedag-storage/http/objectstore/iam/auth"
	"github.com/filedag-project/filedag-storage/http/objectstore/iam/policy"
	"github.com/filedag-project/filedag-storage/http/objectstore/iam/s3action"
	"github.com/filedag-project/filedag-storage/http/objectstore/uleveldb"
	"github.com/filedag-project/filedag-storage/http/objectstore/utils"
	"testing"
)

func TestPolicySys_IsAllowed(t *testing.T) {
	db, _ := uleveldb.OpenDb(utils.TmpDirPath(&testing.T{}))
	iamSys := NewIdentityAMSys(db)
	//poli := NewIPolicySys(db)
	initSys()
	if iamSys.IsAllowed(context.Background(), auth.Args{
		AccountName: auth.DefaultAccessKey,
		Action:      "list",
		BucketName:  "test",
		ObjectName:  "test",
		IsOwner:     false,
	}) {
		// Request is allowed return the appropriate access key.
		fmt.Println(true)
	}
}

func initSys() {
	var states []policy.Statement

	ast := s3action.NewActionSet("list")
	principal := policy.NewPrincipal(auth.DefaultAccessKey)
	resource := policy.NewResourceSet()
	states = append(states, policy.NewStatement("1", policy.Allow, principal, ast, resource))

}

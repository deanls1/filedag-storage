package policy

import (
	"encoding/json"
	"github.com/filedag-project/filedag-storage/http/objectstore/iam/auth"
	"github.com/filedag-project/filedag-storage/http/objectstore/iam/s3action"
	"golang.org/x/xerrors"
	"io"
)

// Policy - iam bucket iamp.
type Policy struct {
	ID         ID `json:"ID,omitempty"`
	Version    string
	Statements []Statement `json:"Statement"`
}
type PolicyDocument struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

// Merge merges two policies documents and drop
// duplicate statements if any.
func (p *PolicyDocument) Merge(input PolicyDocument) PolicyDocument {
	var mergedPolicy PolicyDocument
	for _, st := range p.Statement {
		mergedPolicy.Statement = append(mergedPolicy.Statement, st.Clone())
	}
	for _, st := range input.Statement {
		mergedPolicy.Statement = append(mergedPolicy.Statement, st.Clone())
	}
	mergedPolicy.dropDuplicateStatements()
	return mergedPolicy
}
func (p *PolicyDocument) dropDuplicateStatements() {
redo:
	for i := range p.Statement {
		for _, statement := range p.Statement[i+1:] {
			if !p.Statement[i].Equals(statement) {
				continue
			}
			p.Statement = append(p.Statement[:i], p.Statement[i+1:]...)
			goto redo
		}
	}
}

type Policies struct {
	Policies map[string]PolicyDocument `json:"policies"`
}

func (p PolicyDocument) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// IsAllowed - checks given policy args is allowed to continue the Rest API.
func (p Policy) IsAllowed(args auth.Args) bool {
	// Check all deny statements. If any one statement denies, return false.
	for _, statement := range p.Statements {
		if statement.Effect == Deny {
			if !statement.IsAllowed(args) {
				return false
			}
		}
	}

	// For owner, it allowed by default.
	if args.IsOwner {
		return true
	}

	// Check all allow statements. If anyone statement allows, return true.
	for _, statement := range p.Statements {
		if statement.Effect == Allow {
			if statement.IsAllowed(args) {
				return true
			}
		}
	}

	return false
}

// ParseConfig - parses data in given reader to Policy.
func ParseConfig(reader io.Reader, bucketName string) (*Policy, error) {
	var policy Policy

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&policy); err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	err := policy.Validate(bucketName)
	return &policy, err
}

// Validate - validates all statements are for given bucket or not.
func (p Policy) Validate(bucketName string) error {
	if err := p.isValid(); err != nil {
		return err
	}

	for _, statement := range p.Statements {
		if err := statement.Validate(bucketName); err != nil {
			return err
		}
	}

	return nil
}

// Merge merges two policies documents and drop
// duplicate statements if any.
func (p *Policy) Merge(input Policy) Policy {
	var mergedPolicy Policy
	for _, st := range p.Statements {
		mergedPolicy.Statements = append(mergedPolicy.Statements, st.Clone())
	}
	for _, st := range input.Statements {
		mergedPolicy.Statements = append(mergedPolicy.Statements, st.Clone())
	}
	mergedPolicy.dropDuplicateStatements()
	return mergedPolicy
}
func (p *Policy) dropDuplicateStatements() {
redo:
	for i := range p.Statements {
		for _, statement := range p.Statements[i+1:] {
			if !p.Statements[i].Equals(statement) {
				continue
			}
			p.Statements = append(p.Statements[:i], p.Statements[i+1:]...)
			goto redo
		}
	}
}

// Equals returns true if the two policies are identical
func (p *Policy) Equals(policy Policy) bool {
	if p.ID != policy.ID {
		return false
	}
	if len(p.Statements) != len(policy.Statements) {
		return false
	}
	for i, st := range policy.Statements {
		if !p.Statements[i].Equals(st) {
			return false
		}
	}
	return true
}

// IsEmpty - returns whether policy is empty or not.
func (p Policy) IsEmpty() bool {
	return len(p.Statements) == 0
}

// isValid - checks if Policy is valid or not.
func (p Policy) isValid() error {

	for _, statement := range p.Statements {
		if err := statement.IsValid(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultPolicies - list of canned policies available in FileDagStorage.
var DefaultPolicies = []struct {
	Name       string
	Definition Policy
}{
	// ReadWrite - provides full access to all buckets and all objects.
	{
		Name: "readwrite",
		Definition: Policy{
			Statements: []Statement{
				{
					SID:       "",
					Effect:    Allow,
					Principal: NewPrincipal("*"),
					Actions:   s3action.SupportedActions,
					Resources: NewResourceSet(NewResource("*", "*")),
				},
			},
		},
	},
	// ReadWrite - provides full access to all buckets and all objects.
	{
		Name: "default",
		Definition: Policy{
			Statements: []Statement{
				{
					SID:       "",
					Effect:    Allow,
					Principal: NewPrincipal("*"),
					Actions:   s3action.SupportedActions,
					Resources: NewResourceSet(NewResource("*", "")),
				},
			},
		},
	},

	// ReadOnly - read only.
	{
		Name: "readonly",
		Definition: Policy{
			Statements: []Statement{
				{
					SID:     "",
					Effect:  Allow,
					Actions: s3action.NewActionSet(s3action.GetBucketLocationAction, s3action.GetObjectAction),
				},
			},
		},
	},

	// WriteOnly - provides write access.
	{
		Name: "writeonly",
		Definition: Policy{

			Statements: []Statement{
				{
					SID:     "",
					Effect:  Allow,
					Actions: s3action.NewActionSet(s3action.PutObjectAction),
				},
			},
		},
	},
}

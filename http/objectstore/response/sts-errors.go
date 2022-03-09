package response

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/filedag-project/filedag-storage/http/objectstore/api_errors"
	"github.com/filedag-project/filedag-storage/http/objectstore/consts"
	"github.com/filedag-project/filedag-storage/http/objectstore/iam/auth"
	"net/http"
)

// AssumedRoleUser - The identifiers for the temporary security credentials that
// the operation returns. Please also see https://docs.aws.amazon.com/goto/WebAPI/sts-2011-06-15/AssumedRoleUser
type AssumedRoleUser struct {
	// The ARN of the temporary security credentials that are returned from the
	// AssumeRole action. For more information about ARNs and how to use them in
	// policies, see IAM Identifiers (http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html)
	// in Using IAM.
	//
	// Arn is a required field
	Arn string

	// A unique identifier that contains the role ID and the role session name of
	// the role that is being assumed. The role ID is generated by AWS when the
	// role is created.
	//
	// AssumedRoleId is a required field
	AssumedRoleID string `xml:"AssumeRoleId"`
	// contains filtered or unexported fields
}

// AssumeRoleResult - Contains the response to a successful AssumeRole
// request, including temporary credentials that can be used to make
// MinIO API requests.
type AssumeRoleResult struct {
	// The identifiers for the temporary security credentials that the operation
	// returns.
	AssumedRoleUser AssumedRoleUser `xml:",omitempty"`

	// The temporary security credentials, which include an access key ID, a secret
	// access key, and a security (or session) token.
	//
	// Note: The size of the security token that STS APIs return is not fixed. We
	// strongly recommend that you make no assumptions about the maximum size. As
	// of this writing, the typical size is less than 4096 bytes, but that can vary.
	// Also, future updates to AWS might require larger sizes.
	Credentials auth.Credentials `xml:",omitempty"`

	// A percentage value that indicates the size of the policy in packed form.
	// The service rejects any policy with a packed size greater than 100 percent,
	// which means the policy exceeded the allowed space.
	PackedPolicySize int `xml:",omitempty"`
}

// AssumeRoleResponse contains the result of successful AssumeRole request.
type AssumeRoleResponse struct {
	XMLName xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleResponse" json:"-"`

	Result           AssumeRoleResult `xml:"AssumeRoleResult"`
	ResponseMetadata struct {
		RequestID string `xml:"RequestId,omitempty"`
	} `xml:"ResponseMetadata,omitempty"`
}

// WriteSTSErrorResponse  writes error headers
func WriteSTSErrorResponse(ctx context.Context, w http.ResponseWriter, isErrCodeSTS bool, errCode api_errors.STSErrorCode, errCtxt error) {
	var err api_errors.STSError
	if isErrCodeSTS {
		err = api_errors.StsErrCodes.ToSTSErr(errCode)
	}
	if err.Code == "InternalError" || !isErrCodeSTS {
		aerr := api_errors.GetAPIError(api_errors.ErrorCode(errCode))
		if aerr.Code != "InternalError" {
			err.Code = aerr.Code
			err.Description = aerr.Description
			err.HTTPStatusCode = aerr.HTTPStatusCode
		}
	}
	// Generate error response.
	stsErrorResponse := STSErrorResponse{}
	stsErrorResponse.Error.Code = err.Code
	stsErrorResponse.RequestID = w.Header().Get(consts.AmzRequestID)
	stsErrorResponse.Error.Message = err.Description
	if errCtxt != nil {
		stsErrorResponse.Error.Message = errCtxt.Error()
	}
	encodedErrorResponse := EncodeResponse(stsErrorResponse)
	writeResponse(w, err.HTTPStatusCode, encodedErrorResponse, mimeXML)
}

// Encodes the response headers into XML format.
func EncodeResponse(response interface{}) []byte {
	var bytesBuffer bytes.Buffer
	bytesBuffer.WriteString(xml.Header)
	e := xml.NewEncoder(&bytesBuffer)
	e.Encode(response)
	return bytesBuffer.Bytes()
}

// STSErrorResponse - error response format
type STSErrorResponse struct {
	XMLName xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ ErrorResponse" json:"-"`
	Error   struct {
		Type    string `xml:"Type"`
		Code    string `xml:"Code"`
		Message string `xml:"Message"`
	} `xml:"Error"`
	RequestID string `xml:"RequestId"`
}

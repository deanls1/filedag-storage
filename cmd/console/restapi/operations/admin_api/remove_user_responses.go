// Code generated by go-swagger; DO NOT EDIT.

// This file is part of MinIO Console Server
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package admin_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	models "github.com/filedag-project/filedag-storage/cmd/console/model"
	"net/http"

	"github.com/go-openapi/runtime"

)

// RemoveUserNoContentCode is the HTTP code returned for type RemoveUserNoContent
const RemoveUserNoContentCode int = 204

/*RemoveUserNoContent A successful response.

swagger:response removeUserNoContent
*/
type RemoveUserNoContent struct {
}

// NewRemoveUserNoContent creates RemoveUserNoContent with default headers values
func NewRemoveUserNoContent() *RemoveUserNoContent {

	return &RemoveUserNoContent{}
}

// WriteResponse to the client
func (o *RemoveUserNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

/*RemoveUserDefault Generic error response.

swagger:response removeUserDefault
*/
type RemoveUserDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewRemoveUserDefault creates RemoveUserDefault with default headers values
func NewRemoveUserDefault(code int) *RemoveUserDefault {
	if code <= 0 {
		code = 500
	}

	return &RemoveUserDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the remove user default response
func (o *RemoveUserDefault) WithStatusCode(code int) *RemoveUserDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the remove user default response
func (o *RemoveUserDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the remove user default response
func (o *RemoveUserDefault) WithPayload(payload *models.Error) *RemoveUserDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the remove user default response
func (o *RemoveUserDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *RemoveUserDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

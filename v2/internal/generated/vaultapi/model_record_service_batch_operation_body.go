/*
Skyflow Data API

# Data API  This API inserts, retrieves, and otherwise manages data in a vaultapi.  The Data API is available from two base URIs. *identifier* is the identifier in your vaultapi's URL.<ul><li><b>Sandbox:</b> https://_*identifier*.vaultapi.skyflowapis-preview.com</li><li><b>Production:</b> https://_*identifier*.vaultapi.skyflowapis.com</li></ul>  When you make an API call, you need to add a header: <table><tr><th>Header</th><th>Value</th><th>Example</th></tr><tr><td>Authorization</td><td>A Bearer Token. See <a href='/api-authentication/'>API Authentication</a>.</td><td><code>Authorization: Bearer eyJhbGciOiJSUzI...1NiIsJdfPA</code></td></tr><table/>

API version: v1
Contact: support@skyflow.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package vaultapi

import (
	"encoding/json"
)

// checks if the RecordServiceBatchOperationBody type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &RecordServiceBatchOperationBody{}

// RecordServiceBatchOperationBody struct for RecordServiceBatchOperationBody
type RecordServiceBatchOperationBody struct {
	// Record operations to perform.
	Records []V1BatchRecord `json:"records,omitempty"`
	// Continue performing operations on partial errors.
	ContinueOnError *bool `json:"continueOnError,omitempty"`
	Byot *V1BYOT          `json:"byot,omitempty"`
}

// NewRecordServiceBatchOperationBody instantiates a new RecordServiceBatchOperationBody object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRecordServiceBatchOperationBody() *RecordServiceBatchOperationBody {
	this := RecordServiceBatchOperationBody{}
	var byot V1BYOT = V1BYOT_DISABLE
	this.Byot = &byot
	return &this
}

// NewRecordServiceBatchOperationBodyWithDefaults instantiates a new RecordServiceBatchOperationBody object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRecordServiceBatchOperationBodyWithDefaults() *RecordServiceBatchOperationBody {
	this := RecordServiceBatchOperationBody{}
	var byot V1BYOT = V1BYOT_DISABLE
	this.Byot = &byot
	return &this
}

// GetRecords returns the Records field value if set, zero value otherwise.
func (o *RecordServiceBatchOperationBody) GetRecords() []V1BatchRecord {
	if o == nil || IsNil(o.Records) {
		var ret []V1BatchRecord
		return ret
	}
	return o.Records
}

// GetRecordsOk returns a tuple with the Records field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RecordServiceBatchOperationBody) GetRecordsOk() ([]V1BatchRecord, bool) {
	if o == nil || IsNil(o.Records) {
		return nil, false
	}
	return o.Records, true
}

// HasRecords returns a boolean if a field has been set.
func (o *RecordServiceBatchOperationBody) HasRecords() bool {
	if o != nil && !IsNil(o.Records) {
		return true
	}

	return false
}

// SetRecords gets a reference to the given []V1BatchRecord and assigns it to the Records field.
func (o *RecordServiceBatchOperationBody) SetRecords(v []V1BatchRecord) {
	o.Records = v
}

// GetContinueOnError returns the ContinueOnError field value if set, zero value otherwise.
func (o *RecordServiceBatchOperationBody) GetContinueOnError() bool {
	if o == nil || IsNil(o.ContinueOnError) {
		var ret bool
		return ret
	}
	return *o.ContinueOnError
}

// GetContinueOnErrorOk returns a tuple with the ContinueOnError field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RecordServiceBatchOperationBody) GetContinueOnErrorOk() (*bool, bool) {
	if o == nil || IsNil(o.ContinueOnError) {
		return nil, false
	}
	return o.ContinueOnError, true
}

// HasContinueOnError returns a boolean if a field has been set.
func (o *RecordServiceBatchOperationBody) HasContinueOnError() bool {
	if o != nil && !IsNil(o.ContinueOnError) {
		return true
	}

	return false
}

// SetContinueOnError gets a reference to the given bool and assigns it to the ContinueOnError field.
func (o *RecordServiceBatchOperationBody) SetContinueOnError(v bool) {
	o.ContinueOnError = &v
}

// GetByot returns the Byot field value if set, zero value otherwise.
func (o *RecordServiceBatchOperationBody) GetByot() V1BYOT {
	if o == nil || IsNil(o.Byot) {
		var ret V1BYOT
		return ret
	}
	return *o.Byot
}

// GetByotOk returns a tuple with the Byot field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RecordServiceBatchOperationBody) GetByotOk() (*V1BYOT, bool) {
	if o == nil || IsNil(o.Byot) {
		return nil, false
	}
	return o.Byot, true
}

// HasByot returns a boolean if a field has been set.
func (o *RecordServiceBatchOperationBody) HasByot() bool {
	if o != nil && !IsNil(o.Byot) {
		return true
	}

	return false
}

// SetByot gets a reference to the given V1BYOT and assigns it to the Byot field.
func (o *RecordServiceBatchOperationBody) SetByot(v V1BYOT) {
	o.Byot = &v
}

func (o RecordServiceBatchOperationBody) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o RecordServiceBatchOperationBody) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Records) {
		toSerialize["records"] = o.Records
	}
	if !IsNil(o.ContinueOnError) {
		toSerialize["continueOnError"] = o.ContinueOnError
	}
	if !IsNil(o.Byot) {
		toSerialize["byot"] = o.Byot
	}
	return toSerialize, nil
}

type NullableRecordServiceBatchOperationBody struct {
	value *RecordServiceBatchOperationBody
	isSet bool
}

func (v NullableRecordServiceBatchOperationBody) Get() *RecordServiceBatchOperationBody {
	return v.value
}

func (v *NullableRecordServiceBatchOperationBody) Set(val *RecordServiceBatchOperationBody) {
	v.value = val
	v.isSet = true
}

func (v NullableRecordServiceBatchOperationBody) IsSet() bool {
	return v.isSet
}

func (v *NullableRecordServiceBatchOperationBody) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRecordServiceBatchOperationBody(val *RecordServiceBatchOperationBody) *NullableRecordServiceBatchOperationBody {
	return &NullableRecordServiceBatchOperationBody{value: val, isSet: true}
}

func (v NullableRecordServiceBatchOperationBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRecordServiceBatchOperationBody) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



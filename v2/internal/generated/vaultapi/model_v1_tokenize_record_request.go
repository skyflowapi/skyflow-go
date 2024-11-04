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

// checks if the V1TokenizeRecordRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1TokenizeRecordRequest{}

// V1TokenizeRecordRequest struct for V1TokenizeRecordRequest
type V1TokenizeRecordRequest struct {
	// Existing value to return a token for.
	Value *string `json:"value,omitempty"`
	// Name of the column group that the value belongs to.
	ColumnGroup *string `json:"columnGroup,omitempty"`
}

// NewV1TokenizeRecordRequest instantiates a new V1TokenizeRecordRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1TokenizeRecordRequest() *V1TokenizeRecordRequest {
	this := V1TokenizeRecordRequest{}
	return &this
}

// NewV1TokenizeRecordRequestWithDefaults instantiates a new V1TokenizeRecordRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1TokenizeRecordRequestWithDefaults() *V1TokenizeRecordRequest {
	this := V1TokenizeRecordRequest{}
	return &this
}

// GetValue returns the Value field value if set, zero value otherwise.
func (o *V1TokenizeRecordRequest) GetValue() string {
	if o == nil || IsNil(o.Value) {
		var ret string
		return ret
	}
	return *o.Value
}

// GetValueOk returns a tuple with the Value field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TokenizeRecordRequest) GetValueOk() (*string, bool) {
	if o == nil || IsNil(o.Value) {
		return nil, false
	}
	return o.Value, true
}

// HasValue returns a boolean if a field has been set.
func (o *V1TokenizeRecordRequest) HasValue() bool {
	if o != nil && !IsNil(o.Value) {
		return true
	}

	return false
}

// SetValue gets a reference to the given string and assigns it to the Value field.
func (o *V1TokenizeRecordRequest) SetValue(v string) {
	o.Value = &v
}

// GetColumnGroup returns the ColumnGroup field value if set, zero value otherwise.
func (o *V1TokenizeRecordRequest) GetColumnGroup() string {
	if o == nil || IsNil(o.ColumnGroup) {
		var ret string
		return ret
	}
	return *o.ColumnGroup
}

// GetColumnGroupOk returns a tuple with the ColumnGroup field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TokenizeRecordRequest) GetColumnGroupOk() (*string, bool) {
	if o == nil || IsNil(o.ColumnGroup) {
		return nil, false
	}
	return o.ColumnGroup, true
}

// HasColumnGroup returns a boolean if a field has been set.
func (o *V1TokenizeRecordRequest) HasColumnGroup() bool {
	if o != nil && !IsNil(o.ColumnGroup) {
		return true
	}

	return false
}

// SetColumnGroup gets a reference to the given string and assigns it to the ColumnGroup field.
func (o *V1TokenizeRecordRequest) SetColumnGroup(v string) {
	o.ColumnGroup = &v
}

func (o V1TokenizeRecordRequest) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1TokenizeRecordRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Value) {
		toSerialize["value"] = o.Value
	}
	if !IsNil(o.ColumnGroup) {
		toSerialize["columnGroup"] = o.ColumnGroup
	}
	return toSerialize, nil
}

type NullableV1TokenizeRecordRequest struct {
	value *V1TokenizeRecordRequest
	isSet bool
}

func (v NullableV1TokenizeRecordRequest) Get() *V1TokenizeRecordRequest {
	return v.value
}

func (v *NullableV1TokenizeRecordRequest) Set(val *V1TokenizeRecordRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableV1TokenizeRecordRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableV1TokenizeRecordRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1TokenizeRecordRequest(val *V1TokenizeRecordRequest) *NullableV1TokenizeRecordRequest {
	return &NullableV1TokenizeRecordRequest{value: val, isSet: true}
}

func (v NullableV1TokenizeRecordRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1TokenizeRecordRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



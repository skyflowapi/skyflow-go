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

// checks if the V1AuditAfterOptions type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1AuditAfterOptions{}

// V1AuditAfterOptions struct for V1AuditAfterOptions
type V1AuditAfterOptions struct {
	// Timestamp provided in the previous audit response's `nextOps` attribute. An alternate way to manage response pagination. Can't be used with `sortOps` or `offset`. For the first request in a series of audit requests, leave blank.
	Timestamp *string `json:"timestamp,omitempty"`
	// Change ID provided in the previous audit response's `nextOps` attribute. An alternate way to manage response pagination. Can't be used with `sortOps` or `offset`. For the first request in a series of audit requests, leave blank.
	ChangeID *string `json:"changeID,omitempty"`
}

// NewV1AuditAfterOptions instantiates a new V1AuditAfterOptions object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1AuditAfterOptions() *V1AuditAfterOptions {
	this := V1AuditAfterOptions{}
	return &this
}

// NewV1AuditAfterOptionsWithDefaults instantiates a new V1AuditAfterOptions object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1AuditAfterOptionsWithDefaults() *V1AuditAfterOptions {
	this := V1AuditAfterOptions{}
	return &this
}

// GetTimestamp returns the Timestamp field value if set, zero value otherwise.
func (o *V1AuditAfterOptions) GetTimestamp() string {
	if o == nil || IsNil(o.Timestamp) {
		var ret string
		return ret
	}
	return *o.Timestamp
}

// GetTimestampOk returns a tuple with the Timestamp field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1AuditAfterOptions) GetTimestampOk() (*string, bool) {
	if o == nil || IsNil(o.Timestamp) {
		return nil, false
	}
	return o.Timestamp, true
}

// HasTimestamp returns a boolean if a field has been set.
func (o *V1AuditAfterOptions) HasTimestamp() bool {
	if o != nil && !IsNil(o.Timestamp) {
		return true
	}

	return false
}

// SetTimestamp gets a reference to the given string and assigns it to the Timestamp field.
func (o *V1AuditAfterOptions) SetTimestamp(v string) {
	o.Timestamp = &v
}

// GetChangeID returns the ChangeID field value if set, zero value otherwise.
func (o *V1AuditAfterOptions) GetChangeID() string {
	if o == nil || IsNil(o.ChangeID) {
		var ret string
		return ret
	}
	return *o.ChangeID
}

// GetChangeIDOk returns a tuple with the ChangeID field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1AuditAfterOptions) GetChangeIDOk() (*string, bool) {
	if o == nil || IsNil(o.ChangeID) {
		return nil, false
	}
	return o.ChangeID, true
}

// HasChangeID returns a boolean if a field has been set.
func (o *V1AuditAfterOptions) HasChangeID() bool {
	if o != nil && !IsNil(o.ChangeID) {
		return true
	}

	return false
}

// SetChangeID gets a reference to the given string and assigns it to the ChangeID field.
func (o *V1AuditAfterOptions) SetChangeID(v string) {
	o.ChangeID = &v
}

func (o V1AuditAfterOptions) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1AuditAfterOptions) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Timestamp) {
		toSerialize["timestamp"] = o.Timestamp
	}
	if !IsNil(o.ChangeID) {
		toSerialize["changeID"] = o.ChangeID
	}
	return toSerialize, nil
}

type NullableV1AuditAfterOptions struct {
	value *V1AuditAfterOptions
	isSet bool
}

func (v NullableV1AuditAfterOptions) Get() *V1AuditAfterOptions {
	return v.value
}

func (v *NullableV1AuditAfterOptions) Set(val *V1AuditAfterOptions) {
	v.value = val
	v.isSet = true
}

func (v NullableV1AuditAfterOptions) IsSet() bool {
	return v.isSet
}

func (v *NullableV1AuditAfterOptions) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1AuditAfterOptions(val *V1AuditAfterOptions) *NullableV1AuditAfterOptions {
	return &NullableV1AuditAfterOptions{value: val, isSet: true}
}

func (v NullableV1AuditAfterOptions) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1AuditAfterOptions) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



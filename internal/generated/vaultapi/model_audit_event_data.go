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

// checks if the AuditEventData type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AuditEventData{}

// AuditEventData Any Sensitive data that needs to be wrapped.
type AuditEventData struct {
	// The entire body of the data requested or the query fired.
	Content *string `json:"content,omitempty"`
}

// NewAuditEventData instantiates a new AuditEventData object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAuditEventData() *AuditEventData {
	this := AuditEventData{}
	return &this
}

// NewAuditEventDataWithDefaults instantiates a new AuditEventData object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAuditEventDataWithDefaults() *AuditEventData {
	this := AuditEventData{}
	return &this
}

// GetContent returns the Content field value if set, zero value otherwise.
func (o *AuditEventData) GetContent() string {
	if o == nil || IsNil(o.Content) {
		var ret string
		return ret
	}
	return *o.Content
}

// GetContentOk returns a tuple with the Content field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AuditEventData) GetContentOk() (*string, bool) {
	if o == nil || IsNil(o.Content) {
		return nil, false
	}
	return o.Content, true
}

// HasContent returns a boolean if a field has been set.
func (o *AuditEventData) HasContent() bool {
	if o != nil && !IsNil(o.Content) {
		return true
	}

	return false
}

// SetContent gets a reference to the given string and assigns it to the Content field.
func (o *AuditEventData) SetContent(v string) {
	o.Content = &v
}

func (o AuditEventData) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AuditEventData) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Content) {
		toSerialize["content"] = o.Content
	}
	return toSerialize, nil
}

type NullableAuditEventData struct {
	value *AuditEventData
	isSet bool
}

func (v NullableAuditEventData) Get() *AuditEventData {
	return v.value
}

func (v *NullableAuditEventData) Set(val *AuditEventData) {
	v.value = val
	v.isSet = true
}

func (v NullableAuditEventData) IsSet() bool {
	return v.isSet
}

func (v *NullableAuditEventData) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAuditEventData(val *AuditEventData) *NullableAuditEventData {
	return &NullableAuditEventData{value: val, isSet: true}
}

func (v NullableAuditEventData) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAuditEventData) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



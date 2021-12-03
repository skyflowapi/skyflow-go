package vaultapi

type Client struct {
	configuration Configuration
}

func (client *Client) insert(records map[string]interface{}, options map[string]interface{}, callback Callback) {
	//insert
}

func (client *Client) detokenize(records map[string]interface{}, callback Callback) {
	//detokenize
}

func (client *Client) getById(records map[string]interface{}, callback Callback) {
	//getById
}

func (client *Client) invokeConnection(connectionConfig ConnectionConfig, callback Callback) {
	//invokeConnection
}

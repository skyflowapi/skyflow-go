package vaultapi

type Client struct {
	configuration Configuration
}

func (client *Client) insert(records map[string]interface{}, options map[string]interface{}) (responseBody, error) {
	//insert
	return nil, nil
}

func (client *Client) detokenize(records map[string]interface{}) (responseBody, error) {
	//detokenize
	return nil, nil
}

func (client *Client) getById(records map[string]interface{}) (responseBody, error) {
	//getById
	return nil, nil
}

func (client *Client) invokeConnection(connectionConfig ConnectionConfig) (responseBody, error) {
	//invokeConnection
	return nil, nil
}

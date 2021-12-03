package vaultapi

type ApiClient struct {
	vaultID       string
	vaultURL      string
	tokenProvider TokenProvider
	logLevel      LogLevel
	token         string
}

func (client *ApiClient) insert(records map[string]interface{}, options map[string]interface{}) (responseBody, error) {

	return nil, nil
}

func (client *ApiClient) detokenize(records map[string]interface{}) (responseBody, error) {

	return nil, nil
}

func (client *ApiClient) getById(records map[string]interface{}) (responseBody, error) {

	return nil, nil
}

func (client *ApiClient) invokeConnection(connectionConfig ConnectionConfig) (responseBody, error) {

	return nil, nil
}

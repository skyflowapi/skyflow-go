package vaultapi

func Init(configuration Configuration) Client {
	return Client{configuration}
}

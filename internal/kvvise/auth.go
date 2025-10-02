package kvvise

import "net/http"

func (kv *KVVise) setAuthHeaders(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+kv.apiKey)
	return nil
}

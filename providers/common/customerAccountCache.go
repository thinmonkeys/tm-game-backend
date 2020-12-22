package common

import (
	"encoding/json"
	"fmt"
	"net/http"
)


type CustomerAccountCache struct {
	connection *ConnectionSettings
	cache      map[string]string
}

func NewCache(connection *ConnectionSettings) CustomerAccountCache {
	return CustomerAccountCache {
		connection: connection,
		cache: map[string]string {},
	}
}

func (cache *CustomerAccountCache) GetPrimaryAccountId(customerCif string) (string, error) {
	if accountID, ok := cache.cache[customerCif]; ok {
		return accountID, nil
	}

	response, err := cache.connection.RunRequest(http.MethodGet, fmt.Sprintf("/accounts/%s", customerCif), nil)
	if err != nil {
		return "", fmt.Errorf("Error getting account ID for customer %s: %s", customerCif, err.Error())
	}

	accountIDs := []string{}
	err = json.NewDecoder(response.Body).Decode(&accountIDs)
	if err != nil { return "", fmt.Errorf("Error decoding JSON response: %s", err.Error()) }

	if len(accountIDs) == 0 { return "", fmt.Errorf("No accounts found for customer %s", customerCif)}
	accountID := accountIDs[0]
	cache.cache[customerCif] = accountID
	return accountID, nil
}
package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
	"sarasa/schemas"
)

func (postgres *Client) SaveProvidersList(providers []schemas.Provider, availableZones map[string]int, availableSources map[string]int) error {
	var err error

	log.Printf("Saving %d providers to postgres...\n", len(providers))

	postgres.txn, err = postgres.connection.Begin()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to start a transaction, error: %s", err)
	}

	err = postgres.DeleteProvidersFromSource(availableSources[providers[0].Source])
	if err != nil {
		return err
	}

	err = postgres.SaveZonesFromProviders(providers, availableZones)
	if err != nil {
		return err
	}

	zonesMap, err := postgres.GetZones()
	if err != nil {
		return fmt.Errorf("saveProvidersList - Fail to get zones, error: %s", err)
	}

	err = postgres.SaveSourcesFromProviders(providers, availableSources)
	if err != nil {
		return err
	}

	sourcesMap, err := postgres.GetSources()
	if err != nil {
		return fmt.Errorf("saveProvidersList - Fail to get sources, error: %s", err)
	}

	err = postgres.SaveProviders(providers, zonesMap, sourcesMap)
	if err != nil {
		return fmt.Errorf("saveProvidersList - Fail to save providers, error: %s", err)
	}

	err = postgres.txn.Commit()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to commit txn, error: %s", err)
	}

	log.Printf("%d providers saved to postgres.\n", len(providers))
	postgres.txn = nil

	return nil
}

func (postgres *Client) DeleteProvidersFromSource(sourceID int) error {
	stmt, err := postgres.txn.Prepare(fmt.Sprintf("DELETE FROM providers WHERE source_id = %d", sourceID))
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to prepare CopyIn statement, error: %s", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to exec(final) CopyIn statement, error: %s", err)
	}

	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to close stmt, error: %s", err)
	}

	return nil
}

func (postgres *Client) SaveZonesFromProviders(providers []schemas.Provider, availableZones map[string]int) error {
	zonesMap := make(map[string]bool, 0)
	for i := 0; i < len(providers); i++ {
		zonesMap[providers[i].Place] = true
	}

	var zones []string
	for k := range zonesMap {
		if _, ok := availableZones[k]; !ok {
			zones = append(zones, k)
		}
	}

	stmt, err := postgres.txn.Prepare(pq.CopyIn("zones", "name"))
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to prepare CopyIn statement, error: %s", err)
	}

	for _, zone := range zones {
		_, err := stmt.Exec(zone)
		if err != nil {
			return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to exec(for) CopyIn statement, error: %s", err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to exec(final) CopyIn statement, error: %s", err)
	}

	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveZonesFromProviders - Fail to close stmt, error: %s", err)
	}

	return nil
}

func (postgres *Client) GetZones() (map[string]int, error) {
	var stmt *sql.Stmt
	var err error

	if postgres.txn != nil {
		stmt, err = postgres.txn.Prepare("SELECT id, name FROM zones")
	} else {
		stmt, err = postgres.connection.Prepare("SELECT id, name FROM zones")
	}

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing postgres statement rows - error: %s", err)
		}
	}()

	zones := make(map[string]int, 0)
	for rows.Next() {
		var id int
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, fmt.Errorf("getZones - Fail to scan row, error: %s", err)
		}

		zones[name] = id
	}

	return zones, nil
}

func (postgres *Client) SaveSourcesFromProviders(providers []schemas.Provider, availableSources map[string]int) error {
	sourcesMap := make(map[string]bool, 0)
	for i := 0; i < len(providers); i++ {
		sourcesMap[providers[i].Source] = true
	}

	var sources []string
	for k := range sourcesMap {
		if _, ok := availableSources[k]; !ok {
			sources = append(sources, k)
		}
	}

	stmt, err := postgres.txn.Prepare(pq.CopyIn("sources", "name"))
	if err != nil {
		return fmt.Errorf("postgressClient/SaveSourcesFromProviders - Fail to prepare CopyIn statement, error: %s", err)
	}

	for _, source := range sources {
		_, err := stmt.Exec(source)
		if err != nil {
			return fmt.Errorf("postgressClient/SaveSourcesFromProviders - Fail to exec(for) CopyIn statement, error: %s", err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveSourcesFromProviders - Fail to exec(final) CopyIn statement, error: %s", err)
	}

	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveSourcesFromProviders - Fail to close stmt, error: %s", err)
	}

	return nil
}

func (postgres *Client) GetSources() (map[string]int, error) {
	var stmt *sql.Stmt
	var err error

	if postgres.txn != nil {
		stmt, err = postgres.txn.Prepare("SELECT id, name FROM sources")
	} else {
		stmt, err = postgres.connection.Prepare("SELECT id, name FROM sources")
	}

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing postgres statement rows - error: %s", err)
		}
	}()

	sources := make(map[string]int, 0)
	for rows.Next() {
		var id int
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, fmt.Errorf("GetSources - Fail to scan row, error: %s", err)
		}

		sources[name] = id
	}

	return sources, nil
}

func (postgres *Client) SaveProviders(providers []schemas.Provider, zones map[string]int, sources map[string]int) error {
	stmt, err := postgres.txn.Prepare(pq.CopyIn("providers", "name", "phone", "zone_id", "source_id"))
	if err != nil {
		return fmt.Errorf("postgressClient/SaveProviders - Fail to prepare CopyIn statement, error: %s", err)
	}

	for _, provider := range providers {
		_, err := stmt.Exec(provider.Name, provider.Phone, zones[provider.Place], sources[provider.Source])
		if err != nil {
			return fmt.Errorf("postgressClient/SaveProviders - Fail to exec(for) CopyIn statement, error: %s", err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveProviders - Fail to exec(final) CopyIn statement, error: %s", err)
	}

	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveProviders - Fail to close stmt, error: %s", err)
	}

	providersPhoneIdMap, err := postgres.GetProvidersPhoneIdMap()
	if err != nil {
		return fmt.Errorf("saveProvidersList - Fail to get zones, error: %s", err)
	}

	stmt, err = postgres.txn.Prepare(pq.CopyIn("provider_pics", "provider_id", "pic_url"))
	if err != nil {
		return fmt.Errorf("postgressClient/SaveProviders - Fail to prepare CopyIn statement, error: %s", err)
	}

	for _, provider := range providers {
		for _, pic := range provider.Pics {
			_, err := stmt.Exec(providersPhoneIdMap[provider.Phone], pic)
			if err != nil {
				return fmt.Errorf("postgressClient/SaveProviders - Fail to exec(for) CopyIn statement, error: %s", err)
			}
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveProviders - Fail to exec(final2) CopyIn statement, error: %s", err)
	}

	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("postgressClient/SaveProviders - Fail to close stmt, error: %s", err)
	}

	return nil
}

func (postgres *Client) GetProvidersPhoneIdMap() (map[string]int, error) {
	var stmt *sql.Stmt
	var err error

	query := "SELECT id, phone FROM providers"
	if postgres.txn != nil {
		stmt, err = postgres.txn.Prepare(query)
	} else {
		stmt, err = postgres.connection.Prepare(query)
	}

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing postgres statement rows - error: %s", err)
		}
	}()

	providers := make(map[string]int, 0)
	for rows.Next() {
		var id int
		var phone string

		err = rows.Scan(&id, &phone)
		if err != nil {
			return nil, fmt.Errorf("GetProviders - Fail to scan row, error: %s", err)
		}

		providers[phone] = id
	}

	return providers, nil
}

func (postgres *Client) GetProvidersByZone(zoneID int) ([]schemas.Provider, error) {
	var stmt *sql.Stmt
	var err error

	query := fmt.Sprintf("SELECT id, name, phone FROM providers WHERE zone_id = %d", zoneID)
	if postgres.txn != nil {
		stmt, err = postgres.txn.Prepare(query)
	} else {
		stmt, err = postgres.connection.Prepare(query)
	}

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing postgres statement rows - error: %s", err)
		}
	}()

	providers := make([]schemas.Provider, 0)
	for rows.Next() {
		var id int
		var name string
		var phone string

		err = rows.Scan(&id, &name, &phone)
		if err != nil {
			return nil, fmt.Errorf("GetProvidersByZone - Fail to scan row, error: %s", err)
		}

		providers = append(providers, schemas.Provider{ID: id, Name: name, Phone: phone})
	}

	return providers, nil
}

// GetProviderPics currently not being used.
func (postgres *Client) GetProviderPics(providerID int) ([]string, error) {
	var stmt *sql.Stmt
	var err error

	query := fmt.Sprintf("SELECT pic_url FROM provider_pics WHERE provider_id = %d LIMIT 10", providerID)
	if postgres.txn != nil {
		stmt, err = postgres.txn.Prepare(query)
	} else {
		stmt, err = postgres.connection.Prepare(query)
	}

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing postgres statement rows - error: %s", err)
		}
	}()

	pics := make([]string, 0)
	for rows.Next() {
		var picURL string

		err = rows.Scan(&picURL)
		if err != nil {
			return nil, fmt.Errorf("GetProviderPics - Fail to scan row, error: %s", err)
		}

		pics = append(pics, picURL)
	}

	return pics, nil
}

func (postgres *Client) GetProviders() ([]schemas.Provider, error) {
	var stmt *sql.Stmt
	var err error

	query := `
SELECT 
	providers.id, providers.name, providers.phone, 
	sources.name as source, 
	zones.name as place, 
	provider_pics.pic_url
FROM providers
    JOIN sources ON providers.source_id = sources.id
    JOIN zones ON providers.zone_id = zones.id
    JOIN provider_pics on providers.id = provider_pics.provider_id
ORDER BY providers.id, zones.id
`
	if postgres.txn != nil {
		stmt, err = postgres.txn.Prepare(query)
	} else {
		stmt, err = postgres.connection.Prepare(query)
	}

	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing postgres statement rows - error: %s", err)
		}
	}()
	rows.Next()

	providers := make([]schemas.Provider, 0)

	currentProvider := schemas.Provider{}
	var pic string

	err = rows.Scan(
		&currentProvider.ID,
		&currentProvider.Name,
		&currentProvider.Phone,
		&currentProvider.Source,
		&currentProvider.Place,
		&pic,
	)
	if err != nil {
		return nil, fmt.Errorf("GetProviders - Fail to scan row, error: %s", err)
	}

	currentProvider.Pics = []string{pic}

	newProvider := schemas.Provider{}
	for rows.Next() {
		err = rows.Scan(
			&newProvider.ID,
			&newProvider.Name,
			&newProvider.Phone,
			&newProvider.Source,
			&newProvider.Place,
			&pic,
		)

		if err != nil {
			return nil, fmt.Errorf("GetProviders - Fail to scan row, error: %s", err)
		}

		if currentProvider.ID == newProvider.ID {
			currentProvider.Pics = append(currentProvider.Pics, pic)
		} else {
			providers = append(providers, currentProvider)
			currentProvider = newProvider
			currentProvider.Pics = []string{pic}
		}
	}

	if currentProvider.ID == newProvider.ID {
		providers = append(providers, currentProvider)
	}

	return providers, nil
}

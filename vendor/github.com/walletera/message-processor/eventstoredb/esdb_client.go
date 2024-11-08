package eventstoredb

import "github.com/EventStore/EventStore-Client-Go/v4/esdb"

func GetESDBClient(connectionString string) (*esdb.Client, error) {
    settings, err := esdb.ParseConnectionString(connectionString)
    if err != nil {
        return nil, err
    }

    settings.Logger = esdb.NoopLogging()

    return esdb.NewClient(settings)
}

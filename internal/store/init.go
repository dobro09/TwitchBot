package store

func InitStore(storeType, databaseURL string) (MessageStore, error) {
    switch storeType {
    case "memory":
        return NewInMemoryStore()
    default:
        return NewPostgresStore(databaseURL)
    }
}
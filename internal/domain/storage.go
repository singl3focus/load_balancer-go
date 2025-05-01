package domain

type Storage interface {
    RateLimiterStorage
}

type RateLimiterStorage interface {
	GetConfig(key string) (BucketConfig, error)
    StoreConfig(key string, cfg BucketConfig) error
    UpdateConfig(key string, cfg BucketConfig) error 
    DeleteConfig(key string) error
}
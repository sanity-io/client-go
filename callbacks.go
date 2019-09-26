package sanity

type Callbacks struct {
	OnErrorWillRetry func(error)
	OnQueryResult    func(*QueryResult)
}

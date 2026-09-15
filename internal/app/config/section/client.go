package section

type (
	Client struct {
		Catalog ClientCatalog
	}

	ClientCatalog struct {
		GrpcAddress string `split_words:"true" required:"true"`
	}
)

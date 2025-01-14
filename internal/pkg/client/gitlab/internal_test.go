package gitlab

type (
	StringWriter = stringWriter
)

func NewFetchClient(
	projectProvider projectProvider,
	branchProvider branchProvider,
	commitProvider commitProvider,
	dependencyListExporter dependencyListExporter,
) *Client {
	return &Client{
		projectProvider:        projectProvider,
		branchProvider:         branchProvider,
		commitProvider:         commitProvider,
		dependencyListExporter: dependencyListExporter,
	}
}

func NewPushClient(
	projectProvider projectProvider,
	genericPackagePublisher genericPackagePublisher,
) *Client {
	return &Client{
		projectProvider:         projectProvider,
		genericPackagePublisher: genericPackagePublisher,
	}
}

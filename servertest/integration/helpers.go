package integration

const (
	vaultName     string = "go-vault1"
	vaultItemName string = "secret1"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

package vault

type ItemType string

const (
	ItemTypeVault     ItemType = "vault"
	ItemTypeEncrypted ItemType = "encrypted"
	ItemTypeNormal    ItemType = "normal"
)

func (s *Service) ItemType(name string, item string) (ItemType, error) {
	return ItemTypeVault, nil
}

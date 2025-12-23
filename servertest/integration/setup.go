package integration

import (
	"fmt"
	"os"

	"github.com/go-chef/chef"
)

func RunGoiardiInit(cfg Config) error {
	// do os setup and fake pivotal stuff here
	client := mustCreateClient(cfg)
	_, err := client.Users.Get("rloggia")
	if err == nil {
		return err
	}

	pubKey, err := os.ReadFile(fmt.Sprintf("%s/rloggia.pub", cfg.WorkDir))
	if err != nil {
		return err
	}
	user := chef.User{
		UserName:  "rloggia",
		Email:     "robert.loggia@email.com",
		Password:  "enjoyyourbreakfast",
		PublicKey: string(pubKey),
	}
	users, err := client.Users.Create(user)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully created %s\n", users)
	return nil
}

// create workdir
// create fake pivotal user/keys
// create rloggia
// generate rloggia keys and write to workdir
// write rloggia knife.rb to workdir

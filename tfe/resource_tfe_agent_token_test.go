package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEAgentToken_basic(t *testing.T) {
	skipIfFreeOnly(t)

	agentToken := &tfe.AgentToken{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEAgentTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAgentToken_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEAgentTokenExists(
						"tfe_agent_token.foobar", agentToken),
					testAccCheckTFEAgentTokenAttributes(agentToken),
					resource.TestCheckResourceAttr(
						"tfe_agent_token.foobar", "description", "agent-token-test"),
				),
			},
		},
	})
}

func testAccCheckTFEAgentTokenExists(
	n string, agentToken *tfe.AgentToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		sk, err := tfeClient.AgentTokens.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if sk == nil {
			return fmt.Errorf("agent token not found")
		}

		*agentToken = *sk

		return nil
	}
}

func testAccCheckTFEAgentTokenAttributes(
	agentToken *tfe.AgentToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if agentToken.Description != "agent-token-test" {
			return fmt.Errorf("Bad name: %s", agentToken.Description)
		}
		return nil
	}
}

func testAccCheckTFEAgentTokenDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_agent_token" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.AgentTokens.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("agent token %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEAgentToken_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_agent_token" "foobar" {
	agent_pool_id = tfe_agent_pool.foobar.id
	description   = "agent-token-test"
}`, rInt)
}

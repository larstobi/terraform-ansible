package main
import "fmt"
import "os"
import "strings"
import "github.com/hashicorp/terraform/terraform"
import "github.com/larstobi/terraform-ansible/tory"

func readState(path string) *terraform.State {
    f, _ := os.Open(path)
    if f != nil {
        defer f.Close()
        state, _ := terraform.ReadState(f)
        return state
    }
    return nil
}

func main() {
    // Read Terraform state
    state := readState("terraform.tfstate")
    inv := tory.NewInventory()

    for key, value := range state.Modules[0].Resources {
        if (value.Type == "cloudstack_instance") {
            parts := strings.SplitN(key, ".", 3)
            hostname := value.Primary.Attributes["name"]
            inv.AddHostnameToGroupUnsanitized(parts[1], hostname)
            inv.Meta.AddHostvar(hostname, "ansible_ssh_host",
                value.Primary.Attributes["ipaddress"])

            // TODO: USE Terraform cloudstack_instance connection user.
            // Skip setting it if not existing.
            inv.Meta.AddHostvar(hostname, "ansible_ssh_user", "root")
        }
    }
    jsonMarshalled, _ := inv.MarshalJSON()
    fmt.Println(string(jsonMarshalled))
}

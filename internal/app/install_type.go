package app

import (
        "fmt"
        "os"
        "strings"
        "go-project/internal/ui"
)

type InstallationType string
type OSType string

const (
        RKE2    InstallationType = "rke2"
        K3S     InstallationType = "k3s"
        Kubeadm InstallationType = "kubeadm"
)

const (
        Rocky  OSType = "rocky"
        Ubuntu OSType = "ubuntu"
)

type InstallConfig struct {
        InstallType InstallationType
        OSType     OSType
}

func NewInstallConfig() *InstallConfig {
        return &InstallConfig{}
}

func (ic *InstallConfig) DetectOS() error {
    data, err := os.ReadFile("/etc/os-release")
    if err != nil {
        return fmt.Errorf("failed to read OS release file: %v", err)
    }
    content := string(data)
    contentLower := strings.ToLower(content)
    
    if strings.Contains(contentLower, "rocky") ||
       strings.Contains(contentLower, "amazon linux") ||
       strings.Contains(contentLower, "red hat") ||
       strings.Contains(contentLower, "rhel") {
        ic.OSType = Rocky
    } else if strings.Contains(contentLower, "ubuntu") {
        ic.OSType = Ubuntu
    } else {
        return fmt.Errorf("unsupported OS type")
    }
    return nil
}

func (ic *InstallConfig) SelectInstallationType() error {
        ui.Clear()
        ui.PrintLogo()
        ui.PrintMenuTitle("Select Installation Type:")

        options := []string{
                "1. RKE2",
                "2. K3S",
                "3. Kubeadm",
                "4. Go back",
        }

        choice := ui.ArrowSelect(options)

        switch choice {
        case 0:
                ic.InstallType = RKE2
        case 1:
                ic.InstallType = K3S
        case 2:
                ic.InstallType = Kubeadm
        case 3:
                return fmt.Errorf("user cancelled")
        }

        if err := ic.DetectOS(); err != nil {
                return fmt.Errorf("error detecting OS: %v", err)
        }

        return nil
}

func (ic *InstallConfig) GetPlaybookName() string {
        return fmt.Sprintf("%s-%s-playbook.yaml",
                strings.ToLower(string(ic.OSType)),
                strings.ToLower(string(ic.InstallType)))
}

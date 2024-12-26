package app

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"

    "go-project/internal/ui"
    "golang.org/x/term"
)

// SSHKeyPair 구조체는 SSH 키 관련 정보를 저장
type SSHKeyPair struct {
    PrivateKey string
    PublicKey  string
}

// SetupServers는 모든 서버에 대한 초기 설정을 수행
func (app *App) SetupServers() error {
    ui.PrintMenuTitle("Setting up servers...")

    // SSH 키 생성
    keyPair, err := generateSSHKeyPair()
    if err != nil {
        return fmt.Errorf("failed to generate SSH key pair: %w", err)
    }

    // 모든 서버에 대해 설정 수행
    for _, server := range app.ServerConfig.ALLServers {
        ui.Yellow.Printf("Setting up server: %s (%s)\n", server.Name, server.IP)
        
        if err := app.setupServer(server, keyPair); err != nil {
            ui.Red.Printf("Error setting up server %s: %v\n", server.Name, err)
            continue
        }
        
        ui.Green.Printf("Successfully set up server: %s\n", server.Name)
    }

    return nil
}

func setupAnsibleConfig(server Server) error {
    // Ansible configuration content
    ansibleConfigContent := `[defaults]
host_key_checking = False
deprecation_warnings = False
command_warnings = False
`

    // Create the /etc/ansible directory if it doesn't exist
    mkdirCmd := exec.Command("ssh", fmt.Sprintf("root@%s", server.IP), "mkdir", "-p", "/etc/ansible")
    if err := mkdirCmd.Run(); err != nil {
        return fmt.Errorf("failed to create /etc/ansible directory: %w", err)
    }

    // Create a temporary file with Ansible configuration
    tmpFile := "/tmp/ansible.cfg"
    if err := os.WriteFile(tmpFile, []byte(ansibleConfigContent), 0644); err != nil {
        return fmt.Errorf("failed to create temporary ansible.cfg file: %w", err)
    }
    defer os.Remove(tmpFile)

    // Copy the ansible.cfg file to the server
    scpCmd := exec.Command("scp", tmpFile, fmt.Sprintf("root@%s:/etc/ansible/ansible.cfg", server.IP))
    if err := scpCmd.Run(); err != nil {
        return fmt.Errorf("failed to copy ansible.cfg file: %w", err)
    }

    // Set correct permissions
    chmodCmd := exec.Command("ssh", fmt.Sprintf("root@%s", server.IP), "chmod", "0644", "/etc/ansible/ansible.cfg")
    if err := chmodCmd.Run(); err != nil {
        return fmt.Errorf("failed to set permissions on ansible.cfg file: %w", err)
    }

    return nil
}


// generateSSHKeyPair는 새로운 SSH 키 쌍을 생성
func generateSSHKeyPair() (*SSHKeyPair, error) {
    keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
    
    // 이미 키가 존재하는지 확인
    if _, err := os.Stat(keyPath); err == nil {
        // 기존 키 읽기
        privateKey, err := os.ReadFile(keyPath)
        if err != nil {
            return nil, err
        }
        
        publicKey, err := os.ReadFile(keyPath + ".pub")
        if err != nil {
            return nil, err
        }
        
        return &SSHKeyPair{
            PrivateKey: string(privateKey),
            PublicKey: string(publicKey),
        }, nil
    }

    // 새 SSH 키 생성
    cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", keyPath, "-N", "")
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("failed to generate SSH key: %w", err)
    }

    // 생성된 키 읽기
    privateKey, err := os.ReadFile(keyPath)
    if err != nil {
        return nil, err
    }
    
    publicKey, err := os.ReadFile(keyPath + ".pub")
    if err != nil {
        return nil, err
    }

    return &SSHKeyPair{
        PrivateKey: string(privateKey),
        PublicKey: string(publicKey),
    }, nil
}

// setupServer는 단일 서버에 대한 설정을 수행
func (app *App) setupServer(server Server, keyPair *SSHKeyPair) error {
    // 1. SSH 키 복사
    if err := copySSHKey(server.IP, keyPair.PublicKey); err != nil {
        return fmt.Errorf("failed to copy SSH key: %w", err)
    }

    // 2. 호스트네임 설정
    if err := setHostname(server); err != nil {
        return fmt.Errorf("failed to set hostname: %w", err)
    }

    // 3. Sudoers 설정
    if err := setupSudoers(server); err != nil {
        return fmt.Errorf("failed to setup sudoers: %w", err)
    }

    // 4. Ansible 설정 (새로 추가된 단계)
    if err := setupAnsibleConfig(server); err != nil {
        return fmt.Errorf("failed to setup Ansible configuration: %w", err)
    }

    return nil
}

// copySSHKey는 대상 서버에 SSH 공개키를 복사
func copySSHKey(serverIP, publicKey string) error {
    // 비밀번호 프롬프트
    fmt.Printf("Enter password for server %s: ", serverIP)
    bytePassword, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        return fmt.Errorf("password input error: %w", err)
    }
    fmt.Println() // 줄바꿈

    password := string(bytePassword)
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("failed to get home directory: %w", err)
    }

    keyPath := filepath.Join(homeDir, ".ssh", "id_rsa.pub")

    cmd := exec.Command("sshpass", "-p", password,
        "ssh-copy-id", "-o", "StrictHostKeyChecking=no",
        "-i", keyPath, fmt.Sprintf("root@%s", serverIP))

    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("ssh-copy-id failed: %v, stderr: %s", err, stderr.String())
    }

    return nil
}

// setHostname은 서버의 호스트네임을 설정
func setHostname(server Server) error {
    // 호스트네임 설정 명령 실행
    cmd := exec.Command("ssh", fmt.Sprintf("root@%s", server.IP), "hostnamectl", "set-hostname", server.Name)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to set hostname: %w", err)
    }

    return nil
}

// setupSudoers는 sudoers 설정을 수행
func setupSudoers(server Server) error {
    // sudoers 설정을 위한 템플릿
    sudoersTemplate := `
# Created by CN Studio
%s ALL=(ALL) NOPASSWD: ALL
`
    currentUser := os.Getenv("USER")
    sudoersContent := fmt.Sprintf(sudoersTemplate, currentUser)

    // 임시 파일 생성
    tmpFile := "/tmp/sudoers.tmp"
    if err := os.WriteFile(tmpFile, []byte(sudoersContent), 0644); err != nil {
        return fmt.Errorf("failed to create temporary sudoers file: %w", err)
    }
    defer os.Remove(tmpFile)

    // 원격 서버로 파일 복사
    scpCmd := exec.Command("scp", tmpFile, fmt.Sprintf("root@%s:/etc/sudoers.d/%s", server.IP, currentUser))
    if err := scpCmd.Run(); err != nil {
        return fmt.Errorf("failed to copy sudoers file: %w", err)
    }

    // 권한 설정
    chmodCmd := exec.Command("ssh", fmt.Sprintf("root@%s", server.IP), "chmod", "0440", fmt.Sprintf("/etc/sudoers.d/%s", currentUser))
    if err := chmodCmd.Run(); err != nil {
        return fmt.Errorf("failed to set permissions on sudoers file: %w", err)
    }

    return nil
}

// InitServerSetup은 서버 설정 메뉴를 EditServer 메뉴에 추가
func (app *App) InitServerSetup() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Server Setup")

    options := []string{
        "1. Setup All Servers",
        "2. Back to Server Management",
    }

    choice := ui.ArrowSelect(options)
    ui.Clear()

    switch choice {
    case 0:
        if err := app.SetupServers(); err != nil {
            ui.Red.Printf("Error setting up servers: %v\n", err)
        } else {
            ui.Green.Println("Successfully completed server setup!")
        }
        fmt.Print("\nPress Enter to continue...")
        fmt.Scanln()
    case 1:
        return
    }
}

package app

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "text/template"
    "go-project/internal/config"
    "go-project/internal/ui"

    "github.com/fatih/color"
    "gopkg.in/yaml.v3"
)

var (
    red    = color.New(color.FgRed)
    green  = color.New(color.FgGreen)
    yellow = color.New(color.FgYellow)
)

type Server struct {
    Name   string   `yaml:"name"`
    IP     string   `yaml:"ip"`
    Roles  []string `yaml:"roles"`
    Labels []string `yaml:"labels"`
}

type ServerConfig struct {
    ALLServers []Server `yaml:"ALL_Servers"`
}


// CreateHostsFiles generates both /etc/hosts and /etc/ansible/hosts files
func (app *App) CreateHostsFiles() error {
    // 사용자 정의 함수 맵 생성
    funcMap := template.FuncMap{
        "filterMasterServers": func(servers []Server) []Server {
            var masterServers []Server
            for _, server := range servers {
                if server.Roles[0] == "control-plane" {
                    masterServers = append(masterServers, server)
                }
            }
            return masterServers
        },
    }

    // Create /etc/hosts file
    hostsTempl, err := template.New("hosts").Parse(config.HostsTemplate)
    if err != nil {
        return fmt.Errorf("error parsing hosts template: %w", err)
    }

    hostsFile, err := os.Create("/etc/hosts")
    if err != nil {
        return fmt.Errorf("error creating hosts file: %w", err)
    }
    defer func() {
        if cerr := hostsFile.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("error closing hosts file: %w", cerr)
        }
    }()

    if err := hostsTempl.Execute(hostsFile, app.ServerConfig); err != nil {
        return fmt.Errorf("error writing hosts file: %w", err)
    }

    // Create /etc/ansible directory and hosts file
    if err := os.MkdirAll("/etc/ansible", 0755); err != nil {
        return fmt.Errorf("error creating ansible directory: %w", err)
    }

    // 사용자 정의 함수 맵을 템플릿에 추가
    ansibleHostsTempl, err := template.New("ansible_hosts").Funcs(funcMap).Parse(config.AnsibleHostsTemplate)
    if err != nil {
        return fmt.Errorf("error parsing ansible hosts template: %w", err)
    }

    ansibleHostsFile, err := os.Create("/etc/ansible/hosts")
    if err != nil {
        return fmt.Errorf("error creating ansible hosts file: %w", err)
    }
    defer func() {
        if cerr := ansibleHostsFile.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("error closing ansible hosts file: %w", err)
        }
    }()

    if err := ansibleHostsTempl.Execute(ansibleHostsFile, app.ServerConfig); err != nil {
        return fmt.Errorf("error writing ansible hosts file: %w", err)
    }

    return nil
}
func (app *App) SetupServerConfig() error {
    // 1. 먼저 /etc/openmsa의 기존 서버 설정 파일 확인
    existingConfigPath := "/etc/openmsa/group_vars/all/servers.yaml"
    
    // 2. 기존 설정 파일이 존재하면 로드
    if _, err := os.Stat(existingConfigPath); err == nil {
        // 기존 파일 로드
        configData, err := os.ReadFile(existingConfigPath)
        if err != nil {
            return fmt.Errorf("error reading existing server config: %w", err)
        }

        var config ServerConfig
        if err := yaml.Unmarshal(configData, &config); err != nil {
            return fmt.Errorf("error parsing existing server config: %w", err)
        }

        app.ServerConfig = config

        // hosts 파일 생성
        return app.CreateHostsFiles()
    }

    // 3. 설정 파일이 없는 경우에만 임베디드 설정 사용
    configData, err := playbooks.ReadFile("playbooks/group_vars/all/servers.yaml")
    if err != nil {
        // 파일이 없으면 빈 설정으로 초기화
        defaultConfig := ServerConfig{
            ALLServers: []Server{},
        }
        app.ServerConfig = defaultConfig
    } else {
        // YAML 파싱
        var config ServerConfig
        if err := yaml.Unmarshal(configData, &config); err != nil {
            return fmt.Errorf("error parsing server config: %w", err)
        }
        app.ServerConfig = config
    }

    // 4. 디렉토리 생성
    if err := os.MkdirAll(filepath.Dir(existingConfigPath), 0755); err != nil {
        return fmt.Errorf("error creating config directory: %v", err)
    }

    // 5. 초기 설정 파일 쓰기 (첫 실행 시에만)
    if err := os.WriteFile(existingConfigPath, configData, 0644); err != nil {
        return fmt.Errorf("error creating initial server config: %v", err)
    }

    // 6. hosts 파일 생성
    return app.CreateHostsFiles()
}

func (app *App) saveServerConfig() error {
    // 경로를 /etc/openmsa로 고정
    configPath := "/etc/openmsa/group_vars/all/servers.yaml"

    // 설정 디렉토리 확인 및 생성
    configDir := filepath.Dir(configPath)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return fmt.Errorf("error creating config directory: %w", err)
    }

    // YAML 마샬링
    data, err := yaml.Marshal(app.ServerConfig)
    if err != nil {
        return fmt.Errorf("error marshaling server config: %w", err)
    }

    // 파일 저장
    if err := os.WriteFile(configPath, data, 0644); err != nil {
        return fmt.Errorf("error writing server config: %w", err)
    }

    // hosts 파일 업데이트
    return app.CreateHostsFiles()
}

func (app *App) EditServer() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Server Management")

    options := []string{
        "1. List All Servers",
        "2. Add New Server",
        "3. Edit Existing Server",
        "4. Delete Server",
        "5. Setup Servers",
        "6. Back to Main Menu",
    }

    for {
        choice := ui.ArrowSelect(options)
        ui.Clear()

        switch choice {
        case 0:
            app.listServers()
        case 1:
            app.addServer()
        case 2:
            app.editExistingServer()
        case 3:
            app.deleteServer()
        case 4:
            app.InitServerSetup()
        case 5:
            return
        }
    }
}

func (app *App) listServers() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Server List")

    if len(app.ServerConfig.ALLServers) == 0 {
        yellow.Println("No servers configured.")
    } else {
        for _, server := range app.ServerConfig.ALLServers {
            fmt.Printf("Name: %s\n", server.Name)
            fmt.Printf("IP: %s\n", server.IP)
            fmt.Printf("Roles: %s\n", strings.Join(server.Roles, ", "))
            fmt.Printf("Labels: %s\n", strings.Join(server.Labels, ", "))
            fmt.Println("------------------------")
        }
    }

    fmt.Print("\nPress Enter to continue...")
    bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func validateServerInput(name, ip string) error {
    if strings.TrimSpace(name) == "" {
        return fmt.Errorf("server name cannot be empty")
    }
    if strings.TrimSpace(ip) == "" {
        return fmt.Errorf("server IP cannot be empty")
    }
    return nil
}

func processInputString(input string) []string {
    if strings.TrimSpace(input) == "" {
        return []string{}
    }
    items := strings.Split(strings.TrimSpace(input), ",")
    for i := range items {
        items[i] = strings.TrimSpace(items[i])
    }
    return items
}

func (app *App) addServer() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Add New Server")

    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter server name: ")
    name, _ := reader.ReadString('\n')
    name = strings.TrimSpace(name)

    fmt.Print("Enter server IP: ")
    ip, _ := reader.ReadString('\n')
    ip = strings.TrimSpace(ip)

    if err := validateServerInput(name, ip); err != nil {
        red.Printf("Error: %v\n", err)
        fmt.Print("\nPress Enter to continue...")
        reader.ReadString('\n')
        return
    }

    fmt.Print("Enter roles (comma-separated): ")
    rolesStr, _ := reader.ReadString('\n')
    roles := processInputString(rolesStr)

    fmt.Print("Enter labels (comma-separated): ")
    labelsStr, _ := reader.ReadString('\n')
    labels := processInputString(labelsStr)

    newServer := Server{
        Name:   name,
        IP:     ip,
        Roles:  roles,
        Labels: labels,
    }

    app.ServerConfig.ALLServers = append(app.ServerConfig.ALLServers, newServer)

    if err := app.saveServerConfig(); err != nil {
        red.Printf("Error: %v\n", err)
    } else {
        green.Println("Server added successfully!")
    }

    fmt.Print("\nPress Enter to continue...")
    reader.ReadString('\n')
}

func (app *App) editExistingServer() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Edit Existing Server")

    if len(app.ServerConfig.ALLServers) == 0 {
        yellow.Println("No servers available to edit.")
        fmt.Print("\nPress Enter to continue...")
        bufio.NewReader(os.Stdin).ReadBytes('\n')
        return
    }

    options := make([]string, len(app.ServerConfig.ALLServers)+1)
    for i, server := range app.ServerConfig.ALLServers {
        options[i] = fmt.Sprintf("%s (%s)", server.Name, server.IP)
    }
    options[len(options)-1] = "Back"

    choice := ui.ArrowSelect(options)
    if choice == len(options)-1 {
        return
    }

    server := &app.ServerConfig.ALLServers[choice]
    reader := bufio.NewReader(os.Stdin)

    ui.Clear()
    fmt.Printf("Editing server: %s\n\n", server.Name)

    fmt.Printf("Current name: %s\nEnter new name (or press Enter to keep current): ", server.Name)
    if name, _ := reader.ReadString('\n'); strings.TrimSpace(name) != "" {
        server.Name = strings.TrimSpace(name)
    }

    fmt.Printf("Current IP: %s\nEnter new IP (or press Enter to keep current): ", server.IP)
    if ip, _ := reader.ReadString('\n'); strings.TrimSpace(ip) != "" {
        server.IP = strings.TrimSpace(ip)
    }

    fmt.Printf("Current roles: %s\nEnter new roles (comma-separated, or press Enter to keep current): ",
        strings.Join(server.Roles, ","))
    if rolesStr, _ := reader.ReadString('\n'); strings.TrimSpace(rolesStr) != "" {
        server.Roles = processInputString(rolesStr)
    }

    fmt.Printf("Current labels: %s\nEnter new labels (comma-separated, or press Enter to keep current): ",
        strings.Join(server.Labels, ","))
    if labelsStr, _ := reader.ReadString('\n'); strings.TrimSpace(labelsStr) != "" {
        server.Labels = processInputString(labelsStr)
    }

    if err := app.saveServerConfig(); err != nil {
        red.Printf("Error: %v\n", err)
    } else {
        green.Println("Server updated successfully!")
    }

    fmt.Print("\nPress Enter to continue...")
    reader.ReadString('\n')
}
func (app *App) deleteServer() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Delete Server")

    if len(app.ServerConfig.ALLServers) == 0 {
        yellow.Println("No servers available to delete.")
        fmt.Print("\nPress Enter to continue...")
        bufio.NewReader(os.Stdin).ReadBytes('\n')
        return
    }

    options := make([]string, len(app.ServerConfig.ALLServers)+1)
    for i, server := range app.ServerConfig.ALLServers {
        options[i] = fmt.Sprintf("%s (%s)", server.Name, server.IP)
    }
    options[len(options)-1] = "Back"

    choice := ui.ArrowSelect(options)
    if choice == len(options)-1 {
        return
    }

    reader := bufio.NewReader(os.Stdin)
    fmt.Printf("Are you sure you want to delete server '%s'? (y/n): ",
        app.ServerConfig.ALLServers[choice].Name)
    confirm, _ := reader.ReadString('\n')

    if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
        app.ServerConfig.ALLServers = append(
            app.ServerConfig.ALLServers[:choice],
            app.ServerConfig.ALLServers[choice+1:]...,
        )

        if err := app.saveServerConfig(); err != nil {
            red.Printf("Error: %v\n", err)
        } else {
            green.Println("Server deleted successfully!")
        }
    } else {
        yellow.Println("Deletion cancelled.")
    }

    fmt.Print("\nPress Enter to continue...")
    reader.ReadString('\n')
}

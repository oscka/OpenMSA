package app

import (
        "bufio"
        "embed"
        "fmt"
        "io/fs"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "go-project/internal/ui"
        "go-project/internal/utils"
        "gopkg.in/yaml.v3"
)

//go:embed playbooks/*
var playbooks embed.FS

type App struct {
        PlayNames          []string
        Tags              []string
        AnsiblePath       string
        Playbook          Playbook
        MainPlaybook      []byte
        UninstallPlaybook []byte
        TempDir           string
        ServerConfig      ServerConfig
        ConfigFile        string
        InstallConfig     *InstallConfig
        CatalogConfig     CatalogConfig
        CatalogConfigFile string
}

type Play struct {
        Name string   `yaml:"name"`
        Tags []string `yaml:"tags"`
}

type Playbook []Play


func NewApp() *App {
    app := &App{
        InstallConfig: NewInstallConfig(),
        CatalogConfig: CatalogConfig{
            Catalogs: make(map[string]bool),
        },
    }
    return app
}


func (app *App) Cleanup() {
    // /etc/openmsa 폴더를 완전히 삭제하지 않음
    // 필요한 경우 특정 파일만 정리할 수 있음
    ui.Yellow.Println("Cleanup method called. Keeping /etc/openmsa directory.")
}


func (app *App) InitializeAnsible() error {
        path, err := utils.FindExecutable("ansible-playbook")
        if err != nil {
                return fmt.Errorf("ansible-playbook is not installed: %v", err)
        }
        app.AnsiblePath = path
        return nil
}

func (app *App) CopyPlaybookStructure() error {
    // 고정된 디렉토리 경로 설정
    targetDir := "/etc/openmsa"

    // 디렉토리가 존재하지 않으면 생성
    if _, err := os.Stat(targetDir); os.IsNotExist(err) {
        if err := os.MkdirAll(targetDir, 0755); err != nil {
            return fmt.Errorf("failed to create /etc/openmsa directory: %v", err)
        }
    } else if err != nil {
        return fmt.Errorf("error checking /etc/openmsa directory: %v", err)
    }

    app.TempDir = targetDir

    // playbooks 디렉토리의 기본 경로 설정
    playbooksRoot := "playbooks"

    var content []byte
    var err error
    // InstallType이 설정되어 있는 경우 해당 플레이북을 사용
    if app.InstallConfig != nil && app.InstallConfig.InstallType != "" {
        playbookName := app.InstallConfig.GetPlaybookName()
        content, err = playbooks.ReadFile(filepath.Join(playbooksRoot, playbookName))
        if err != nil {
            ui.Yellow.Printf("Specific playbook not found, using default playbook\n")
            // 기본 플레이북으로 폴백
            content, err = playbooks.ReadFile(filepath.Join(playbooksRoot, "playbook.yaml"))
            if err != nil {
                return fmt.Errorf("failed to read default playbook: %v", err)
            }
        }
    } else {
        // InstallType이 설정되어 있지 않은 경우 기본 플레이북 사용
        content, err = playbooks.ReadFile(filepath.Join(playbooksRoot, "playbook.yaml"))
        if err != nil {
            return fmt.Errorf("failed to read default playbook: %v", err)
        }
    }

    app.MainPlaybook = content

    // 메인 플레이북 파일이 이미 존재하지 않는 경우에만 쓰기
    playbookPath := filepath.Join(targetDir, "playbook.yaml")
    if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
        if err := os.WriteFile(playbookPath, content, 0644); err != nil {
            return fmt.Errorf("failed to write main playbook: %v", err)
        }
    }

    // Copy other necessary files from embedded FS
    err = fs.WalkDir(playbooks, playbooksRoot, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        relPath, err := filepath.Rel(playbooksRoot, path)
        if err != nil {
            return err
        }


        // playbooks 디렉토리를 제거한 상대 경로로 대상 경로 구성
        destPath := filepath.Join(targetDir, relPath)
        if filepath.HasPrefix(relPath, "playbooks") {
        relPath = relPath[len("playbooks/"):] // "playbooks/" 제거
        destPath = filepath.Join(targetDir, relPath)
        }
        // 디렉토리의 경우 존재하지 않으면 생성
        if d.IsDir() {
            return os.MkdirAll(destPath, 0755)
        }

        content, err := playbooks.ReadFile(path)
        if err != nil {
            return fmt.Errorf("failed to read embedded file %s: %v", path, err)
        }

        if relPath == "uninstall-playbook.yaml" {
            app.UninstallPlaybook = content
        }

        // 파일이 존재하지 않는 경우에만 쓰기
        if _, err := os.Stat(destPath); os.IsNotExist(err) {
            dir := filepath.Dir(destPath)
            if err := os.MkdirAll(dir, 0755); err != nil {
                return err
            }

            if err := os.WriteFile(destPath, content, 0644); err != nil {
                return err
            }
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("failed to copy playbook structure: %v", err)
    }

    // Parse main playbook
    if err := yaml.Unmarshal(app.MainPlaybook, &app.Playbook); err != nil {
        return fmt.Errorf("failed to parse playbook: %v", err)
    }

    // 기존의 Setup 로직들 유지
    if err := app.SetupCatalogConfig(); err != nil {
        ui.Red.Printf("Error setting up catalog configuration: %v\n", err)
    }

    // Extract names and tags
    app.PlayNames = nil // Reset PlayNames
    app.Tags = nil      // Reset Tags
    seen := make(map[string]bool)
    for _, play := range app.Playbook {
        app.PlayNames = append(app.PlayNames, play.Name)
        for _, tag := range play.Tags {
            if !seen[tag] {
                app.Tags = append(app.Tags, tag)
                seen[tag] = true
            }
        }
    }

    if err := app.SetupServerConfig(); err != nil {
        ui.Red.Printf("Error setting up server configuration: %v\n", err)
    }

    if err := app.SetupCatalogConfig(); err != nil {
        ui.Red.Printf("Error setting up catalog configuration: %v\n", err)
    }

    return nil
}

func (app *App) RunPlaybook(isUninstall bool, args ...string) error {
        ui.Green.Println("Running Ansible Playbook...")

        var playbookPath string
        if isUninstall {
                playbookPath = filepath.Join(app.TempDir, "uninstall-playbook.yaml")
        } else {
                // Dynamically determine the playbook name based on installation type
                if app.InstallConfig != nil && app.InstallConfig.InstallType != "" {
                        playbookName := app.InstallConfig.GetPlaybookName()
                        playbookPath = filepath.Join(app.TempDir, playbookName)
                        
                        // Fallback to default if specific playbook doesn't exist
                        if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
                                ui.Yellow.Printf("Specific playbook %s not found, falling back to default playbook\n", playbookName)
                                playbookPath = filepath.Join(app.TempDir, "playbook.yaml")
                        }
                } else {
                        playbookPath = filepath.Join(app.TempDir, "playbook.yaml")
                }
        }

        // 플레이북 파일 존재 확인
        if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
                return fmt.Errorf("playbook file not found at %s", playbookPath)
        }

        ui.Yellow.Printf("Executing playbook: %s\n", playbookPath)

        cmdArgs := append([]string{
                "sudo",
                app.AnsiblePath,
                "-i", "/etc/ansible/hosts",
                playbookPath,
        }, args...)

        cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

        // Ansible 환경 변수 설정
        callbackPath := filepath.Join(app.TempDir, "callback_plugins")
        if _, err := os.Stat(callbackPath); !os.IsNotExist(err) {
                cmd.Env = append(os.Environ(),
                        fmt.Sprintf("ANSIBLE_CALLBACK_PLUGINS=%s", callbackPath))
        }

        if err := cmd.Run(); err != nil {
                ui.Red.Printf("Error occurred while running playbook: %v\n", err)
                return err
        }

        ui.Green.Println("Playbook execution completed successfully.")
        return nil
}

func (app *App) GetTagForPlay(playName string) string {
        for _, play := range app.Playbook {
                if play.Name == playName && len(play.Tags) > 0 {
                        return play.Tags[0]
                }
        }
        return ""
}

func (app *App) SelectAction() {
        // 초기 플레이북 구조 복사
        if err := app.LoadCatalogConfig(); err != nil {
          ui.Red.Printf("Error loading catalog configuration: %v\n", err)
        }
        if err := app.CopyPlaybookStructure(); err != nil {
                ui.Red.Printf("Error copying playbook structure: %v\n", err)
                os.Exit(1)
        }

        for {
                ui.Clear()
                ui.PrintLogo()
                ui.PrintMenuTitle("Select an action to execute:")

                options := []string{
                        "1. Select Installation Type",
                        "2. Install All",
                        "3. Install Single Task",
                        "4. Install from Specific Task",
                        "5. Server Management",
                        "6. Catalog Management",
                        "7. Uninstall All",
                        "8. Exit",
                }

                choice := ui.ArrowSelect(options)
                ui.Clear()

                switch choice {
                case 0:
                        if err := app.InstallConfig.SelectInstallationType(); err != nil {
                                ui.Red.Printf("Error selecting installation type: %v\n", err)
                                continue
                        }
                        if err := app.CopyPlaybookStructure(); err != nil {
                                ui.Red.Printf("Error copying playbook: %v\n", err)
                                continue
                        }
                case 1:
                        if app.InstallConfig.InstallType == "" {
                                ui.Yellow.Println("Please select installation type first")
                                continue
                        }
                        app.RunPlaybook(false)
                        ui.Yellow.Println("Complete installation finished. Exiting OpenMSA.")
                        os.Exit(0)
                case 2:
                        if app.InstallConfig.InstallType == "" {
                                ui.Yellow.Println("Please select installation type first")
                                continue
                        }
                        app.SelectPlay("single")
                case 3:
                        if app.InstallConfig.InstallType == "" {
                                ui.Yellow.Println("Please select installation type first")
                                continue
                        }
                        app.SelectPlay("after")
                case 4:
                        app.EditServer()
                case 5:
                        app.ManageCatalogs()
                case 6:
                        app.ConfirmUninstall()
                case 7:
                        ui.Yellow.Println("Exiting OpenMSA.")
                        os.Exit(0)
                }
        }
}

func (app *App) ConfirmUninstall() {
        ui.Clear()
        reader := bufio.NewReader(os.Stdin)
        ui.Yellow.Print("Are you sure you want to proceed with complete uninstallation? (y/n) ")
        text, _ := reader.ReadString('\n')

        if strings.ToLower(strings.TrimSpace(text)) == "y" {
                app.RunPlaybook(true)
                os.Exit(0)
        } else {
                ui.Yellow.Println("Uninstallation cancelled.")
        }
}

func (app *App) SelectPlay(mode string) {
        ui.Clear()
        ui.PrintLogo()
        ui.PrintMenuTitle("Select a task to install:")

        options := make([]string, len(app.PlayNames)+1)
        for i, name := range app.PlayNames {
                options[i] = fmt.Sprintf("%d. %s", i+1, name)
        }
        options[len(options)-1] = "Go back"

        choice := ui.ArrowSelect(options)
        if choice == len(options)-1 {
                return
        }

        selectedPlay := app.PlayNames[choice]
        selectedTag := app.GetTagForPlay(selectedPlay)

        ui.Clear()

        if mode == "single" {
                ui.Cyan.Printf("Selected task: %s\n", selectedPlay)
                app.RunPlaybook(false, "--tags", selectedTag, "-v")
                os.Exit(0)
        } else {
                var tagsToRun []string
                startFound := false

                for _, play := range app.PlayNames {
                        if startFound || play == selectedPlay {
                                startFound = true
                                playTag := app.GetTagForPlay(play)
                                tagsToRun = append(tagsToRun, playTag)
                        }
                }

                if len(tagsToRun) == 0 {
                        ui.Red.Println("Error: Cannot find tags after selected task.")
                        return
                }

                tagList := strings.Join(tagsToRun, ",")
                ui.Green.Printf("Installing all tasks after '%s'.\n", selectedPlay)
                app.RunPlaybook(false, "--tags", tagList)
                os.Exit(0)
        }
}

package app

import (
        "bufio"
        "fmt"
        "sort"
        "os"
        "path/filepath"
        "go-project/internal/ui"
        "github.com/fatih/color"
        "gopkg.in/yaml.v3"
)

type CatalogConfig struct {
    Catalogs       map[string]bool `yaml:"DEPLOY_CATALOGS,omitempty"`
    PersistentFile string          `yaml:"-"`
}

func (app *App) LoadCatalogConfig() error {
    // 대상 경로를 /etc/openmsa로 직접 지정
    persistentConfigPath := "/etc/openmsa/group_vars/all/catalogs.yaml"

    // 디렉토리 존재 확인 및 생성
    if err := os.MkdirAll(filepath.Dir(persistentConfigPath), 0755); err != nil {
        return fmt.Errorf("could not create config directory: %v", err)
    }

    // 파일이 없으면 embedded 설정 복사
    if _, err := os.Stat(persistentConfigPath); os.IsNotExist(err) {
        configData, err := playbooks.ReadFile("playbooks/group_vars/all/catalogs.yaml")
        if err != nil {
            return fmt.Errorf("error reading embedded catalog config: %v", err)
        }

        if err := os.WriteFile(persistentConfigPath, configData, 0644); err != nil {
            return fmt.Errorf("error creating initial persistent config: %v", err)
        }
    }

    // 설정 파일 읽기
    configData, err := os.ReadFile(persistentConfigPath)
    if err != nil {
        return fmt.Errorf("error reading persistent catalog config: %v", err)
    }

    var fullConfig map[string]interface{}
    if err := yaml.Unmarshal(configData, &fullConfig); err != nil {
        return fmt.Errorf("error parsing catalog config: %v", err)
    }

    // 카탈로그 추출
    catalogs := make(map[string]bool)
    for key, value := range fullConfig {
        if key != "DEPLOY_CATALOGS" &&
           len(key) > 7 &&
           key[:7] == "DEPLOY_" &&
           value != nil {

            switch v := value.(type) {
            case bool:
                catalogs[key[7:]] = v
            }
        }
    }

    // 앱의 카탈로그 설정 업데이트
    app.CatalogConfig = CatalogConfig{
        Catalogs:       catalogs,
        PersistentFile: persistentConfigPath,
    }

    return nil
}

func (app *App) SetupCatalogConfig() error {
    // 1. 먼저 기존 설정 파일 확인
    existingConfigPath := "/etc/openmsa/group_vars/all/catalogs.yaml"
    
    // 기존 설정 파일이 존재하면 로드
    if _, err := os.Stat(existingConfigPath); err == nil {
        return app.LoadCatalogConfig()
    }

    // 2. 설정 파일이 없는 경우에만 embedded 설정 사용
    configData, err := playbooks.ReadFile("playbooks/group_vars/all/catalogs.yaml")
    if err != nil {
        return fmt.Errorf("error reading catalog config: %v", err)
    }

    // 3. 디렉토리 생성
    if err := os.MkdirAll(filepath.Dir(existingConfigPath), 0755); err != nil {
        return fmt.Errorf("error creating config directory: %v", err)
    }

    // 4. 초기 설정 파일 쓰기
    if err := os.WriteFile(existingConfigPath, configData, 0644); err != nil {
        return fmt.Errorf("error copying catalog config: %v", err)
    }

    // 5. 설정 로드
    return app.LoadCatalogConfig()
}

func (app *App) saveCatalogConfig() error {
    // /etc/openmsa의 catalogs.yaml 파일 업데이트
    persistentConfigPath := "/etc/openmsa/group_vars/all/catalogs.yaml"

    // 디렉토리 존재 확인 및 생성
    if err := os.MkdirAll(filepath.Dir(persistentConfigPath), 0755); err != nil {
        return fmt.Errorf("error creating config directory: %v", err)
    }

    // 전체 설정 생성
    fullConfig := make(map[string]interface{})
    for name, enabled := range app.CatalogConfig.Catalogs {
        fullConfig[fmt.Sprintf("DEPLOY_%s", name)] = enabled
    }

    // YAML로 마샬링
    persistentData, err := yaml.Marshal(&fullConfig)
    if err != nil {
        return fmt.Errorf("error marshaling catalog config: %w", err)
    }

    // /etc/openmsa/group_vars/all/catalogs.yaml에 쓰기
    if err := os.WriteFile(persistentConfigPath, persistentData, 0644); err != nil {
        return fmt.Errorf("error writing catalog config: %w", err)
    }

    ui.Green.Println("Catalog configuration updated successfully!")
    return nil
}


func (app *App) ManageCatalogs() {
        ui.Clear()
        ui.PrintLogo()
        ui.PrintMenuTitle("Catalog Management")

        options := []string{
                "1. List Catalogs",
                "2. Toggle Catalog Deployment",
                "3. Back to Main Menu",
        }

        for {
                choice := ui.ArrowSelect(options)
                ui.Clear()

                switch choice {
                case 0:
                        app.listCatalogs()
                case 1:
                        app.toggleCatalogDeployment()
                case 2:
                        return
                }
        }
}

func (app *App) listCatalogs() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Catalog List")

    if len(app.CatalogConfig.Catalogs) == 0 {
        color.Yellow("No catalogs configured.")
    } else {
        // 카탈로그 이름을 정렬하여 출력
        catalogNames := make([]string, 0, len(app.CatalogConfig.Catalogs))
        for name := range app.CatalogConfig.Catalogs {
            catalogNames = append(catalogNames, name)
        }
        sort.Strings(catalogNames)

        for _, name := range catalogNames {
            status := "Disabled"
            if app.CatalogConfig.Catalogs[name] {
                status = "Enabled"
            }
            fmt.Printf("%s: %s\n", name, status)
        }
    }

    fmt.Print("\nPress Enter to continue...")
    bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func (app *App) toggleCatalogDeployment() {
    ui.Clear()
    ui.PrintLogo()
    ui.PrintMenuTitle("Toggle Catalog Deployment")

    // 카탈로그 이름을 미리 정렬해서 생성
    catalogNames := make([]string, 0, len(app.CatalogConfig.Catalogs))
    for name := range app.CatalogConfig.Catalogs {
        catalogNames = append(catalogNames, name)
    }
    sort.Strings(catalogNames) // 알파벳 순으로 정렬

    currentIndex := 0

    for {
        // 옵션 생성 (정렬된 catalogNames 사용)
        options := make([]string, len(catalogNames)+1)
        
        for i, name := range catalogNames {
            status := color.YellowString("[ ]")
            if app.CatalogConfig.Catalogs[name] {
                status = color.GreenString("[✓]")
            }
            options[i] = fmt.Sprintf("%s %s", status, color.CyanString(name))
        }
        options[len(options)-1] = color.RedString("Save and Exit")

        // 현재 인덱스 유지하며 선택
        choice := ui.ArrowSelect(options, currentIndex)

        // 종료 조건
        if choice == len(options)-1 {
            if err := app.saveCatalogConfig(); err != nil {
                color.Red("Error saving catalog config: %v", err)
                fmt.Print("\nPress Enter to continue...")
                bufio.NewReader(os.Stdin).ReadBytes('\n')
            } else {
                color.Green("Catalog configuration saved successfully!")
            }
            return
        }

        // 카탈로그 상태 토글
        catalogName := catalogNames[choice]
        app.CatalogConfig.Catalogs[catalogName] = !app.CatalogConfig.Catalogs[catalogName]
        
        // 현재 인덱스 유지
        currentIndex = choice
    }
}

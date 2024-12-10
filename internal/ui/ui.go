package ui

import (
        "fmt"
        "os"
        "os/exec"
        "github.com/fatih/color"
        "github.com/nsf/termbox-go"
)

var (
        Red    = color.New(color.FgRed)
        Green  = color.New(color.FgGreen)
        Yellow = color.New(color.FgYellow)
        Blue   = color.New(color.FgBlue)
        Cyan   = color.New(color.FgCyan)
)

func Clear() {
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
}

func PrintLogo() {
        Clear()
        Blue.Add(color.Bold).Println(`
⠀⢀⣤⣶⣶⣦⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⢠⣿⠏⠁⠀⠙⢿⣧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⣿⡄⠀⠀⢸⣿⡀⣴⣿⠿⠿⠿⠿⠿⢣⣿⣷⡀⠀⠀⠀
⢸⣿⠀⠀⠀⠀⢸⣿⣴⣶⣶⣄⣠⣴⣶⣦⡀⣠⣶⣶⣄⠀⠀⣾⣿⣷⠀⢠⣿⢿⡇⣿⣇⠀⠀⠀⠀⢀⣾⡟⢿⣧⠀⠀⠀
⠸⣿⣄⠀⠀⢀⣾⣿⠋⠀⠙⣿⣿⣁⣀⣹⣿⣿⠁⠈⣿⡇⢀⣿⠉⣿⡆⣼⡟⠸⣷⠙⠿⠿⠿⠿⣷⣾⡟⠀⠈⢿⣧⠀⠀
⠀⠙⠿⣿⣾⡿⢿⣿⣄⠀⢀⣿⣿⡛⠛⠛⢻⣿⠀⠀⣿⡇⣸⡟⠀⠸⣿⣿⠁⠀⣿⡆⠀⠀⠀⢀⣿⡿⠁⠀⠀⠘⣿⣆⠀
⠀⠀⠀⠀⠀⠀⢸⣿⠻⠿⠿⠋⠙⠿⠿⠿⠸⠿⠀⠀⠿⠧⠿⠇⠀⠀⠀⠀⠀⠀⠹⠿⠿⠿⠿⠿⠟⠱⠿⠿⠿⠿⠿⠿⠄
⠀⠀⠀⠀⠀⠀⠸⠿`)
        Cyan.Add(color.Bold).Println("(Open MSA)")
        fmt.Println()
}

func PrintMenuTitle(title string) {
        fmt.Println()
        Green.Println(title)
        fmt.Println()
}

func ArrowSelect(options []string, startIndex ...int) int {
    if err := termbox.Init(); err != nil {
        panic(err)
    }
    defer termbox.Close()

    // 시작 인덱스 설정 (기본값 0)
    selected := 0
    if len(startIndex) > 0 {
        selected = startIndex[0]
        // 유효한 범위로 제한
        if selected < 0 || selected >= len(options) {
            selected = 0
        }
    }

    printOptions := func() {
        termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
        PrintLogo()
        for i, opt := range options {
            if i == selected {
                Green.Printf("> %s <\n", opt)
            } else {
                fmt.Printf("  %s  \n", opt)
            }
        }
        fmt.Printf("\033[?25l")
        termbox.Flush()
    }

    printOptions()
    for {
        switch ev := termbox.PollEvent(); ev.Type {
        case termbox.EventKey:
            switch ev.Key {
            case termbox.KeyArrowUp:
                selected--
                if selected < 0 {
                    selected = len(options) - 1
                }
                printOptions()
            case termbox.KeyArrowDown:
                selected++
                if selected >= len(options) {
                    selected = 0
                }
                printOptions()
            case termbox.KeyEnter:
                fmt.Printf("\033[?25h")
                return selected
            case termbox.KeyEsc:
                fmt.Printf("\033[?25h")
                return len(options) - 1
            }
        }
    }
}

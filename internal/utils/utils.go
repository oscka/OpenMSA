package utils

import (
        "fmt"
        "os/exec"
)

func FindExecutable(name string) (string, error) {
        path, err := exec.LookPath(name)
        if err == nil {
                return path, nil
        }
        return "", fmt.Errorf("%s not found in system", name)
}

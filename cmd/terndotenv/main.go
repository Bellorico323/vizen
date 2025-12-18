package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

func main() {
	envFile := ".env"

	if len(os.Args) > 1 {
		envFile = os.Args[1]
	}

	fmt.Printf("Carregando variáveis de: %s\n", envFile)

	if err := godotenv.Load(envFile); err != nil {
		panic(err)
	}

	cmd := exec.Command("tern", "migrate", "--migrations", "./internal/store/pgstore/migrations", "--config", "./internal/store/pgstore/migrations/tern.conf")

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Command execution failed ", err)
		fmt.Println("Output: ", string(output))
		return
	}

	fmt.Println("Command executed successfully ", string(output))
}

package main

import (
	"flag"
	"fmt"
	"os/exec"

	"github.com/joho/godotenv"
)

func main() {
	envFile := flag.String("env", ".env", "Caminho para o arquivo .env")
	shouldDrop := flag.Bool("drop", false, "Se verdadeiro, reverte todas as migrações (Drop All)")
	flag.Parse()

	fmt.Printf("Carregando variáveis de: %s\n", *envFile)

	if err := godotenv.Load(*envFile); err != nil {
		panic(fmt.Sprintf("Erro ao carregar .env: %v", err))
	}

	cmdArgs := []string{
		"migrate",
		"--migrations", "./internal/store/pgstore/migrations",
		"--config", "./internal/store/pgstore/migrations/tern.conf",
	}

	if *shouldDrop {
		fmt.Println("⚠️  MODO DROP ATIVADO: Revertendo todas as tabelas (destination 0)...")
		cmdArgs = append(cmdArgs, "--destination", "0")
	} else {
		fmt.Println("🚀 Rodando migrações (UP)...")
	}

	cmd := exec.Command("tern", cmdArgs...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("❌ Falha na execução do comando:", err)
		fmt.Println("Output detalhado:\n", string(output))
		return
	}

	fmt.Println("✅ Comando executado com sucesso!\n", string(output))
}

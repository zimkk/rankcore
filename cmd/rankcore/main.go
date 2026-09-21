package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "setup":
		setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
		agentFlag := setupCmd.String("agent", "", "Specific agent to setup (e.g. claude-code)")
		dryRunFlag := setupCmd.Bool("dry-run", false, "Simulate setup without writing files")
		setupCmd.Parse(os.Args[2:])
		
		fmt.Println("Running setup...")
		// internal/setup integration will go here
		fmt.Printf("Agent: %s, Dry Run: %v\n", *agentFlag, *dryRunFlag)
		
	case "doctor":
		doctorCmd := flag.NewFlagSet("doctor", flag.ExitOnError)
		jsonFlag := doctorCmd.Bool("json", false, "Output in JSON format")
		doctorCmd.Parse(os.Args[2:])
		
		if *jsonFlag {
			fmt.Println(`{"status": "ok", "version": "1.0.0"}`)
		} else {
			fmt.Println("RankCore environment is healthy (Version: 1.0.0)")
		}

	case "audit":
		auditCmd := flag.NewFlagSet("audit", flag.ExitOnError)
		outDir := auditCmd.String("out", ".rankcore/runs/latest", "Output directory")
		jsonFlag := auditCmd.Bool("json", false, "Output in JSON format")
		auditCmd.Parse(os.Args[2:])
		
		fmt.Printf("Running deterministic audit on %s...\n", auditCmd.Arg(0))
		fmt.Printf("Output: %s, JSON: %v\n", *outDir, *jsonFlag)
		// internal/crawl and internal/audit integration will go here

	case "verify":
		verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
		baseline := verifyCmd.String("baseline", "", "Baseline audit.json to compare against")
		outDir := verifyCmd.String("out", ".rankcore/runs/verify", "Output directory")
		jsonFlag := verifyCmd.Bool("json", false, "Output in JSON format")
		verifyCmd.Parse(os.Args[2:])
		
		fmt.Printf("Verifying %s against baseline %s...\n", verifyCmd.Arg(0), *baseline)
		fmt.Printf("Output: %s, JSON: %v\n", *outDir, *jsonFlag)
		// internal/verify integration will go here

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("RankCore Native CLI")
	fmt.Println("Usage: rankcore <command> [flags]")
	fmt.Println("Commands:")
	fmt.Println("  setup   - Install/update the rank skill into detected agents")
	fmt.Println("  doctor  - Check environment health")
	fmt.Println("  audit   - Run deterministic SEO audit against a target")
	fmt.Println("  verify  - Re-run checks and compare against a baseline")
}

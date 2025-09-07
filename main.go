package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
)

var page = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Go Simulator</title>
</head>
<body>
    <h1>Simulation Control</h1>
    <form action="/run" method="POST">
        <label>Difficulty:
            <input type="number" name="difficulty" value="1" min="1" max="4">
            (1 - Uktril, 2 - Geraktril, 3 - Reishid, 4 - Custom JSON)
        </label><br>
        <label>Arena:
            <input type="number" name="arena" value="1" min="1" max="3">
        </label><br>
        <label>Runmode:
            <input type="number" name="runmode" value="0" min="0" max="3">
        </label><br>
        <label>Runs:
            <input type="number" name="runs" value="1000">
        </label><br>
        <label>Player JSON Input:</label><br>
        <textarea name="player_json" rows="15" cols="80">
{
  "name": "Bob",
  "hp": 205,
  "ac": 43,
  "dr": 0,
  "fort": 100,
  "curaAcelerada": 0,
  "duroDeMatar": 0,
  "duroDeFerir": 0,
  "cleave": 0,
  "flankImmune": false,
  "rigidezRaivosa": true,
  "perfectMobility": false,
  "vampiricWeapon": false,
  "erosion": false,
  "attacks": [
    {
      "name": "SwordAttack1",
      "attackBonus": 21,
      "damageDice": "5d6+36",
      "critRange": 19,
      "critBonus": 2
    },
    {
      "name": "SwordAttack2",
      "attackBonus": 21,
      "damageDice": "5d6+36",
      "critRange": 19,
      "critBonus": 2
    }
  ]
}
        </textarea><br>
        <label>Custom enemy JSON Input:</label><br>
        <textarea name="enemy_json" rows="15" cols="80">
{
  "name": "Darius",
  "hp": 130,
  "ac": 27,
  "dr": 13,
  "fort": 30,
  "curaAcelerada": 5,
  "duroDeMatar": 0,
  "duroDeFerir": 0,
  "cleave": 0,
  "flankImmune": false,
  "rigidezRaivosa": true,
  "perfectMobility": false,
  "vampiricWeapon": false,
  "erosion": false,
  "isNPC": true,
  "attacks": [
    {
      "name": "Machandejante",
      "attackBonus": 20,
      "damageDice": "4d6+27",
      "critRange": 20,
      "critBonus": 3
    }
  ]
}

        </textarea><br>
        <button type="submit">Run Simulation</button>
    </form>
</body>
</html>
`))

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page.Execute(w, nil)
	})

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Save JSON input to file
		playerJsonData := r.FormValue("player_json")
		err := os.WriteFile("player.json", []byte(playerJsonData), 0644)
		if err != nil {
			http.Error(w, "Failed to write player.json", http.StatusInternalServerError)
			return
		}

		enemyJsonData := r.FormValue("enemy_json")
		err = os.WriteFile("custom_enemy.json", []byte(enemyJsonData), 0644)
		if err != nil {
			http.Error(w, "Failed to write custom_enemy.json", http.StatusInternalServerError)
			return
		}

		// Build command with args
		cmd := exec.Command("./ETTES",
			"--difficulty="+r.FormValue("difficulty"),
			"--arena="+r.FormValue("arena"),
			"--runmode="+r.FormValue("runmode"),
			"--runs="+r.FormValue("runs"),
			"--json=player.json")

		// Pipe output back to the browser
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		if err := cmd.Start(); err != nil {
			http.Error(w, "Failed to run simulator", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		io.Copy(w, stdout)
		io.Copy(w, stderr)
		cmd.Wait()
	})

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

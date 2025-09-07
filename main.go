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
    <title>ExperimentalTabletopEncounterSimulator Online Server</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background: #f4f6f9;
            margin: 0;
            padding: 0;
            color: #333;
        }
        header {
            background: #2c3e50;
            color: #ecf0f1;
            padding: 20px;
            text-align: center;
            font-size: 24px;
        }
        main {
            max-width: 900px;
            margin: 30px auto;
            background: #fff;
            padding: 20px 30px;
            border-radius: 8px;
            box-shadow: 0px 2px 6px rgba(0,0,0,0.15);
        }
        h1 {
            text-align: center;
            margin-bottom: 25px;
            color: #2c3e50;
        }
        form label {
            display: block;
            margin: 10px 0 6px;
            font-weight: bold;
        }
        input[type="number"], textarea {
            width: 100%;
            padding: 8px;
            margin-bottom: 12px;
            border: 1px solid #ccc;
            border-radius: 4px;
            font-family: monospace;
        }
        textarea {
            resize: vertical;
        }
        button {
            display: block;
            width: 100%;
            background: #27ae60;
            color: white;
            border: none;
            padding: 12px;
            font-size: 16px;
            border-radius: 6px;
            cursor: pointer;
            margin-top: 15px;
        }
        button:hover {
            background: #219150;
        }
        .hint {
            font-size: 0.9em;
            color: #555;
        }
    </style>
</head>
<body>
    <header>ExperimentalTabletopEncounterSimulator Online Server</header>
    <main>
        <h1>Simulation Control</h1>
        <form action="/run" method="POST">
            <label>Difficulty:</label>
            <input type="number" name="difficulty" value="1" min="1" max="4">
            <div class="hint">(1 - Uktril, 2 - Geraktril, 3 - Reishid, 4 - Custom JSON)</div>

            <label>Arena:</label>
            <input type="number" name="arena" value="1" min="1" max="3">

            <label>Runmode:</label>
            <input type="number" name="runmode" value="0" min="0" max="3">

            <label>Runs:</label>
            <input type="number" name="runs" value="1000">

            <label>Player JSON Input:</label>
            <textarea name="player_json" rows="15">
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
            </textarea>

            <label>Custom Enemy JSON Input:</label>
            <textarea name="enemy_json" rows="15">
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
            </textarea>

            <button type="submit">Run Simulation</button>
        </form>
    </main>
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

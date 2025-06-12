package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Game struct {
	Board  [4][4]int `json:"board"`
	Score  int       `json:"score"`
	GameOn bool      `json:"gameOn"`
}

func NewGame() *Game {
	g := &Game{
		Board:  [4][4]int{},
		Score:  0,
		GameOn: true,
	}
	g.addNewTile()
	g.addNewTile()
	return g
}

func (g *Game) addNewTile() bool {
	empty := []struct{ row, col int }{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if g.Board[i][j] == 0 {
				empty = append(empty, struct{ row, col int }{i, j})
			}
		}
	}
	
	if len(empty) == 0 {
		return false
	}
	
	pos := empty[rand.Intn(len(empty))]
	if rand.Float32() < 0.9 {
		g.Board[pos.row][pos.col] = 2
	} else {
		g.Board[pos.row][pos.col] = 4
	}
	return true
}

func (g *Game) move(dir string) bool {
	moved := false
	
	switch dir {
	case "left":
		for i := 0; i < 4; i++ {
			row := g.Board[i][:]
			newRow, changed := slideAndMerge(row)
			if changed {
				moved = true
				copy(g.Board[i][:], newRow)
			}
		}
	case "right":
		for i := 0; i < 4; i++ {
			row := reverseArray(g.Board[i][:])
			newRow, changed := slideAndMerge(row)
			if changed {
				moved = true
				copy(g.Board[i][:], reverseArray(newRow))
			}
		}
	case "up":
		for j := 0; j < 4; j++ {
			col := []int{g.Board[0][j], g.Board[1][j], g.Board[2][j], g.Board[3][j]}
			newCol, changed := slideAndMerge(col)
			if changed {
				moved = true
				for i := 0; i < 4; i++ {
					g.Board[i][j] = newCol[i]
				}
			}
		}
	case "down":
		for j := 0; j < 4; j++ {
			col := []int{g.Board[3][j], g.Board[2][j], g.Board[1][j], g.Board[0][j]}
			newCol, changed := slideAndMerge(col)
			if changed {
				moved = true
				for i := 0; i < 4; i++ {
					g.Board[3-i][j] = newCol[i]
				}
			}
		}
	}
	
	if moved {
		g.addNewTile()
		g.checkGameOver()
	}
	
	return moved
}

func slideAndMerge(row []int) ([]int, bool) {
	result := make([]int, 4)
	idx := 0
	changed := false
	
	for i := 0; i < 4; i++ {
		if row[i] != 0 {
			if idx > 0 && result[idx-1] == row[i] {
				result[idx-1] *= 2
				changed = true
			} else {
				result[idx] = row[i]
				if idx != i {
					changed = true
				}
				idx++
			}
		}
	}
	
	return result, changed
}

func reverseArray(arr []int) []int {
	result := make([]int, len(arr))
	for i := 0; i < len(arr); i++ {
		result[i] = arr[len(arr)-1-i]
	}
	return result
}

func (g *Game) checkGameOver() {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if g.Board[i][j] == 0 {
				return
			}
			if j < 3 && g.Board[i][j] == g.Board[i][j+1] {
				return
			}
			if i < 3 && g.Board[i][j] == g.Board[i+1][j] {
				return
			}
		}
	}
	g.GameOn = false
}

func (g *Game) calculateScore() {
	g.Score = 0
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			g.Score += g.Board[i][j]
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/new", handleNewGame)
	http.HandleFunc("/api/move", handleMove)
	http.HandleFunc("/api/save", handleSave)
	http.HandleFunc("/api/load", handleLoad)
	
	fmt.Println("2048 Game server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleNewGame(w http.ResponseWriter, r *http.Request) {
	game := NewGame()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func handleMove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Direction string `json:"direction"`
		Game      *Game  `json:"game"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	req.Game.move(req.Direction)
	req.Game.calculateScore()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req.Game)
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	data, _ := json.Marshal(game)
	http.SetCookie(w, &http.Cookie{
		Name:     "game_state",
		Value:    string(data),
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: false,
	})
	
	w.WriteHeader(http.StatusOK)
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("game_state")
	if err != nil {
		game := NewGame()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
		return
	}
	
	var game Game
	if err := json.Unmarshal([]byte(cookie.Value), &game); err != nil {
		game = *NewGame()
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}
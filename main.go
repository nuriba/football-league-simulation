package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Interfaces
type TeamInterface interface {
	GetName() string
	GetStrength() int
	GetStats() TeamStats
	UpdateStats(goalsFor, goalsAgainst int, result string)
}

type MatchInterface interface {
	GetHomeTeam() string
	GetAwayTeam() string
	GetResult() string
	IsPlayed() bool
	PlayMatch(homeTeam, awayTeam *Team)
}

type LeagueInterface interface {
	GetTeams() []TeamInterface
	GetMatches() []MatchInterface
	GetTable() []TeamInterface
	PlayNextWeek() error
	PlayAllMatches()
	PredictChampionship() map[string]float64
}

// Structs
type TeamStats struct {
	Played       int `json:"played"`
	Won          int `json:"won"`
	Drawn        int `json:"drawn"`
	Lost         int `json:"lost"`
	GoalsFor     int `json:"goals_for"`
	GoalsAgainst int `json:"goals_against"`
	GoalDiff     int `json:"goal_difference"`
	Points       int `json:"points"`
}

type Team struct {
	Name         string    `json:"name"`
	Strength     int       `json:"strength"`
	BaseStrength int       `json:"base_strength"`
	Stats        TeamStats `json:"stats"`
}

func (t *Team) GetName() string {
	return t.Name
}

func (t *Team) GetStrength() int {
	return t.Strength
}

func (t *Team) GetStats() TeamStats {
	return t.Stats
}

func (t *Team) UpdateStats(goalsFor, goalsAgainst int, result string) {
	t.Stats.Played++
	t.Stats.GoalsFor += goalsFor
	t.Stats.GoalsAgainst += goalsAgainst
	t.Stats.GoalDiff = t.Stats.GoalsFor - t.Stats.GoalsAgainst

	switch result {
	case "W":
		t.Stats.Won++
		t.Stats.Points += 3
	case "D":
		t.Stats.Drawn++
		t.Stats.Points += 1
	case "L":
		t.Stats.Lost++
	}
}

type Match struct {
	Week      int    `json:"week"`
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeGoals int    `json:"home_goals"`
	AwayGoals int    `json:"away_goals"`
	Played    bool   `json:"played"`
	Result    string `json:"result"`
}

func (m *Match) GetHomeTeam() string {
	return m.HomeTeam
}

func (m *Match) GetAwayTeam() string {
	return m.AwayTeam
}

func (m *Match) GetResult() string {
	return m.Result
}

func (m *Match) IsPlayed() bool {
	return m.Played
}

func (m *Match) PlayMatch(homeTeam, awayTeam *Team) {
	if m.Played {
		return //Match already played
	}

	// Simulate the match
	homeGoals, awayGoals := simulateMatchResult(homeTeam, awayTeam)

	// Update match data
	m.HomeGoals = homeGoals
	m.AwayGoals = awayGoals
	m.Played = true
	m.Result = fmt.Sprintf("%s %d-%d %s", m.HomeTeam, homeGoals, awayGoals, m.AwayTeam)

	// Update team statistics
	if homeGoals > awayGoals {
		homeTeam.UpdateStats(homeGoals, awayGoals, "W")
		awayTeam.UpdateStats(awayGoals, homeGoals, "L")
	} else if homeGoals < awayGoals {
		homeTeam.UpdateStats(homeGoals, awayGoals, "L")
		awayTeam.UpdateStats(awayGoals, homeGoals, "W")
	} else {
		homeTeam.UpdateStats(homeGoals, awayGoals, "D")
		awayTeam.UpdateStats(awayGoals, homeGoals, "D")
	}
}

type League struct {
	Teams          []*Team  `json:"teams"`
	Matches        []*Match `json:"matches"`
	CurrentWeek    int      `json:"current_week"`
	TotalWeeks     int      `json:"total_weeks"`
	MatchesPerWeek int      `json:"matches_per_week"`
}

func NewLeague() *League {
	league := &League{
		Teams:          make([]*Team, 0),
		Matches:        make([]*Match, 0),
		CurrentWeek:    0,
		TotalWeeks:     6,
		MatchesPerWeek: 2,
	}

	// Initialize teams with different strengths based on the Premier League 2024-2025 season
	teams := []*Team{
		{Name: "Chelsea", Strength: 80, BaseStrength: 80, Stats: TeamStats{}},
		{Name: "Arsenal", Strength: 85, BaseStrength: 85, Stats: TeamStats{}},
		{Name: "Manchester City", Strength: 82, BaseStrength: 82, Stats: TeamStats{}},
		{Name: "Liverpool", Strength: 90, BaseStrength: 90, Stats: TeamStats{}},
	}

	league.Teams = teams
	league.generateFixtures()
	return league
}

func (l *League) generateFixtures() {
	// Define matches ensuring proper home/away alternation and no team playing twice in a row
	fixtures := []struct {
		week      int
		homeTeam1 string
		awayTeam1 string
		homeTeam2 string
		awayTeam2 string
	}{
		// Initialize fixtures for the first 6 weeks by hand
		{1, "Chelsea", "Arsenal", "Manchester City", "Liverpool"},
		{2, "Arsenal", "Manchester City", "Liverpool", "Chelsea"},
		{3, "Chelsea", "Liverpool", "Manchester City", "Arsenal"},
		{4, "Arsenal", "Chelsea", "Liverpool", "Manchester City"},
		{5, "Manchester City", "Chelsea", "Liverpool", "Arsenal"},
		{6, "Arsenal", "Liverpool", "Chelsea", "Manchester City"},
	}
	// Generate all matches based on the defined fixtures
	for _, fixture := range fixtures {
		// Match 1
		match1 := &Match{
			Week:     fixture.week,
			HomeTeam: fixture.homeTeam1,
			AwayTeam: fixture.awayTeam1,
			Played:   false,
		}
		l.Matches = append(l.Matches, match1)

		// Match 2
		match2 := &Match{
			Week:     fixture.week,
			HomeTeam: fixture.homeTeam2,
			AwayTeam: fixture.awayTeam2,
			Played:   false,
		}
		l.Matches = append(l.Matches, match2)
	}
}

func (l *League) GetTeams() []TeamInterface {
	teams := make([]TeamInterface, len(l.Teams))
	for i, team := range l.Teams {
		teams[i] = team
	}
	return teams
}

func (l *League) GetMatches() []MatchInterface {
	matches := make([]MatchInterface, len(l.Matches))
	for i, match := range l.Matches {
		matches[i] = match
	}
	return matches
}

func (l *League) GetTable() []TeamInterface {
	// Sort teams by points, then goal difference, then goals for
	sort.Slice(l.Teams, func(i, j int) bool {
		if l.Teams[i].Stats.Points != l.Teams[j].Stats.Points {
			return l.Teams[i].Stats.Points > l.Teams[j].Stats.Points
		}
		if l.Teams[i].Stats.GoalDiff != l.Teams[j].Stats.GoalDiff {
			return l.Teams[i].Stats.GoalDiff > l.Teams[j].Stats.GoalDiff
		}
		return l.Teams[i].Stats.GoalsFor > l.Teams[j].Stats.GoalsFor
	})

	teams := make([]TeamInterface, len(l.Teams))
	for i, team := range l.Teams {
		teams[i] = team
	}
	return teams
}

func (l *League) getTeamByName(name string) *Team {
	for _, team := range l.Teams {
		if team.Name == name {
			return team
		}
	}
	return nil
}

// Match simulation function
func simulateMatchResult(homeTeam, awayTeam *Team) (int, int) {
	rand.Seed(time.Now().UnixNano())

	// Calculate probability based on team strengths
	homeAdvantage := 5 // Give Home team a slight advantage
	homeStrength := homeTeam.Strength + homeAdvantage
	totalStrength := homeStrength + awayTeam.Strength
	homeProbability := float64(homeStrength) / float64(totalStrength)
	awayProbability := float64(awayTeam.Strength) / float64(totalStrength)

	// Generate goals based on strength (0-5 goals typical match score range to make it more realistic)
	homeGoals := 0
	awayGoals := 0

	for i := 0; i < 6; i++ {
		if rand.Float64() < homeProbability*0.4 {
			homeGoals++
		}
		if rand.Float64() < awayProbability*0.4 {
			awayGoals++
		}
	}

	return homeGoals, awayGoals
}

// UpdateTeamStrength adjusts team strength based on performance
func (l *League) UpdateTeamStrength(team *Team) {
	if team.Stats.Played == 0 {
		return
	}
	pointsPerGame := float64(team.Stats.Points) / float64(team.Stats.Played)

	currentPosition := l.getTeamPosition(team.Name)

	// Base strength adjustment calculation
	strengthChange := 0

	// Performance-based adjustment
	if pointsPerGame >= 2.5 {
		strengthChange += 2
	} else if pointsPerGame >= 2.0 {
		strengthChange += 1
	} else if pointsPerGame <= 1.0 {
		strengthChange -= 2
	} else if pointsPerGame <= 1.5 {
		strengthChange -= 1
	}

	// League position adjustment
	switch currentPosition {
	case 1:
		strengthChange += 1
	case 4:
		strengthChange -= 1
	}

	// Apply strength change with limits
	newStrength := team.Strength + strengthChange

	// Keep strength within reasonable bounds
	minStrength := team.BaseStrength - 15
	maxStrength := team.BaseStrength + 15

	if newStrength < minStrength {
		newStrength = minStrength
	} else if newStrength > maxStrength {
		newStrength = maxStrength
	}

	// Ensure strength stays within 1-100 range
	if newStrength < 1 {
		newStrength = 1
	} else if newStrength > 100 {
		newStrength = 100
	}

	team.Strength = newStrength
}

func (l *League) getTeamPosition(teamName string) int {
	table := l.GetTable()
	for i, teamInterface := range table {
		if teamInterface.GetName() == teamName {
			return i + 1
		}
	}
	return 4
}

// PredictChampionship calculates championship probability for each team
func (l *League) PredictChampionship() map[string]float64 {
	predictions := make(map[string]float64)

	// If no matches played yet, equal chances
	if l.CurrentWeek == 0 {
		for _, team := range l.Teams {
			predictions[team.Name] = 25.0 // Equal 25% for each team
		}
		return predictions
	}

	// Calculate base score for each team
	teamScores := make(map[string]float64)
	totalScore := 0.0

	for _, team := range l.Teams {
		score := 0.0

		// Current points
		score += float64(team.Stats.Points) * 10.0

		// Goal difference
		score += float64(team.Stats.GoalDiff) * 2.0

		// Current form
		if team.Stats.Played > 0 {
			pointsPerGame := float64(team.Stats.Points) / float64(team.Stats.Played)
			score += pointsPerGame * 15.0
		}

		// Current strength
		score += float64(team.Strength) * 1.0

		// League position bonus
		position := l.getTeamPosition(team.Name)
		switch position {
		case 1:
			score += 20.0
		case 2:
			score += 10.0
		case 3:
			score += 0.0
		case 4:
			score -= 10.0
		}

		// Remaining matches potential
		remainingMatches := (l.TotalWeeks - l.CurrentWeek) * 2
		if remainingMatches > 0 {
			maxPossiblePoints := remainingMatches * 3
			score += float64(maxPossiblePoints) * 0.3
		}

		teamScores[team.Name] = score
		totalScore += score
	}

	// Convert scores to percentages
	for teamName, score := range teamScores {
		if totalScore > 0 {
			predictions[teamName] = (score / totalScore) * 100.0
		} else {
			predictions[teamName] = 25.0 // Equal chances
		}

		// Ensure reasonable bounds (minimum 1%, maximum 95%)
		if predictions[teamName] < 1.0 {
			predictions[teamName] = 1.0
		} else if predictions[teamName] > 95.0 {
			predictions[teamName] = 95.0
		}
	}

	// Normalize to ensure total is 100%
	total := 0.0
	for _, percentage := range predictions {
		total += percentage
	}

	if total > 0 {
		for teamName := range predictions {
			predictions[teamName] = (predictions[teamName] / total) * 100.0
		}
	}

	return predictions
}

func (l *League) PlayNextWeek() error {
	if l.CurrentWeek >= l.TotalWeeks {
		return fmt.Errorf("season is complete")
	}

	l.CurrentWeek++

	// Play all matches for current week
	for _, match := range l.Matches {
		if match.Week == l.CurrentWeek && !match.Played {
			homeTeam := l.getTeamByName(match.HomeTeam)
			awayTeam := l.getTeamByName(match.AwayTeam)
			match.PlayMatch(homeTeam, awayTeam)
		}
	}

	// Update team strengths after all matches in the week are played
	for _, team := range l.Teams {
		l.UpdateTeamStrength(team)
	}

	return nil
}

func (l *League) PlayAllMatches() {
	for l.CurrentWeek < l.TotalWeeks {
		l.PlayNextWeek()
	}
}

// Global league instance
var league *League

func getLeagueTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	table := league.GetTable()

	response := map[string]interface{}{
		"current_week":             league.CurrentWeek,
		"table":                    table,
		"championship_predictions": league.PredictChampionship(),
	}

	json.NewEncoder(w).Encode(response)
}

func getMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matches":      league.Matches,
		"current_week": league.CurrentWeek,
	})
}

func getChampionshipPredictions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	predictions := league.PredictChampionship()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_week": league.CurrentWeek,
		"predictions":  predictions,
	})
}

func playNextWeek(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := league.PlayNextWeek()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message":                  fmt.Sprintf("Week %d completed", league.CurrentWeek),
		"current_week":             league.CurrentWeek,
		"table":                    league.GetTable(),
		"championship_predictions": league.PredictChampionship(),
	}

	json.NewEncoder(w).Encode(response)
}

func playAllMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	league.PlayAllMatches()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "All matches completed",
		"final_table": league.GetTable(),
		"all_matches": league.Matches,
	})
}

func editMatchResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	weekStr := vars["week"]
	week, err := strconv.Atoi(weekStr)
	if err != nil {
		http.Error(w, "Invalid week number", http.StatusBadRequest)
		return
	}

	var editRequest struct {
		HomeTeam  string `json:"home_team"`
		AwayTeam  string `json:"away_team"`
		HomeGoals int    `json:"home_goals"`
		AwayGoals int    `json:"away_goals"`
	}

	if err := json.NewDecoder(r.Body).Decode(&editRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find and update the match
	for _, match := range league.Matches {
		if match.Week == week && match.HomeTeam == editRequest.HomeTeam && match.AwayTeam == editRequest.AwayTeam {
			// Reset team stats if match was already played
			if match.Played {
				homeTeam := league.getTeamByName(match.HomeTeam)
				awayTeam := league.getTeamByName(match.AwayTeam)

				// Reverse previous result
				homeTeam.reverseMatchResult(match.HomeGoals, match.AwayGoals)
				awayTeam.reverseMatchResult(match.AwayGoals, match.HomeGoals)
			}

			// Apply new result
			match.HomeGoals = editRequest.HomeGoals
			match.AwayGoals = editRequest.AwayGoals
			match.Played = true

			homeTeam := league.getTeamByName(match.HomeTeam)
			awayTeam := league.getTeamByName(match.AwayTeam)

			if editRequest.HomeGoals > editRequest.AwayGoals {
				match.Result = fmt.Sprintf("%s %d-%d %s", match.HomeTeam, editRequest.HomeGoals, editRequest.AwayGoals, match.AwayTeam)
				homeTeam.UpdateStats(editRequest.HomeGoals, editRequest.AwayGoals, "W")
				awayTeam.UpdateStats(editRequest.AwayGoals, editRequest.HomeGoals, "L")
			} else if editRequest.HomeGoals < editRequest.AwayGoals {
				match.Result = fmt.Sprintf("%s %d-%d %s", match.HomeTeam, editRequest.HomeGoals, editRequest.AwayGoals, match.AwayTeam)
				homeTeam.UpdateStats(editRequest.HomeGoals, editRequest.AwayGoals, "L")
				awayTeam.UpdateStats(editRequest.AwayGoals, editRequest.HomeGoals, "W")
			} else {
				match.Result = fmt.Sprintf("%s %d-%d %s", match.HomeTeam, editRequest.HomeGoals, editRequest.AwayGoals, match.AwayTeam)
				homeTeam.UpdateStats(editRequest.HomeGoals, editRequest.AwayGoals, "D")
				awayTeam.UpdateStats(editRequest.AwayGoals, editRequest.HomeGoals, "D")
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":       "Match result updated",
				"match":         match,
				"updated_table": league.GetTable(),
			})
			return
		}
	}

	http.Error(w, "Match not found", http.StatusNotFound)
}

func (t *Team) reverseMatchResult(goalsFor, goalsAgainst int) {
	t.Stats.Played--
	t.Stats.GoalsFor -= goalsFor
	t.Stats.GoalsAgainst -= goalsAgainst
	t.Stats.GoalDiff = t.Stats.GoalsFor - t.Stats.GoalsAgainst

	if goalsFor > goalsAgainst {
		t.Stats.Won--
		t.Stats.Points -= 3
	} else if goalsFor < goalsAgainst {
		t.Stats.Lost--
	} else {
		t.Stats.Drawn--
		t.Stats.Points -= 1
	}
}

func resetLeague(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	league = NewLeague()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "League reset successfully",
		"teams":   league.Teams,
	})
}

func main() {
	// Initialize league
	league = NewLeague()

	// Setup routes
	r := mux.NewRouter()

	// API endpoints
	r.HandleFunc("/api/table", getLeagueTable).Methods("GET")
	r.HandleFunc("/api/matches", getMatches).Methods("GET")
	r.HandleFunc("/api/predictions", getChampionshipPredictions).Methods("GET")
	r.HandleFunc("/api/play-next-week", playNextWeek).Methods("POST")
	r.HandleFunc("/api/play-all", playAllMatches).Methods("POST")
	r.HandleFunc("/api/edit-match/{week}", editMatchResult).Methods("PUT")
	r.HandleFunc("/api/reset", resetLeague).Methods("POST")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}

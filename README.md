## Features
- **Premier League rules**: 4 teams, double round-robin, 3 pts win, 1 pt draw
- **Dynamic team strengths**: Teams get stronger/weaker based on results and table position
- **Live championship predictions**: Probabilities updated after every week
- **REST API**: Simulate, edit, and analyze the league from any client
- **SQL schema & queries**: For persistence, reporting, and analytics
- **Easy reset & replay**: Start over, edit results, or play all matches instantly


---

## Quick Start

### Prerequisites
- Go 1.21+
- SQLite3 or PostgreSQL 

### 1. Install dependencies
```bash
go mod tidy
```

### 2. Initialize the database
- **SQLite:**
  ```bash
  sqlite3 league.db < schema.sql
  ```
- **PostgreSQL:**
  ```bash
  psql -U <user> -d <dbname> -f schema.sql
  ```

### 3. Run the server
```bash
go run main.go
```
Server runs at `http://localhost:8000`

---

##  API Endpoints
| Method | Endpoint                | Description                        |
|--------|-------------------------|------------------------------------|
| GET    | `/api/table`            | Current league table & predictions |
| GET    | `/api/matches`          | All matches & results              |
| GET    | `/api/predictions`      | Championship predictions           |
| POST   | `/api/play-next-week`   | Play next week's matches           |
| POST   | `/api/play-all`         | Play all remaining matches         |
| PUT    | `/api/edit-match/{week}`| Edit a match result                |
| POST   | `/api/reset`            | Reset league to initial state      |
| GET    | `/health`               | Health check                       |

---

## Usage Examples

**Get League Table & Predictions**
```bash
curl http://localhost:8080/api/table
```

**Play Next Week**
```bash
curl -X POST http://localhost:8080/api/play-next-week
```

**Get All Matches**
```bash
curl http://localhost:8080/api/matches
```

**Get Predictions**
```bash
curl http://localhost:8080/api/predictions
```

**Edit a Match Result**
```bash
curl -X PUT http://localhost:8080/api/edit-match/1 \
  -H "Content-Type: application/json" \
  -d '{"home_team": "Chelsea", "away_team": "Arsenal", "home_goals": 3, "away_goals": 1}'
```

**Play All Matches**
```bash
curl -X POST http://localhost:8080/api/play-all
```

**Reset League**
```bash
curl -X POST http://localhost:8080/api/reset
```

---

## League & Simulation Logic
- **Teams:** Chelsea, Arsenal, Manchester City, Liverpool
- **Strengths:** Dynamic, change after each week based on form & table
- **Home advantage:** +5 strength for home team
- **Match simulation:** Based on relative strengths, with randomness
- **Predictions:** Calculated from points, form, goal difference, and strength
- **Table sorting:** Points > Goal difference > Goals for

---

##  SQL Database
- **schema.sql:**
  - Tables: leagues, teams, team_stats, matches, championship_predictions
  - Views: league_table, matches_view, current_predictions
  - Triggers: auto-update timestamps
  - Inserts: initial teams, fixtures, stats
- **queries.sql:**
  - League table, match results, predictions, stats, reporting

**To initialize:**
```bash
sqlite3 league.db < schema.sql
```

**To use queries:**
```bash
sqlite3 league.db < queries.sql
```

---

## Docker (Optional)
```bash
 `docker build -t football-league .`
 `docker run -p 8080:8080 football-league`
```

---

## Sample Workflow
1. Start league: server auto-initializes 4 teams
2. Check state: `GET /api/table`
3. Play week: `POST /api/play-next-week`
4. View results: `GET /api/matches`
5. Get predictions: `GET /api/predictions`
6. Edit result: `PUT /api/edit-match/{week}`
7. Repeat for weeks
8. Complete season: `POST /api/play-all`

---

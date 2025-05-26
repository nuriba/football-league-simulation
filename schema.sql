-- Drop tables if they exist
DROP TABLE IF EXISTS championship_predictions;
DROP TABLE IF EXISTS match_results;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS team_stats;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS leagues;

-- Create leagues table
CREATE TABLE leagues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    current_week INTEGER DEFAULT 0,
    total_weeks INTEGER DEFAULT 6,
    matches_per_week INTEGER DEFAULT 2,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create teams table
CREATE TABLE teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL UNIQUE,
    strength INTEGER NOT NULL CHECK (strength >= 1 AND strength <= 100),
    base_strength INTEGER NOT NULL CHECK (base_strength >= 1 AND base_strength <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE
);

-- Create team_stats table (current season statistics)
CREATE TABLE team_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    played INTEGER DEFAULT 0,
    won INTEGER DEFAULT 0,
    drawn INTEGER DEFAULT 0,
    lost INTEGER DEFAULT 0,
    goals_for INTEGER DEFAULT 0,
    goals_against INTEGER DEFAULT 0,
    goal_difference INTEGER GENERATED ALWAYS AS (goals_for - goals_against) STORED,
    points INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

-- Create matches table
CREATE TABLE matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id INTEGER NOT NULL,
    week INTEGER NOT NULL,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    home_goals INTEGER DEFAULT NULL,
    away_goals INTEGER DEFAULT NULL,
    played BOOLEAN DEFAULT FALSE,
    result VARCHAR(100) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE,
    FOREIGN KEY (home_team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (away_team_id) REFERENCES teams(id) ON DELETE CASCADE,
    UNIQUE(league_id, week, home_team_id, away_team_id)
);

-- Create championship_predictions table
CREATE TABLE championship_predictions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    week INTEGER NOT NULL,
    prediction_percentage DECIMAL(5,2) NOT NULL CHECK (prediction_percentage >= 0 AND prediction_percentage <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    UNIQUE(league_id, team_id, week)
);

-- Create indexes for better performance
CREATE INDEX idx_teams_league_id ON teams(league_id);
CREATE INDEX idx_team_stats_team_id ON team_stats(team_id);
CREATE INDEX idx_matches_league_week ON matches(league_id, week);
CREATE INDEX idx_matches_home_team ON matches(home_team_id);
CREATE INDEX idx_matches_away_team ON matches(away_team_id);
CREATE INDEX idx_predictions_league_week ON championship_predictions(league_id, week);
CREATE INDEX idx_predictions_team ON championship_predictions(team_id);

-- Insert initial data
INSERT INTO leagues (name, current_week, total_weeks, matches_per_week) 
VALUES ('Premier League 2024-2025', 0, 6, 2);

-- Insert teams with their strengths
INSERT INTO teams (league_id, name, strength, base_strength) VALUES
(1, 'Chelsea', 80, 80),
(1, 'Arsenal', 85, 85),
(1, 'Manchester City', 82, 82),
(1, 'Liverpool', 90, 90);

-- Initialize team stats
INSERT INTO team_stats (team_id, played, won, drawn, lost, goals_for, goals_against, points)
SELECT id, 0, 0, 0, 0, 0, 0, 0 FROM teams WHERE league_id = 1;

-- Insert all matches for the 6-week season
INSERT INTO matches (league_id, week, home_team_id, away_team_id) VALUES
-- Week 1
(1, 1, (SELECT id FROM teams WHERE name = 'Chelsea'), (SELECT id FROM teams WHERE name = 'Arsenal')),
(1, 1, (SELECT id FROM teams WHERE name = 'Manchester City'), (SELECT id FROM teams WHERE name = 'Liverpool')),
-- Week 2
(1, 2, (SELECT id FROM teams WHERE name = 'Arsenal'), (SELECT id FROM teams WHERE name = 'Manchester City')),
(1, 2, (SELECT id FROM teams WHERE name = 'Liverpool'), (SELECT id FROM teams WHERE name = 'Chelsea')),
-- Week 3
(1, 3, (SELECT id FROM teams WHERE name = 'Chelsea'), (SELECT id FROM teams WHERE name = 'Liverpool')),
(1, 3, (SELECT id FROM teams WHERE name = 'Manchester City'), (SELECT id FROM teams WHERE name = 'Arsenal')),
-- Week 4
(1, 4, (SELECT id FROM teams WHERE name = 'Arsenal'), (SELECT id FROM teams WHERE name = 'Chelsea')),
(1, 4, (SELECT id FROM teams WHERE name = 'Liverpool'), (SELECT id FROM teams WHERE name = 'Manchester City')),
-- Week 5
(1, 5, (SELECT id FROM teams WHERE name = 'Manchester City'), (SELECT id FROM teams WHERE name = 'Chelsea')),
(1, 5, (SELECT id FROM teams WHERE name = 'Liverpool'), (SELECT id FROM teams WHERE name = 'Arsenal')),
-- Week 6
(1, 6, (SELECT id FROM teams WHERE name = 'Arsenal'), (SELECT id FROM teams WHERE name = 'Liverpool')),
(1, 6, (SELECT id FROM teams WHERE name = 'Chelsea'), (SELECT id FROM teams WHERE name = 'Manchester City'));

-- League table view
CREATE VIEW league_table AS
SELECT  t.id, t.name, t.strength, ts.played, ts.won, ts.drawn, ts.lost, ts.goals_for, ts.goals_against, ts.goal_difference, ts.points, RANK() OVER (ORDER BY ts.points DESC, ts.goal_difference DESC, ts.goals_for DESC) as position
FROM teams t
JOIN team_stats ts ON t.id = ts.team_id
WHERE t.league_id = 1
ORDER BY ts.points DESC, ts.goal_difference DESC, ts.goals_for DESC;

-- Matches with team names view
CREATE VIEW matches_view AS
SELECT  m.id, m.week, ht.name as home_team, at.name as away_team, m.home_goals, m.away_goals, m.played, m.result, m.created_at
FROM matches m
JOIN teams ht ON m.home_team_id = ht.id
JOIN teams at ON m.away_team_id = at.id
WHERE m.league_id = 1
ORDER BY m.week, m.id;

-- Championship predictions view
CREATE VIEW current_predictions AS
SELECT  t.name as team_name, cp.prediction_percentage, cp.week
FROM championship_predictions cp
JOIN teams t ON cp.team_id = t.id
WHERE cp.league_id = 1
ORDER BY cp.week DESC, cp.prediction_percentage DESC;

-- Update team stats timestamp on changes
CREATE TRIGGER update_team_stats_timestamp AFTER UPDATE ON team_stats
BEGIN
    UPDATE team_stats SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Update match timestamp on changes
CREATE TRIGGER update_match_timestamp AFTER UPDATE ON matches
BEGIN
    UPDATE matches SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Update league timestamp when current week changes
CREATE TRIGGER update_league_timestamp AFTER UPDATE ON leagues
BEGIN
    UPDATE leagues SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
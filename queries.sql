-- 1. Get current league table (ordered by position)
SELECT name, played,won,drawn,lost,goals_for,goals_against,goal_difference,points,position
FROM league_table;

-- 2. Get current week and league status
SELECT name as league_name,current_week,total_weeks,matches_per_week,
    CASE WHEN current_week >= total_weeks THEN 'Season Complete' ELSE 'Season Active' END as status
FROM leagues WHERE id = 1;

-- 3. Get all matches with results
SELECT * FROM matches_view ORDER BY week, id;

-- 4. Get matches for a specific week
SELECT week,home_team,away_team,home_goals,away_goals,played,result
FROM matches_view 
WHERE week = 1 -- Replace 1 with week number you want to get matches for
ORDER BY id;

-- 5. Get upcoming/unplayed matches
SELECT week,home_team,away_team
FROM matches_view 
WHERE played = FALSE
ORDER BY week, id;

-- 6. Update match result
UPDATE matches 
SET home_goals = 1,away_goals = 2,played = TRUE,result = '1-2',updated_at = CURRENT_TIMESTAMP
WHERE league_id = 1 AND week = 1 AND home_team_id = 1 AND away_team_id = 2; --Adjust the variables as needed

-- 7. Update team statistics after match
-- For home team 
UPDATE team_stats 
SET  played = played + 1, won = won + 1, goals_for = goals_for + 1, goals_against = goals_against + 2, points = points + 3, updated_at = CURRENT_TIMESTAMP
WHERE team_id = 1; --Adjust the variables as needed

-- For away team 
UPDATE team_stats 
SET  played = played + 1, lost = lost + 1, goals_for = goals_for + 2, goals_against = goals_against + 1, updated_at = CURRENT_TIMESTAMP
WHERE team_id = 2; --Adjust the variables as needed

-- For draw result, use drawn = drawn + 1 and points = points + 1 for both teams

-- 8. Insert championship predictions for current week
INSERT OR REPLACE INTO championship_predictions (league_id, team_id, week, prediction_percentage)
VALUES 
(1, (SELECT id FROM teams WHERE name = 'Chelsea'), 1, 10),
(1, (SELECT id FROM teams WHERE name = 'Arsenal'), 1, 30),
(1, (SELECT id FROM teams WHERE name = 'Manchester City'), 1, 20),
(1, (SELECT id FROM teams WHERE name = 'Liverpool'), 1, 50);

-- 9. Get latest championship predictions
SELECT team_name, prediction_percentage 
FROM current_predictions 
WHERE week = (SELECT MAX(week) FROM championship_predictions WHERE league_id = 1)
ORDER BY prediction_percentage DESC;

-- 10. Get prediction history for a specific team
SELECT  t.name, cp.week, cp.prediction_percentage, cp.created_at
FROM championship_predictions cp
JOIN teams t ON cp.team_id = t.id
WHERE t.name = ? AND cp.league_id = 1
ORDER BY cp.week;

-- 11. Get team performance metrics
SELECT  t.name, t.strength, ts.played, ROUND(CAST(ts.won AS REAL) / NULLIF(ts.played, 0) * 100, 2) as win_percentage, ROUND(CAST(ts.points AS REAL) / NULLIF(ts.played, 0), 2) as points_per_game, ROUND(CAST(ts.goals_for AS REAL) / NULLIF(ts.played, 0), 2) as goals_per_game, ROUND(CAST(ts.goals_against AS REAL) / NULLIF(ts.played, 0), 2) as goals_conceded_per_game
FROM teams t
JOIN team_stats ts ON t.id = ts.team_id
WHERE t.league_id = 1
ORDER BY ts.points DESC;

-- 12. Head-to-head record between two teams
SELECT  m.week, m.home_team, m.away_team, m.home_goals, m.away_goals, m.result,
    CASE 
        WHEN m.home_team = ? AND m.home_goals > m.away_goals THEN 'Win'
        WHEN m.away_team = ? AND m.away_goals > m.home_goals THEN 'Win'
        WHEN m.home_goals = m.away_goals THEN 'Draw'
        ELSE 'Loss'
    END as result_for_team
FROM matches_view m
WHERE (m.home_team = 'Arsenal' OR m.away_team = 'Arsenal') AND (m.home_team = 'Manchester City' OR m.away_team = 'Manchester City') AND m.played = TRUE
ORDER BY m.week;

-- 13. Get top scorers 
SELECT  name, goals_for, played, ROUND(CAST(goals_for AS REAL) / NULLIF(played, 0), 2) as avg_goals_per_game
FROM league_table
ORDER BY goals_for DESC;

-- 14. Get best defense
SELECT 
    name,
    goals_against,
    played,
    ROUND(CAST(goals_against AS REAL) / NULLIF(played, 0), 2) as avg_goals_conceded_per_game
FROM league_table
ORDER BY goals_against ASC;

-- ==========================================
-- LEAGUE MANAGEMENT
-- ==========================================

-- 15. Advance to next week
UPDATE leagues 
SET current_week = current_week + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = 1 AND current_week < total_weeks;

-- 16. Reset league to beginning
UPDATE leagues SET current_week = 0, updated_at = CURRENT_TIMESTAMP WHERE id = 1;

UPDATE team_stats SET   played = 0, won = 0, drawn = 0, lost = 0,  goals_for = 0, goals_against = 0, points = 0,  updated_at = CURRENT_TIMESTAMP;

UPDATE matches SET  home_goals = NULL, away_goals = NULL,  played = FALSE, result = NULL, updated_at = CURRENT_TIMESTAMP
WHERE league_id = 1;

DELETE FROM championship_predictions WHERE league_id = 1;

-- 17. Update team strength
UPDATE teams 
SET strength = 10, updated_at = CURRENT_TIMESTAMP
WHERE name = 'Arsenal' AND league_id = 1;

-- 18. Season summary report
SELECT 'League Table' as report_section, '' as team_name, '' as details
UNION ALL
SELECT 'Position ' || position, name, points || ' pts, ' || won || 'W-' || drawn || 'D-' || lost || 'L, GD:' || goal_difference
FROM league_table
UNION ALL
SELECT '', '', ''
UNION ALL
SELECT 'Championship Predictions', team_name, ROUND(prediction_percentage, 1) || '%'
FROM current_predictions 
WHERE week = (SELECT MAX(week) FROM championship_predictions WHERE league_id = 1)
ORDER BY report_section, team_name;

-- 19. Match results summary by week
SELECT 'Week ' || week as week_header,home_team || ' vs ' || away_team as fixture,
    CASE 
        WHEN played = TRUE THEN home_goals || '-' || away_goals
        ELSE 'Not Played'
    END as score
FROM matches_view
ORDER BY week, id;

-- 20. Find closest title race 
SELECT  t1.name as team1, t1.points as points1, t2.name as team2, t2.points as points2, ABS(t1.points - t2.points) as points_difference
FROM league_table t1
CROSS JOIN league_table t2
WHERE t1.name < t2.name
ORDER BY points_difference, t1.points DESC
LIMIT 5; 
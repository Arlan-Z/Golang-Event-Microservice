CREATE TABLE events (
    id TEXT PRIMARY KEY,
    event_name TEXT NOT NULL,
    home_team TEXT NOT NULL,
    away_team TEXT NOT NULL,
    home_win_chance REAL NOT NULL,
    away_win_chance REAL NOT NULL,
    draw_chance REAL NOT NULL,
    event_start_date DATETIME NOT NULL,
    event_end_date DATETIME NOT NULL,   
    event_result TEXT, 
    type TEXT NOT NULL,
    is_active INTEGER DEFAULT 1 
);
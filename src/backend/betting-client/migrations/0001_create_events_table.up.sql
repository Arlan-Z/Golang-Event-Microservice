CREATE TABLE events (
    id TEXT PRIMARY KEY,
    event_name TEXT NOT NULL,
    home_team TEXT NOT NULL,
    away_team TEXT NOT NULL,
    home_win_chance REAL NOT NULL,
    away_win_chance REAL NOT NULL,
    draw_chance REAL NOT NULL,
    event_start_date TEXT NOT NULL, -- Store as ISO 8601 string or INTEGER timestamp
    event_end_date TEXT NOT NULL,
    event_result TEXT, -- e.g., 'HomeWin', 'AwayWin', 'Draw', NULL if not finished
    type TEXT NOT NULL,
    is_active INTEGER DEFAULT 1 -- 1 for active, 0 for inactive/finished
);
CREATE TABLE bets (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL, -- The GUID from the request
    event_id TEXT NOT NULL,
    amount REAL NOT NULL,
    predicted_outcome TEXT NOT NULL, -- 'HomeWin', 'AwayWin', 'Draw'
    recorded_home_win_chance REAL NOT NULL,
    recorded_away_win_chance REAL NOT NULL,
    recorded_draw_chance REAL NOT NULL,
    placed_at DATETIME NOT NULL, -- ISO 8601 timestamp
    status TEXT NOT NULL DEFAULT 'Pending', -- 'Pending', 'Won', 'Lost', 'Paid', 'Failed'
    payout_amount REAL DEFAULT 0,
    FOREIGN KEY (event_id) REFERENCES events(id)
);
CREATE INDEX idx_bets_event_id ON bets(event_id);
CREATE INDEX idx_bets_user_id ON bets(user_id);
CREATE INDEX idx_bets_status ON bets(status);
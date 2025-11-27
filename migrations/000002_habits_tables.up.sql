CREATE TABLE IF NOT EXISTS habits (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name VARCHAR(255) NOT NULL,
    description TEXT,
    goal VARCHAR(255),

    frequency VARCHAR(50) NOT NULL,
    weekly_days VARCHAR(13),
    monthly_days VARCHAR(255),

    current_streak INTEGER DEFAULT 0,
    best_streak INTEGER DEFAULT 0,
    last_completed_date DATE,

    last_checked_date DATE,

    is_active BOOLEAN DEFAULT TRUE,
    is_completed BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,

    CONSTRAINT valid_frequency CHECK (frequency IN ('daily', 'weekly', 'monthly'))
);

CREATE INDEX idx_habits_user_id ON habits(user_id);
CREATE INDEX idx_habits_is_active ON habits(is_active);

CREATE TABLE IF NOT EXISTS habit_logs (
    id SERIAL PRIMARY KEY,
    habit_id INTEGER NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    comment TEXT,

    logged_date DATE NOT NULL,
    logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_habit_log_per_day UNIQUE(habit_id, logged_date)
);

CREATE INDEX idx_habit_logs_habit_id ON habit_logs(habit_id);
CREATE INDEX idx_habit_logs_user_id ON habit_logs(user_id);
CREATE INDEX idx_habit_logs_logged_date ON habit_logs(logged_date);

CREATE TABLE IF NOT EXISTS habit_reminders (
    id SERIAL PRIMARY KEY,
    habit_id INTEGER NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    reminder_date DATE NOT NULL,

    is_completed BOOLEAN DEFAULT FALSE,

    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_reminder_per_day UNIQUE(habit_id, reminder_date)
);

CREATE INDEX idx_habit_reminders_user_id ON habit_reminders(user_id);
CREATE INDEX idx_habit_reminders_reminder_date ON habit_reminders(reminder_date);

CREATE TABLE IF NOT EXISTS streak_reset_queue (
    id SERIAL PRIMARY KEY,
    habit_id INTEGER NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    reset_date DATE NOT NULL,

    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP,

    previous_streak INTEGER,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_reset_per_day UNIQUE(habit_id, reset_date)
);

CREATE INDEX idx_streak_reset_queue_reset_date ON streak_reset_queue(reset_date);
CREATE INDEX idx_streak_reset_queue_processed ON streak_reset_queue(processed);

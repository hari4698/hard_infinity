CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clerk_id TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    current_day INTEGER DEFAULT 1,
    status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE sections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    section_id UUID NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    task_type TEXT NOT NULL CHECK (task_type IN ('boolean', 'number', 'text', 'select')),
    required BOOLEAN DEFAULT TRUE,
    restart_on_fail BOOLEAN DEFAULT FALSE,
    strikes_enabled BOOLEAN DEFAULT FALSE,
    strikes_limit INTEGER,
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE daily_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    day_number INTEGER NOT NULL,
    date DATE NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    notes TEXT,
    progress_photo_url TEXT,
    energy_level INTEGER,
    mood_level INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (challenge_id, day_number)
);

CREATE TABLE task_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    daily_entry_id UUID NOT NULL REFERENCES daily_entries(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    completed BOOLEAN DEFAULT FALSE,
    value JSONB,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (daily_entry_id, task_id)
);

CREATE TABLE measurements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    day_number INTEGER NOT NULL,
    date DATE NOT NULL,
    weight DECIMAL,
    chest DECIMAL,
    waist DECIMAL,
    hips DECIMAL,
    arms DECIMAL,
    thighs DECIMAL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (challenge_id, day_number)
);

-- Add indexes for performance
CREATE INDEX idx_users_clerk_id ON users(clerk_id);
CREATE INDEX idx_challenges_user_id ON challenges(user_id);
CREATE INDEX idx_sections_challenge_id ON sections(challenge_id);
CREATE INDEX idx_tasks_section_id ON tasks(section_id);
CREATE INDEX idx_daily_entries_challenge_id ON daily_entries(challenge_id);
CREATE INDEX idx_task_entries_daily_entry_id ON task_entries(daily_entry_id);
CREATE INDEX idx_task_entries_task_id ON task_entries(task_id);
CREATE INDEX idx_measurements_challenge_id ON measurements(challenge_id);
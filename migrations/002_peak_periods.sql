-- ============================================================
-- Peak Periods Table - Stores pre-computed snowiest periods
-- ============================================================

CREATE TABLE IF NOT EXISTS resort_peak_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resort_id UUID NOT NULL,
    peak_rank INTEGER NOT NULL CHECK (peak_rank BETWEEN 1 AND 5),
    start_date VARCHAR(5) NOT NULL,     -- MM-DD format
    end_date VARCHAR(5) NOT NULL,       -- MM-DD format
    center_date VARCHAR(5) NOT NULL,    -- Peak center MM-DD
    avg_daily_snowfall NUMERIC(6,2) NOT NULL,
    total_period_snowfall NUMERIC(6,2) NOT NULL,
    prominence_score NUMERIC(6,2) NOT NULL,
    years_of_data INTEGER NOT NULL,
    confidence_level VARCHAR(20) NOT NULL CHECK (confidence_level IN ('high', 'medium', 'low')),
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(resort_id, peak_rank),
    FOREIGN KEY (resort_id) REFERENCES resorts(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_resort_peak_periods_resort ON resort_peak_periods(resort_id);
CREATE INDEX idx_resort_peak_periods_calculated ON resort_peak_periods(calculated_at);

-- Verification
SELECT 'Peak periods table created successfully!' as status;

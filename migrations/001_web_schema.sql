-- ============================================================
-- Web Interface Schema - Minimal tables for read-only queries
-- ============================================================

-- ============================================================
-- Table: resorts
-- Stores ski resort metadata
-- ============================================================
CREATE TABLE resorts (
    id UUID PRIMARY KEY,
    slug VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    prefecture VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    top_elevation_m INTEGER,
    base_elevation_m INTEGER,
    vertical_m INTEGER,
    num_courses INTEGER,
    longest_course_km NUMERIC(5,2),
    steepest_course_deg NUMERIC(4,1),
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================
-- Table: daily_snowfall
-- Stores calculated daily snowfall for date range queries
-- ============================================================
CREATE TABLE daily_snowfall (
    resort_id UUID NOT NULL,
    date DATE NOT NULL,
    snowfall_cm INTEGER NOT NULL,
    season VARCHAR(20),
    PRIMARY KEY (resort_id, date),
    FOREIGN KEY (resort_id) REFERENCES resorts(id) ON DELETE CASCADE
);

-- ============================================================
-- Indexes for query performance
-- ============================================================
-- slug index not needed: already has unique constraint index
CREATE INDEX idx_resorts_prefecture ON resorts(prefecture);
CREATE INDEX idx_daily_snowfall_date ON daily_snowfall(date);
CREATE INDEX idx_daily_snowfall_season ON daily_snowfall(season);

-- ============================================================
-- Verification
-- ============================================================
SELECT 'Schema created successfully!' as status;
SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;

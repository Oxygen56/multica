-- squad.capability stores structured squad expertise metadata for cross-squad
-- discovery (Feature #4106). Each squad leader declares what their squad can do:
--   domains      — high-level category tags (e.g. "tech_architecture", "data_engineering")
--   keywords     — lower-level match tokens used by `squad route` for keyword overlap scoring
--   description  — free-text summary of the squad's expertise
--
-- The column is JSONB so the route command can index into it directly in SQL
-- (e.g. jsonb_array_length) without pulling every row into Go. NULL means the
-- squad has not declared capability yet — `squad route` will append a nudge
-- listing these squads.
ALTER TABLE squad
    ADD COLUMN capability JSONB NULL;
